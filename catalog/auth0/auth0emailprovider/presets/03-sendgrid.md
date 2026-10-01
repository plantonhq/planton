# SendGrid

This preset sends the tenant's emails through Twilio SendGrid, from `no-reply@planton.ai`, with one API key.

## When to Use

- Your product already sends through SendGrid
- You want the smallest configuration: one key, no region, no server details

## Key Configuration Choices

- **API key** (`sendgrid.apiKey`) -- a restricted key with only Mail Send access, read from an organization secret (`planton secret set sendgrid-api-key --string`)
- **Sender** (`defaultFromAddress`) -- an address on a domain authenticated in SendGrid (Sender Authentication), so the emails are signed for your domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.defaultFromAddress` | The sender, on an authenticated domain | SendGrid, Settings, Sender Authentication |
| `spec.sendgrid.apiKey` | The secret holding your SendGrid API key | SendGrid, Settings, API Keys (Restricted Access, Mail Send); store it as an organization secret |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-resend-over-smtp** -- Resend through its SMTP interface
- **02-amazon-ses** -- Amazon SES with an IAM access key
