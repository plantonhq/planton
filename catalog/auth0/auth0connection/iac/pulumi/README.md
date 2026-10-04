# Auth0Connection Pulumi Module

This directory contains the Pulumi implementation for the Auth0Connection kind.

## Overview

The Auth0Connection Pulumi module creates and manages Auth0 identity connections, including:
- Database connections (Auth0 hosted)
- Social identity providers (Google, Facebook, GitHub, etc.)
- Enterprise SSO (SAML, OIDC, Azure AD)

## Prerequisites

1. **Pulumi CLI**: Install from https://www.pulumi.com/docs/install/
2. **Go 1.21+**: Required for building the module
3. **Auth0 Account**: With a Machine-to-Machine application configured

## Environment Variables

The module reads IaC input from the `IAC_INPUT_FILE` environment variable:

```bash
export IAC_INPUT_FILE=/path/to/manifest.yaml
```

Alternatively, Auth0 credentials can be provided via environment variables:
- `AUTH0_DOMAIN`: Your Auth0 tenant domain
- `AUTH0_CLIENT_ID`: M2M application client ID
- `AUTH0_CLIENT_SECRET`: M2M application client secret

## Usage

### Build the Module

```bash
make build
```

### Install Pulumi Plugins

```bash
make install-pulumi-plugins
```

### Run with Test Manifest

```bash
make test
```

### Direct Pulumi Commands

```bash
# Initialize stack
pulumi stack init local

# Preview changes
IAC_INPUT_FILE=../../e2e/manifest.yaml pulumi preview

# Apply changes
IAC_INPUT_FILE=../../e2e/manifest.yaml pulumi up

# Destroy resources
IAC_INPUT_FILE=../../e2e/manifest.yaml pulumi destroy
```

## Module Structure

```
pulumi/
├── main.go           # Entry point, loads IaC input and calls module
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and test automation
├── debug.sh          # Debug helper script
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Resources orchestration
    ├── locals.go     # Local value initialization
    ├── outputs.go    # Output exports
    └── connection.go # Auth0 connection resource creation
```

## Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Auth0 connection ID |
| `name` | Connection name |
| `strategy` | Connection strategy type |
| `is_enabled` | Whether the connection has enabled clients |
| `enabled_client_ids` | List of enabled client IDs |
| `realms` | Connection realms |

## Troubleshooting

### "failed to create Auth0 provider"

Ensure Auth0 credentials are correctly configured either via:
- Provider config in IaC input
- Environment variables

### "connection already exists"

Auth0 connection names must be unique within a tenant. Either:
- Delete the existing connection
- Use a different name in the manifest

### Plugin Not Found

Run `make install-pulumi-plugins` to install the Auth0 provider plugin.

## Related Documentation

- [Auth0Connection spec.proto](../../spec.proto)

