.PHONY: fmt test test-fast lint run demo demo-12 api-docs mysql-test-up mysql-test-down test-mysql

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

# Rewrite docs/api/schema and docs/api/examples from the server (JSON Schema
# from the Go types, examples from the example games). The same test without
# -update runs in make test/test-fast and fails when they are stale.
api-docs:
	go test ./internal/transport/ws -run '^TestAPIDocs$$' -count=1 -update

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

# A throwaway MySQL 8.4 for the store tests (docs/database.md), on port 3307 so
# it never meets a MySQL already on the machine. Its data is gone with the
# container.
MYSQL_TEST_DSN ?= root:masoi-test@tcp(127.0.0.1:3307)/masoi_test
mysql-test-up:
	docker run -d --rm --name masoi-mysql-test -p 127.0.0.1:3307:3306 \
		-e MYSQL_ROOT_PASSWORD=masoi-test -e MYSQL_DATABASE=masoi_test mysql:8.4
	@echo "waiting for MySQL..."; for i in $$(seq 60); do \
		docker exec masoi-mysql-test mysqladmin ping -h127.0.0.1 -uroot -pmasoi-test --silent 2>/dev/null && exit 0; sleep 1; \
	done; echo "MySQL did not start"; exit 1

mysql-test-down:
	docker stop masoi-mysql-test

# The store tests against that MySQL (skipped in the other targets without
# MASOI_MYSQL_DSN).
test-mysql:
	MASOI_MYSQL_DSN='$(MYSQL_TEST_DSN)' go test ./internal/store/... -count=1 -v -run 'MySQL'
