.PHONY: tests lint fmt

SOURCE_FILES?=.
TEST_PATTERN?=.

fmt:
	gofumpt -w -l .

lint:
	golangci-lint run ./...

tests:
	go test -v -failfast -race -coverpkg=./... -covermode=atomic\
	 -coverprofile=coverage.out $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m