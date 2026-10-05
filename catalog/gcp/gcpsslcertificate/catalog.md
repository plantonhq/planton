# GCP SSL Certificate

Deploys a self-managed Compute Engine SSL certificate — you bring the PEM chain and private key (your own CA, a commercial purchase, or ACME automation outside GCP) and the load balancer presents it to clients. Choose self-managed when Google-managed certificates cannot do the job: wildcard domains, EV/OV issuance, internal load balancers, or serving TLS before public DNS cutover. Every field is immutable in GCP — rotation is create-before-destroy under a versioned name. The private key travels as a managed-secret reference (`$secret/<slug>`) resolved just-in-time at deploy; the manifest never carries the key.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine API enablement** -- the module enables `compute.googleapis.com` on the target project before creating the certificate, so a fresh project works without manual API setup
- **Compute Engine SSL Certificate** -- global (blank `region`) for global external Application Load Balancer proxies, or regional for regional external and internal ALB proxies
- **Validated key pair** -- GCP verifies the certificate chain matches the private key at create time and stores the key write-only (it never appears in outputs or the console)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **An org secret holding the private key** -- the `privateKey` field is sensitive: it carries a `$secret/<slug>` reference, never plaintext. Store the unencrypted PEM key as a managed secret first.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the certificate will be created. Provide the project ID directly, reference a GcpProject Infra Component via ValueFromRef, or omit `projectId` to use the provider connection's default project. The module enables the Compute Engine API itself — no manual API activation is needed.
- **The issued certificate chain** -- leaf first, then intermediates (at least one intermediate, at most 5 certificates total), and its unencrypted RSA-2048+/ECDSA P-256 private key.

## Deploy

### Console

Open the deployment store, find **GCP SSL Certificate**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the private key goes through the managed-secret picker. Start from the **Imported CA-Issued Certificate** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSslCertificate
metadata:
  name: prod-app-cert-2026
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  certificateName: "prod-app-cert-2026"
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...leaf, then intermediates...
    -----END CERTIFICATE-----
  privateKey: $secret/prod-app-cert-key
```

```shell
planton apply -f ssl-certificate.yaml
```

This creates a global certificate ready for a target HTTPS proxy's certificate list. The runner resolves the `$secret/` reference just-in-time — the manifest never carries the key. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the certificate to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the certificate — and downstream target proxies reference its `self_link` output.

## Key Configuration

These are the most important decisions when configuring a self-managed SSL certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Serving scope** -- Leave `region` empty for a GLOBAL certificate (global external load balancers); set it for a REGIONAL one (regional external and internal ALB proxies). A certificate cannot move between scopes or regions.

**Versioned name** -- `certificateName` defaults to the resource name; bake a version into it (`prod-app-cert-2026`). Every field is immutable, so rotation is create-before-destroy: create the replacement under the next version, repoint the proxy's certificate list, then destroy the old one — GCP refuses to delete a certificate a proxy still references. Self-managed and Google-managed certificates share one name namespace per scope.

**Certificate chain** -- Leaf certificate FIRST, then intermediates; at most 5 certificates total. The chain is public handshake material — only the key is secret.

**Private key** -- Unencrypted (no passphrase) RSA-2048+ or ECDSA P-256, referenced as `$secret/<slug>`. Write-only in GCP: keep the secret-store copy as the source of truth.

**Deletion policy** -- `DELETE` (the default) already fails rather than dropping TLS while any proxy references the certificate. Set `PREVENT` as a guard rail for a certificate whose replacement is not yet serving, or `ABANDON` to remove it from management while it keeps serving in GCP — the mid-rotation handoff.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the certificate | GcpTargetHttpsProxy `sslCertificates` list via ValueFromRef |
| `certificate_name` | Name as it exists in GCP | Audit, rotation planning |
| `certificate_id` | Server-assigned numeric ID | GCP console links, API references |
| `expire_time` | RFC3339 expiry parsed from the chain | The rotation clock — self-managed certificates never renew themselves |
| `region` | Region of a regional certificate (empty for global) | Scope verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Imported commercial certificate** -- A purchased or corporate-CA certificate serving a global external load balancer. Start from the **Imported CA-Issued Certificate** preset.

**Regional internal TLS** -- A certificate for a regional internal ALB proxy — the case Google-managed certificates cannot serve. Start from the **Regional ALB Certificate** preset.

**Versioned rotation** -- An explicit `certificateName` decoupled from the resource name demonstrates the create-before-destroy rotation workflow. Start from the **Rotation with a Versioned Name** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the certificate is created
- [**GCP Target HTTPS Proxy**](/infra-catalog/gcp-target-https-proxy) -- consumes the certificate's `self_link` in its certificate list
- [**GCP Managed SSL Certificate**](/infra-catalog/gcp-managed-ssl-certificate) -- the hands-off alternative when Google-issued renewal fits
- [**GCP SSL Policy**](/infra-catalog/gcp-ssl-policy) -- controls which TLS versions and ciphers the proxy negotiates alongside this certificate
