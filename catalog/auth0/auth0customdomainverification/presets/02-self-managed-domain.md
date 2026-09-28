# Self-Managed Domain

This preset verifies a custom domain behind your own reverse proxy. It waits for the TXT record that proves control of the domain, verifies it, and hands back `cname_api_key`, the key your proxy sends to Auth0 in the `cname-api-key` header. Auth0 returns the key once; Planton keeps it as a secret output.

## When to Use

- A custom domain with `type: self_managed_certs` (Enterprise plan)

## Key Configuration Choices

- **Ordered after the TXT record** (`relationships: depends_on`) -- the record Auth0 checks for a self-managed domain
- **The proxy key as a secret** (`status.outputs.cname_api_key`) -- feed it to your proxy through your secret store

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `relationships[0].name` | The DNS record resource that publishes the TXT record | Your chart or manifest set |
| `spec.customDomainId.valueFrom.name` | The self-managed Auth0 Custom Domain resource | Your chart or manifest set |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-verify-after-the-record** -- the verification of an Auth0-managed domain
