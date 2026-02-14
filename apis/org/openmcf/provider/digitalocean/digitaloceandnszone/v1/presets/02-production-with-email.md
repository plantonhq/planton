# Production Zone with Email Records

This preset creates a DNS zone with A, MX, and TXT (SPF) records for a production website that also receives email. The A record points the domain to your web server; MX directs mail to your mail provider; SPF authorizes that provider to send on behalf of the domain.

## When to Use

- Production domains that receive and send email
- Using Google Workspace, Microsoft 365, or other third-party mail
- Need SPF record to improve deliverability and prevent spoofing

## Key Configuration Choices

- **A record** (`name: "@"`, `type: A`) -- root domain points to web server or load balancer.
- **MX record** (`type: MX`, `priority: 10`) -- mail for `@your-domain.com` goes to the specified mail server; lower priority = higher preference.
- **SPF TXT record** (`type: TXT`, `value: "v=spf1 include:_spf.google.com ~all"`) -- authorizes Google to send mail; adjust `include` for your provider (e.g., `_spf.protonmail.ch`, `spf.protection.outlook.com`).
- **TTL 3600** (`ttlSeconds: 3600`) -- 1-hour cache for all records.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-domain.com>` | Your registered domain name | Domain registrar |
| `<web-server-ip>` | IPv4 address of web server or load balancer | DigitalOcean dashboard or resource outputs |
| `<mail-server-hostname>` | Mail server hostname (e.g., `aspmx.l.google.com` for Google) | Your email provider's setup instructions |

## Related Presets

- **01-simple-website** -- Use when email records are not needed
