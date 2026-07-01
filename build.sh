#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o dist/draftcat-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o dist/draftcat-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o dist/draftcat-windows-amd64.exe
