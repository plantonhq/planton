# SRV Record for a Service

Creates an SRV record that advertises the host and port of a service (SIP, XMPP,
Minecraft, etc.). SRV records are structured: their priority, weight, port, and
target are supplied through the `data.srv` block rather than a flat content string.

## When to Use

- Publishing a service endpoint for protocols that use SRV discovery (SIP, XMPP, LDAP)
- Advertising a non-standard port for a service under a well-known name

## Key Configuration Choices

- **type SRV** (`type: SRV`) -- Service locator record; uses the `data.srv` block.
- **name** (`name: _sip._tcp`) -- The `_service._proto` label the record answers for.
- **data.srv.priority / weight** -- Lower priority is preferred; weight distributes load among equal priorities.
- **data.srv.port / target** -- The TCP/UDP port and hostname of the machine providing the service.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the DNS zone | CloudflareDnsZone status.outputs.zone_id or Dashboard |
| `<service-hostname>` | Hostname of the machine providing the service | Your service deployment (e.g., sip.example.com) |

## Related Presets

- **04-caa-certificate-authority** -- Another structured record type using a `data` block
