#!/bin/bash

releasePath="release"
mkdir -p releasePath

GOOS=windows GOARCH=386 go build -o $releasePath/rectail.exe ./cmd/rectail
GOOS=windows GOARCH=amd64 go build -o $releasePath/rectail64.exe ./cmd/rectail

GOOS=linux GOARCH=386 go build -o $releasePath/rectail_linux ./cmd/rectail
GOOS=linux GOARCH=amd64 go build -o $releasePath/rectail64_linux ./cmd/rectail

GOOS=darwin GOARCH=amd64 go build -o $releasePath/rectail_osx ./cmd/rectail

echo "generated"