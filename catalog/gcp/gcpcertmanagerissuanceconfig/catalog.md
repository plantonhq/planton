# GCP Cert Manager Issuance Config

Creates one Certificate Manager certificate issuance config — how Google-managed certificates that name it are issued from your own Certificate Authority Service pool instead of a public CA. A GcpCertManagerCert opts in through `managed.issuance_config`, and many certificates share one config; Google then requests each certificate from the pool with the key algorithm and lifetime set here and renews it automatically. Private PKI with public-CA convenience, for internal load balancers and service-to-service TLS that must chain to your own root.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Certificate Manager API enablement** (`certificatemanager.googleapis.com`) on the target project (never disabled on destroy)
- **Certificate Manager Certificate Issuance Config** -- a `google_certificate_manager_certificate_issuance_config` naming the CA pool, key algorithm, certificate lifetime, and rotation window

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.

### GCP Project

- **A Certificate Authority Service pool** with an enabled authority -- a [GcpPrivateCaPool](/infra-catalog/gcp-private-ca-pool) and a GcpPrivateCaCertificateAuthority inside it.
- **The Certificate Manager service agent** (`service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com`) holding `roles/privateca.certificateRequester` on the pool. The config creates without it, but no certificate issues until it is granted.

## Deploy

### Console

Open the deployment store, find **GCP Cert Manager Issuance Config**, and click **Deploy**. The wizard walks the config's envelope (project, name, location), then the pool and the certificate policy. Start from the **Thirty-Day ECDSA Issuance** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCertManagerIssuanceConfig
metadata:
  name: internal-tls-issuance
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-prod-12345
  caPool:
    value: projects/acme-prod-12345/locations/us-central1/caPools/internal-pool
  keyAlgorithm: ECDSA_P256
  lifetime: 2592000s
  rotationWindowPercentage: 66
  deletionPolicy: PREVENT
```

```shell
planton apply -f issuance-config.yaml
```

This creates a global issuance config whose `issuance_config_id` output a GcpCertManagerCert's `managed.issuance_config` takes. An Infra Job tracks the provisioning in real time.

### InfraChart

The pool arrives by reference, and the certificate references this config in turn:

```yaml
spec:
  caPool:
    valueFrom:
      kind: GcpPrivateCaPool
      name: internal-pool
      fieldPath: status.outputs.name
  keyAlgorithm: ECDSA_P256
  lifetime: 2592000s
  rotationWindowPercentage: 66
```

The InfraPipeline provisions the pool (and its authority) first, then this config, then the certificates that name it.

## Key Configuration

These are the most important decisions when configuring an issuance config. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**CA pool** (`caPool`) -- the pool's full name, `projects/{project}/locations/{location}/caPools/{pool}`. Reference a GcpPrivateCaPool's `name` output. A global config may name a regional pool. Immutable.

**Key algorithm** -- `RSA_2048` for the widest client compatibility, `ECDSA_P256` for smaller keys and faster handshakes on every modern client. Immutable.

**Lifetime** -- a duration in seconds from `1814400s` (21 days) to `2592000s` (30 days). Shorter lifetimes limit the damage of a leaked key; renewal is automatic either way. Immutable.

**Rotation window** (`rotationWindowPercentage`) -- how far into a certificate's life Google renews it. Google requires renewal at least 7 days after issuance and 7 days before expiry, so the usable range is 34–66 for 21 days and 24–76 for 30 days. Validation enforces the rule for any lifetime. Immutable.

**Location** -- empty means `global`, which serves global certificates; a regional certificate needs a config in its own region. Immutable.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpPrivateCaPool** | `caPool` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `issuance_config_id` | Full resource name (`projects/*/locations/*/certificateIssuanceConfigs/*`) | A GcpCertManagerCert's `managed.issuance_config` |
| `issuance_config_name` | The config's name in GCP | Auditing |
| `location` | The Certificate Manager location (`global` unless set) | Matching certificate scope |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Thirty-day ECDSA issuance** -- P-256 keys, the 30-day maximum lifetime, renewal at two thirds of life. The everyday shape for internal service TLS. Start from the **Thirty-Day ECDSA Issuance** preset.

**Short-lived RSA issuance** -- RSA-2048 for older clients, the 21-day minimum lifetime, renewal at the latest allowed point. Start from the **Short-Lived RSA Issuance** preset.

## Works With

- [**GCP Cert Manager Cert**](/infra-catalog/gcp-cert-manager-cert) -- names this config in `managed.issuance_config` and is issued from the pool
- [**GCP Private CA Pool**](/infra-catalog/gcp-private-ca-pool) -- its `name` output feeds `caPool`
- [**GCP Private CA Certificate Authority**](/infra-catalog/gcp-private-ca-certificate-authority) -- the enabled authority the pool needs before anything issues
