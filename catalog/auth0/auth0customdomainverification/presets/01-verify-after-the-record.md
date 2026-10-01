# Verify After the Record

This preset completes an Auth0-managed custom domain. It waits for the Cloudflare CNAME that proves control of the domain (a `depends_on` relationship), asks Auth0 to verify the domain, and waits until the domain is ready to serve sign-in. It is the last of the three resources that bring a sign-in domain live.

## When to Use

- Every custom domain: nothing serves on it until it is verified
- Installing the domain, its record and its verification together

## Key Configuration Choices

- **Ordered after the record** (`relationships: depends_on`) -- Auth0 can verify only a record that already resolves publicly
- **The domain by reference** (`customDomainId.valueFrom`) -- reads the Auth0 Custom Domain's `status.outputs.id`
- **Any DNS kind** -- point the relationship at whichever record publishes the domain's `dns_record_*` outputs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `relationships[0].name` | The DNS record resource that publishes the domain's record | Your chart or manifest set |
| `spec.customDomainId.valueFrom.name` | The Auth0 Custom Domain resource | Your chart or manifest set |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **02-self-managed-domain** -- the verification of a domain behind your own proxy
