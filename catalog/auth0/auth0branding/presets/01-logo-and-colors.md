# Logo and Colors

This preset puts your logo and colors on every login, signup, reset and consent page of the tenant: Planton's own look as the sample, the Planton mark on a black page with black buttons. It manages only these settings, so the tenant's favicon, font, page template and theme are left as they are.

## When to Use

- Any tenant real people sign in through, on any plan, the Free plan included
- The first branding step, before a theme or a page template (neither needs this one, and all three combine)

## Key Configuration Choices

- **Logo** (`logoUrl`) -- a public HTTPS image of about 150 x 150 pixels, PNG or SVG; it takes precedence over the tenant's `pictureUrl` (Auth0 Tenant Settings) on the login pages
- **Primary color** (`colors.primary`) -- the accent of buttons and links, a hex color
- **Page background** (`colors.pageBackground`) -- the page around the login box: a hex color, or Auth0's gradient JSON
- **Unset is unmanaged** -- add `faviconUrl` or `fontUrl` to manage them too; leave them out and the tenant keeps its current values

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.logoUrl` | A public HTTPS URL of your logo (sample: Planton's mark) | Your brand assets host |
| `spec.colors.primary` | Your accent color (sample: `#000000`) | Your brand guidelines |
| `spec.colors.pageBackground` | The page's background (sample: `#000000`) | Your brand guidelines |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-page-template-on-custom-domain** -- the login box inside your own page, on your own domain
- **03-themed-login-box** -- the login box's buttons, inputs and layout
