# NSW Sri Lanka Platform

[![Go Version](https://img.shields.io/badge/Go-1.25.0-blue.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/NSW-Platform-green.svg)](#)

This repository (`nsw-srilanka`) serves as the deployer-specific application repository for the **Sri Lanka instance** of the National Single Window (NSW) Platform.

It houses custom integrations, domain-specific services, and local plug-ins tailored specifically for Sri Lanka's regulatory and payment ecosystems.

---

## Repository Architecture

The platform follows a decoupled, plugin-based architecture, dividing responsibilities between the generic core engine and deployer-specific implementations.

```
nsw-srilanka/
├── cmd/
│   └── server/
│       ├── db.go                   # File-backed task DB implementation
│       ├── main.go                 # Sri Lanka server wiring & bootstrapper
│       ├── server.go               # HTTP API handler routing
│       └── templates.go            # Graph/Task template loader
├── internal/
│   └── plugins/
│       └── payments/
│           └── payment.go          # Sri Lanka-specific payment plugins
├── static/                         # Portal frontend assets (UI)
├── templates/                      # Workflow definitions & task schemas
└── README.md                       # This project documentation
```

---

## How to Run Locally

To run the custom Sri Lanka platform instance locally, follow these steps:

### 1. Prerequisites
- **Go**: Version `1.25.0` or higher.
- **Temporal Server**: The local orchestrator requires a running Temporal instance on the default port (`localhost:7233`).

If you do not have Temporal installed, you can start the local development server via the Temporal CLI:
```bash
temporal server start-dev
```

### 2. Establish Workspace Symlink
Because the custom code imports core modules from `github.com/OpenNSW/nsw`, you must compile from the symlinked path under the core backend repository to satisfy internal package restrictions:
```bash
ln -sfn /Users/tmp/nsw-srilanka /Users/tmp/nsw/backend/srilanka
```

### 3. Build & Run the Server
From the core backend's symlinked workspace directory, execute the Go runner:
```bash
cd /Users/tmp/nsw/backend/srilanka
go run ./cmd/server
```

*Optional Flags:*
- Use the `-real` flag if you want to bypass mock/demo dispatchers and connect to actual external endpoints:
  ```bash
  go run ./cmd/server -real
  ```

### 4. Verify & Interact
Once the server is running:
- **Web UI Dashboard**: Navigate to [http://localhost:8080](http://localhost:8080) to interact with the split-pane Portal and Reviewer dashboard.
- **Task persistence**: Local mock task state is backed up to `/tmp/nsw_task_db.json`.

---

## Core Components

### 1. Payments (`internal/plugins/payments`)
Contains localized payment gateways, webhook handlers, and reference validators integrated specifically with Sri Lankan financial entities (e.g., local banks, LankaPay, and customized demo workflows).

### 2. Services (`internal/services`)
Custom workflows, document validation, and agency-specific routers tailored to the Sri Lankan single-window guidelines.
