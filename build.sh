#!/bin/sh
which webrpc-gen &>/dev/null || go install github.com/webrpc/webrpc/cmd/webrpc-gen@v0.13.0
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=golang -pkg=ctl -server -client -out=./internal/api/ctl/proto.gen.go
go build -o forester-controller ./cmd/controller
go build -o forester-cli ./cmd/cli
