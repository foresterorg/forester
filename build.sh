#!/bin/sh
which webrpc-gen &>/dev/null || go install github.com/webrpc/webrpc/cmd/webrpc-gen@latest
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=golang -pkg=ctl -server -client -out=./internal/api/ctl/proto.gen.go
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=github.com/webrpc/gen-openapi@v0.11.2 -out=openapi.gen.yaml \
  -title="Forester API" -apiVersion="$(git describe --abbrev=0 --tags)" -serverUrl=https://forester.example.com:8000 -serverDescription="Forester service"
go build -o forester-controller ./cmd/controller
go build -o forester-cli ./cmd/cli
