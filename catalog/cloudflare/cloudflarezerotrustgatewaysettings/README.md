# Cloudflare Zero Trust Gateway Settings

## Overview

`CloudflareZeroTrustGatewaySettings` configures the account's Secure Web Gateway: the settings panel behind every Gateway policy (TLS inspection, antivirus scanning, block-page branding, browser isolation, sandboxing), the activity-logging controls, and the account's proxy auto-config (PAC) files.

The spec folds three Cloudflare surfaces with different lifecycles: the configuration SINGLETON (upsert, NO-OP destroy, unset sub-objects unmanaged), the logging SINGLETON (upsert, no-op destroy, the COMPLETE tree always sent), and the PAC-file COLLECTION (real per-file lifecycle).

## Key Features

- **The whole Gateway panel** -- antivirus, block page, body scanning, browser isolation, TLS inspection, sandboxing, FIPS, DNS TTL cap
- **Partial adoption** -- manage only the sub-objects you declare; dashboard-set values elsewhere survive
- **Logging without drift** -- the full per-rule-type tree ships explicitly (partial sends drift forever at Cloudflare)
- **PAC files as rows** -- each file a real resource with its own lifecycle

## Use Cases

**Ideal for:**

- The account's Gateway posture as code next to its Gateway policies
- Branding the block page and standardizing logging redaction
- Serving PAC files for managed-device proxy configuration

**Not ideal for:**

- The filtering rules themselves -- those are `CloudflareZeroTrustGatewayPolicy`
- DNS locations -- `CloudflareZeroTrustDnsLocation`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |

### Optional Surfaces

| Field | Type | Description |
|-------|------|-------------|
| `settings` | object | The configuration singleton's tree (14 managed sub-objects; `custom_certificate` deliberately unmodeled -- deprecated at the provider). |
| `logging` | object | redact_pii + per-rule-type (dns/http/l4) log switches. |
| `pac_files` | list | PAC files: name + contents required; slug immutable. |

### Outputs

| Field | Description |
|-------|-------------|
| `account_id` | The account the configuration was applied to (the singleton's identity) |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewaySettings
metadata:
  name: acme-gateway
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  settings:
    activity_log:
      enabled: true
  logging:
    redact_pii: true
    settings_by_rule_type:
      dns: { log_blocks: true }
      http: { log_blocks: true }
      l4: {}
```

## Destroy Semantics

The settings and logging singletons: destroy is a NO-OP -- the live configuration stays exactly as last applied. PAC files: destroy is a real delete per file.

## The TLS-inspection gate

Enabling `tls_decrypt` (or FIPS TLS, or deep body scanning) before a Gateway certificate is ACTIVATED on the account fails with error 2211 -- and leaves the account's Gateway erroring until reverted. The certificate lifecycle is not yet a catalog kind; until it is, these switches belong in change windows with a certificate already activated.

## Related Resources

- **CloudflareZeroTrustGatewayPolicy** -- the filtering rules this panel configures behavior for
- **CloudflareZeroTrustDnsLocation** -- where DNS traffic enters Gateway

## Further Reading

For operational judgment -- the no-op destroy discipline, the block-page drift fields, the 2211 hazard -- see GUIDE.md.

## References

- [Cloudflare Gateway](https://developers.cloudflare.com/cloudflare-one/policies/gateway/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
