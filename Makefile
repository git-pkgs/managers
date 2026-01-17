.PHONY: test test-docker test-docker-all lint fmt

# Run tests locally
test:
	go test ./... -v

# Run tests against all Docker versions
test-docker-all:
	./scripts/docker-test.sh

# Run tests against specific manager version
# Usage: make test-docker MANAGER=npm VERSION=10
test-docker:
	./scripts/docker-test.sh $(MANAGER) $(VERSION)

# List available Docker versions
test-docker-list:
	./scripts/docker-test.sh --list

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean test cache
clean:
	go clean -testcache
