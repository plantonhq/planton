# GcpCloudArmorPolicy — Research Documentation

This document captures research and design decisions for the GcpCloudArmorPolicy Planton component. It provides background on Google Cloud Armor, the 80/20 scoping rationale, and deliberate exclusions.

---

## Google Cloud Armor Overview

Google Cloud Armor is a distributed denial-of-service (DDoS) and web application firewall (WAF) service that protects applications and APIs behind Google Cloud HTTP(S) load balancers, Cloud CDN, and backend buckets. It operates at the edge of Google’s network, filtering traffic before it reaches backends.

Cloud Armor policies consist of:

1. **Policy-level configuration** — Type, adaptive protection, advanced options
2. **Rules** — Prioritized list of match conditions and actions
3. **Default rule** — Always present at priority 2147483647; typically "allow all"

Rules are evaluated in ascending priority order (lowest number first). The first matching rule’s action is applied. If no rule matches, the default rule applies.

---

## Policy Types and Their Use Cases

| Type | Attachments | Use Case |
|------|-------------|----------|
| **CLOUD_ARMOR** | Backend services of HTTP(S) load balancers | Full WAF, rate limiting, redirects, headers; backend security |
| **CLOUD_ARMOR_EDGE** | Cloud CDN, backend buckets | IP/geo rules only; edge-level filtering |
| **CLOUD_ARMOR_INTERNAL_SERVICE** | Internal Traffic Director | Limited feature set; internal service mesh |

The policy type is immutable after creation. Choosing the wrong type requires recreating the policy.

---

## Rule Evaluation Order and Priority System

- Priority is an integer from 0 to 2147483647.
- Lower values = higher priority (evaluated first).
- Each rule must have a unique priority.
- Priority 2147483647 is reserved for the default rule.

**Evaluation flow**: For each request, Cloud Armor evaluates rules from lowest priority number to highest. The first rule whose match condition is satisfied triggers that rule’s action. If no rule matches, the default rule (2147483647) applies.

---

## Match Conditions

### IP-Based Matching

- **versioned_expr**: `SRC_IPS_V1` (only supported value)
- **src_ip_ranges**: List of CIDR ranges (max 10 per rule)
- Use `["*"]` to match all IPs

Example: Allow office and VPN ranges, deny rest.

### CEL Expression Matching

Common Expression Language (CEL) provides flexible matching on request attributes:

- **origin.region_code** — Client country/region
- **origin.ip** — Client IP
- **request.path** — URL path
- **request.headers['X-Custom']** — Header values
- **inIpRange(origin.ip, '1.2.3.0/24')** — IP in CIDR

Example: `origin.region_code == 'US'` for geo-allowlist; `request.path.matches('/api/.*')` for path-based rules.

IP-based and CEL are mutually exclusive per rule.

---

## Rate Limiting: Throttle vs Rate-Based Ban

### Throttle

When traffic exceeds the threshold, the configured `exceed_action` is applied (e.g., deny(429), redirect). Below the threshold, traffic is allowed.

### Rate-Based Ban

Two thresholds:

1. **rate_limit_threshold** — When exceeded, apply `exceed_action`
2. **ban_threshold** — When exceeded again, ban the source for `ban_duration_sec` seconds

Useful for escalating from throttle to full block.

### enforce_on_key Options

Determines how requests are grouped for counting:

| Key | Behavior |
|-----|----------|
| `ALL` | Single counter for all matched traffic |
| `IP` | Per source IP |
| `HTTP_HEADER` | Per value of a header (set `enforce_on_key_name`) |
| `XFF_IP` | Per IP from X-Forwarded-For |
| `HTTP_COOKIE` | Per cookie value |
| `HTTP_PATH` | Per URL path |
| `SNI` | Per TLS Server Name Indication |
| `REGION_CODE` | Per client country/region |

---

## Adaptive Protection (CAAP)

Cloud Armor Adaptive Protection (CAAP) uses machine learning to detect Layer 7 DDoS and anomalous traffic. When enabled:

- Traffic patterns are analyzed in real time
- Anomalies generate alerts
- Optional auto-mitigation creates adaptive rules

**rule_visibility**: `STANDARD` (default) or `PREMIUM` (requires Managed Protection Plus).

---

## Preconfigured WAF Rules (OWASP ModSecurity Core Rule Set)

Cloud Armor includes preconfigured WAF rules based on the OWASP ModSecurity Core Rule Set. Common rule set identifiers:

