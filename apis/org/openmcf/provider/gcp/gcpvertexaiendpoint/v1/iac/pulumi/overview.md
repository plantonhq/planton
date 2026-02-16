# Pulumi Module: GcpVertexAiEndpoint

## Architecture

The Pulumi module creates a single GCP resource:

- **`vertex.AiEndpoint`** -- The Vertex AI Endpoint serving surface

## File Structure

```
module/
├── main.go          # Entry point: Resources()
├── locals.go        # Variable initialization and GCP labels
├── ai_endpoint.go   # Vertex AI Endpoint resource creation
└── outputs.go       # Output key constants
```

## Resource Flow

1. `main.go` initializes locals and obtains the GCP provider
2. `ai_endpoint.go` creates the endpoint with conditional networking and encryption
3. Outputs are exported: endpoint_id, display_name, dedicated_endpoint_dns, create_time

## Networking Modes

The module supports three mutually exclusive networking modes:

- **Public** (default): No network or PSC config set
- **VPC-peered**: `network` field set to a VPC self-link
- **Private Service Connect**: `private_service_connect_config` block present

## Key Implementation Details

- Framework GCP labels are applied automatically
- CMEK encryption is optional via `kms_key_name`
- The `endpoint_name` (numeric GCP resource ID) is optional; Pulumi auto-generates when omitted
- PSC block implies `enable_private_service_connect = true` internally
