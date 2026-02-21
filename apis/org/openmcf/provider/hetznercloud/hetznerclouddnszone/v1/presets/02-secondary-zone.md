# Secondary Zone

This preset creates a DNS zone in secondary mode where Hetzner Cloud acts as a secondary (slave) nameserver, synchronizing records from your external primary nameserver via zone transfer (AXFR/IXFR). No records are managed in this manifest -- they are pulled automatically from the primary. TSIG authentication is included to secure the zone transfer channel.

This is the correct choice when your authoritative DNS is hosted elsewhere and you want Hetzner Cloud as a geographically distributed secondary for redundancy or latency improvement.

## When to Use

- You already manage DNS on an external primary nameserver (e.g., BIND, PowerDNS, another cloud provider) and want Hetzner Cloud as a secondary
- Geographic redundancy for DNS resolution -- Hetzner's anycast nameservers supplement your primary
- Migration scenarios where you want Hetzner Cloud to mirror your existing zone before cutting over to primary mode

## Key Configuration Choices

- **Secondary mode** (`mode: secondary`) -- Hetzner Cloud pulls records from the primary nameserver; the `recordSets` field is forbidden in this mode (enforced by spec validation)
- **TSIG authentication** (`tsigAlgorithm` + `tsigKey`) -- secures the zone transfer channel between primary and secondary; without TSIG, any host could request a full copy of your zone data via AXFR
- **No TTL** -- TTL values are controlled by the primary nameserver and replicated during zone transfer
- **Delete protection enabled** (`deleteProtection: true`) -- prevents accidental deletion of the secondary zone, which would break DNS redundancy
- **Single primary nameserver** -- add additional entries to the `primaryNameservers` list if your primary DNS runs on multiple addresses for failover

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-domain>` | The DNS domain name to mirror (e.g., `example.com`) | Your domain registrar |
| `<primary-ns-ip-address>` | Public IPv4 or IPv6 address of your primary nameserver | Your primary DNS server configuration or hosting provider |
| `<tsig-algorithm>` | TSIG algorithm for zone transfer authentication (e.g., `hmac-sha256`, `hmac-sha512`) | Your primary nameserver's TSIG key configuration |
| `<tsig-key>` | Base64-encoded TSIG shared secret | Generated with `tsig-keygen` or `dnssec-keygen` on the primary nameserver; must match the key configured there |

## Related Presets

- **01-web-domain** -- use instead when Hetzner Cloud should be the primary (authoritative) nameserver with records managed directly in the manifest
- **03-simple-zone** -- use instead for a minimal primary zone without email or security records
