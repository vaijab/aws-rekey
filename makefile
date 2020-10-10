VERSION = $(shell git describe --always --tags --dirty)

aws-rekey: main.go
	go build -mod=vendor -ldflags "-X main.Version=${VERSION}"

.PHONY: clean
clean:
	go clean
	go mod tidy
