# External DNS (Manual Validation)

This preset creates a Google-managed TLS certificate without auto-creating DNS validation records in Cloud DNS. The `cloudDnsZoneId` field is intentionally omitted so the module exports the required CNAME validation records as stack outputs (`dns-validation-records`). You then manually insert those records into your DNS provider (e.g. AWS Route 53, Cloudflare, etc.).

## When to Use

- Your DNS zone is hosted outside GCP (e.g. AWS Route 53, Cloudflare, Azure DNS)
- You need a GCP Certificate Manager certificate but cannot or prefer not to use Cloud DNS for validation
- You want full control over which DNS records are created and where

## Key Configuration Choices

- **No `cloudDnsZoneId`** -- omitting this field skips automatic DNS record creation in Cloud DNS
- **Managed certificate** (`certificateType: MANAGED`) -- Google provisions and renews the certificate automatically
- **DNS validation** (`validationMethod: DNS`) -- validation records are exported as stack outputs for manual insertion

## After Deployment

After running `pulumi up` or `terraform apply`, read the `dns-validation-records` output:

```shell
pulumi stack output dns-validation-records
```

Each record contains `record_name`, `record_type`, `record_data`, and `domain`. Create these CNAME records in your external DNS provider. Certificate provisioning completes once the records propagate.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID (plain string) | `GcpProject` outputs |
| `<your-domain.com>` | Primary domain for the certificate | Your domain registrar |

## Related Presets

- **01-single-domain-dns** -- Use when your DNS zone is in Cloud DNS (auto-creates validation records)
- **02-wildcard-domain** -- Use for wildcard certificates with Cloud DNS
