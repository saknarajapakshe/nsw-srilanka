# NSW Sri Lanka Platform

[![Go Version](https://img.shields.io/badge/Go-1.25.7-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/NSW-Platform-green.svg)](#)

`nsw-srilanka` is the deployer-specific application repository for the **Sri Lanka instance** of the National Single Window (NSW) Platform.

It depends on the core engine published as `github.com/OpenNSW/nsw/backend` (taskv2 branch) and wires Sri Lanka–specific service endpoints, payment gateways, and FCAU workflow configurations on top of it.

---

## Repository Layout

```
nsw-srilanka/
├── cmd/
│   └── server/
│       └── main.go                       # Entry point: loads config, builds the app, runs the HTTP server
├── internal/
│   └── bootstrap/
│       └── app.go                        # Wires DB, Temporal, taskv2, auth, storage, notifications, routes
├── configs/
│   ├── services.json                     # Remote service endpoints (gitignored)
│   ├── services.example.json             # Template for services.json
│   ├── payment_methods.json              # Payment gateway catalogue (gitignored)
│   ├── payment_methods.example.json      # Template for payment_methods.json
│   └── fcau/                             # FCAU health-certificate workflow + JSONForms (gitignored)
├── .env.example                          # Template for environment variables
├── .gitignore
├── Dockerfile
├── go.mod
└── go.sum
```

The Sri Lanka–specific FCAU workflow lives under `configs/fcau/` as a set of JSON files (workflow graph, JSONForms, render configs). The Go server itself is intentionally thin — most behaviour is configured via these JSON files and the `github.com/OpenNSW/nsw/backend` module.

---

## How to Run Locally

### 1. Prerequisites
- **Go** `1.25.7` or higher
- **PostgreSQL** running on `localhost:5432` (default values match `.env.example`)
- **Temporal** server on `localhost:7233` — start a dev instance with:
  ```bash
  temporal server start-dev
  ```
- **NSW Agency** backend running on `http://localhost:8082` if you want to exercise the FCAU officer-review flow end-to-end. See [OpenNSW/nsw-agency](https://github.com/OpenNSW/nsw-agency).

### 2. Prepare local config files

Copy the templates and edit each one for your environment:

```bash
cp .env.example .env
cp configs/services.example.json configs/services.json
cp configs/payment_methods.example.json configs/payment_methods.json
```

The FCAU workflow JSON tree under `configs/fcau/` is also gitignored — pull it from your team's secure store or generate it from your workflow templates before the first run.

### 3. Run the server

```bash
set -a; source .env; set +a
go run ./cmd/server
```

`go run` reads from the process environment; `set -a; source .env; set +a` exports every variable from `.env` so the upstream `internal/config` loader sees them.

Or one-shot, without polluting your shell:

```bash
env $(grep -v '^#' .env | xargs) go run ./cmd/server
```

### 4. Verify

- Health check: `curl http://localhost:8080/health` should return `{"status":"ok","service":"nsw-backend"}`.
- Logs will report DB connection, Temporal worker startup, and the FCAU workflow registrations from `configs/fcau/`.

### 5. Simulating a payment webhook (dev only)

INFO-type gateways (e.g. `govpay`) don't fire a real callback. To advance a `PENDING_PAYMENT` task manually:

```bash
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "reference_number": "TNSW-XXXXXXXX",
    "session_id": "manual-test-1",
    "gateway_transaction_id": "MOCK-001",
    "status": "SUCCESS",
    "amount": "1500",
    "currency": "LKR",
    "payment_method": "govpay",
    "timestamp": "2026-01-01T00:00:00Z",
    "metadata": {}
  }'
```

REDIRECT-type gateways (e.g. `lankapay`) fire this webhook on their own.

---

## Upstream Dependency

The core engine is pulled directly from GitHub via Go modules:

```
github.com/OpenNSW/nsw/backend v0.0.0-…  // pinned to a taskv2 commit
```

To pull the latest taskv2:

```bash
go get github.com/OpenNSW/nsw/backend@taskv2
go mod tidy
```

There is **no** `replace` directive and **no** sibling clone of `OpenNSW/nsw` required to build.

> The local module path declares itself as `github.com/OpenNSW/nsw/backend/srilanka` so Go's `internal/...` visibility rule allows importing `github.com/OpenNSW/nsw/backend/internal/*`. The physical directory name (`nsw-srilanka`) is independent of the module path.

---

## Configuration Reference

| File | Purpose | Source of truth |
|------|---------|-----------------|
| `.env` | Runtime environment (DB, Temporal, CORS, auth, storage, config paths) | `.env.example` |
| `configs/services.json` | Outbound service endpoint registry (FCAU, NPQS, IRD, customs, …) | `configs/services.example.json` |
| `configs/payment_methods.json` | Payment gateway catalogue (id, type, gateway URL, instruction template) | `configs/payment_methods.example.json` |
| `configs/fcau/` | FCAU health-certificate workflow definition + JSONForms + render configs | Per-team workflow store |

Workflow execution mechanics (input/output mappings, task plugins, render projections) are documented in the upstream `github.com/OpenNSW/nsw/backend` README and the FCAU configs themselves.
