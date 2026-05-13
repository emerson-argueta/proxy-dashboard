.PHONY: build build-linux install-server run run-remote clean

SERVER_USER := devadmin
SERVER_HOST := 100.97.193.52
SERVER_KEY  := ~/.ssh/id_ed25519_mac
CONTAINER   := budget-clear-proxy

# Build for macOS
build:
	go build -o bin/proxy-dashboard .

# Build for Linux (amd64) to deploy on the server
build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/proxy-dashboard-linux .

# Copy Linux binary to server
install-server: build-linux
	scp -i $(SERVER_KEY) bin/proxy-dashboard-linux $(SERVER_USER)@$(SERVER_HOST):~/proxy-dashboard
	@echo "Installed at ~/proxy-dashboard on $(SERVER_HOST)"
	@echo "Run on server: ~/proxy-dashboard --container $(CONTAINER)"

# Run locally (Mac connecting to server over SSH)
run: build
	./bin/proxy-dashboard \
		--host $(SERVER_USER)@$(SERVER_HOST) \
		--key $(SERVER_KEY) \
		--container $(CONTAINER)

# Run on the server itself (local mode, no SSH)
run-local: build
	./bin/proxy-dashboard --container $(CONTAINER)

clean:
	rm -rf bin/
