.PHONY: fmt test lint run demo demo-12

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

# make demo [PLAYERS=12 ROLES=CursedWolf=1,WolfSeer=1,Seer=1]; villagers fill the rest.
demo:
	go run ./cmd/demo $(if $(PLAYERS),-players $(PLAYERS)) $(if $(ROLES),-roles $(ROLES))

demo-12:
	go run ./cmd/demo -players 12 -roles CursedWolf=1,WolfSeer=1,Seer=1,Guard=1,Witch=1,Hunter=1,Visitor=1
