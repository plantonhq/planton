# GCP KMS Key Ring

Deploys a Cloud KMS key ring -- a permanent organizational container for cryptographic keys in a GCP project. Key rings are scoped to a specific location (region, multi-region, or global) and group CryptoKeys by geographic or organizational boundaries. Key rings can never be deleted from GCP, so the name and location are permanent decisions.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **KMS Key Ring** -- a `kms.KeyRing` in the specified GCP project and location, serving as the container for CryptoKeys created afterward
- **Cloud KMS API enablement** -- `cloudkms.googleapis.com` enabled in the target project (never disabled on destroy)

Key rings are permanent GCP resources -- they cannot be deleted. On destroy, the IaC module removes the resource from state but the key ring persists in GCP.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the key ring will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Cloud KMS API itself.

## Deploy

### Console

Open the deployment store, find **GCP KMS Key Ring**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Regional Key Ring** preset in the [Presets](#presets) tab to pre-populate a key ring co-located with your workload region.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyRing
metadata:
  name: prod-encryption
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  keyRingName: prod-encryption
  location: us-central1
```

```shell
planton apply -f gcp-kms-key-ring.yaml
```

This creates a regional key ring in `us-central1`. The key ring is empty -- CryptoKeys are added separately as GcpKmsKey Infra Components. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the key ring to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the key ring with the resolved project ID.

## Key Configuration

These are the most important decisions when configuring a key ring. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Location** -- Set `location` to a GCP region (`us-central1`), multi-region (`us`, `europe`, `asia`), or `global`. Regional locations co-locate keys with data for lowest latency and strictest data residency. Multi-region locations provide continental availability with geographic boundaries. Global provides universal access with no residency guarantees. Location is immutable after creation.

**Key ring name** -- `keyRingName` is the permanent GCP resource name (1-63 characters, letters, digits, hyphens, or underscores). Key rings cannot be deleted from GCP, so choose a name that reflects long-term organizational structure (e.g., `prod-encryption`, `compliance-keys`).

**Project placement** -- Key rings belong to a single GCP project. Use a dedicated security or shared-services project for centralized key management, or place key rings in the same project as the workloads they protect for simpler IAM.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `key_ring_id` | Fully qualified key ring path (`projects/{p}/locations/{l}/keyRings/{name}`) | GcpKmsKey `keyRingId` field for creating CryptoKeys in this ring |
| `key_ring_name` | Short name of the key ring | Display, logging, human-readable references |
| `location` | The location the ring resides in, exactly as GCP resolved it | Consumers that take a bare ring name plus a location compose from `key_ring_name` + this |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Regional key ring** -- Key ring in a specific GCP region, co-locating encryption keys with the workloads they protect. The standard choice for production deployments with data residency requirements. Start from the **Regional Key Ring** preset.

**Global key ring** -- Key ring in the `global` location, accessible from any region. Suitable for shared signing keys or application-level encryption without data residency constraints. Start from the **Global Key Ring** preset.

**Multi-region key ring** -- Key ring in a multi-region location (`us`, `europe`, `asia`), replicated across all regions in the geography. Provides high availability while maintaining continental data residency for GDPR or similar requirements. Start from the **Multi-Region Key Ring** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the key ring is created
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- the CryptoKeys created inside this ring; its `keyRingId` field consumes the `key_ring_id` output