---
title: "Preset: WAF OWASP Protection"
description: "Protect web applications and APIs against OWASP Top 10–style attacks: SQL injection (SQLi), cross-site scripting (XSS), and Layer 7 DDoS. Uses Cloud Armor preconfigured WAF rules derived from OWASP..."
type: "preset"
rank: "03"
presetSlug: "03-waf-owasp-protection"
componentSlug: "cloud-armor-policy"
componentTitle: "Cloud Armor Policy"
provider: "gcp"
icon: "package"
order: 3
---

# Preset: WAF OWASP Protection

## Use Case

Protect web applications and APIs against OWASP Top 10–style attacks: SQL injection (SQLi), cross-site scripting (XSS), and Layer 7 DDoS. Uses Cloud Armor preconfigured WAF rules derived from OWASP ModSecurity CRS 3.3.2, with adaptive protection and verbose logging for debugging.

## What This Creates

- A CLOUD_ARMOR policy with L7 DDoS adaptive protection (STANDARD visibility)
- JSON body parsing (STANDARD) and verbose logging
- SQLi rule (priority 1000): blocks requests matching `sqli-v33-stable` at sensitivity 1
- XSS rule (priority 2000): blocks requests matching `xss-v33-stable` at sensitivity 1
- Default allow rule (priority 2147483647) for all other traffic

## OWASP ModSecurity Rules

The rules reference Cloud Armor preconfigured expression sets built from OWASP Core Rule Set (CRS). Each set has multiple signatures; sensitivity levels control which signatures run. Higher sensitivity catches more attacks but increases false positives.

| Rule Set | Sensitivity | Effect |
|----------|-------------|--------|
| `sqli-v33-stable` | 1–4 | SQL injection signatures (1 = fewer false positives) |
| `xss-v33-stable` | 1–2 | Cross-site scripting signatures |

## Adding Exclusions for False Positives

Legitimate traffic (e.g., search with SQL-like terms, rich text with HTML) can trigger WAF rules. Add `preconfiguredWafConfig` exclusions to the matching rule:

```yaml
preconfiguredWafConfig:
  exclusions:
    - targetRuleSet: sqli-v33-stable
      targetRuleIds: ["owasp-crs-v030301-id942100-sqli"]  # optional: specific rule IDs
      requestUris:
        - operator: STARTS_WITH
          value: /api/search
```

Exclude by `requestHeaders`, `requestCookies`, `requestUris`, or `requestQueryParams` using operators: `EQUALS`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `EQUALS_ANY`.
