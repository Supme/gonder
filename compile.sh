#!/bin/sh

GO111MODULES=on go build -ldflags "-s -w" -o gonder main.go && upx -9 gonder
GO111MODULES=on GOOS=windows go build -ldflags "-s -w" -o gonder.exe main.go && upx -9 gonder.exe
