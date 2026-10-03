# GCP Secret Manager Secret

Creates a Secret Manager secret — the container for versioned secret payloads (API keys, passwords, certificates) — with replication or regional residency, rotation reminders, expiry, CMEK, an optional first version, and secret-scoped IAM grants. One manifest takes a workload from nothing to a readable secret: container, payload, and the `secretAccessor` grant to the consuming service account.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Secret** -- global (with replication control) or regional (payloads never leave the region)
- **Secret Version** (optional) -- version 1 seeded from `initialVersion.data`
- **IAM Members** (optional) -- additive secret-scoped grants
- **Secret Manager API enablement** -- `secretmanager.googleapis.com` (never disabled on destroy)
- **GCP Labels** -- resource metadata labels applied automatically

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** (directly or via a GcpProject reference).
- **IAM**: `roles/secretmanager.admin` for the deploying identity.
- **For CMEK**: a GcpKmsKey with the Secret Manager service agent granted `roles/cloudkms.cryptoKeyEncrypterDecrypter`.

## Deploy

### Console

Open the deployment store, find **GCP Secret Manager Secret**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the secret's scope and replication, the first payload, and the access grants. Start from the **App Secret with Access** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSecretManagerSecret
metadata:
  name: db-password
  org: acme-corp
  env: prod
spec:
  initialVersion:
    data:
      value: super-secret-value
  iamMembers:
    - role: roles/secretmanager.secretAccessor
      member:
        value: serviceAccount:app@acme-prod.iam.gserviceaccount.com
```

```shell
planton apply -f secret.yaml
```

This creates the secret with automatic replication, stores the payload as version 1, and grants the workload read access — a READABLE secret from one manifest. The payload literal rides the platform's managed-secret handling; it never sits in plaintext in the control plane.

### InfraChart

The chart-safe payload story — wire the data from a producing resource instead of a literal:

```yaml
spec:
  initialVersion:
    data:
      valueFrom:
        kind: GcpServiceAccount
        name: ci-signer
        fieldPath: status.outputs.key_base64
  iamMembers:
    - role: roles/secretmanager.secretAccessor
      member:
        valueFrom:
          kind: GcpServiceAccount
          name: app-runtime
          fieldPath: status.outputs.member
```

The InfraPipeline resolves the dependency graph, deploys the producing service accounts first, then provisions the secret with the resolved payload and grant.

## Key Configuration

These are the most important decisions when configuring a secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- omit `region` for a GLOBAL secret (replication control: automatic by default, or pinned `userManaged` replicas, each optionally CMEK-encrypted); set it for a REGIONAL secret whose payloads never leave the region — the data-residency posture, with CMEK attached directly.

**The first version** -- `initialVersion.data` seeds version 1 so consumers can read immediately. It is immutable by design: rotations add NEW versions, and `versionAliases` re-points consumers without touching them.

**Access** -- `iamMembers` grants are secret-SCOPED and additive: `roles/secretmanager.secretAccessor` to each consuming workload's service account, composing safely with grants made elsewhere.

**Version aliases have a temporal constraint** -- GCP validates `versionAliases` against EXISTING versions, so a first apply that both seeds `initialVersion` and aliases it is rejected ("Aliases cannot be assigned to versions that don't exist"). Deploy first, then add the alias on a subsequent apply once version 1 exists.

**Safety rails** -- `deletionProtection` (engine-side) plus `deletionPolicy: PREVENT` (API-side) for production credentials; `versionDestroyTtl` adds a restore window to version destruction; `expireTime`/`ttl` auto-delete short-lived secrets. Whole-secret deletion removes every version regardless of `versionDestroyTtl`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpServiceAccount** | `iamMembers[].member` | `status.outputs.member` |
| **GcpKmsKey** (optional) | CMEK fields | `status.outputs.key_id` |
| **GcpPubSubTopic** (optional) | `topics[]` | `status.outputs.topic_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_name` | Full resource name | Cloud Run `valueFromSecret` mounts, GKE CSI driver, anything that reads secrets by name |
| `secret_id` | The short ID | CLI and console references |
| `latest_version_name` | Version 1's resource name | Consumers pinning an exact version instead of the `latest` alias |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**App secret with access** -- payload + `secretAccessor` grant in one manifest: the standard workload-credential shape. Start from the **App Secret with Access** preset.

**Regional residency secret** -- a regional secret with CMEK for data-residency regimes. Start from the **Regional CMEK Secret** preset.

**Rotated secret** -- rotation reminders through Pub/Sub driving a rotation pipeline, with the version-destroy safety window armed. Start from the **Rotated Secret** preset.

## Works With

- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the workload identity granted read access
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- customer-managed payload encryption
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- rotation-reminder delivery
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- mounts the secret via `valueFromSecret`
