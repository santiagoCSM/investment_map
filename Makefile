BINARY := bin/server

.PHONY: run build docker-up docker-down smoketest-yahoo smoketest-store smoketest-market-status fmt vet

# `go run` only forwards SIGINT/SIGQUIT to its child process, never SIGTERM,
# and always leaves an intermediary process in between. That combination
# orphans the server on a plain `kill` instead of triggering its graceful
# shutdown. This target builds a real binary and execs it directly (no `go
# run`, no extra shell in between) so Ctrl+C and `kill` both reach the
# server's own signal handling.
run: build
	exec ./$(BINARY)

build:
	go build -o $(BINARY) ./cmd/server

docker-up:
	docker compose up -d

docker-down:
	docker compose down

smoketest-yahoo:
	go run ./cmd/yahoo_smoketest

smoketest-store:
	go run ./cmd/store_smoketest

smoketest-market-status:
	go run ./cmd/market_status_smoketest

fmt:
	gofmt -l -w .

vet:
	go vet ./...
