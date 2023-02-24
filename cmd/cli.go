package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/eu-erwin/nextcloud-cli/pkg/cloud"
	"github.com/eu-erwin/nextcloud-cli/pkg/nextcloud"
)

var (
	client     cloud.Storage
	targetPath *string
	wd         string
)

func init() {
	cloudUrl := flag.String("url", "", "Please enter url")
	username := flag.String("username", "", "Please enter username")
	password := flag.String("password", "", "Please enter password")
	targetPath = flag.String("path", "", "Target path to upload file")
	flag.Parse()

	storage, err := nextcloud.NewStorage(*cloudUrl, *username, *password)
	if nil != err {
		log.Println("storage can't be created. reason: ", err.Error())
		return
	}

	client = storage
	wd, _ = os.Getwd()
	_ = []string{wd}
}

func printHelp() {
	log.Println(`Available command:
upload

Available flags:
--url 		url of the nextcloud host (*)
--username 	your username (*)
--password 	your password (*)
--path 		target path

Ex.: nextcloud-cli upload --username=john --password=supersecret --url=https://cloud.example.com hello.text`)
}

func Run() {
	args := flag.Args()
	if 0 == len(args) {
		log.Println(`missing command`)
		printHelp()
		return
	}

	switch args[0] {
	case "upload":
		upload(args[1:]...)
	default:
		log.Println(`missing command`)
		printHelp()
	}
}

func upload(sources ...string) {
	for _, source := range sources {
		fmt.Printf("Uploading %s\r\n", source)

		content, err := os.ReadFile(strings.Join([]string{wd, source}, "/"))
		if nil != err {
			log.Println("Can't upload", source, err.Error())
			continue
		}

		target := source
		if "" != *targetPath {
			if err = client.Mkdir(*targetPath); nil == err {
				log.Println("New directory created", source)
			}
			target = strings.Join([]string{*targetPath, source}, "/")
		}

		err = client.Upload(content, target)
		if nil != err {
			log.Println("Upload failed", source, err.Error())
			continue
		}
	}
}
