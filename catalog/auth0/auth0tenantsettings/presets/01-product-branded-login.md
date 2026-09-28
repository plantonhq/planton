# Product-Branded Login

This preset makes the tenant introduce itself as your product. Universal Login reads "Log in to Example to continue to *application*", your logo replaces Auth0's on the login and consent pages, and people who can't sign in see your support address and page.

## When to Use

- Any tenant real people sign in through: a product's production tenant, or a pre-production tenant prospects or customers see

## Key Configuration Choices

- **Friendly name** (`friendlyName`) -- the first half of Universal Login's sentence; name each application for the second half with the Auth0 Client kind's `name`
- **Logo** (`pictureUrl`) -- a public HTTPS image of about 150 x 150 pixels
- **Support contacts** (`supportEmail`, `supportUrl`) -- where people who can't sign in get help
- **Unset is unmanaged** -- leave out any field and the tenant keeps its current value

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.friendlyName` | Your product's name | Your brand |
| `spec.pictureUrl` | A public HTTPS URL of your logo | Your brand assets host |
| `spec.supportEmail` | Where people ask for help | Your support mailbox |
| `spec.supportUrl` | Your help page | Your support site |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-your-domain-in-emails** -- email links on your own verified domain
- **03-mcp-ready-tenant** -- the settings MCP clients rely on to register and ask for tokens
- **04-short-idle-sessions** -- sessions that end when left unused
