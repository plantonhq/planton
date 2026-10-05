# Public Zone with DNSSEC

Creates a production public Cloud DNS zone with DNSSEC signing and query logging enabled. DNSSEC applies to public zones only in Cloud DNS.

## When to Use

- Production domains that need cryptographic DNS authentication
- Compliance or security policies requiring DNSSEC and query audit logs

## Key Configuration Choices

- **visibility: public** — DNSSEC requires a public managed zone
- **dnssecConfig.state: on** — Cloud DNS generates and rotates signing keys
- **cloudLoggingConfig.enableLogging: true** — logs every query for audit and troubleshooting

## Post-Deploy Step

After deploy, add the DS record from Cloud DNS to your domain registrar to complete the DNSSEC chain of trust.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `my-gcp-project-123` | GCP project ID |
| `example.com` | Your public domain |

## Related Presets

- **01-public-zone** — minimal public zone without DNSSEC
- **02-private-vpc** — internal VPC-only zone

## Related Kinds

- [GcpDnsRecord](/docs/catalog/gcp/gcpdnsrecord) — records within the signed zone
- [GcpCertManagerCert](/docs/catalog/gcp/gcpcertmanagercert) — TLS certs that may use DNS-01 against this zone
