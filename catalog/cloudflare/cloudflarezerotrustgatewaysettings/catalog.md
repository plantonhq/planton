# Cloudflare Zero Trust Gateway Settings

Configures the account's Cloudflare Secure Web Gateway: the settings panel behind every Gateway policy (TLS inspection, antivirus scanning, block-page branding, browser isolation, sandboxing), the activity-logging controls, and the account's proxy auto-config (PAC) files. Three Cloudflare surfaces with different lifecycles fold into one component: the settings and logging singletons upsert in place and survive destroy, while PAC files are real per-row resources that create and delete like normal objects.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Gateway Configuration** — created only when `settings` is declared, a `cloudflare_zero_trust_gateway_settings` PUT against the account singleton. An unset sub-object is never sent, so dashboard-set values survive
- **Gateway Logging** — created only when `logging` is declared, a `cloudflare_zero_trust_gateway_logging` that always sends the complete logging tree (Cloudflare reports drift on partial sends)
- **PAC Files** — one `cloudflare_zero_trust_gateway_pacfile` per `pacFiles` row; removing a row deletes the file

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **An activated Gateway certificate** (only for `tlsDecrypt`, `fips.tls`, and deep `bodyScanning`) — Cloudflare rejects those writes with error 2211 until a Gateway certificate exists and is activated on the account. The certificate lifecycle is not a catalog kind yet; activate one out-of-band first.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust Gateway Settings**, and click **Deploy**. The creation wizard walks you through the owning account, the settings tree (each surface independently managed or left alone), the logging controls, and the PAC file rows. Start from the **Logging and branded block page** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewaySettings
metadata:
  name: acme-gateway
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  settings:
    activityLog:
      enabled: true
  logging:
    redactPii: true
    settingsByRuleType:
      dns:
        logBlocks: true
      http:
        logBlocks: true
      l4:
        logBlocks: true
```

```shell
planton apply -f gateway-settings.yaml
```

This turns on activity logging, redacts PII, and logs blocked requests per firewall type — while every undeclared settings surface (TLS inspection, antivirus, isolation) stays exactly as the dashboard has it. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring Gateway settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Three lifecycles under one spec** — `settings` and `logging` are account singletons: create and update are the same PUT, and destroy is a no-op that abandons the live configuration. `pacFiles` rows are real resources. Destroying this component removes the PAC files but leaves every Gateway setting exactly as last applied — plan teardown accordingly.

**Unset means unmanaged (settings only)** — a settings sub-object you don't declare is never sent, so partial adoption is safe. But removing a sub-object from the manifest does not revert it; apply the previous value explicitly. The `logging` surface is the deliberate opposite: when declared, the complete tree ships, and unset switches become false.

**The 2211 trap: certificate before decryption** — `tlsDecrypt`, `fips.tls`, and deep `bodyScanning` all require an activated Gateway certificate on the account, or the API rejects the write with error 2211. Activate the certificate first, and flip these switches in a change window — TLS inspection changes what every proxied HTTPS connection experiences.

**Declare the block-page drift fields explicitly** — the provider has a recorded defect where `blockPage.mode`, `includeContext`, `suppressFooter`, and `targetUri` drop from state on refresh when absent from configuration. If you manage the block page, declare those four fields even at their defaults so refresh has nothing to drop. `redirect_uri` mode requires `targetUri`; validation enforces the pairing.

**PAC slugs are URLs** — a PAC file's `slug` is baked into its public download URL and forces replacement on change; every device configured with the old URL breaks. Set it deliberately on day one (or accept the name-derived one) and never touch it.

**Antivirus fail posture** — `antivirus.failClosed` decides what happens to files that cannot be scanned (encrypted archives, oversize): block them or let them through. Blocking is the safer posture and the more surprising one for users — pair it with `notificationSettings` so blocked transfers explain themselves.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The TLS-inspection certificate (`settings.certificate.id`) takes a literal certificate UUID because no catalog kind manages Gateway certificates yet; the nil UUID selects the Cloudflare Root CA.

### What This Kind Provides

This component's `status.outputs` carries only `account_id` — the account the configuration was applied to, echoed back as the singleton's identity for the harness and import recipes. There is nothing here for downstream resources to consume: Gateway policies and DNS locations attach to the same account by ID, not by referencing this resource.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Logging and branded block page** — the everyday shape: blocks logged (not all traffic), PII redacted, and a branded block page users can report false positives from. Everything undeclared stays as the dashboard has it. Start from the **Logging and branded block page** preset.

**PAC files only** — manage proxy auto-config files as rows without touching the settings or logging singletons at all — the safe first footprint on an account whose Gateway settings are still dashboard-managed. Start from the **PAC files only** preset.

**TLS inspection rollout** — activate the Gateway certificate out-of-band, declare `certificate.id`, then enable `tlsDecrypt` in a change window. Only after decryption is on do HTTP policies see full request detail, and antivirus and body scanning become meaningful.

## Works With

- [**Cloudflare Zero Trust Gateway Policy**](/infra-catalog/cloudflare-zero-trust-gateway-policy) — the filtering rules this panel sets the behavior for; block-page branding and TLS inspection change what those policies can see and show.
- [**Cloudflare Zero Trust DNS Location**](/infra-catalog/cloudflare-zero-trust-dns-location) — the entry points DNS filtering runs against.
- [**Cloudflare Zero Trust Organization**](/infra-catalog/cloudflare-zero-trust-organization) — the login half of Zero Trust; this kind is the traffic half.
