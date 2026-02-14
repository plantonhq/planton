# Wildcard Domain Certificate

This preset creates a Google-managed wildcard TLS certificate (`*.example.com`) with the apex domain (`example.com`) as a Subject Alternative Name. This covers all subdomains under a single certificate.

## When to Use

- Multi-tenant SaaS platforms with dynamic subdomains (e.g., `tenant1.example.com`, `tenant2.example.com`)
- Environments with many services under a shared domain
- Simplifying certificate management by covering all subdomains with one certificate

## Key Configuration Choices

- **Wildcard primary domain** (`*.example.com`) -- covers all first-level subdomains
- **Apex domain as SAN** (`example.com`) -- includes the bare domain alongside the wildcard
- **Managed certificate with DNS validation** -- automatic provisioning and renewal via Cloud DNS
- **Note**: Wildcard only covers one level of subdomain; `*.*.example.com` is not supported

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID (plain string) | `GcpProject` outputs |
| `<dns-zone-name>` | Cloud DNS managed zone name for `example.com` | `GcpDnsZone` status outputs |

## Related Presets

- **01-single-domain-dns** -- Use for certificates covering a single specific domain
