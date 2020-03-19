# SPDX-License-Identifier: AGPL-3.0-only

export GO111MODULE=on
BINARY_NAME=drlm-core

all: deps build docker
install:
	go install drlm-core.go
build:
	go build drlm-core.go
test:
	go test -cover -race ./...
clean:
	go clean
	rm -f $(BINARY_NAME)
deps:
	go build -v ./...
upgrade:
	go get -u
run:
	go run drlm-core.go

docker:
	docker build -t drlm-core:1.0.0 .