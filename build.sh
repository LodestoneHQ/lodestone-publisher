#!/bin/bash -e

dep ensure

# go test -v ./...

go build -o lodestone-fs-publisher-linux-amd64 ./cmd/fs-publisher/fs-publisher.go
go build -o lodestone-email-publisher-linux-amd64 ./cmd/email-publisher/email-publisher.go

./lodestone-fs-publisher-linux-amd64 --help
./lodestone-email-publisher-linux-amd64 --help
