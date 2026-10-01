# Your Domain in Emails

This preset makes a verified custom domain the tenant's default, so the verification, password-reset and invitation emails Auth0 sends link to your domain rather than to the tenant's `auth0.com` address. The reference reads the Auth0 Custom Domain Verification's `domain`, so the default is set only after Auth0 has verified the domain.

## When to Use

- Any tenant with a custom domain: without a default, Auth0 uses the canonical domain for email links

## Key Configuration Choices

- **The default by reference** (`defaultCustomDomain.valueFrom`) -- only a verified domain can become the default
- **The canonical domain as a literal** -- set `defaultCustomDomain.value` to the tenant's `auth0.com` domain to make it the default again
- **Destroy keeps the default** -- Auth0 has no way to unset one

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.friendlyName` | Your product's name | Your brand |
| `spec.defaultCustomDomain.valueFrom.name` | Your Auth0 Custom Domain Verification resource | Your chart or manifest set |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-product-branded-login** -- the tenant's name, logo and support contacts
