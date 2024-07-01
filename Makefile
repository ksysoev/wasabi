test:
	go test -v --race ./...

test-norace:
	go test -v ./...

lint:
	golangci-lint run

mocks:
	mockery --all --keeptree

install:
	cp ./scripts/pre-commit ./.git/hooks/pre-commit

fmt-all:
	gofmt -w .
