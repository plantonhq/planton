# Terms on Signup

This preset adds a required terms-of-service checkbox to the sign-up screen, below the fields and above the submit button. The browser will not submit the form until it is ticked, and the field reaches Actions as `ulp-terms-of-service` on `event.request.body`, so a pre-user-registration Action can record the acceptance on the new user (or refuse the sign-up with `api.validation.error`).

The screens and insertion points each prompt offers come from Auth0's customization reference: [Customize Signup and Login Prompts](https://auth0.com/docs/customize/login-pages/universal-login/customize-signup-and-login-prompts). Partials render only inside a page template -- Auth0 Branding's `universal_login_template` -- and a page template needs a custom domain on the tenant (Auth0 Custom Domain).

## When to Use

- A product whose terms must be accepted before an account exists
- Any other field sign-up should collect (a company name, a marketing opt-in), in the same place and the same way

## Key Configuration Choices

- **The one-screen sign-up** (`promptType: signup`, `screenName: signup`) -- with Auth0 Prompt's `identifierFirst`, sign-up runs on the `signup-id` prompt instead; move the entry there (`promptType: signup-id`, `screenName: signup-id`)
- **Below the fields** (`formContentEnd`) -- above the submit button, where a person reads it last before continuing
- **Auth0's field styles** (`class="ulp-field"`) -- the checkbox looks like the rest of the form
- **The `ulp-` prefix** -- only fields named `ulp-...` are passed to Actions
- **One label in every language** -- replace the label with `{{ prompt.screen.text.varTos }}` and define `var-tos` per language with the Auth0 Prompt Custom Text kind

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `https://planton.ai/terms` | Your terms of service page | Your legal pages |
| `Planton Terms of Service` | Your product's name in the label | Your brand |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-notice-above-login** -- a notice at the top of the login form
