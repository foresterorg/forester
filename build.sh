#!/bin/sh
which webrpc-gen &>/dev/null || go install github.com/webrpc/webrpc/cmd/webrpc-gen@latest
which tern &>/dev/null || go github.com/jackc/tern@latest
webrpc-gen -silent -schema=internal/api/ctl/controller.ridl -target=golang -pkg=ctl -server -client -out=./internal/api/ctl/proto.gen.go
go build -o forester-controller ./cmd/controller
go build -o forester-cli ./cmd/cli
