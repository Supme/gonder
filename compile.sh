#!/bin/sh

go build -ldflags "-s -w" -o gonder main.go && upx -9 gonder
GOOS=windows go build -ldflags "-s -w" -o gonder.exe main.go && upx -9 gonder.exe
