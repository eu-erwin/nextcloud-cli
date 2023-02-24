package nextcloud

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/remogatto/prettytest"
)

type testSuite struct {
	prettytest.Suite
}

var (
	client  *Client
	testDir = "test/testdata"
)

func TestRunner(t *testing.T) {
	prettytest.Run(
		t,
		new(testSuite),
	)
}

func (t *testSuite) BeforeAll() {
	var err error
	client, err = Dial(
		"http://localhost:18080/",
		"admin",
		"password",
	)
	if err != nil {
		panic(err)
	}
}

func (t *testSuite) After() {
	_ = client.Delete("Test")
}

func (t *testSuite) TestMkDir() {
	err := client.Mkdir("Test")
	t.Nil(err)
}

func (t *testSuite) TestDelete() {
	err := client.Mkdir("Test")
	t.Nil(err)
	err = client.Delete("Test")
	t.Nil(err)
}

func (t *testSuite) TestDownloadUpload() {
	mkdirErr := client.Mkdir("Test")
	t.Nil(mkdirErr)

	src, readErr := os.ReadFile(filepath.Join(testDir, "test.txt"))
	t.Nil(readErr)
	readErr = client.Upload(src, "Test/test.txt")
	t.Nil(readErr)

	data, err := client.Download("Test/test.txt")
	t.Nil(err)

	t.Equal("Hello World!\n", string(data))
}

func (t *testSuite) TestUploadDir() {
	err := client.Mkdir("Test")
	t.Nil(err)

	err = client.Mkdir("Test/Folder")
	t.Nil(err)

	files, err := client.UploadDir(filepath.Join(testDir, "Folder/*"), "Test/Folder/")
	t.Nil(err)

	t.Equal(1, len(files))

	data, err := client.Download("Test/Folder/test.txt")
	t.Nil(err)

	t.Equal("Hello World!\n", string(data))
}

func (t *testSuite) TestExists() {
	err := client.Mkdir("Test")
	t.Nil(err)
	t.True(client.Exists("Test"))
}

func (t *testSuite) TestCreateGroupFolder() {
	groupFolder, err := client.CreateGroupFolder("GroupFolder")
	t.Nil(err)
	if groupFolder != nil {
		t.Equal(uint(100), groupFolder.StatusCode)

		result, err := client.AddGroupToGroupFolder("admin", groupFolder.Id)
		t.Nil(err)
		if result != nil {
			t.Equal(uint(100), result.StatusCode)
		}

		result, err = client.SetGroupPermissionsForGroupFolder(31, "admin", groupFolder.Id)
		t.Nil(err)
		if result != nil {
			t.Equal(uint(100), result.StatusCode)
		}
	}

}

func (t *testSuite) TestCreateFileDropShare() {
	err := client.Mkdir("ShareTest")
	t.Nil(err)

	result, err := client.CreateFileDropShare("ShareTest")
	t.Nil(err)

	if result != nil {
		t.Equal(uint(200), result.StatusCode)
		t.True(len(result.Url) > 0)
	}

	err = client.Delete("ShareTest")
	if err != nil {
		panic("can't delete ShareTest on TestCreateFileDropShare")
	}
}

func (t *testSuite) TestGetShare() {
	err := client.Mkdir("ShareTest")
	t.Nil(err)

	result, err := client.CreateFileDropShare("ShareTest")
	t.Nil(err)
	t.Not(t.Nil(result))

	result, err = client.GetShare("ShareTest")
	t.Nil(err)
	t.True(len(result.Elements) > 0)
	t.Not(t.Nil(result))

	err = client.Delete("ShareTest")
	if err != nil {
		panic("can't delete ShareTest on TestGetShare")
	}
}

func (t *testSuite) TestDeleteShare() {
	err := client.Mkdir("ShareTest")
	t.Nil(err)

	result, err := client.CreateFileDropShare("ShareTest")
	t.Nil(err)
	t.Not(t.Nil(result))

	result, err = client.GetShare("ShareTest")
	t.Nil(err)
	t.True(len(result.Elements) > 0)
	t.Not(t.Nil(result))

	result, err = client.DeleteShare(result.Elements[0].Id)
	t.Nil(err)
	t.Not(t.Nil(result))

	result, err = client.GetShare("ShareTest")
	t.Nil(err)
	t.True(len(result.Elements) == 0)
	t.Not(t.Nil(result))

	err = client.Delete("ShareTest")
	if err != nil {
		panic("can't delete ShareTest on TestDeleteShare")
	}
	_ = result
}

func (t *testSuite) TestCreateReadOnlyShare() {
	err := client.Mkdir("ShareTest")
	t.Nil(err)

	result, err := client.CreateReadOnlyShare("ShareTest")
	t.Nil(err)
	t.Not(t.Nil(result))

	result, err = client.GetShare("ShareTest")
	t.Nil(err)
	t.True(len(result.Elements) > 0)
	t.Not(t.Nil(result))

	result, err = client.DeleteShare(result.Elements[0].Id)
	t.Nil(err)
	t.Not(t.Nil(result))

	result, err = client.GetShare("ShareTest")
	t.Nil(err)
	t.True(len(result.Elements) == 0)
	t.Not(t.Nil(result))

	err = client.Delete("ShareTest")
	if err != nil {
		panic("can't delete ShareTest on TestCreateReadOnlyShare")
	}
}