- **sqli-v33-stable** — SQL injection
- **xss-v33-stable** — Cross-site scripting
- **rce-v33-stable** — Remote code execution
- **lfi-v33-stable** — Local file inclusion

Without exclusions, these rules can cause false positives (e.g., SQL-like content in search, HTML in rich text, GraphQL syntax).

---

## WAF Exclusions and False Positive Handling

WAF exclusions carve out specific request components from rule evaluation:

- **target_rule_set** — Which OWASP rule set to exclude from
- **target_rule_ids** — Optional list of specific rule IDs
- **request_headers** — Exclude headers from WAF inspection
- **request_cookies** — Exclude cookies
- **request_uris** — Exclude URI paths
- **request_query_params** — Exclude query parameters

Each exclusion field uses **operator** (EQUALS, STARTS_WITH, ENDS_WITH, CONTAINS, EQUALS_ANY) and **value**.

Example: Exclude `?search=` from SQLi rules so search boxes with "SELECT" are not blocked.

---

## Advanced Options

### JSON Parsing

- **DISABLED** — No JSON body inspection
- **STANDARD** — Parse JSON for WAF rules
- **STANDARD_WITH_GRAPHQL** — Parse JSON and GraphQL

Needed when WAF rules inspect request bodies.

### Logging

- **NORMAL** — Standard logging
- **VERBOSE** — Matched rule and request details

### IP Resolution

**user_ip_request_headers** — Custom headers to use for client IP when traffic passes through a CDN or proxy (e.g., `X-Forwarded-For`, `CF-Connecting-IP`).

### Request Body Inspection Size

Maximum size of request body to inspect (8KB, 16KB, 32KB, 48KB, 64KB). Larger values increase inspection coverage but may add latency. Supported in Pulumi; not in Terraform.

---

## 80/20 Scoping Rationale

### What Was Included

The component covers the majority of production use cases:

1. **Core policy** — Name, type, description
2. **Rules** — All actions (allow, deny, redirect, throttle, rate_based_ban)
3. **Match** — IP-based and CEL expressions
4. **Rate limiting** — Full throttle and rate-based ban with thresholds, ban duration, enforce-on-key, exceed redirect
5. **Redirect** — EXTERNAL_302 and GOOGLE_RECAPTCHA
6. **Header injection** — Custom headers for matching requests
7. **Preconfigured WAF** — Exclusions for all four field types (headers, cookies, URIs, query params)
8. **Adaptive Protection** — Layer 7 DDoS defense and rule visibility
9. **Advanced options** — JSON parsing, log level, user IP headers, request body inspection size (Pulumi)
10. **Preview mode** — Log without enforcing
11. **Labels** — Via Pulumi (Terraform does not support labels on security policies)
12. **Default rule** — Auto-added at 2147483647 if not specified

### What Was Excluded (Lower Priority)

- **Custom WAF rules** — ModSecurity rules beyond preconfigured sets; advanced use case
- **Threshold configs** — Fine-grained threshold overrides; niche
- **enforce_on_key_configs** — Complex key configuration; uncommon
- **recaptcha_options_config** — reCAPTCHA Enterprise integration; separate product surface
- **json_custom_config** — Custom JSON parsing; rarely needed

---

## Deliberate Exclusions

The following GCP Cloud Armor features are intentionally not exposed in the spec:

### recaptcha_options_config

reCAPTCHA Enterprise integration for key configuration and site keys. Requires reCAPTCHA Enterprise setup. Handled separately from the base policy; users can extend the component if needed.

### threshold_configs

Per-rule or per-policy threshold overrides for rate limiting. Most users rely on the standard `rate_limit_threshold` and `ban_threshold`. Advanced tuning is possible via raw provider if required.

### enforce_on_key_configs

Complex configuration for custom enforce-on-key behavior. The flat `enforce_on_key` and `enforce_on_key_name` cover the common cases (IP, header, cookie, path).

### json_custom_config

Custom JSON parsing configuration for non-standard JSON structures. `STANDARD` and `STANDARD_WITH_GRAPHQL` cover the vast majority of APIs. Custom configs are rarely needed.

---

## References

- [Cloud Armor overview](https://cloud.google.com/armor/docs/overview)
- [Security policies](https://cloud.google.com/armor/docs/security-policy-overview)
- [Preconfigured WAF rules](https://cloud.google.com/armor/docs/waf-rules)
- [CEL expressions](https://cloud.google.com/armor/docs/rules-language-reference)
- [Adaptive Protection](https://cloud.google.com/armor/docs/adaptive-protection-overview)
