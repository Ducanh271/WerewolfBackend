.PHONY: fmt test lint run

fmt:
	gofmt -w .
	go run golang.org/x/tools/cmd/goimports@latest -w .

test:
	go test ./... -race -count=1

lint:
	go vet ./...
	golangci-lint run

run:
	go run ./cmd/server
