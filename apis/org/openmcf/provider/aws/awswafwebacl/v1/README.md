# AwsWafWebAcl

An AWS WAFv2 Web Access Control List (Web ACL) that protects web applications from common exploits, bots, and volumetric attacks.

## What It Does

Creates a WAFv2 Web ACL with an ordered set of rules that inspect incoming web requests and take actions (allow, block, count, CAPTCHA, challenge) based on matching conditions. Rules are evaluated by priority (lowest first); when a rule matches, its action is taken and evaluation stops.

## When to Use

Use this component when you need to:

- Protect ALBs, API Gateways, CloudFront distributions, or App Runner services from web attacks
- Block SQL injection, cross-site scripting, and other OWASP Top 10 threats using AWS Managed Rules
- Rate-limit requests to prevent DDoS and brute-force attacks
- Block or allow traffic from specific countries
- Filter requests by IP allowlists or blocklists

## Scope

- **REGIONAL** — protects ALB, API Gateway REST/HTTP, AppSync, Cognito User Pools, App Runner. Created in the same region as the protected resource.
- **CLOUDFRONT** — protects CloudFront distributions. Must be created in **us-east-1**.

## What is NOT Bundled

- **Associations** — The Web ACL ARN is the primary output. Protected resources (ALB, API Gateway, etc.) reference it via StringValueOrRef. CloudFront associations are configured on the CloudFront distribution itself.
- **IP Sets** — WAFv2 IP Sets are separate resources. Reference them by ARN in `ipSetReference` rules.
- **Regex Pattern Sets** — Separate WAFv2 resources. Use `customStatement` for regex-based rules.

## Statement Types

This component models the four most common WAF rule types as first-class proto messages:

| Statement | Description | Action Type |
|-----------|-------------|-------------|
| `managedRuleGroup` | AWS Managed Rules or marketplace rule groups | `overrideAction` |
| `rateBased` | Rate limiting by IP, forwarded IP, or constant | `action` |
| `geoMatch` | Geographic blocking/allowing by country code | `action` |
| `ipSetReference` | IP allowlist/blocklist via WAFv2 IP Set ARN | `action` |

For all other statement types (SQL injection, XSS, byte match, regex, AND/OR/NOT compound, size constraint, label match, rule group reference), use the `customStatement` escape hatch with the raw AWS WAFv2 JSON structure.

## Prerequisites

- AWS credentials with `wafv2:*` permissions
- For CLOUDFRONT scope: provider configured for us-east-1
- For IP set rules: existing WAFv2 IP Sets created separately
- For logging: a destination resource named starting with `aws-waf-logs-`

## Spec Fields

### Top Level

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `scope` | string | Yes | `REGIONAL` or `CLOUDFRONT` |
| `defaultAction` | message | Yes | Action when no rule matches (allow or block) |
| `description` | string | No | Human-readable description (max 256 chars) |
| `rules` | repeated | No | Ordered rule set |
| `visibilityConfig` | message | No | CloudWatch metrics (defaults: enabled, metric=name) |
| `customResponseBodies` | repeated | No | Reusable response body templates |
| `tokenDomains` | repeated string | No | Domains for CAPTCHA/Challenge tokens |
| `logging` | message | No | Logging destination and redacted fields |

### Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Rule name (1-128 chars, unique) |
| `priority` | int32 | Evaluation order (lower = first) |
| `action` | string | For custom rules: allow/block/count/captcha/challenge |
| `overrideAction` | string | For managed groups: count/none |
| `managedRuleGroup` | message | AWS Managed Rule Group statement |
| `rateBased` | message | Rate limiting statement |
| `geoMatch` | message | Geographic match statement |
| `ipSetReference` | message | IP set reference statement |
| `customStatement` | Struct | Escape hatch for any WAFv2 statement |
| `customResponse` | message | Custom block response |
| `customRequestHeaders` | repeated | Headers to add on allow/count |
| `ruleLabels` | repeated string | Labels for downstream rules |
| `visibilityConfig` | message | Rule-level CloudWatch metrics |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `web_acl_arn` | Web ACL ARN for associations |
| `web_acl_id` | Web ACL unique ID |
| `web_acl_name` | Web ACL name |
| `capacity` | WCUs consumed (max 5,000) |

## Capacity (WCU) Limits

Each rule consumes Web ACL Capacity Units. The total must not exceed 5,000 WCUs per Web ACL. Managed rule groups consume varying amounts (e.g., AWSManagedRulesCommonRuleSet: 700 WCUs). Monitor the `capacity` output to plan rule additions.

## Deliberately Omitted (v1)

| Feature | Reason | Adoption |
|---------|--------|----------|
| CAPTCHA/Challenge config | Niche; use defaults | <15% |
| Association config | Body inspection limits; rare | <10% |
| Data protection config | Field-level masking; very new | <5% |
| Custom keys (rate-based) | Complex nested structure | <10% |
| ASN match | Very new statement type | <5% |
| JA3/JA4 fingerprint | Niche TLS fingerprinting | <5% |
| Log filter conditions | Complex filter structure | <15% |
