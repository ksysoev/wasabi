test:
	go test -v --race ./...

lint:
	golangci-lint run

mocks:
	mockery --all --keeptree