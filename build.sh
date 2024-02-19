#!/bin/sh
which webrpc-gen &>/dev/null || go install github.com/webrpc/webrpc/cmd/webrpc-gen@latest
which goimports-reviser &>/dev/null || go install -v github.com/incu6us/goimports-reviser/v3@latest
goimports-reviser -rm-unused -format ./...
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=golang -pkg=ctl -server -client -out=./internal/api/ctl/proto.gen.go
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=github.com/webrpc/gen-openapi@v0.11.3 -out=openapi.gen.yaml \
  -title="Forester API" -apiVersion="$(git describe --abbrev=0 --tags)" -serverUrl=https://forester.example.com:8000 -serverDescription="Forester service"
go build -o forester-controller ./cmd/controller
go build -o forester-cli ./cmd/cli
