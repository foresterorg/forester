#!/bin/bash

go install github.com/webrpc/webrpc/cmd/webrpc-gen@latest
go install -v github.com/incu6us/goimports-reviser/v3@latest
go get -u github.com/webrpc/gen-golang
go get -u github.com/webrpc/gen-openapi
go get -u ./...