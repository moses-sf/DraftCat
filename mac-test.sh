#!/bin/bash
set -e

./build.sh

sudo install -m 755 dist/draftcat-darwin-arm64 /usr/local/bin/draftcat

draftcat --help
