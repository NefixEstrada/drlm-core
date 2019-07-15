export GO111MODULE=on
BINARY_NAME=drlm-core

all: deps build
install:
	go install drlm-core.go
build:
	go build drlm-core.go
test:
	go test -cover ./...
clean:
	go clean
	rm -f $(BINARY_NAME)
deps:
	go build -v ./...
upgrade:
	go get -u
run:
	go run drlm-core.go