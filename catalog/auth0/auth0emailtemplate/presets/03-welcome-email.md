# Welcome Email

This preset turns on Auth0's welcome email in Planton's voice: from `Planton <no-reply@planton.ai>`, sent once a new person has verified their address (or at their first sign-in, when the tenant sends no verification email), with one button that opens `https://planton.ai`.

## When to Use

- Tenants whose people sign up themselves, where the first email after verification should point them at the product
- Beside the Verify Email preset, so the tenant's emails share one layout

## Key Configuration Choices

- **Template** (`template: welcome_email`) -- sent when a person verifies their address, or at their first sign-up or login when the tenant sends no verification email; it carries no action link, so it sets no `resultUrl` or `urlLifetimeInSeconds`
- **Body** (`body`) -- `{{ user.email }}` is trusted; add `{{ user.name | escape }}` if you greet people by name, since the name is user-supplied
- **One call to action** -- the button leads to the product, written into the body

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.from` | Your sender, on your email provider's verified domain | Your Auth0 Email Provider's `defaultFromAddress` |
| `spec.body` | Your product name, first step, colors, link and support address | Your brand and onboarding |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-verify-email** -- the verification email that comes first
- **02-reset-password** -- the password-reset email in the same layout
