# Verify Email

This preset replaces Auth0's default verification email with one in Planton's voice: from `Planton <no-reply@planton.ai>`, a short HTML message naming the address being confirmed (`{{ user.email }}`), one button carrying the verification link (`{{ url }}`), and a support address. After verifying, Auth0 shows its own result page: the preset leaves `resultUrl` out, because Auth0 refuses a custom one on non-Enterprise tenants created on or after May 5, 2026.

## When to Use

- Any tenant whose database connection asks people to verify their email: the verification email is the first thing a new user receives from you
- As the starting point for the tenant's other emails, which can share its layout

## Key Configuration Choices

- **Sender** (`from`) -- on a domain the tenant's email provider may send for; it overrides the provider's `defaultFromAddress` for this email
- **Body** (`body`) -- a single-column, table-based layout with inline styles, which email clients render consistently. `{{ user.email }}` is trusted; untrusted values such as `{{ user.name }}` need the `escape` filter.
- **Where the person lands** (`resultUrl`) -- add it to send people to your application after verifying; Auth0 appends `success` and `message` query parameters. Auth0 refuses a custom `resultUrl` (403) on non-Enterprise tenants created on or after May 5, 2026, which is why this preset leaves it out.
- **Link lifetime** (`urlLifetimeInSeconds: 86400`) -- one day, shorter than Auth0's five-day default; the body says so
- **Destroy disables** -- Auth0 cannot delete a template, so destroy turns it off and the default email returns

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.from` | Your sender, on your email provider's verified domain | Your Auth0 Email Provider's `defaultFromAddress` |
| `spec.body` | Your product name, colors and support address | Your brand |
| `spec.resultUrl` | Where people land after verifying | Your application's home or welcome page |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-reset-password** -- the password-reset email in the same layout
- **03-welcome-email** -- the welcome email sent after the first verification
