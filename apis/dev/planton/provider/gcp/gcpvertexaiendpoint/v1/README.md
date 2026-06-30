# GcpVertexAiEndpoint

A GCP Vertex AI Endpoint is a stable serving surface for deploying machine learning models in production. It provides a durable prediction URL and configurable networking boundary so that teams can deploy, update, and route traffic to ML models independently of the infrastructure layer.

## When to Use

Use `GcpVertexAiEndpoint` when you need:

- A persistent prediction URL for serving ML models via Vertex AI
- Private networking (VPC peering or Private Service Connect) for model serving
- Customer-managed encryption (CMEK) for endpoint data at rest
- Dedicated DNS for isolated, higher-performance prediction traffic
- Infrastructure-as-code management of the endpoint lifecycle

## What This Component Creates

This component provisions a single Vertex AI Endpoint. Model deployment to the endpoint is an operational step performed separately via the Vertex AI API, SDK, or console -- it is not managed by this infrastructure component.

## Key Configuration Options

### Networking

Three mutually exclusive networking modes:

1. **Public** (default) -- The endpoint is accessible via the shared regional DNS (`{region}-aiplatform.googleapis.com`). No additional configuration needed.

2. **VPC-peered** -- The endpoint is accessible only within a peered VPC network. Requires Private Services Access configured on the VPC. Set the `network` field.

3. **Private Service Connect** -- The endpoint is exposed via a PSC service attachment. Provides the strongest network isolation without VPC peering. Set `privateServiceConnectConfig`.

### Dedicated Endpoint DNS

Setting `dedicatedEndpointEnabled: true` provisions a dedicated DNS name (`https://{endpointId}.{region}-{projectNumber}.prediction.vertexai.goog`) for better performance and traffic isolation. Not available with Private Service Connect.

### Encryption

By default, Google-managed encryption is used. For customer-managed encryption (CMEK), set `kmsKeyName` to a Cloud KMS key resource path. This is immutable after creation.

### Endpoint Name

Vertex AI endpoints use **numeric-only** identifiers (max 10 digits). Most users should omit `endpointName` and let the IaC module auto-generate it. The `displayName` field serves as the human-readable identifier.

## Outputs

| Output | Description |
|--------|-------------|
| `endpoint_id` | Fully qualified endpoint path |
| `display_name` | Human-readable display name |
| `dedicated_endpoint_dns` | Dedicated DNS (if enabled) |
| `create_time` | Creation timestamp |

## Presets

- **basic-public** -- Minimal public endpoint
- **private-vpc-peered** -- VPC-peered with CMEK encryption
- **private-psc** -- Private Service Connect with project allowlist
