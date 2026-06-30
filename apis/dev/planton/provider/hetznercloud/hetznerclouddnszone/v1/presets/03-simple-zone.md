# Simple Zone

This preset creates a minimal primary DNS zone with just an apex A record and a www CNAME -- the bare minimum to make a domain resolve to a server. No email records, no security records, no delete protection. It inherits the provider's default TTL of 3600 seconds.

Use this for infrastructure domains, staging environments, or any domain where you only need basic name resolution without email routing or certificate policy.

## When to Use

- Staging or development domains that just need to point at a server
- Infrastructure subdomains (e.g., `infra.example.com`) used internally
- Quick prototyping where you want DNS resolution without the overhead of a full production record set
- Domains that do not send email and do not need SPF, DMARC, or MX records

## Key Configuration Choices

- **Primary mode** (`mode: primary`) -- Hetzner Cloud is the authoritative nameserver with records managed in this manifest
- **No explicit TTL** -- inherits the provider default of 3600 seconds; override by adding `ttl:` at the spec level or per record set
- **No delete protection** (`deleteProtection` omitted, defaults to `false`) -- appropriate for ephemeral or non-critical zones; set to `true` if the domain matters
- **Apex A record only** (`@` / `A`) -- single server; add a second record entry for round-robin across multiple servers
- **WWW CNAME** (`www` / `CNAME`) -- aliases `www.<your-domain>` to the apex; remove if you do not use the `www` subdomain
- **No email records** -- MX, SPF, and DMARC are intentionally omitted; add them from the `01-web-domain` preset if you later need email

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-domain>` | The DNS domain name (e.g., `staging.example.com`) | Your domain registrar or parent zone configuration |
| `<server-ipv4-address>` | Public IPv4 address of the target server | The `status.outputs.ipv4_address` of your HetznerCloudServer resource, or the Hetzner Cloud Console |

## Related Presets

- **01-web-domain** -- use instead for a production website with email authentication (MX, SPF, DMARC) and certificate policy (CAA) records
- **02-secondary-zone** -- use instead when Hetzner Cloud should act as a secondary nameserver syncing from an external primary
