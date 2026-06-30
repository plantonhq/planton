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
    project_id   = { value = "my-gcp-project" }
    location     = "us-central1"
    display_name = "My ML Endpoint"
  }
}
```

## Inputs

| Name | Type | Required | Description |
|------|------|----------|-------------|
| metadata | object | yes | Planton resource metadata |
| spec | object | yes | GcpVertexAiEndpoint specification |
| provider_config | object | no | GCP provider configuration |

## Outputs

| Name | Description |
|------|-------------|
| endpoint_id | Fully qualified endpoint resource path |
| display_name | Display name of the endpoint |
| dedicated_endpoint_dns | DNS of the dedicated endpoint (if enabled) |
| create_time | RFC3339 creation timestamp |

## Networking

Three mutually exclusive modes:

1. **Public** (default) -- No network or PSC config
2. **VPC-peered** -- Set `spec.network`
3. **Private Service Connect** -- Set `spec.private_service_connect_config`

## Endpoint Name

Vertex AI endpoints require a **numeric-only** name (max 10 digits, no leading zeros).
When `spec.endpoint_name` is not provided, the module auto-generates a stable 10-digit
numeric identifier using `random_integer` with keepers tied to the display name and location.

## Provider Requirements

- `hashicorp/google` ~> 6.0
- `hashicorp/random` ~> 3.0 (for endpoint name auto-generation)
