# Managed Let's Encrypt Certificate

This preset creates a Hetzner Cloud managed certificate that automatically obtains and renews a TLS certificate from Let's Encrypt. You specify one or more domain names, and Hetzner Cloud handles issuance, validation (via ACME HTTP-01 challenge), and renewal -- no manual certificate management required. The resulting certificate is referenced by HetznerCloudLoadBalancer HTTPS services via its certificate ID.

This is the recommended certificate type for new deployments. It eliminates certificate expiry as an operational concern entirely.

## When to Use

- Any HTTPS service behind a Hetzner Cloud load balancer where you control the DNS for the domain
- Standard web applications, APIs, or services that need TLS termination at the load balancer
- Environments where minimizing operational overhead (no manual renewal) is a priority

## Key Configuration Choices

- **Managed variant** (`managed`) -- Hetzner Cloud provisions and auto-renews a Let's Encrypt certificate; zero ongoing maintenance
- **Multi-domain SAN** (`domainNames` with 2 entries) -- a single certificate covers both the apex domain and a subdomain (e.g., `example.com` + `www.example.com`); add or remove entries as needed
- **No uploaded material** -- no PEM files to manage, rotate, or store securely

## Prerequisites

Before provisioning, each domain in `domainNames` must satisfy:

1. A DNS A or AAAA record pointing to a Hetzner Cloud load balancer
2. The load balancer must have an HTTPS service configured that references this certificate

These are required for the ACME HTTP-01 challenge to succeed. If DNS is not configured, certificate issuance will fail.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<primary-domain>` | Primary domain name for the certificate (e.g., `example.com`) | Your domain registrar or DNS configuration |
| `<additional-domain>` | Additional domain for the SAN certificate (e.g., `www.example.com`); remove this line if only one domain is needed | Your domain registrar or DNS configuration |

## Related Presets

- **02-uploaded-certificate** -- use instead when you have an existing certificate from a commercial CA, internal CA, or need a wildcard certificate (Let's Encrypt HTTP-01 cannot validate wildcards)
