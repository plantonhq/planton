# Company-Wide Email Domain Access

Allows access to a protected hostname for anyone with an email from your company domain (e.g., @company.com). Simple Zero Trust policy for internal tools when the entire organization should have access.

## When to Use

- Internal dashboards, wikis, or tools for all company employees
- Restricting access to verified company emails only
- Quick setup without per-user or group management

## Key Configuration Choices

- **allowedEmails** (`allowedEmails: ["*@<company-domain>"]`) -- Wildcard matches any user at your domain.
- **policyType ALLOW** (`policyType: ALLOW`) -- Allowlist; use BLOCK to block specific patterns.
- **sessionDurationMinutes 1440** (`sessionDurationMinutes: 1440`) -- 24 hours; adjust for shorter/longer sessions.
- **requireMfa: false** (`requireMfa: false`) -- No MFA; set true for higher security.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for the protected hostname's domain | CloudflareDnsZone status.outputs.zone_id |
| `<company-domain>` | Your company email domain | Your organization's email domain (e.g., company.com) |
| `app.example.com` | Hostname to protect | Your application's FQDN within the zone |
