.PHONY: test
test:
	@echo 'Running tests...'
	@go test ./...

.PHONY: fmt
fmt:
	@go fmt -x ./...
