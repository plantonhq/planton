---
title: "Single Domain Managed Certificate"
description: "This preset creates a Google-managed TLS certificate for a single domain using DNS validation via Cloud DNS. The certificate is automatically provisioned and renewed by Google Certificate Manager."
type: "preset"
rank: "01"
presetSlug: "01-single-domain-dns"
componentSlug: "certificate-manager-cert"
componentTitle: "Certificate Manager Cert"
provider: "gcp"
icon: "package"
order: 1
---

# Single Domain Managed Certificate

This preset creates a Google-managed TLS certificate for a single domain using DNS validation via Cloud DNS. The certificate is automatically provisioned and renewed by Google Certificate Manager.

## When to Use

- HTTPS for a single domain on a Google Cloud load balancer
- Any service that needs an auto-renewing TLS certificate without manual intervention
- Domains whose DNS is managed in Cloud DNS (required for automatic validation)

## Key Configuration Choices

- **Managed certificate** (`certificateType: MANAGED`) -- Google provisions and renews the certificate automatically
- **DNS validation** (`validationMethod: DNS`) -- validation records are created in the specified Cloud DNS zone
- **Single domain** -- for multiple domains or SANs, add entries to `alternateDomainNames`
- **Note**: `gcpProjectId` is a plain string (not `StringValueOrRef`) in this component

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID (plain string, not `value:` wrapped) | `GcpProject` outputs |
| `<your-domain.com>` | Primary domain for the certificate (e.g., `api.example.com`) | Your domain registrar |
| `<dns-zone-name>` | Cloud DNS managed zone name for validation | `GcpDnsZone` status outputs |

## Related Presets

- **02-wildcard-domain** -- Use for wildcard certificates covering all subdomains
