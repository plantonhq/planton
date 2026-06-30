# Auth0EventStream Terraform Module

This Terraform module creates and manages Auth0 Event Streams for real-time event delivery.

## Overview

The module creates an `auth0_event_stream` resource with support for both EventBridge and Webhook destinations.

## Prerequisites

- Terraform >= 1.0
- Auth0 Provider >= 1.0
- Auth0 tenant with Management API access

## Usage

### Basic Webhook Example

```hcl
module "auth0_event_stream" {
  source = "./path/to/module"

  auth0_credential = {
    domain        = "your-tenant.auth0.com"
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
  }

  metadata = {
    name = "user-events"
  }

  spec = {
    destination_type = "webhook"
    subscriptions    = ["user.created", "user.updated"]
    
    webhook_configuration = {
      webhook_endpoint = "https://api.example.com/webhooks/auth0"
      webhook_authorization = {
        method = "bearer"
        token  = "your-secret-token"
      }
    }
  }
}
```

### EventBridge Example

```hcl
module "auth0_event_stream" {
  source = "./path/to/module"

  auth0_credential = {
    domain        = "your-tenant.auth0.com"
    client_id     = "your-client-id"
    client_secret = "your-client-secret"
  }

  metadata = {
    name = "security-events"
  }

  spec = {
    destination_type = "eventbridge"
    subscriptions    = ["authentication.success", "authentication.failure"]
    
    eventbridge_configuration = {
      aws_account_id = "123456789012"
      aws_region     = "us-east-1"
    }
  }
}
```

## Inputs

| Name | Description | Type | Required |
|------|-------------|------|----------|
| `auth0_credential` | Auth0 API credentials | object | Yes |
| `metadata` | Resource metadata including name | object | Yes |
| `spec` | Event stream specification | object | Yes |

### auth0_credential Object

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `domain` | Auth0 tenant domain | string | Yes |
| `client_id` | M2M application client ID | string | Yes |
| `client_secret` | M2M application client secret | string | Yes |

### spec Object

| Field | Description | Type | Required |
|-------|-------------|------|----------|
| `destination_type` | Destination type (eventbridge, webhook) | string | Yes |
| `subscriptions` | List of event types to subscribe | list(string) | Yes |
| `eventbridge_configuration` | EventBridge config | object | Conditional |
| `webhook_configuration` | Webhook config | object | Conditional |

## Outputs

| Name | Description |
|------|-------------|
| `id` | Event stream ID |
| `name` | Stream name |
| `status` | Current status |
| `destination_type` | Destination type |
| `subscriptions` | Subscribed event types |
| `created_at` | Creation timestamp |
| `updated_at` | Last update timestamp |
| `aws_partner_event_source` | AWS partner event source (EventBridge only) |

## Files

| File | Purpose |
|------|---------|
| `provider.tf` | Auth0 provider configuration |
| `variables.tf` | Input variable definitions |
| `locals.tf` | Local value computations |
| `main.tf` | Event stream resource |
| `outputs.tf` | Output definitions |

## Important Notes

### EventBridge Immutability

EventBridge configuration (`aws_account_id`, `aws_region`) cannot be changed after creation. Any modification will force resource recreation.

### Webhook Updates

Webhook configuration can be updated after creation, including:
- `webhook_endpoint`
- `webhook_authorization` settings

### Sensitive Values

The following values are marked as sensitive:
- `auth0_credential.client_secret`
- `webhook_authorization.password`
- `webhook_authorization.token`

