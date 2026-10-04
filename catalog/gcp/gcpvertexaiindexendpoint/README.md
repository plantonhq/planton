# GcpVertexAiIndexEndpoint

A GCP Vertex AI Vector Search index endpoint is the serving surface that deployed indexes answer nearest-neighbor queries through. It owns connectivity (public, VPC-peered, or Private Service Connect); the indexes themselves are separate `GcpVertexAiIndex` resources placed onto the endpoint by `GcpVertexAiDeployedIndex`.

This is a **different GCP resource** from the online-prediction `GcpVertexAiEndpoint` (which serves models): an IndexEndpoint serves vector-search queries only.

## When to Use

Use `GcpVertexAiIndexEndpoint` when you need:

- A serving surface for one or more Vector Search indexes
- Public querying over a managed domain name
- Private querying inside a peered VPC or through Private Service Connect
- Infrastructure-as-code management of the vector-search serving lifecycle

## What This Kind Creates

This kind provisions a single Vector Search index endpoint. Placing an index onto it is modeled by `GcpVertexAiDeployedIndex` -- the endpoint alone serves nothing.

## Key Configuration Options

### Connectivity (immutable)

Three mutually exclusive modes:

1. **Public** -- set `publicEndpointEnabled: true`. Queries go to the `public_endpoint_domain_name` output.
2. **VPC-peered** -- set `network`. Reachable only inside the peered VPC; requires Private Services Access on the network. The IaC modules normalize a network self-link (the `GcpVpcNetwork` reference's canonical output) to the relative form the Vertex AI API expects.
3. **Private Service Connect** -- set `privateServiceConnectConfig`. Consumers connect through a service attachment (surfaced on the deployed index's outputs); no peering needed. Wire consumers through `projectAllowlist` (they create their own forwarding rules): `pscAutomationConfigs` is refused by validation on this kind because the live Vertex AI API accepts and silently discards it on index endpoints -- nothing is stored and no consumer-side endpoint is ever provisioned (automation works on online-prediction endpoints only).

Omitting all three creates an endpoint that only Google-internal networks can reach -- almost always you want one of the arms.

### Encryption (immutable)

`kmsKeyName` (reference a `GcpKmsKey`) turns on customer-managed encryption (CMEK) for data on the endpoint's serving replicas; the Vertex AI service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key. Empty means Google-managed keys.

### Destroy behavior

`deletionPolicy` is the client-side lever: empty/`DELETE` deletes the endpoint (and stops every deployment on it), `PREVENT` makes destroy fail, `ABANDON` removes it from management but keeps it serving (and billing).

### Labels

User-defined `labels` organize the endpoint for cost attribution and ownership; they merge beneath the platform's attribution labels identically on both engines.

## Outputs

| Output | Description |
|--------|-------------|
| `index_endpoint_id` | Fully qualified endpoint path -- the deployed index's composition key |
| `index_endpoint_name` | The GCP-assigned numeric endpoint ID |
| `public_endpoint_domain_name` | Public query domain (public arm only) |
| `create_time` | Creation timestamp |
| `update_time` | Last-update timestamp |

## Presets

- **public** -- Public endpoint for straightforward querying
- **vpc-peered** -- Private endpoint inside a peered VPC
- **psc** -- Private Service Connect with consumer allowlist

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
