# DigitalOcean Spaces Access Key

Creates the S3-style credential workloads use against Spaces, DigitalOcean's object storage -- scoped to exactly the buckets each workload needs through per-bucket read/readwrite grants, or account-wide with a single fullaccess grant. Grants are the sole authorization: a key with no grants is valid and unlocks nothing. The secret key is returned exactly once, in the create response, and can never be fetched from DigitalOcean again -- this component's outputs are the only place it ever exists.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Spaces Access Key** -- the access-key/secret-key pair carrying the grant rows you declare; both engines mark the secret sensitive in state

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.
- **Buckets (optional)** -- DigitalOceanBucket resources for per-bucket grants; a fullaccess grant names none.

### DigitalOcean Account

- **Nothing extra** -- keys are free API objects in any quantity, and key management rides the same account API token (not the S3-compatible endpoint); the key this component creates is itself the Spaces credential.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Spaces Access Key**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **App Read-Write Key** preset in the [Presets](#presets) tab to mint one application's credential for its own bucket.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanSpacesKey
metadata:
  name: app-uploads-key
  org: acme-corp
  env: prod
spec:
  keyName: app-uploads
  grants:
    - bucket:
        value: app-assets
      permission: readwrite
```

```shell
planton apply -f do-spaces-key.yaml
```

This mints a key pair that can read and write exactly one bucket and touch nothing else in the account; the pair lands in `status.outputs`, with the secret held as a sensitive value. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to scope a grant to a bucket deployed in the same InfraPipeline:

```yaml
spec:
  keyName: app-uploads
  grants:
    - bucket:
        valueFrom:
          kind: DigitalOceanBucket
          name: app-assets
          fieldPath: status.outputs.bucket_id
      permission: readwrite
```

The InfraPipeline resolves the dependency graph, deploys the bucket first, then mints the key already scoped to it.

## Key Configuration

These are the most important decisions when configuring a Spaces access key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Grants are the only thing that authorizes** -- With `grants` empty, DigitalOcean creates the key with NO access to anything: a valid, safe placeholder state. Each grant is either `read`/`readwrite` scoped to one named bucket, or `fullaccess` with no bucket at all -- the pairing is enforced at validation time, because that is the provider's own grant grammar.

**Treat `fullaccess` as an admin credential** -- A fullaccess grant covers every bucket in the account, including buckets created AFTER the key. Occasionally right (a backup operator, an admin migration job), usually wrong for workloads. Per-bucket grants exist precisely so each workload's key unlocks only its own data.

**The permission wall here is the only one anywhere** -- DigitalOcean's provider performs no validation on `permission`: a typo like `write` becomes an EMPTY permission upstream, producing a grant that silently authorizes nothing -- the failure surfaces as mysterious 403s at the workload, far from the cause. The spec rejects anything outside `read`/`readwrite`/`fullaccess` before it can reach DigitalOcean.

**Rotation is destroy-and-recreate, and consumers must follow** -- The key material is immutable; rotating changes BOTH the access key and the secret. Plan it as a two-step move: create the replacement key, re-point every consumer, then destroy the old one. Deleting a key invalidates it immediately -- anything still signing with it starts failing on the spot.

**Grant updates replace the whole list** -- Every update sends the complete grant list in one call; there is no add-one-grant API. Invisible in normal declarative use (the spec IS the whole list), but worth knowing when diagnosing: a grant removed from the spec is genuinely revoked on the next apply, never left behind. Only `keyName` and the grants update in place -- the key material never changes.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanBucket** (optional, per grant) | `grants[].bucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `access_key` | The access key ID (also the resource's API identity) | Workload S3 client configuration against the bucket's regional Spaces endpoint |
| `secret_key` | The secret access key -- a SECRET, held sensitive in state | Paired with `access_key` as standard S3 credentials |

Design your handoff around the write-once secret: DigitalOcean returns `secret_key` only in the create response and never again, from any API. Consumers should read it from these outputs (or through a reference) rather than expecting to look it up later. If the secret is lost, there is no recovery -- destroy the key and create a new one.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One key per workload, scoped to its bucket** -- readwrite on exactly the bucket the application owns, nothing else: the workload uploads and serves its files and cannot touch any other data in the account. The default shape for application credentials. Start from the **App Read-Write Key** preset.

**Account-wide operator key** -- a single fullaccess grant for the rare tool that legitimately needs every bucket, such as a backup system sweeping all storage. Accept that it also covers buckets that do not exist yet, and guard it accordingly. Start from the **Backup Operator Key** preset.

## Works With

- [**DigitalOcean Spaces Bucket**](/infra-catalog/digital-ocean-bucket) -- the buckets grants scope to; workloads use this key against the bucket's regional Spaces endpoint
- [**DigitalOcean App Platform App**](/infra-catalog/digital-ocean-app) -- app services consume the key pair as environment variables for S3-style access to Spaces
