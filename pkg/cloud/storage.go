package cloud

import "encoding/xml"

type ShareElement struct {
	Id  uint   `xml:"id"`
	Url string `xml:"url"`
}

type ShareResult struct {
	XMLName    xml.Name       `xml:"ocs"`
	Status     string         `xml:"meta>status"`
	StatusCode uint           `xml:"meta>statuscode"`
	Message    string         `xml:"meta>message"`
	Id         uint           `xml:"data>id"`
	Url        string         `xml:"data>url"`
	Elements   []ShareElement `xml:"data>element"`
}

type Storage interface {
	Mkdir(path string) error
	Delete(path string) error
	Upload(src []byte, dest string) error
	UploadDir(src string, dest string) ([]string, error)
	Download(path string) ([]byte, error)
	Exists(path string) bool
	CreateGroupFolder(mountPoint string) (*ShareResult, error)
	AddGroupToGroupFolder(group string, folderId uint) (*ShareResult, error)
	SetGroupPermissionsForGroupFolder(permissions int, group string, folderId uint) (*ShareResult, error)
	CreateShare(path string, shareType int, publicUpload string, permissions int) (*ShareResult, error)
	GetShare(path string) (*ShareResult, error)
	DeleteShare(id uint) (*ShareResult, error)
	CreateFileDropShare(path string) (*ShareResult, error)
	CreateReadOnlyShare(path string) (*ShareResult, error)
}
