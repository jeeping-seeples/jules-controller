.PHONY: help fmt test vet check build

help:
	@printf '%s\\n' 'Targets: fmt test vet check build'

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

test:
	go test ./...

vet:
	go vet ./...

check: test vet

build:
	go build ./cmd/jules-tui
