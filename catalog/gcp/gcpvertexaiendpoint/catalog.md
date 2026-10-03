# GCP Vertex AI Endpoint

Deploys a Vertex AI Endpoint -- a stable serving surface for deploying machine learning models -- with configurable networking (public, VPC-peered, or Private Service Connect), optional customer-managed encryption via Cloud KMS, dedicated DNS for traffic isolation, and request/response logging into BigQuery. The networking mode is a create-time decision: moving between public, peered, and PSC recreates the endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Vertex AI Endpoint** -- a managed `vertex.AiEndpoint` in the specified GCP project and region, configured with the chosen networking mode and display name
- **Encryption Configuration** -- when `kmsKeyName` is provided, configures customer-managed encryption (CMEK) for data at rest on the endpoint; otherwise uses Google-managed encryption
- **VPC-Peered Networking** -- created only when `network` is set; places the endpoint behind a VPC peering connection, accessible only from within the peered network (requires Private Services Access on the VPC)
- **Private Service Connect Configuration** -- created only when `privateServiceConnectConfig` is set; exposes the endpoint via a PSC service attachment with an optional project allowlist for fine-grained access control
- **Dedicated Endpoint DNS** -- created only when `dedicatedEndpointEnabled` is true; provisions a dedicated prediction URL (`{endpointId}.{region}-{projectNumber}.prediction.vertexai.goog`) for isolated traffic
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance
- **Vertex AI API enablement** -- `aiplatform.googleapis.com` is enabled in the target project; tearing down the endpoint never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the endpoint will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Vertex AI API itself, so the connection's principal needs permission to enable services on a fresh project.
- **Private Services Access** (if using VPC peering) -- the VPC network must have a private services connection configured for the `servicenetworking.googleapis.com` API.
- **Cloud KMS key** (if using CMEK) -- a key in the same region as the endpoint, with the Vertex AI service agent granted the `cloudkms.cryptoKeyEncrypterDecrypter` role.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Public Endpoint** preset in the [Presets](#presets) tab to pre-populate a minimal public endpoint configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiEndpoint
metadata:
  name: ml-serving
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  location: us-central1
  displayName: ML Serving Endpoint
```

```shell
planton apply -f vertex-ai-endpoint.yaml
```

This creates a public Vertex AI Endpoint with Google-managed encryption and no private networking. Models are deployed to the endpoint separately via the Vertex AI API or console. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the endpoint to a GCP project, VPC, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: ml-encryption-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project, VPC, and KMS key first, then provisions the Vertex AI Endpoint with VPC-peered private networking and CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Vertex AI Endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Networking mode** -- Three mutually exclusive modes are available. Public (default) exposes the endpoint via the shared regional DNS. VPC-peered (`network`) restricts access to within a peered VPC and requires Private Services Access. Private Service Connect (`privateServiceConnectConfig`) provides the strongest isolation via a PSC service attachment with a project allowlist.

**Dedicated endpoint DNS** -- Set `dedicatedEndpointEnabled: true` to provision a dedicated prediction URL instead of the shared regional endpoint. Dedicated DNS provides better performance and traffic isolation. Not available with Private Service Connect.

**Customer-managed encryption** -- Set `kmsKeyName` to a Cloud KMS key resource path for CMEK encryption. The key must be in the same region as the endpoint. If omitted, Google-managed encryption is used. Immutable after creation.

**Display name and endpoint name** -- `displayName` is the human-readable identifier (up to 128 characters). `endpointName` is an optional numeric GCP resource identifier (1-10 digits). Most users should omit `endpointName` and use `displayName` for identification. Both fields serve different purposes -- `displayName` is mutable while `endpointName` is immutable.

**Request/response logging** -- `requestResponseLoggingConfig` samples prediction traffic into a BigQuery table (`bq://project`, `bq://project.dataset`, or `bq://project.dataset.table`), the raw material for drift monitoring and audit. Tune `samplingRate` (in `(0, 1]`) to bound BigQuery cost on high-QPS endpoints. Mutable in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional) | `network` | `status.outputs.network_self_link` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpVpcNetwork** (optional, per PSC automation entry) | `privateServiceConnectConfig.pscAutomationConfigs[].network` | `status.outputs.network_self_link` |
| **GcpProject** (optional, per PSC automation entry) | `privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint_id` | Fully qualified endpoint resource path (`projects/{project}/locations/{location}/endpoints/{name}`) | Model deployment targets, monitoring dashboards |
| `display_name` | Display name of the endpoint | Application configuration, inventory tracking |
| `dedicated_endpoint_dns` | Dedicated prediction URL (populated only when `dedicatedEndpointEnabled` is true) | Application prediction client configuration |
| `endpoint_name` | The numeric endpoint ID (explicit or derived from the resource identity) | Model-deploy tooling that addresses the endpoint by ID |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic public endpoint** -- A public prediction URL with Google-managed encryption and no private networking. The simplest configuration for development, testing, or workloads where network isolation is not required. Start from the **Basic Public Endpoint** preset.

**Private VPC-peered endpoint** -- VPC peering for network isolation, CMEK encryption for data protection, and dedicated DNS for performance. Suitable for regulated environments (HIPAA, PCI, SOC 2). Start from the **Private VPC-Peered Endpoint** preset.

**Private Service Connect endpoint** -- The strongest network isolation using PSC with an explicit project allowlist. Suited to multi-tenant environments or cross-project model serving without VPC peering. Start from the **Private PSC Endpoint** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the endpoint is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for private endpoint access via VPC peering
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the customer-managed encryption key for data at rest