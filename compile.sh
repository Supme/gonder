#!/bin/sh
go generate
GO111MODULES=on go build -ldflags "-s -w" -o gonder main.go && upx -9 gonder
