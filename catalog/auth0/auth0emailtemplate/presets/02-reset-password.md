# Reset Password

This preset replaces Auth0's default password-reset email (the link version, `reset_email`) with one in Planton's voice: from `Planton <no-reply@planton.ai>`, naming the account (`{{ user.email }}`), with one button carrying the reset link (`{{ url }}`) that works for one hour.

## When to Use

- Any tenant with a database connection whose people reset their own passwords
- Beside the Verify Email preset, so the tenant's emails share one layout

## Key Configuration Choices

- **Template** (`template: reset_email`) -- the link version; `reset_email_by_code` sends a code instead (its body uses `{{ code }}`)
- **Link lifetime** (`urlLifetimeInSeconds: 3600`) -- one hour: a reset link acts for the person, so keep it short; the body says so
- **No redirect** -- Universal Login ignores the template's `resultUrl` after a password reset and sends people to the default login route, so the preset leaves it out; to choose where people land, use the `password-reset-post-challenge` Action trigger's `api.transaction.setResultURL()`
- **Reassurance** -- the body tells a person who did not ask for the reset that nothing changes

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.from` | Your sender, on your email provider's verified domain | Your Auth0 Email Provider's `defaultFromAddress` |
| `spec.body` | Your product name, colors and support address | Your brand |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-verify-email** -- the verification email in the same layout
- **03-welcome-email** -- the welcome email sent after the first verification
