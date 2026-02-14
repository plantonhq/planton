# Single Domain DNS-Validated Certificate

This preset provisions an ACM certificate for a single domain using automated DNS validation via Route53. DNS validation is the recommended method because it requires no manual intervention -- ACM automatically creates and manages the validation CNAME records in your Route53 hosted zone.

## When to Use

- HTTPS for a single domain or subdomain (e.g., `api.example.com` or `example.com`)
- Standard production certificates where DNS is managed in Route53
- Any ACM certificate use case where you do not need wildcard or multi-domain coverage

## Key Configuration Choices

- **Single domain** (`primaryDomainName`) -- One certificate for one domain; simplest and most common pattern
- **DNS validation** (`validationMethod: DNS`) -- Fully automated when using Route53; no email approval needed, auto-renews before expiration
- **No alternate domains** -- Certificate covers exactly one domain; add `alternateDomainNames` if you need SANs

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-domain.com>` | The domain name to secure (e.g., `api.example.com`) | Your domain registrar or DNS provider |
| `<route53-hosted-zone-id>` | ID of the Route53 hosted zone matching your domain | AWS Route53 console or `AwsRoute53Zone` status outputs |

## Related Presets

- **02-wildcard-domain** -- Use instead when you need a certificate covering all subdomains (`*.example.com`)
