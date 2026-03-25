# Managed WAF — Cloudflare + OWASP Rulesets

Enable Cloudflare's managed WAF rulesets to protect against common web attacks including XSS, SQL injection, and OWASP Top 10 vulnerabilities.

## When to Use

- Production websites and APIs that need baseline security protection
- Compliance requirements that mandate WAF protection
- Sites receiving untrusted user input (forms, file uploads, API endpoints)

## Key Configuration Choices

- **`phase: http_request_firewall_managed`** — The phase for executing managed rulesets
- **Cloudflare Managed Ruleset** (`efb7b8c949ac4650a09736fc376e9aee`) — Cloudflare's curated set of rules for common threats
- **OWASP Core Ruleset** (`4814384a9e5d4991b9815dcfc25d2f1f`) — Industry-standard OWASP ModSecurity Core Rule Set
- **`sensitivity_level: medium`** — Balanced between security and false positives; adjust to `low` if you see false positives or `high` for stricter enforcement

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID for your domain | Cloudflare dashboard > Overview > Zone ID |

## Related Presets

- **01-origin-rule** — Route traffic to the correct origin before WAF evaluates
- **03-cache-settings** — Cache responses after WAF processing
