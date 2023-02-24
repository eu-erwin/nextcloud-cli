package nextcloud

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/eu-erwin/nextcloud-cli/pkg/cloud"
)

func NewStorage(
	cloudUrl,
	username,
	password string,
) (cloud.Storage, error) {
	if cloudUrl == "" {
		return nil, errors.New("missing url")
	}
	if username == "" || password == "" {
		return nil, errors.New("missing credentials")
	}

	parsedUrl, err := url.Parse(cloudUrl)
	if nil != err {
		return nil, errors.New("invalid url")
	}
	return &Client{
		Url:      parsedUrl,
		Username: username,
		Password: password,
	}, nil
}

// Client represents a client connection to a {own|next}cloud
type Client struct {
	Url      *url.URL
	Username string
	Password string
}

// Error type encapsulates the returned error messages from the
// server.
type Error struct {
	// Exception contains the type of the exception returned by
	// the server.
	Exception string `xml:"exception"`

	// Message contains the error message string from the server.
	Message string `xml:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("Exception: %s, Message: %s", e.Exception, e.Message)
}

// Dial connects to an {own|next}Cloud instance at the specified
// address using the given credentials.
func Dial(host, username, password string) (*Client, error) {
	parsedUrl, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &Client{
		Url:      parsedUrl,
		Username: username,
		Password: password,
	}, nil
}

// Mkdir creates a new directory on the cloud with the specified name.
func (c *Client) Mkdir(path string) error {
	_, err := c.sendWebDavRequest("MKCOL", path, nil)
	return err

}

// Delete removes the specified folder from the cloud.
func (c *Client) Delete(path string) error {
	_, err := c.sendWebDavRequest("DELETE", path, nil)
	return err
}

// Upload uploads the specified source to the specified destination
// path on the cloud.
func (c *Client) Upload(src []byte, dest string) error {
	_, err := c.sendWebDavRequest("PUT", dest, src)
	return err
}

// UploadDir uploads an entire directory on the cloud. It returns the
// path of uploaded files or error. It uses glob pattern in src.
func (c *Client) UploadDir(src string, dest string) ([]string, error) {
	files, err := filepath.Glob(src)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		err = c.Upload(data, filepath.Join(dest, filepath.Base(file)))
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// Download downloads a file from the specified path.
func (c *Client) Download(path string) ([]byte, error) {
	return c.sendWebDavRequest("GET", path, nil)
}

func (c *Client) Exists(path string) bool {
	_, err := c.sendWebDavRequest("PROPFIND", path, nil)
	return err == nil
}

func (c *Client) CreateGroupFolder(mountPoint string) (*cloud.ShareResult, error) {
	return c.sendAppsRequest("POST", "groupfolders/folders", fmt.Sprintf("mountpoint=%s", mountPoint))
}

func (c *Client) AddGroupToGroupFolder(group string, folderId uint) (*cloud.ShareResult, error) {
	return c.sendAppsRequest("POST", fmt.Sprintf("groupfolders/folders/%d/groups", folderId), fmt.Sprintf("group=%s", group))
}

func (c *Client) SetGroupPermissionsForGroupFolder(permissions int, group string, folderId uint) (*cloud.ShareResult, error) {
	return c.sendAppsRequest("POST", fmt.Sprintf("apps/groupfolders/folders/%d/groups/%s", folderId, group), fmt.Sprintf("permissions=%d", permissions))
}

func (c *Client) CreateShare(path string, shareType int, publicUpload string, permissions int) (*cloud.ShareResult, error) {
	return c.sendOCSRequest("POST", "shares", fmt.Sprintf("path=%s&shareType=%d&publicUpload=%s&permissions=%d", path, shareType, publicUpload, permissions))
}

func (c *Client) GetShare(path string) (*cloud.ShareResult, error) {
	return c.sendOCSRequest("GET", fmt.Sprintf("shares?path=%s", path), "")
}

func (c *Client) DeleteShare(id uint) (*cloud.ShareResult, error) {
	return c.sendOCSRequest("DELETE", fmt.Sprintf("shares/%d", id), "")
}

func (c *Client) CreateFileDropShare(path string) (*cloud.ShareResult, error) {
	result, err := c.CreateShare(path, 3, "true", 4)
	if err != nil {
		return nil, err
	}
	id := result.Id
	return c.sendOCSRequest("PUT", fmt.Sprintf("shares/%d", id), "permissions=4")
}

func (c *Client) CreateReadOnlyShare(path string) (*cloud.ShareResult, error) {
	result, err := c.CreateShare(path, 3, "true", 4)
	if err != nil {
		return nil, err
	}
	id := result.Id
	return c.sendOCSRequest("PUT", fmt.Sprintf("shares/%d", id), "permissions=1")
}

func (c *Client) sendWebDavRequest(request string, path string, data []byte) ([]byte, error) {
	// Create the https request

	webdavPath := filepath.Join("remote.php/webdav", path)

	folderUrl, err := url.Parse(webdavPath)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(request, c.Url.ResolveReference(folderUrl).String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(body) > 0 {
		if body[0] == '<' {
			reqErr := &Error{}
			decodeErr := xml.Unmarshal(body, reqErr)
			if decodeErr != nil {
				return body, decodeErr
			}
			if reqErr.Exception != "" {
				return nil, reqErr
			}
		}

	}

	return body, nil
}

func (c *Client) sendAppsRequest(request string, path string, data string) (*cloud.ShareResult, error) {
	// Create the https request

	appsPath := filepath.Join("apps", path)

	folderUrl, err := url.Parse(appsPath)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(request, c.Url.ResolveReference(folderUrl).String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("OCS-APIRequest", "true")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := cloud.ShareResult{}
	err = xml.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != 100 {
		return nil, fmt.Errorf("share API returned an unsuccessful status code %d", result.StatusCode)
	}

	return &result, nil
}

func (c *Client) sendOCSRequest(request string, path string, data string) (*cloud.ShareResult, error) {
	// Create the https request

	appsPath := filepath.Join("ocs/v2.php/apps/files_sharing/api/v1", path)

	folderUrl, err := url.Parse(appsPath)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(request, c.Url.ResolveReference(folderUrl).String(), strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("OCS-APIRequest", "true")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.SetBasicAuth(c.Username, c.Password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &cloud.ShareResult{}

	err = xml.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != 200 {
		return nil, fmt.Errorf("share API returned an unsuccessful status code %d", result.StatusCode)
	}

	return result, nil
}
