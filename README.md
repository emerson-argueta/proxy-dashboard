# proxy-dashboard

Terminal UI dashboard for monitoring a [browser-metered-proxy](https://github.com/emerson-argueta/browser-metered-proxy) service running in Docker.

Built with Go + [bubbletea](https://github.com/charmbracelet/bubbletea) + [lipgloss](https://github.com/charmbracelet/lipgloss).

## What it shows

| Panel | Metrics |
|-------|---------|
| **System** | Container CPU, memory, disk usage |
| **Database** | SQLite file size, actor count, total capability log count, calls today |
| **Billing** | Total actor balance, container name |
| **Revenue** | Total charged, provider cost, profit (markup) |
| **Recent Requests** | Last 12 requests — time, status, method, path, duration |

Refreshes automatically every 5 seconds. Press `r` to force refresh, `q` to quit.

## Requirements

- Go 1.21+
- Docker running on the target machine
- `sqlite3` available inside the proxy container

## Installation

```bash
git clone git@github.com:emerson-argueta/proxy-dashboard.git
cd proxy-dashboard
```

## Usage

### From your Mac (connects to server over SSH)

```bash
make run
```

Or manually:

```bash
./bin/proxy-dashboard \
  --host devadmin@100.97.193.52 \
  --key ~/.ssh/id_ed25519_mac \
  --container budget-clear-proxy
```

### On the server itself (local mode, no SSH)

```bash
~/proxy-dashboard --container budget-clear-proxy
```

### Install binary on server

```bash
make install-server
```

This cross-compiles for Linux (amd64) and copies the binary to `~/proxy-dashboard` on the server.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | _(local)_ | SSH target: `user@host`. Omit to run in local mode. |
| `--key` | — | Path to SSH private key. Required when `--host` is set. |
| `--container` | `budget-clear-proxy` | Docker container name prefix to monitor. |

## Makefile targets

```bash
make build          # Build for macOS
make build-linux    # Cross-compile for Linux (amd64)
make install-server # Build for Linux and scp to server
make run            # Build and run in remote SSH mode
make clean          # Remove bin/
```

## Adding a new proxy service

Point the dashboard at any other `browser-metered-proxy` container by changing the `--container` flag and `--host`:

```bash
./bin/proxy-dashboard \
  --host devadmin@100.97.193.52 \
  --key ~/.ssh/id_ed25519_mac \
  --container my-other-app-proxy
```
