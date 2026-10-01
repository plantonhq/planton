# Themed Login Box

This preset styles the login box itself without writing CSS: pill-shaped buttons, rounder corners, your accent on buttons, links and focus rings, a bolder title, a dark page, and the logo and header aligned left. The theme is applied whole, so every field it leaves out takes Auth0's default and the box always matches the declaration.

## When to Use

- The login box should carry your product's look, on any plan, the Free plan included
- You want the look without a page template (no custom domain needed), or inside one

## Key Configuration Choices

- **Only the overrides** -- the theme lists only what differs from Auth0's look; borders, colors, fonts, the page background and the widget are always sent, each unset field at Auth0's default
- **Accent in three places** (`colors.primaryButton`, `colors.linksFocusedComponents`, `colors.baseFocusColor`) -- the button fill, links and focused fields, and the focus outline
- **Title** (`fonts.title`) -- bold and at 125% of the reference text size (Auth0 accepts 75 to 150 for the title)
- **Layout** (`widget.logoPosition`, `widget.headerTextAlignment`) -- the logo and title aligned left; `logoPosition: none` hides the logo
- **Destroy deletes the theme** -- the login box returns to Auth0's look; the logo stays in place

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.logoUrl` | A public HTTPS URL of your logo | Your brand assets host |
| `spec.theme.displayName` | The theme's name in the Auth0 dashboard | Your product's name |
| `spec.theme.colors` | Your accent color | Your brand guidelines |
| `spec.theme.pageBackground.backgroundColor` | The page around the login box | Your brand guidelines |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-logo-and-colors** -- the logo and colors every page shares
- **02-page-template-on-custom-domain** -- the login box inside your own page
