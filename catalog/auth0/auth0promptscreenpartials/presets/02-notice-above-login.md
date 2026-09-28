# Notice Above Login

This preset shows a notice at the start of the login form, above the email field: a maintenance window and the address to write to. It extends the `login-id` prompt, the first screen of identifier-first login, so it pairs with Auth0 Prompt's **Identifier-First Login** preset.

The screens and insertion points each prompt offers come from Auth0's customization reference: [Customize Signup and Login Prompts](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts). Partials render only inside a page template -- Auth0 Branding's `universal_login_template` -- and a page template needs a custom domain on the tenant (Auth0 Custom Domain).

## When to Use

- A maintenance window, an incident, or a change people should hear about before they sign in
- A standing note: which account to use, or where to get an invitation

## Key Configuration Choices

- **The identifier screen** (`promptType: login-id`, `screenName: login-id`) -- on the identifier-and-password flow, use `promptType: login` and `screenName: login` instead
- **Above the fields** (`formContentStart`) -- the first thing a person reads on the form
- **Announced, not hidden** (`role="status"`) -- screen readers read the notice with the page
- **Remove it when it is over** -- the notice stays until the next apply; delete the entry (or the resource) once the window has passed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.screenPartials[0].insertionPoints.formContentStart` | Your notice | Your status page or change calendar |
| `support@planton.ai` | Where people ask for help | Your support mailbox (the Auth0 Tenant Settings kind's `supportEmail`) |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-terms-on-signup** -- a required terms checkbox on the sign-up screen
