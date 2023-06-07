GOLANDCI_LINT_VERSION ?= v1.53.1

build:
	mkdir -p bin
	go build -o bin/lvctl ./cmd

sanity: goimport
	go version
	go fmt ./...
	go mod tidy -v
	go mod vendor
	git add -N vendor
	git difftool -y --trust-exit-code --extcmd=./hack/diff-csv.sh

goimport:
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w -local="github.com/vladikr/logsviewer/tools/lvctl"  $(shell find . -type f -name '*.go' ! -path "*/vendor/*" )

test:
	go test -v ./pkg/...

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANDCI_LINT_VERSION}
	golangci-lint run

.PHONY: sanity goimport test lint
