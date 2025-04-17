.PHONY: test unit-test integration-test

test: unit-test integration-test

unit-test:
	go test ./... -short

integration-test:
	go test ./tests/... -tags=integration
