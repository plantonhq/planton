# Cloudflare mTLS Certificate

Uploads a certificate to Cloudflare's account-level mTLS certificate store — the trust material that Authenticated Origin Pulls rows, zone TLS CA associations, and Workers mTLS bindings reference by certificate ID. A CA upload (`ca: true`) validates client certificates; a leaf upload (`ca: false`, with its private key) is a certificate Cloudflare presents itself as a client. Self-signed CAs are the normal case here: your infrastructure validates this material, not the public. Every field is create-only — any change replaces the upload and the certificate ID changes, so rotation is always replace-and-repoint, never in-place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **mTLS Certificate** -- one `cloudflare_mtls_certificate` in the account-level store, CA or leaf per the `ca` flag, with the private key attached only when one is provided

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has SSL and Certificates Edit on the account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **The PEM material to upload** -- a CA certificate (or chain) for `ca: true`, minted by your own PKI or openssl; the CA's signing key stays outside Cloudflare entirely. The store has no plan gate — free accounts included.
- **A private key, only for leaf uploads** -- needed only when Cloudflare must present the certificate itself (`ca: false`). Provide it as a managed-secret reference for the `privateKey` field; the API never returns it, so your secret store is the real system of record.

## Deploy

### Console

Open the deployment store, find **Cloudflare mTLS Certificate**, and click **Deploy**. The creation wizard walks you through the target account, the CA-or-leaf decision, the PEM material, and the optional private key. Start from the **Origin-pull CA upload** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareMtlsCertificate
metadata:
  name: origin-pull-ca
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-pull-ca
  ca: true
  certificates: |
    -----BEGIN CERTIFICATE-----
    MIIDdzCCAl+gAwIBAgIEbGyhVTANBgkqhkiG9w0BAQsFADBsMQswCQYDVQQGEwJV
    -----END CERTIFICATE-----
```

```shell
planton apply -f mtls-ca.yaml
```

This uploads a CA certificate to the account store with no private key — the shape Authenticated Origin Pulls and CA hostname associations consume. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an mTLS certificate upload. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**CA or leaf decides the consumer — and cannot change.** `ca: true` is trust material for validating clients: Authenticated Origin Pulls and API Shield reference it, and it needs no private key — uploading one anyway spreads a secret for nothing. `ca: false` is a leaf certificate Cloudflare presents as a client (Workers mTLS bindings), and that one needs its key. The spec makes you state the flag explicitly because the API cannot change it after upload.

**Everything is create-only: rotate by replace.** Any field change — even a formatting-only edit to the PEM — replaces the upload and mints a new `certificate_id`. The rotation discipline is a three-step dance in strict order: upload the new certificate, re-point every consumer (CA associations, Workers bindings, AOP rows) at the new ID, then destroy the old upload. Skipping the re-point step leaves consumers referencing a deleted certificate, and client validation starts failing at the edge, not at apply time.

**Keep the PEM byte-stable.** Because a whitespace change replaces the upload, treat `certificates` as an artifact, not editable text — store it verbatim and paste it verbatim.

**The key never comes back.** The API never returns `privateKey`; an imported or refreshed resource re-asserts it from configuration. An upload whose key you lost cannot be re-presented, only replaced — keep the key in a managed secret and reference it, so the platform resolves it just-in-time at deploy.

**Destroy is a real delete.** Re-point consumers at replacement material before destroying, or they lose their trust anchor with no warning at destroy time.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The PEM material and account ID travel as literal values, and the private key arrives as a managed-secret reference rather than a typed component reference.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | The uploaded certificate's ID — the identity every consumer holds | Zone TLS CA associations, Workers mTLS bindings, AOP rows |
| `expires_on` | Expiry timestamp (RFC3339) | Rotation planning and expiry alerting |
| `serial_number` | The certificate's serial number | Matching the upload against your PKI's records |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Origin-pull CA** -- A self-signed CA uploaded with `ca: true` and no private key — the trust anchor per-hostname Authenticated Origin Pulls and zone TLS CA associations validate client certificates against. The signing key never leaves your PKI. Start from the **Origin-pull CA upload** preset.

**Leaf certificate for Workers mTLS** -- A client certificate with `ca: false` and its `privateKey` supplied as a managed-secret reference, presented by Cloudflare when a Worker's mTLS binding calls an origin that demands client authentication.

**Zero-downtime rotation** -- Deploy the replacement as a second Infra Component, move consumers to its `certificate_id`, then destroy the original. Running both uploads in parallel during the switch is what makes the rotation invisible at the edge.

## Works With

- [**Cloudflare Zone TLS Settings**](/infra-catalog/cloudflare-zone-tls-settings) -- the CA hostname associations that scope an uploaded CA to specific hostnames.
- [**Cloudflare Authenticated Origin Pulls**](/infra-catalog/cloudflare-authenticated-origin-pulls) -- the zone-level enablement this CA validates client certificates for.
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- mTLS bindings that present leaf uploads from this store.
