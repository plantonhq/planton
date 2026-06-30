# Let's Encrypt Wildcard Certificate

> **IaC Support Pending**: As of this writing, the Civo provider does not support certificate management through Terraform or Pulumi. This preset defines the correct YAML structure for when IaC support is added. In the meantime, certificates can be managed through the Civo API or dashboard directly.

This preset requests a free Let's Encrypt certificate covering both the apex domain and all subdomains via a wildcard. Auto-renewal is enabled by default, so the certificate renews automatically before expiration.

## When to Use

- Any domain on Civo that needs TLS (HTTPS)
- Wildcard certificates covering all subdomains (e.g., `api.example.com`, `app.example.com`)
- Free, automated TLS without managing certificate files manually

## Key Configuration Choices

- **Let's Encrypt** (`type: letsEncrypt`) -- free, automated, trusted by all browsers
- **Wildcard + apex** (`domains: [example.com, *.example.com]`) -- covers the root domain and all subdomains with a single certificate
- **Auto-renewal enabled** (`disableAutoRenew` omitted) -- Let's Encrypt certificates expire every 90 days; auto-renewal prevents outages

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `example.com` | Your domain name | Your domain registrar |
| `*.example.com` | Wildcard for all subdomains | Same domain as above |
