# GcpCloudArmorPolicy Examples

## 1. IP Allowlist with Default Deny

Allow only traffic from specified CIDR ranges; deny everything else.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: ip-allowlist-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges:
          - 192.168.1.0/24
          - 10.0.0.0/8
      description: Allow office and VPN IP ranges
    - priority: 2147483647
      action: deny(403)
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Default deny all other traffic
```

## 2. Geo-blocking with CEL Expression

Block traffic from specific regions using Common Expression Language.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: geo-block-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: deny(403)
      match:
        expression: "origin.region_code == 'RU' || origin.region_code == 'CN'"
      description: Block Russia and China
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Allow all other traffic
```

## 3. Rate Limiting per IP

Throttle requests when a single IP exceeds 100 requests per minute.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: rate-limit-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: throttle
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      rateLimitOptions:
        conformAction: allow
        exceedAction: deny(429)
        enforceOnKey: IP
        rateLimitThreshold:
          count: 100
          intervalSec: 60
      description: 100 req/min per IP
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Default allow
```

## 4. Rate-based Ban

Escalate from throttle to full ban when traffic exceeds a second threshold.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: rate-ban-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: rate_based_ban
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      rateLimitOptions:
        conformAction: allow
        exceedAction: deny(429)
        enforceOnKey: IP
        rateLimitThreshold:
          count: 100
          intervalSec: 60
        banThreshold:
          count: 200
          intervalSec: 60
        banDurationSec: 600
      description: Throttle at 100/min, ban at 200/min for 10 min
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Default allow
```

## 5. WAF OWASP Protection with Exclusions

Enable OWASP rules and exclude false positives for specific request fields.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: waf-owasp-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: deny(403)
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      preconfiguredWafConfig:
        exclusions:
          - targetRuleSet: sqli-v33-stable
            requestQueryParams:
              - operator: CONTAINS
                value: search
          - targetRuleSet: xss-v33-stable
            requestUris:
              - operator: STARTS_WITH
                value: /api/richtext/
      description: WAF with SQL/search and XSS/richtext exclusions
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Default allow
```

## 6. Redirect to reCAPTCHA Challenge

Send suspicious traffic to a Google reCAPTCHA challenge page.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: recaptcha-policy
spec:
  projectId:
    value: my-gcp-project
  rules:
    - priority: 1000
      action: redirect
      match:
        expression: "evaluateThreatIntelligence('bot-management-verification')"
      redirectOptions:
        type: GOOGLE_RECAPTCHA
      description: reCAPTCHA challenge for bot detection
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      description: Default allow
```

## 7. Full-featured Production WAF Policy

Combines multiple features: Adaptive Protection, JSON parsing, verbose logging, and a layered rule set.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpCloudArmorPolicy
metadata:
  name: prod-waf-policy
spec:
  projectId:
    value: my-gcp-project
  description: Production WAF for e-commerce load balancer
  type: CLOUD_ARMOR
  adaptiveProtectionConfig:
    enableLayer7DdosDefense: true
    ruleVisibility: STANDARD
  advancedOptionsConfig:
    jsonParsing: STANDARD_WITH_GRAPHQL
    logLevel: VERBOSE
  rules:
    - priority: 100
      action: deny(403)
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges:
          - 203.0.113.0/24  # Example blocklist - replace with actual bad-actor ranges
      description: Deny known bad actor IPs
      preview: true
    - priority: 500
      action: throttle
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      rateLimitOptions:
        conformAction: allow
        exceedAction: deny(429)
        enforceOnKey: IP
        rateLimitThreshold:
          count: 500
          intervalSec: 60
      description: 500 req/min per IP
    - priority: 1000
      action: deny(403)
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      preconfiguredWafConfig:
        exclusions:
          - targetRuleSet: sqli-v33-stable
            requestQueryParams:
              - operator: EQUALS
                value: q
      description: OWASP WAF with search param exclusion
    - priority: 2147483647
      action: allow
      match:
        versioned_expr: SRC_IPS_V1
        src_ip_ranges: ["*"]
      headerAction:
        requestHeadersToAdds:
          - headerName: X-CloudArmor-Processed
            headerValue: "true"
      description: Default allow with header injection
```
