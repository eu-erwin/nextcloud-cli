# {own|next}cloud Golang client
[![GoDoc](https://godoc.org/github.com/remogatto/cloud?status.png)](http://godoc.org/github.com/remogatto/cloud)

This is a golang client for [ownCloud](https://owncloud.com) and
[NextCloud](https://nextcloud.com).

# Usages
Installation
```shell
go install github.com/eu-erwin/nextcloud-cli
```

# Usages
Available cli command
- upload

## Upload command
To upload a file for example `README.md` from current working directory to directory `Notes` in nextcloud. simply use
```shell
nextcloud-cli --username=hello --password=world --url=http://localhost:18080/ --path=Notes README.md
```


# LICENSE

MIT
