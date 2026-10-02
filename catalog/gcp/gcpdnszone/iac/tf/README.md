# Terraform Module: GcpDnsZone

Provisions `google_dns_managed_zone` plus `google_project_service` for `dns.googleapis.com`.

## Resources

| Resource | Purpose |
|----------|---------|
| `google_project_service.dns_api` | Enables Cloud DNS API |
| `google_dns_managed_zone.managed_zone` | The managed zone |

## Inputs

See `variables.tf`. Key spec fields map 1:1 to the protobuf spec.

## Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | Managed zone ID |
| `zone_name` | Zone name for GcpDnsRecord FK |
| `nameservers` | Delegation NS set |

## Notes

- Provider pin: `~> 8.3`
- `dns_name` defaults to `${metadata.name}.` when spec.dns_name is empty
- No `google_dns_record_set` or `google_project_iam_binding` in this module
