# GCP Private CA Pool

Run your own private certificate authority on Google Cloud: one trust anchor your services trust, with the policy every certificate it issues must follow -- key types, lifetimes, allowed identities, and the X.509 fields stamped on each certificate.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `privateca.googleapis.com` on the project (never disabled on destroy)
- **CA pool** -- a `privateca_ca_pool` with its tier, issuance policy, publishing options, and optional at-rest encryption key

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with CA Service admin permissions (`roles/privateca.caManager`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- None. A **`GcpKmsKey`** is optional, to encrypt stored certificates at rest.

## Deploy

### Console

Open the deployment store, find **GCP Private CA Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **DevOps Workload TLS** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPrivateCaPool
metadata:
  name: internal-tls
  org: acme-corp
  env: prod
spec:
  location: us-central1
  tier: DEVOPS
  issuancePolicy:
    maximumLifetime: 2592000s
    allowedKeyTypes:
      - ellipticCurveSignatureAlgorithm: ECDSA_P256
```

```shell
planton apply -f private-ca-pool.yaml
```

This creates a DevOps-tier pool for 30-day certificates on P-256 keys. Add a `GcpPrivateCaCertificateAuthority` to it before it can issue. A Stack Job tracks the provisioning in real time.

### InfraChart

Authorities and certificates reference the pool from `pool`; Kafka and Redis clusters reference it for mTLS and server certificates.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tier** -- `ENTERPRISE` stores, lists, and revokes certificates and allows your own KMS keys; `DEVOPS` issues faster and tracks nothing. Immutable.

**Issuance policy** -- the guardrails: key types, maximum lifetime, which request forms are accepted, a CEL expression over the identities, and baseline X.509 values (key usage, CA constraints, name constraints) stamped on every certificate.

**Publishing** -- whether each authority publishes its CA certificate and CRL and advertises their URLs in issued certificates.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The pool's full resource name | `GcpPrivateCaCertificateAuthority.pool`, `GcpPrivateCaCertificate.pool`, `GcpManagedKafkaCluster.tlsConfig.caPools`, `GcpRedisCluster.serverCaPool` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Workload TLS** -- a DevOps pool for short-lived service certificates. Start from the **DevOps Workload TLS** preset.

**Governed issuance** -- an Enterprise pool with a strict policy and published CRLs. Start from the **Enterprise Issuance Policy** preset.

**Root of a hierarchy** -- a pool whose only job is to sign subordinate authorities. Start from the **Root Pool** preset.

## Works With

- [**GCP Private CA Certificate Authority**](/cloud-catalog/gcp-private-ca-certificate-authority) -- the authorities inside the pool
- [**GCP Private CA Certificate**](/cloud-catalog/gcp-private-ca-certificate) -- certificates issued from it
- [**GCP Private CA Certificate Template**](/cloud-catalog/gcp-private-ca-certificate-template) -- reusable certificate shapes
- [**GCP Managed Kafka Cluster**](/cloud-catalog/gcp-managed-kafka-cluster) -- mTLS clients trusted by pool
