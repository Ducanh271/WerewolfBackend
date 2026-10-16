.PHONY: fmt test test-fast lint run demo demo-12

fmt:
	gofmt -w .
	go run golang.org/x/tools/cmd/goimports@latest -w .

# Full check: race detector + the simulation at its default SIM_GAMES=300. Run
# this once before committing.
test:
	go test ./... -race -count=1

# Fast loop while iterating: no race detector, skips the simulation and other
# slow tests (testing.Short()). Prefer `go test ./internal/<pkg> -run <Name>`
# for a single test.
test-fast:
	go test ./... -short

lint:
	go vet ./...
	golangci-lint run

run:
	go run ./cmd/server

# make demo [PLAYERS=12 ROLES=CursedWolf=1,WolfSeer=1,Seer=1 SPEED=20]; villagers fill the rest.
demo:
	go run ./cmd/demo $(if $(PLAYERS),-players $(PLAYERS)) $(if $(ROLES),-roles $(ROLES)) $(if $(SPEED),-speed $(SPEED))

demo-12:
	go run ./cmd/demo -speed $(or $(SPEED),20) -players 12 -roles CursedWolf=1,WolfSeer=1,Seer=1,Guard=1,Witch=1,Hunter=1,Visitor=1
