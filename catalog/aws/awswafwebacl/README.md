# AwsWafWebAcl

An AWS WAFv2 Web Access Control List (Web ACL) that protects web applications from common exploits, bots, and volumetric attacks.

## What It Does

Creates a WAFv2 Web ACL with an ordered set of rules that inspect incoming web requests and take actions (allow, block, count, CAPTCHA, challenge) based on matching conditions. Rules are evaluated by priority (lowest first); when a rule matches, its action is taken and evaluation stops (count/CAPTCHA-passed requests continue to later rules).

The full WAFv2 statement language is modeled as a typed, recursive tree — managed rule groups (with ATP/ACFP/Bot Control/anti-DDoS configs), rate limiting (including custom aggregation keys), IP set / regex pattern set / rule group references, geo/byte/SQLi/XSS/size/regex/label/ASN matching, and AND/OR/NOT composition. A raw-JSON `customStatement` escape hatch remains for anything AWS ships before this spec models it.

## When to Use

Use this component when you need to:

- Protect ALBs, API Gateways, CloudFront distributions, or App Runner services from web attacks
- Block SQL injection, cross-site scripting, and other OWASP Top 10 threats using AWS Managed Rules
- Rate-limit requests with custom aggregation keys (header, cookie, query, IP, forwarded IP)
- Block or allow traffic from specific countries, ASNs, or IP sets
- Match request components against shared regex pattern sets

## Scope

- **REGIONAL** — protects ALB, API Gateway REST/HTTP, AppSync, Cognito User Pools, App Runner, Verified Access. Created in the same region as the protected resource. Associate from [AwsAlb](../awsalb/README.md) via `web_acl_arn`.
- **CLOUDFRONT** — protects CloudFront distributions. Must be created in **us-east-1**. Associate from [AwsCloudFront](../awscloudfront/README.md) via `web_acl_arn`.

## What is NOT Bundled

- **Associations** — The Web ACL ARN is the primary output. Protected resources reference it via `StringValueOrRef`; the web ACL never points at the resources it protects.
- **IP Sets** — First-class [AwsWafIpSet](../awswafipset/README.md) resources. Reference by ARN in `ip_set_reference` statements.
- **Regex Pattern Sets** — First-class [AwsWafRegexPatternSet](../awswafregexpatternset/README.md) resources. Reference by ARN in `regex_pattern_set_reference` statements.
- **Rule groups** — Customer-managed rule groups are separate AWS resources; reference by ARN in `rule_group_reference` statements.

## Statement Tree

Each rule carries exactly one root `statement`. Statements compose recursively:

| Category | Statement kinds |
|----------|-----------------|
| **AWS-managed** | `managed_rule_group` (with optional scope-down, rule action overrides, managed configs for ATP/ACFP/Bot Control/anti-DDoS) |
| **Rate limiting** | `rate_based` (IP, forwarded IP, constant, or custom aggregation keys) |
| **Reusable sets** | `ip_set_reference`, `regex_pattern_set_reference` |
| **References** | `rule_group_reference` |
| **Match primitives** | `geo_match`, `byte_match`, `sqli_match`, `xss_match`, `size_constraint`, `regex_match`, `label_match`, `asn_match`, IP/subnet, JA3/JA4 fingerprint |
| **Composition** | `and`, `or`, `not` (nested statements) |

Managed rule groups use `override_action` (count/none); custom match rules use `action` (allow/block/count/captcha/challenge).

## Top-Level Configuration

Beyond rules, the spec models:

- **CAPTCHA/challenge immunity** — `captcha_config`, `challenge_config` for token replay windows
- **Association config** — per-resource body inspection size limits (ALB, AppSync, Cognito, Verified Access)
- **Data protection** — field-level masking for logged or inspected data; a
  `SINGLE_*` protection without `field_keys` protects ALL keys of that type
- **Logging** — single bundled logging configuration (AWS allows exactly one
  destination per web ACL) with full field redaction (headers, URI path,
  query string, method) and keep/drop log filtering by applied action or
  request label — e.g. keep only BLOCK and COUNT records to cut logging
  cost on high-traffic ACLs
- **Custom response bodies** — reusable templates for block actions

## Prerequisites

- AWS credentials with `wafv2:*` permissions
- For CLOUDFRONT scope: provider configured for us-east-1
- For set-reference rules: [AwsWafIpSet](../awswafipset/README.md) and/or [AwsWafRegexPatternSet](../awswafregexpatternset/README.md) in matching scope
- For logging: a destination resource named starting with `aws-waf-logs-`

## Outputs

| Output | Description |
|--------|-------------|
| `web_acl_arn` | Web ACL ARN for associations |
| `web_acl_id` | Web ACL unique ID |
| `web_acl_name` | Web ACL name |
| `capacity` | WCUs consumed (max 5,000 per ACL) |
| `application_integration_url` | JavaScript integration endpoint for CAPTCHA/challenge clients |

## Capacity (WCU) Limits

Each rule consumes Web ACL Capacity Units. The total must not exceed 5,000 WCUs per Web ACL. Managed rule groups consume varying amounts (e.g., AWSManagedRulesCommonRuleSet: 700 WCUs). Monitor the `capacity` output to plan rule additions.

## Nesting Depth

The Terraform module unrolls the statement tree to three levels of AND/OR/NOT nesting. Deeper trees fail the plan with an explicit precondition — extend the module or flatten with `customStatement` if you need more.

See [reference.md](./v1alpha1/reference.md) for the full field reference and [catalog.md](./catalog.md) for deployment examples.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
