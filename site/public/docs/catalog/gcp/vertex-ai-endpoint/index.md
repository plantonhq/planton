---
title: "Vertex AI Endpoint"
description: "Vertex AI Endpoint deployment documentation"
icon: "package"
order: 100
componentName: "gcpvertexaiendpoint"
---

# GCP Vertex AI Endpoint

Deploys a GCP Vertex AI Endpoint — a stable serving surface for machine learning models with configurable networking (public, VPC-peered, or Private Service Connect), optional CMEK encryption, and optional dedicated DNS for isolated prediction traffic. Model deployment to the endpoint is a separate operational step.

## What Gets Created

When you deploy a GcpVertexAiEndpoint resource, OpenMCF provisions:

- **Vertex AI Endpoint** — a `google_vertex_ai_endpoint` resource in the specified region with framework GCP labels applied automatically
- **Random Endpoint Name** (Terraform only) — a `random_integer` resource to generate the required numeric endpoint identifier, created only when `endpointName` is not specified

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** with the Vertex AI API enabled
- **A VPC network with Private Services Access** if using VPC-peered networking (`network` field)
- **A Cloud KMS key** in the same region as the endpoint if using CMEK encryption (`kmsKeyName` field)
- **IAM permissions** — the Vertex AI service agent must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the KMS key if CMEK is enabled

## Quick Start

Create a file `endpoint.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: my-endpoint
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpVertexAiEndpoint.my-endpoint
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: My ML Endpoint
```

Deploy:

```shell
openmcf apply -f endpoint.yaml
```

This creates a public Vertex AI Endpoint accessible via the shared regional DNS with Google-managed encryption.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the endpoint is created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `location` | `string` | Region for the endpoint (e.g., `us-central1`). Immutable after creation. | Required, min length 1 |
| `displayName` | `string` | Human-readable name for the endpoint. | Required, 1-128 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Description of the endpoint. |
| `network` | `StringValueOrRef` | — | VPC network for private endpoints via VPC peering. Format: `projects/{project}/global/networks/{network}`. Mutually exclusive with `privateServiceConnectConfig`. Immutable. Can reference a GcpVpc resource via `valueFrom`. |
| `kmsKeyName` | `StringValueOrRef` | — | Cloud KMS key for CMEK encryption. Format: `projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}`. Immutable. Can reference a GcpKmsKey resource via `valueFrom`. |
| `dedicatedEndpointEnabled` | `bool` | `false` | Enables a dedicated DNS name for better performance and traffic isolation. Mutually exclusive with `privateServiceConnectConfig`. |
| `privateServiceConnectConfig` | `object` | — | Private Service Connect configuration. Mutually exclusive with `network` and `dedicatedEndpointEnabled`. |
| `privateServiceConnectConfig.projectAllowlist` | `string[]` | `[]` | Projects allowed to create forwarding rules targeting this endpoint. |
| `endpointName` | `string` | auto-generated | Numeric-only GCP resource identifier (1-10 digits, no leading zeros). Most users should omit this and use `displayName` for identification. Immutable. |

## Examples

### Public Endpoint with Description

A public endpoint for development or non-sensitive workloads:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: dev-recommendations
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: ml-platform
    pulumi.openmcf.org/stack.name: dev.GcpVertexAiEndpoint.dev-recommendations
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Dev Recommendation Engine
  description: Development endpoint for A/B testing recommendation models
```

### VPC-Peered Private Endpoint with CMEK

Production endpoint with network isolation and customer-managed encryption:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: prod-scoring
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: ml-platform
    pulumi.openmcf.org/stack.name: prod.GcpVertexAiEndpoint.prod-scoring
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Production Scoring Endpoint
  network:
    value: projects/123456789/global/networks/prod-vpc
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/ml-ring/cryptoKeys/endpoint-key
  dedicatedEndpointEnabled: true
```

### Private Service Connect Endpoint

Strongest network isolation using PSC with an explicit project allowlist:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: psc-inference
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: ml-platform
    pulumi.openmcf.org/stack.name: prod.GcpVertexAiEndpoint.psc-inference
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: PSC Inference Endpoint
  privateServiceConnectConfig:
    projectAllowlist:
      - consumer-project-a
      - consumer-project-b
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/ml-ring/cryptoKeys/endpoint-key
```

### Using Foreign Key References

Reference other OpenMCF-managed resources for composable infrastructure:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpVertexAiEndpoint
metadata:
  name: composed-endpoint
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: ml-platform
    pulumi.openmcf.org/stack.name: prod.GcpVertexAiEndpoint.composed-endpoint
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: ml-project
      field: status.outputs.project_id
  location: us-central1
  displayName: Composed ML Endpoint
  network:
    valueFrom:
      kind: GcpVpc
      name: ml-vpc
      field: status.outputs.network_self_link
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: ml-encryption-key
      field: status.outputs.key_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `endpoint_id` | `string` | Fully qualified endpoint resource path: `projects/{project}/locations/{location}/endpoints/{name}` |
| `display_name` | `string` | Display name of the endpoint |
| `dedicated_endpoint_dns` | `string` | DNS of the dedicated endpoint. Populated only when `dedicatedEndpointEnabled` is `true`. Format: `https://{endpointId}.{region}-{projectNumber}.prediction.vertexai.goog` |
| `create_time` | `string` | RFC3339 timestamp of when the endpoint was created |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project for the endpoint
- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network for VPC-peered private endpoints
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — provides the encryption key for CMEK
- [GcpVertexAiNotebook](/docs/catalog/gcp/vertex-ai-notebook) — commonly co-deployed for ML development workflows
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — provides the service identity for model serving
