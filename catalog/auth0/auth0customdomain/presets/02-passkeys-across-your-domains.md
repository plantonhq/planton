# Passkeys Across Your Domains

This preset creates an Auth0-managed custom domain whose passkeys bind to the parent domain. A passkey is usable on its relying party and every subdomain of it, so a passkey created at `id.example.com` also signs a person in on `app.example.com`. Auth0 recommends the parent domain for exactly this.

## When to Use

- Your applications live on several subdomains of one domain, and you offer passkeys
- You are setting up a custom domain before anyone has enrolled a passkey (the relying party is fixed at enrollment)

## Key Configuration Choices

- **Parent-domain relying party** (`relyingPartyIdentifier: example.com`) -- one passkey for every subdomain; leave it out to bind passkeys to the custom domain itself
- **Auth0-managed certificate** (`type: auth0_managed_certs`) -- as in the Auth0-Managed Domain preset
- **Recommended TLS** (`tlsPolicy: recommended`) -- the only policy Auth0 accepts

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.domain` | The domain to sign in on (sample: `id.example.com`) | A subdomain of a zone you control |
| `spec.relyingPartyIdentifier` | The parent domain passkeys bind to (sample: `example.com`) | The registrable domain your applications share |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-auth0-managed-domain-on-cloudflare** -- the record and verification that complete the domain
