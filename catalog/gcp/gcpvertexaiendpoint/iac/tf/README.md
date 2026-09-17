# GcpVertexAiEndpoint Terraform Module

This directory contains the Terraform module for provisioning a GCP Vertex AI Endpoint.

## Usage

```hcl
module "vertex_ai_endpoint" {
  source = "./path/to/module"

  metadata = {
    name = "my-endpoint"
  }

  spec = {
    location     = "us-central1"
    display_name = "My ML Endpoint"
  }
}
```

`spec.project_id` is optional: when empty, the endpoint lands in the provider's
default project.

## Inputs

| Name | Type | Required | Description |
|------|------|----------|-------------|
| metadata | object | yes | Planton resource metadata |
| spec | object | yes | GcpVertexAiEndpoint specification |

## Outputs

| Name | Description |
|------|-------------|
| endpoint_id | Fully qualified endpoint resource path |
| display_name | Display name of the endpoint |
| dedicated_endpoint_dns | DNS of the dedicated endpoint (if enabled) |
| create_time | RFC3339 creation timestamp |
| endpoint_name | The numeric endpoint ID (explicit or identity-derived) |

## Networking

Three mutually exclusive modes:

1. **Public** (default) -- No network or PSC config
2. **VPC-peered** -- Set `spec.network`
3. **Private Service Connect** -- Set `spec.private_service_connect_config`

## Endpoint Name

Vertex AI endpoints require a **numeric-only** name (max 10 digits, no leading
zeros) and the API never generates one. When `spec.endpoint_name` is empty, the
module derives a stable identifier from the resource identity (sha256 of
`"org/env/name"`, first 48 bits, mapped into `[1000000000, 9999999999]`). The
Pulumi module implements the identical derivation, so the same manifest yields
the same endpoint ID on either engine.

## Provider Requirements

- `hashicorp/google` ~> 8.3
