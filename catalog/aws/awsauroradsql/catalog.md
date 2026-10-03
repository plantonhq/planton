# AWS Aurora DSQL

Deploys an Aurora DSQL cluster — AWS's serverless, PostgreSQL-compatible distributed SQL database with no instances to size, no capacity dials, and synchronous active-active writes across regions when clusters are paired. AWS names the cluster with a generated identifier and derives the connection endpoint from it; every connection authenticates with IAM auth tokens, because DSQL has no native database passwords. Multi-region topology is a create-time decision: peering happens in a one-shot window while a fresh cluster is still in PENDING_SETUP, and a live single-region cluster cannot be upgraded into a pair.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DSQL Cluster** — the serverless cluster with deletion protection, force-destroy posture, optional customer-managed KMS encryption, and (for multi-region halves) the witness region baked in at create
- **Cluster Peering** — created only when `multiRegion` is set; joins this cluster to its named peers to form one logical active-active database. The peering has no update path and a no-op delete at the provider — changing peers means recreating the cluster

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Aurora DSQL permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Nothing for a single-region cluster** — the spec's defaults are a complete deployment.
- **Peer clusters** (only for multi-region) — the peers must already exist in their own regions, each deployed as its own instance of this kind naming the others' `cluster_arn` and the same witness region.
- **A supported region** — Aurora DSQL runs in a subset of regions (us-east-1, us-east-2, us-west-2, eu-west-1/2/3, ap-northeast-1/2, ap-southeast-1/2 as of early 2026).

## Deploy

### Console

Open the deployment store, find **AWS Aurora DSQL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and deletion posture, then the optional encryption key and multi-region pairing. Start from the **Single-Region Production Cluster** preset in the [Presets](#presets) tab — the whole production posture in two fields.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAuroraDsql
metadata:
  name: orders-dsql
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  deletionProtectionEnabled: true
```

```shell
planton apply -f aws-aurora-dsql.yaml
```

This creates a delete-protected single-region cluster with AWS-owned encryption; connect any PostgreSQL driver to the `endpoint` output using an IAM auth token as the password. An Infra Job tracks the provisioning in real time.

### InfraChart

When the cluster deploys alongside its encryption key in one chart, wire the key via ValueFromRef:

```yaml
spec:
  region: us-east-1
  deletionProtectionEnabled: true
  kmsEncryptionKey:
    valueFrom:
      kind: AwsKmsKey
      name: orders-data-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, creates the KMS key first, then encrypts the cluster with it from the first byte.

## Key Configuration

These are the most important decisions when configuring a cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**IAM tokens are the only door** — DSQL has no CREATE USER WITH PASSWORD. Every connection authenticates with a short-lived IAM auth token as the password (`aws dsql generate-db-connect-admin-auth-token`, or the non-admin variant for roles mapped inside the database). Applications need token-refresh plumbing, not credential storage — there is no secret to rotate, and no password field anywhere in this spec.

**Design multi-region at day zero** — the pairing window is PENDING_SETUP, a freshly created cluster before its first activation completes. The modules order the peering correctly, but the peers must already exist, so an active-active database is one instance of this kind per region, each naming the others in `multiRegion.peerClusterArns`. A live single-region cluster cannot be retrofitted — plan the topology before data lands.

**The witness region is a third region, deliberately** — it stores transaction logs and arbitrates during a region failure; it runs no queries and holds no full data. Pick a region distinct from both peers. Changing `witnessRegion` later replaces the cluster.

**Not all of PostgreSQL is here** — DSQL speaks the PostgreSQL wire protocol but is a distributed engine underneath: no extensions, and optimistic concurrency instead of long locks — conflicting transactions retry rather than block. Port schema and load tests before committing a migration; "PostgreSQL-compatible" is a dialect claim, not a drop-in guarantee.

**Deletes are gated twice by design** — `deletionProtectionEnabled` refuses deletes at AWS; `forceDestroy` makes the module disable protection first and then delete. Keep `forceDestroy` false in production — it converts deletion protection from a wall into a speed bump. Both toggles update in place.

**Encryption can change, cheaply** — leave `kmsEncryptionKey` empty for the AWS-owned key, or reference an AwsKmsKey for your own governance. Switching keys later re-encrypts in place with no replacement, so starting on the AWS-owned key is not a one-way door.

**Cost follows usage, not capacity** — there is nothing to size: the drivers are request volume and bytes stored. That makes DSQL forgiving for spiky and idle workloads, and makes a runaway query loop a billing event rather than a saturated instance — put application-side guardrails where you would once have put a connection pool.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsEncryptionKey` | `status.outputs.key_arn` |
| **AwsAuroraDsql** | `multiRegion.peerClusterArns[]` | `status.outputs.cluster_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `endpoint` | The PostgreSQL connection host, `{identifier}.dsql.{region}.on.aws` | Application database-host configuration — the chart-ready join key |
| `cluster_arn` | The cluster's ARN | A peer cluster's `multiRegion.peerClusterArns`; IAM policies granting `dsql:DbConnect` |
| `vpc_endpoint_service_name` | The PrivateLink service name | An interface VPC endpoint (AwsVpcEndpoint) for private connectivity without public egress |

`identifier` (the AWS-generated cluster ID and import address) and `encryption_type` (AWS's report of `AWS_OWNED_KMS_KEY` or `CUSTOMER_MANAGED_KMS_KEY`) are also present — audit echoes, not composition inputs.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-region production cluster** — deletion protection on, AWS-owned encryption, nothing to size. The right default for new services: the endpoint and IAM-token auth work identically if the topology later grows, and the only thing that cannot grow in place is multi-region. Start from the **Single-Region Production Cluster** preset.

**Multi-region active-active pair** — one instance of this kind per region, each naming the other's `cluster_arn` and the same third witness region; both halves write and read synchronously. Pair fresh clusters only — the window is create-time. Start from the **Multi-Region Active-Active Half** preset and deploy its mirror image in the peer region.

**Private-only connectivity** — create an interface VPC endpoint against `vpc_endpoint_service_name` so application traffic rides PrivateLink instead of public egress. Pairs naturally with workloads in private subnets that have no NAT path.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption via `kmsEncryptionKey`; swappable in place
- [**AWS VPC Endpoint**](/infra-catalog/aws-vpc-endpoint) — the PrivateLink door to the cluster, created against `vpc_endpoint_service_name`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the application identity that generates DSQL auth tokens; grant it `dsql:DbConnect` on the `cluster_arn`
