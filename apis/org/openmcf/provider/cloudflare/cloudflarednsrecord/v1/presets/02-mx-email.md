# MX Record for Email

Creates an MX record for email delivery. Priority is required; MX records cannot be proxied. Use for configuring mail servers (Google Workspace, Microsoft 365, custom mail) for your domain.

## When to Use

- Connecting domain to Google Workspace, Microsoft 365, or other mail providers
- Custom mail server configuration
- Email routing for your domain

## Key Configuration Choices

- **type MX** (`type: MX`) -- Mail exchange record; priority is required.
- **priority** (`priority: 10`) -- Lower = higher priority; 10 is common for primary mail.
- **proxied: false** (`proxied: false`) -- MX records must be DNS-only; proxying not supported.
- **zoneId** (`zoneId`) -- Cloudflare zone ID; use value wrapper or reference to CloudflareDnsZone.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the DNS zone | CloudflareDnsZone status.outputs.zone_id or Dashboard |
| `<mail-server-hostname>` | Mail server hostname (e.g., mail.example.com) | Your mail provider's MX target (e.g., aspmx.l.google.com) |
| `10` | MX priority; lower = higher priority | Mail provider documentation (10–50 typical) |

## Related Presets

- **01-proxied-a-record** -- Use when pointing hostnames to IPs with Cloudflare proxy
