.PHONY: test test-unit test-integration test-coverage

# Default test command runs all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	go test -v ./order-service/tests/...

# Run integration tests
test-integration:
	go test -v ./order-service/tests/integration/...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./order-service/tests/... ./order-service/tests/integration/...
	go tool cover -html=coverage.out -o coverage.html

# Clean test cache and coverage files
clean:
	go clean -testcache
	rm -f coverage.out coverage.html 