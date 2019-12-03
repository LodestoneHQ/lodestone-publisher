#!/bin/bash -e

dep ensure

# go test -v ./...

go build -o lodestone-fs-publisher-linux-amd64 ./cmd/fs-publisher/fs-publisher.go

./lodestone-fs-publisher-linux-amd64 --help
