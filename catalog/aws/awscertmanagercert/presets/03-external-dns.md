# DNS-Validated Certificate with External DNS

This preset requests an ACM certificate for a domain whose DNS lives outside Route53 (Cloudflare, a registrar's DNS, an on-prem zone). The deployment creates the certificate and finishes without waiting: the certificate rests in `PENDING_VALIDATION`, and the exact CNAME records to create are exported as the `domain_validation_records` output. Once the records exist in your DNS, ACM issues the certificate -- and renews it automatically for as long as the records stay in place.

## When to Use

- Your domain's DNS is not hosted in Route53
- DNS changes go through another team or system, and you want the certificate (and its stable ARN) created ahead of time
- You are migrating DNS and want the validation records staged before the cutover

## Key Configuration Choices

- **No `route53HostedZoneId`** -- the module manages no DNS; validation records are handed to you as outputs instead
- **DNS validation (the default)** -- the records are one-time setup; leaving them in place makes every future renewal automatic
- **Stable ARN immediately** -- downstream resources can reference `status.outputs.cert_arn` right away, but TLS consumers (listeners, CloudFront) will fail to attach the certificate until it is ISSUED

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | Region for the certificate (`us-east-1` if this certificate fronts CloudFront) | Your architecture |
| `replaceme.example.com` | The domain name to secure — a real DNS shape; the field's pattern rejects placeholders | Your DNS provider |

## After Deploying

1. Read the `domain_validation_records` output -- one CNAME per domain.
2. Create each record in your DNS provider exactly as given.
3. ACM detects the records (typically within minutes) and flips the certificate to `ISSUED`.

## Related Presets

- **01-single-domain-dns** -- Use instead when the domain's DNS is a Route53 hosted zone (fully automated validation)
- **02-wildcard-domain** -- Wildcard coverage with automated Route53 validation
