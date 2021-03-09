#!/bin/sh
go get -u github.com/go-bindata/go-bindata/...
go generate
GO111MODULES=on go build -ldflags "-s -w" -o start main.go
#upx -9 start
