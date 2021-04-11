#!/bin/bash

tag=$(git describe --tags)

mkdir -p release/$tag

GOOS=windows GOARCH=386 go build -o release/$tag/rectail.exe ./cmd/rectail
GOOS=windows GOARCH=amd64 go build -o release/$tag/rectail64.exe ./cmd/rectail

GOOS=linux GOARCH=386 go build -o release/$tag/rectail_linux ./cmd/rectail
GOOS=linux GOARCH=amd64 go build -o release/$tag/rectail64_linux ./cmd/rectail

GOOS=darwin GOARCH=amd64 go build -o release/$tag/rectail_osx ./cmd/rectail