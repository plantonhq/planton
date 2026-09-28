---
display_name: Subdomain Delegation
---

# NS Record Delegating a Subdomain

Delegates a subdomain of a Cloudflare zone to a zone hosted elsewhere (here an AWS Route 53 zone) by publishing one of that zone's name servers. The content is read from the child zone's outputs, so the record is created after the zone exists and always names the zone's current name servers. There is no copying of values between two installs, and a recreated zone cannot leave the delegation pointing at dead name servers.

## When to Use

- Handing `aws.example.com` to Route 53 (or `gcp.example.com` to Cloud DNS) while `example.com` stays on Cloudflare
- Giving a team or an environment its own zone under the company domain
- Any delegation that must survive the child zone being destroyed and recreated

## Key Configuration Choices

- **type NS** (`type: NS`) -- A delegation is a set of NS records at the subdomain's name, one per name server of the child zone.
- **content from the child zone** (`content.valueFrom`) -- `status.outputs.nameservers.0` names the first name server; declare one record per name server (`.0`, `.1`, `.2`, `.3` for Route 53's four), each with its own `metadata.name`.
- **proxied: false** (`proxied: false`) -- NS records are DNS-only.
- **ttl** (`ttl: 3600`) -- Delegations change rarely; an hour is typical.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | The parent zone's ID | CloudflareDnsZone status.outputs.zone_id, or the Cloudflare dashboard |
| `<subdomain-label>` | The delegated label (e.g. `aws` for `aws.example.com`) | Your naming choice; it must match the child zone's domain |
| `<child-zone-resource-name>` | The child zone's resource name | The `metadata.name` of the AwsRoute53Zone (or GcpDnsZone, CloudflareDnsZone) |

## Related Presets

- **02-mx-email** -- Another record whose target is a hostname
