# Resend over SMTP

This preset sends the tenant's emails -- verification, password reset, invitations, multi-factor codes -- through [Resend](https://resend.com), from `no-reply@planton.ai`. Auth0's API accepts Resend by name, but the Terraform provider this kind deploys through (terraform-provider-auth0 v1.58.0) does not list it among the services it accepts, so the preset reaches Resend through its SMTP interface, which Resend offers for exactly this.

## When to Use

- Your product already sends through Resend, and Auth0's emails should leave from the same verified domain
- You want delivery insights for Auth0's emails in Resend's Emails page beside your own

## Key Configuration Choices

- **Host, user and port** (`smtp.host: smtp.resend.com`, `smtp.user: resend`, `smtp.port: 587`) -- the values Resend publishes for SMTP ([Send emails with SMTP](https://resend.com/docs/send-with-smtp)). Resend offers 25, 587 and 2587 with STARTTLS and 465 and 2465 with implicit TLS; 587 is the standard submission port and the one Auth0 recommends.
- **Password** (`smtp.password`) -- a Resend API key with sending access, read from an organization secret (`planton secret set resend-api-key --string`)
- **Sender** (`defaultFromAddress`) -- an address on a domain you have verified with Resend; a template can name its own sender (Auth0 Email Template's `from`)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.defaultFromAddress` | The sender, on a domain verified with Resend | Resend dashboard, Domains |
| `spec.smtp.password` | The secret holding your Resend API key | Resend dashboard, API Keys (create one with sending access); store it as an organization secret |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-amazon-ses** -- Amazon SES with an IAM access key
- **03-sendgrid** -- Twilio SendGrid with an API key
