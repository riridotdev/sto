git_info=$(shell git describe --always --dirty)
linker_flags='-s -X main.version=${git_info}'

.PHONY: build
build:
	@echo 'Building Sto'
	@go build -ldflags=${linker_flags} -o=./bin/sto ./cmd/
