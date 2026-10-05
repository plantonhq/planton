# Cloudflare Turnstile Widget

Provisions a Cloudflare Turnstile widget: a privacy-preserving CAPTCHA alternative that protects forms and endpoints from bots without the friction of traditional CAPTCHAs. A widget yields a public **site key** you embed in your page and a sensitive **secret key** your backend uses to verify tokens via the `/siteverify` endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Turnstile Widget** -- a configured widget scoped to your account and domains
- **Site key and secret key** -- exported as outputs (the secret is sensitive)

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Turnstile edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **An account** -- the widget is created under an existing Cloudflare account (selected from the connection's accounts).

## Deploy

### Console

Open the deployment store, find **Cloudflare Turnstile Widget**, and click **Deploy**. The creation wizard captures the account, name, served domains, challenge mode, and region. Start from the **Managed Widget** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareTurnstileWidget
metadata:
  name: signup-widget
  org: acme-corp
  env: prod
spec:
  accountId: a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4
  name: signup
  domains:
    - example.com
    - www.example.com
  mode: managed
```

```shell
planton apply -f cloudflare-turnstile-widget.yaml
```

This creates a managed-mode widget for the listed domains. An Infra Job tracks the provisioning in real time; the site and secret keys appear in `status.outputs` once issued.

## Key Configuration

These are the most important decisions when configuring a Turnstile widget. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domains (`domains`)** -- The hostnames allowed to serve the widget. Tokens are only issued on listed domains, so a forgotten domain silently breaks the form served there; include `localhost` for local development.

**Mode (`mode`)** -- `managed` (Cloudflare picks the challenge -- recommended), `non-interactive` (visible but never interactive), or `invisible` (no UI). The lower-friction modes trade challenge strength for UX, so pair them with server-side verification you actually enforce.

**Clearance Level (`clearanceLevel`)** -- Optional; the clearance granted on a Cloudflare-proxied site (`no_clearance`, `jschallenge`, `managed`, `interactive`). Setting a clearance lets a passed widget also satisfy the zone's challenge rules, so visitors are not challenged twice.

**Region (`region`)** -- `world` (default) or `china`. **Immutable** -- changing it replaces the widget, which issues new site and secret keys that every embedding page and verifying backend must pick up.

**Enterprise flags (`botFightMode`, `ephemeralId`, `offlabel`)** -- All three require a Cloudflare Enterprise plan; enabling them on a non-Enterprise account fails the deployment.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies -- the widget is account-scoped, and its served domains travel as plain strings.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `sitekey` | The public site key | Embed in the frontend Turnstile widget |
| `secret` | The sensitive secret key | Server-side token verification via `/siteverify`, e.g. a Worker's secret binding |

`status.outputs` also carries `created_on` and `modified_on` timestamps.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed challenge** -- the recommended default; Cloudflare adapts the challenge to risk. Start from the **Managed Widget** preset.

**Invisible widget** -- zero visible UI for low-friction flows, paired with server-side verification. Start from the **Invisible Widget** preset.

## Works With

- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- verifies Turnstile tokens server-side by referencing the widget's `secret` output as a secret binding
