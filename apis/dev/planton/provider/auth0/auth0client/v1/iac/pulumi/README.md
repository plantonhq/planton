# Auth0Client Pulumi Module

This directory contains the Pulumi implementation for the Auth0Client deployment component.

## Overview

The Auth0Client Pulumi module creates and manages Auth0 applications (clients), including:
- Single Page Applications (SPAs)
- Native applications (mobile/desktop)
- Regular web applications (server-side)
- Machine-to-Machine (M2M) applications

## Prerequisites

1. **Pulumi CLI**: Install from https://www.pulumi.com/docs/install/
2. **Go 1.21+**: Required for building the module
3. **Auth0 Account**: With a Machine-to-Machine application configured

## Environment Variables

The module reads stack input from the `STACK_INPUT_FILE` environment variable:

```bash
export STACK_INPUT_FILE=/path/to/manifest.yaml
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
STACK_INPUT_FILE=../hack/manifest.yaml pulumi preview

# Apply changes
STACK_INPUT_FILE=../hack/manifest.yaml pulumi up

# Destroy resources
STACK_INPUT_FILE=../hack/manifest.yaml pulumi destroy
```

## Module Structure

```
pulumi/
├── main.go           # Entry point, loads stack input and calls module
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and test automation
├── debug.sh          # Debug helper script
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Resources orchestration
    ├── locals.go     # Local value initialization
    ├── outputs.go    # Stack output exports
    └── client.go     # Auth0 client resource creation
```

## Stack Outputs

After deployment, the following outputs are available:

| Output | Description |
|--------|-------------|
| `id` | Auth0 client internal ID |
| `client_id` | OAuth 2.0 client identifier (public) |
| `client_secret` | OAuth 2.0 client secret (confidential) |
| `name` | Application name |
| `application_type` | Application type (spa, native, etc.) |
| `signing_keys` | JWT signing keys |
| `token_endpoint_auth_method` | Token endpoint authentication method |

## Troubleshooting

### "failed to create Auth0 provider"

Ensure Auth0 credentials are correctly configured either via:
- Provider config in stack input
- Environment variables

### "client already exists"

Auth0 client names don't need to be unique, but you may want to check for duplicates. Either:
- Delete the existing client if intended
- Verify you're not creating a duplicate

### Plugin Not Found

Run `make install-pulumi-plugins` to install the Auth0 provider plugin.

## Related Documentation

- [Auth0Client spec.proto](../../spec.proto)
- [Examples](../../examples.md)
- [Research Documentation](../../docs/README.md)


