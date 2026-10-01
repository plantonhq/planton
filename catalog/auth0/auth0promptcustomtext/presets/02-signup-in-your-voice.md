# Signup in Your Voice

This preset re-words the English sign-up screen: a title that names your product, the sentence that names the application ("Sign up to Planton to continue to ${clientName}."), the log-in nudge for people who already have an account, and the message a person sees when their email is already registered. Every key it leaves out keeps Auth0's default English.

The screens and keys come from Auth0's customization reference: [Customize Universal Login Text Elements](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements#prompt-values). A key may use Universal Login's variables, such as `${clientName}` and `${companyName}`.

## When to Use

- A tenant that lets people sign up themselves, on the one-screen `signup` prompt
- Paired with **01-login-in-your-voice**, so logging in and signing up read as one product

## Key Configuration Choices

- **One prompt, one language** (`prompt: signup`, `language: en`) -- this resource owns all of the sign-up prompt's English words
- **The already-registered message** (`email-in-use`) -- sends the person to log in instead of leaving them at a dead end
- **Identifier-first tenants** -- with Auth0 Prompt's `identifierFirst`, sign-up runs on the `signup-id` and `signup-password` prompts; re-word those instead
- **Custom text variables** -- a key named `var-<name>` defines words a screen partial reads from `prompt.screen.text` (`var-tos` is read as `prompt.screen.text.varTos`), so a terms-of-service label added by the Auth0 Prompt Screen Partials kind is translated here. Partials render only with Auth0 Branding's `universal_login_template`, which needs a custom domain

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.screens.signup.texts` | Your product's name and voice in each text | Your brand |
| `support@planton.ai` | Where people ask for help | Your support mailbox |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-login-in-your-voice** -- the login screen in the same voice
