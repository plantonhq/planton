# Self-Managed Certificate

This preset puts the tenant's sign-in behind your own reverse proxy: your proxy terminates TLS with your certificate and forwards to the tenant's origin. Control of the domain is proven with a TXT record, and the proxy authenticates to Auth0 with the key the verification returns. It needs Auth0's Enterprise plan.

## When to Use

- Every endpoint you serve sits behind your own CDN or WAF, and the sign-in domain must too
- You must hold the sign-in domain's certificate yourself

## Key Configuration Choices

- **Self-managed certificate** (`type: self_managed_certs`) -- your proxy terminates TLS; `tlsPolicy` does not apply
- **Client IP from the proxy** (`customClientIpHeader: cf-connecting-ip`) -- Cloudflare's header, so attack protection and IP allow lists see the person rather than the proxy
- **Origin and key** -- forward to the domain's `origin_domain_name`, and send the verification's `cname_api_key` in the `cname-api-key` header

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `spec.domain` | The domain your proxy serves (sample: `login.example.com`) | A subdomain of a zone you control |
| `spec.customClientIpHeader` | The header your proxy puts the person's IP in | Your proxy's documentation: `x-forwarded-for`, `cf-connecting-ip`, `true-client-ip`, or `x-azure-clientip` |
| `metadata.org` | Your Planton organization | The Planton console's organization switcher |

## Related Presets

- **01-auth0-managed-domain-on-cloudflare** -- the same domain with Auth0 holding the certificate, on any plan
