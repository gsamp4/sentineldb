.PHONY: test test-verbose test-watch test-race test-cover test-integration

## test: run all unit tests with clean output
test:
	gotestsum --format testdox -- ./...

## test-verbose: show each test name and result
test-verbose:
	gotestsum --format standard-verbose -- ./...

## test-watch: re-run tests on file changes
test-watch:
	gotestsum --watch --format testdox -- ./...

## test-race: run with race detector
test-race:
	gotestsum --format testdox -- -race ./...

## test-cover: run with coverage report
test-cover:
	gotestsum --format testdox -- -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## test-integration: include integration tests (requires running PostgreSQL)
test-integration:
	INTEGRATION_TEST=1 gotestsum --format testdox -- ./...
