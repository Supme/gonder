#!/bin/sh
go build -ldflags "-s -w -X gonder/models.AppVersion=v0.22.3 -X gonder/models.AppCommit=`git describe --always` -X gonder/models.AppDate=`date -u +%FT%TZ`" -o start main.go
