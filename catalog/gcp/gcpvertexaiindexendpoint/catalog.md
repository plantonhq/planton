# GCP Vertex AI Index Endpoint

Deploys a Vertex AI Vector Search index endpoint: the serving surface deployed indexes answer nearest-neighbor queries through. The endpoint owns connectivity — public, VPC-peered, or Private Service Connect — while the indexes themselves are separate GcpVertexAiIndex resources placed onto it by GcpVertexAiDeployedIndex. This is a different GCP resource from the online-prediction GcpVertexAiEndpoint (which serves models). Connectivity is a one-way door: every mode is immutable, and switching means recreating the endpoint and re-deploying every index on it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Vertex AI Index Endpoint** -- a regional Vector Search endpoint with the configured connectivity mode
- **Public Serving Domain** -- when `publicEndpointEnabled` is true, a GCP-managed public domain name (surfaced as the `public_endpoint_domain_name` output) that authenticated clients query over the internet
- **VPC Peering Attachment** -- when `network` is set, the endpoint becomes reachable only inside the peered VPC via Private Services Access; both IaC modules normalize the network self-link to the relative form the Vertex AI API expects
- **PSC Service Attachment Surface** -- when `privateServiceConnectConfig` is present, allowlisted consumer projects create forwarding rules to each deployment's service attachment; no peering needed
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance
- **Vertex AI API enablement** -- `aiplatform.googleapis.com` is enabled in the target project; tearing down the endpoint never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the endpoint will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Vertex AI API itself, so the connection's principal needs permission to enable services on a fresh project.
- **For VPC-peered serving** -- Private Services Access configured on the network first: a GcpGlobalAddress (purpose VPC_PEERING) composed with a GcpServiceNetworkingConnection.
- **For PSC serving** -- the list of consumer project IDs, decided up front (the allowlist is immutable).
- **Cloud KMS key** (only for CMEK) -- a key in the same region as the endpoint, with the Vertex AI service agent granted `roles/cloudkms.cryptoKeyEncrypterDecrypter`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Index Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public Index Endpoint** preset in the [Presets](#presets) tab for the simplest path to production.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiIndexEndpoint
metadata:
  name: catalog-search
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  location: us-central1
  displayName: catalog-search-endpoint
  publicEndpointEnabled: true
```

```shell
planton apply -f index-endpoint.yaml
```

This creates a public serving endpoint: queries hit a GCP-managed domain over the internet, authenticated with Google Cloud credentials. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the full vector-search composition, deploy the endpoint (and index) before the deployment that joins them:

```yaml
# In a GcpVertexAiDeployedIndex spec:
indexEndpoint:
  valueFrom:
    kind: GcpVertexAiIndexEndpoint
    name: catalog-search
    fieldPath: status.outputs.index_endpoint_id
```

The InfraPipeline resolves the dependency graph, provisions the endpoint first, then the deployment that serves through it.

## Key Configuration

These are the most important decisions when configuring an index endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Connectivity (the one-way door)** -- exactly one of `publicEndpointEnabled: true`, `network`, or `privateServiceConnectConfig`. All three are mutually exclusive and every choice is immutable — switching modes means recreating the endpoint and re-deploying every index on it.

**Public serving** -- queries hit a GCP-managed domain over the internet, authenticated with Google Cloud credentials. Public reachability, not public access — IAM still gates every query.

**VPC peering** -- `network` references a GcpVpcNetwork (its `network_self_link` output). Requires Private Services Access on the network before the endpoint deploys. Private clients call each deployment's `match_grpc_address`.

**Private Service Connect** -- `projectAllowlist` names the consumer projects allowed to create forwarding rules; include every consumer up front (immutable). Consumers connect through each deployment's `service_attachment`.

**Encryption** -- `kmsKeyName` pins data on the endpoint's serving replicas to a customer-managed key in the same region; the Vertex AI service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it. Immutable after creation.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (optional — VPC-peered mode) | `network` | `status.outputs.network_self_link` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpVpcNetwork** (optional, per PSC automation entry) | `privateServiceConnectConfig.pscAutomationConfigs[].network` | `status.outputs.network_self_link` |
| **GcpProject** (optional, per PSC automation entry) | `privateServiceConnectConfig.pscAutomationConfigs[].projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `index_endpoint_id` | Fully qualified endpoint path (`projects/{p}/locations/{l}/indexEndpoints/{id}`) | The exact value a GcpVertexAiDeployedIndex's `indexEndpoint` join consumes |
| `index_endpoint_name` | The GCP-assigned numeric endpoint ID | Display, logging |
| `public_endpoint_domain_name` | Public query domain (populated only when public serving is enabled) | Query clients over the internet |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public serving** -- The simplest path to production: one flag, a public domain output, IAM-authenticated queries. Start from the **Public Index Endpoint** preset.

**VPC-peered private serving** -- Query traffic never leaves private address space; compose with the Private Services Access pair first. Start from the **VPC-Peered Index Endpoint** preset.

**Private Service Connect** -- Private connectivity across project boundaries without peering — the modern model for multi-project estates. Start from the **Private Service Connect Index Endpoint** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the endpoint is created
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- the VPC a peered endpoint serves inside
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- the customer-managed key for data at rest on the serving replicas
- [**GCP Global Address**](/infra-catalog/gcp-global-address) + [**GCP Service Networking Connection**](/infra-catalog/gcp-service-networking-connection) -- the Private Services Access pair a peered endpoint requires
- [**GCP Vertex AI Index**](/infra-catalog/gcp-vertex-ai-index) -- the vector index deployed onto this endpoint
- [**GCP Vertex AI Deployed Index**](/infra-catalog/gcp-vertex-ai-deployed-index) -- joins an index to this endpoint via `indexEndpoint`
