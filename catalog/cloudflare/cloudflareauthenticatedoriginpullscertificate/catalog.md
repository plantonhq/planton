# Cloudflare Authenticated Origin Pulls Certificate

Uploads the client certificate Cloudflare presents to your origin under Authenticated Origin Pulls. Self-signed certificates are the designed case — the origin validates this certificate, not the public. The `scope` decides the blast radius: a `zone` upload replaces Cloudflare's shared client certificate for the entire zone, while a `hostname` upload only stages material that per-hostname associations pin explicitly. Rotation is replacement, and key and certificate must always change together — a key-only change against the zone surface silently does nothing.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one of:

- **Zone-scoped Upload** — one `cloudflare_authenticated_origin_pulls_certificate` when `scope` is `zone` (the default). Every origin pull in the zone presents it from then on.
- **Hostname-scoped Upload** — one `cloudflare_authenticated_origin_pulls_hostname_certificate` when `scope` is `hostname`. Nothing changes at the edge until a Cloudflare Authenticated Origin Pulls association pins a hostname to the resulting certificate ID.

Destroy is a real delete on both surfaces, settling asynchronously — the API answers with `pending_deletion` before the record actually goes.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Zone → SSL and Certificates → Edit**. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone** for `zoneId` — free plans included.
- **A client certificate and private key in PEM form** — mint a self-signed pair with openssl; the origin's trust store validates it. Store the private key as a managed secret and reference it from `privateKey` — the API never returns the key, so configuration is its only source.

## Deploy

### Console

Open the deployment store, find **Cloudflare Authenticated Origin Pulls Certificate**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, the scope choice, and the certificate and private-key material. Start from the **Hostname-scoped client certificate** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPullsCertificate
metadata:
  name: app-client-certificate
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  scope: hostname
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  privateKey:
    value: ${secrets-group/prod-aop/app-client-key}
```

```shell
planton apply -f aop-certificate.yaml
```

This uploads a hostname-scoped client certificate and returns its `certificate_id` — no hostname presents it until an association pins one. An Infra Job tracks the provisioning in real time.

### InfraChart

When composing the zone and the upload in one InfraPipeline, wire the zone reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  scope: zone
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  privateKey:
    value: ${secrets-group/prod-aop/zone-client-key}
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then uploads the certificate to the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring an Authenticated Origin Pulls certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope is blast radius** — `scope: zone` replaces Cloudflare's shared client certificate for the entire zone the moment it deploys. `scope: hostname` only uploads material; nothing changes until an association pins a hostname to the `certificate_id`. When in doubt, upload hostname-scoped and pin explicitly.

**Never rotate only the private key** — at provider v5.23.0 the zone-scoped surface has an empty Update and its `privateKey` does not force replacement: a plan that changes only the key applies "successfully" while Cloudflare keeps the old key, and the presented certificate and signing key silently diverge. The discipline is a real re-issue — key and certificate always change together, which forces the replacement the provider honors. The hostname-scoped surface handles key changes correctly, but follow the same discipline everywhere.

**Keep the PEM byte-stable** — the provider replaces the upload on any change to the `certificate` value, including formatting-only churn (a trailing newline counts). Store the PEM exactly as issued.

**Deletion is asynchronous** — the API answers 200 with `pending_deletion` before the record goes, so automation that deletes and immediately re-uploads the same certificate can race the pending state. Associations referencing a hostname certificate stop authenticating when it dies — revert or re-point them before destroying the upload.

**The key is never readable back** — Cloudflare never returns the private key, so an imported or refreshed resource re-asserts it from configuration. Keep the managed-secret reference authoritative.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | The uploaded certificate's ID | `hostnameAssociations[].certificateId` on Cloudflare Authenticated Origin Pulls |
| `zone_id` | The zone the certificate belongs to | Confirming the upload landed in the intended zone |
| `expires_on` | Expiry timestamp (RFC3339) | Rotation scheduling and expiry alerting |

Deployment status is deliberately not an output: it transitions asynchronously (`pending_deployment` to `active` seconds after create), so a point-in-time phase would flip on the first refresh and re-plan forever. Read deployment status from the Cloudflare API or dashboard.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Stage, then pin** — upload hostname-scoped material first, then pin hostnames to it through associations. The upload alone changes nothing, so new certificates can be staged safely before any traffic depends on them. Start from the **Hostname-scoped client certificate** preset.

**Zone-wide replacement** — a `scope: zone` upload for zones where one client certificate serves every hostname. Deploying it is the cutover — every origin pull presents the new certificate immediately.

**Rotation as replacement** — mint a new certificate/key pair, upload it as a new resource, re-point the associations to the new `certificate_id`, then destroy the old upload after the asynchronous deletion window.

## Works With

- [**Cloudflare Authenticated Origin Pulls**](/infra-catalog/cloudflare-authenticated-origin-pulls) — the enablement whose association rows pin hostnames to this upload's `certificate_id`
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone the certificate is uploaded to; wire `zoneId` via ValueFromRef
- [**Cloudflare mTLS Certificate**](/infra-catalog/cloudflare-mtls-certificate) — the account-level CA side of per-hostname client-certificate validation
