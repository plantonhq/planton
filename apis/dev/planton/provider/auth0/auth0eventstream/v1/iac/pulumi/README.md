# Auth0EventStream Pulumi Module

This Pulumi module creates and manages Auth0 Event Streams for real-time event delivery.

## Overview

The module creates an `auth0.EventStream` resource with support for both EventBridge and Webhook destinations.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- Auth0 tenant with Management API access
- For EventBridge: AWS account with EventBridge access

## Usage

### Install Dependencies

```bash
make deps
```

### Build

```bash
make build
```

### Run Tests

```bash
# Set up Auth0 credentials
export AUTH0_DOMAIN="your-tenant.auth0.com"
export AUTH0_CLIENT_ID="your-client-id"
export AUTH0_CLIENT_SECRET="your-client-secret"

# Run with test manifest
make test
```

### Debug Locally

```bash
./debug.sh
```

## Stack Input

The module expects a stack input with the following structure:

```yaml
target:
  apiVersion: auth0.planton.dev/v1
  kind: Auth0EventStream
  metadata:
    name: my-event-stream
  spec:
    destination_type: webhook
    subscriptions:
      - user.created
    webhook_configuration:
      webhook_endpoint: "https://example.com/webhook"
      webhook_authorization:
        method: bearer
        token: "secret-token"

auth0_provider_config:
  domain: your-tenant.auth0.com
  client_id: your-client-id
  client_secret: your-client-secret
```

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Event stream ID |
| `name` | Stream name |
| `status` | Current status |
| `destination_type` | Destination type |
| `subscriptions` | Subscribed event types |
| `created_at` | Creation timestamp |
| `updated_at` | Last update timestamp |
| `eventbridge_configuration` | EventBridge config (if applicable) |

## Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point |
| `Pulumi.yaml` | Project configuration |
| `Makefile` | Build targets |
| `debug.sh` | Debug script |
| `module/main.go` | Provider setup |
| `module/locals.go` | Input processing |
| `module/eventstream.go` | Event stream creation |
| `module/outputs.go` | Output exports |

