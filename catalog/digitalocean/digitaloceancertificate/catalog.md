# DigitalOcean Certificate

Deploys an SSL/TLS certificate on DigitalOcean, supporting both free auto-renewing Let's Encrypt certificates and user-provided custom certificates. The certificate name can be referenced by DigitalOcean Load Balancers for HTTPS termination. Every field is create-only -- any change replaces the certificate, with the replacement created before the old one is destroyed, so consumers referencing the stable name never observe a gap.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **DigitalOcean Certificate** -- a `digitalocean_certificate` resource; the manifest's source branch (`letsEncrypt` or `custom`) determines the certificate type
- **Let's Encrypt Certificate** -- created when the `letsEncrypt` branch is set; includes the specified domains with automatic ACME validation and renewal managed by DigitalOcean
- **Custom Certificate** -- created when the `custom` branch is set; uploads the provided PEM-encoded leaf certificate, private key, and optional intermediate chain with `create_before_destroy` lifecycle for zero-downtime rotation

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **For Let's Encrypt certificates** -- every listed domain must be managed by DigitalOcean DNS in the SAME account, or issuance fails -- no challenge records get planted. Wildcards (e.g., `*.example.com`) work under the same condition.
- **For custom certificates** -- have the PEM-encoded leaf certificate, private key, and optional intermediate chain ready. Store the private key as a managed secret and reference it as `$secret/<slug>` in the manifest -- never paste PEM key material.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Certificate**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Let's Encrypt Certificate** preset in the [Presets](#presets) tab for a free, auto-renewing certificate.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanCertificate
metadata:
  name: my-cert
  org: acme-corp
  env: prod
spec:
  certificateName: my-cert
  letsEncrypt:
    domains:
      - example.com
      - www.example.com
```

```shell
planton apply -f certificate.yaml
```

This creates a free Let's Encrypt certificate covering `example.com` and `www.example.com`, with renewal handled by DigitalOcean. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a DigitalOcean certificate. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Certificate source** -- Set exactly one branch: `letsEncrypt` for a free, auto-renewing certificate managed by DigitalOcean, or `custom` when providing your own PEM-encoded certificate from an enterprise CA or purchased provider. The branch determines the certificate type; there is no separate type field to keep consistent.

**Let's Encrypt domains** -- The `letsEncrypt.domains` field accepts multiple FQDNs and wildcard domains (e.g., `*.example.com`). All domains must have DNS managed by DigitalOcean in the same account for ACME validation. DigitalOcean always auto-renews Let's Encrypt certificates.

**Custom certificate materials** -- When using the `custom` branch, provide `leafCertificate` (server cert), `privateKey` (matching key), and optionally `certificateChain` (intermediate certs). `privateKey` is a sensitive field: bind it as a managed-secret reference (`privateKey: $secret/<slug>`), never pasted PEM. The DigitalOcean API never returns certificate material, so the manifest (or the secret store feeding it) is the only home the key has -- there is no "download my key" recovery path.

**Rotation and renewal** -- Custom certificates have no renewal machinery: DigitalOcean serves exactly what you uploaded until it expires. Rotation is editing the manifest with new PEM material and re-applying; the replacement is created before the old certificate is destroyed, so load balancers referencing the name never observe a gap. Watch the `expiry_rfc3339` output and alert well before it.

**Certificate naming** -- Use a stable, descriptive `certificateName` since load balancers reference certificates by name, and the name IS the certificate's resource identity (a Let's Encrypt certificate's UUID rotates on every auto-renewal; the name never does). One deletion-order caveat: DigitalOcean refuses to delete a certificate while a load balancer still references it, so when restructuring HTTPS termination, update the forwarding rule first -- the reverse order looks hung.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `certificate_id` | The certificate's resource identifier -- the NAME, not a UUID | Load balancer forwarding-rule `certificateName` references, API operations |
| `expiry_rfc3339` | Expiration timestamp in RFC 3339 format (moves forward on auto-renewal) | Certificate rotation monitoring, alerting dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Let's Encrypt certificate** -- Free, auto-renewing certificate for one or more domains managed by DigitalOcean DNS. Covers the common case of HTTPS for public websites and APIs. Start from the **Let's Encrypt Certificate** preset.

**Custom certificate** -- User-provided PEM certificate for enterprise CAs, purchased certificates, or EV certificates not available through Let's Encrypt. Requires manual rotation before expiry. Start from the **Custom Certificate** preset.

## Works With

- [**DigitalOcean Load Balancer**](/infra-catalog/digital-ocean-load-balancer) -- HTTPS forwarding rules reference this certificate by name via the `certificate_id` output