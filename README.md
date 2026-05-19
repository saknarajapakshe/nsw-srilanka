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
│       └── main.go                 # Sri Lanka customized server entrypoint
├── internal/
│   ├── plugins/
│   │   └── payments/               # Sri Lanka specific payment plugin integrations
│   └── services/                   # Sri Lanka custom services & business logic
├── .gitignore                      # Standard Git exclusion file
└── README.md                       # This project documentation
```

---

## Development & Compilation

To build and compile the Sri Lanka customized platform server, the workspace utilizes the core generic platform repository (`nsw`) under the parent path `/Users/tmp/nsw`.

### Prerequisites
- **Go**: Version `1.25.0` or higher.

### Linking & Compilation
Because the server leverages the generic bootstrap and configuration modules from `nsw/backend`, compiling the application requires establishing a symbolic link within the core module tree. This resolves standard Go internal package restrictions seamlessly.

1. **Establish the symlink** (if not already present):
   ```bash
   ln -sfn /Users/tmp/nsw-srilanka /Users/tmp/nsw/backend/srilanka
   ```

2. **Compile the server binary**:
   ```bash
   cd /Users/tmp/nsw/backend/srilanka/cmd/server
   go build -o /Users/tmp/nsw-srilanka/server main.go
   ```

3. **Run the server**:
   ```bash
   cd /Users/tmp/nsw-srilanka
   ./server
   ```

---

## Core Components

### 1. Payments (`internal/plugins/payments`)
Contains localized payment gateways, webhook handlers, and reference validators integrated specifically with Sri Lankan financial entities (e.g., local banks, LankaPay, etc.).

### 2. Services (`internal/services`)
Custom workflows, document validation, and agency-specific routers tailored to the Sri Lankan regulatory guidelines.
