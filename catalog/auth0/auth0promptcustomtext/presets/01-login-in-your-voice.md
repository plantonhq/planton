# Login in Your Voice

This preset re-words the English login screen: a welcome in your product's name, the sentence that names the application ("Log in to Planton to continue to ${clientName}."), the sign-up nudge, and two error messages that tell people what to do next, including where to write when their account is blocked. Every key it leaves out keeps Auth0's default English.

The screens and keys come from Auth0's customization reference: [Customize Universal Login Text Elements](https://auth0.com/docs/customize/login-pages/universal-login/customize-text-elements#prompt-values). A key may use Universal Login's variables, such as `${clientName}` (the application's name) and `${companyName}` (the tenant's friendly name).

## When to Use

- A tenant on the identifier-and-password flow, where the `login` prompt is the one page people sign in on
- Before a launch, so the first page a customer sees speaks like the rest of your product

## Key Configuration Choices

- **One prompt, one language** (`prompt: login`, `language: en`) -- Auth0 stores one custom text per pair and replaces it whole, so this resource owns all of the login prompt's English words
- **The application's name** (`${clientName}`) -- filled in per application, from the Auth0 Client kind's `name`
- **Errors that lead somewhere** (`wrong-credentials`, `user-blocked`) -- the message a stuck person reads names the next step and the support address
- **Identifier-first tenants** -- with Auth0 Prompt's `identifierFirst`, the first screen is the `login-id` prompt and the second the `login-password` prompt; re-word those two instead

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.screens.login.texts` | Your product's name and voice in each text | Your brand |
| `support@planton.ai` | Where people ask for help | Your support mailbox (the Auth0 Tenant Settings kind's `supportEmail`) |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-signup-in-your-voice** -- the sign-up screen in the same voice
