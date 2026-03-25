# CloudflareRuleset Examples

## Origin Rule — Split Traffic Between Two Origins

Route non-marketing paths to a Kubernetes backend while the default origin (GitHub Pages) serves static content:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: planton-origin-routing
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_origin
  name: "Route app traffic to K8s"
  rules:
    - ref: "route-app-to-k8s"
      expression: >-
        not (
          http.request.uri.path eq "/" or
          http.request.uri.path starts_with "/docs" or
          http.request.uri.path starts_with "/features" or
          http.request.uri.path starts_with "/pricing" or
          http.request.uri.path starts_with "/changelog" or
          http.request.uri.path starts_with "/_site"
        )
      action: route
      description: "Route non-marketing paths to K8s origin"
      actionParameters:
        hostHeader: "planton.ai"
        origin:
          host: "<k8s-lb-hostname>"
          port: 443
```

## Origin Rule — Zone ID from CloudflareDnsZone Reference

Use `valueFrom` to wire the zone ID from an existing CloudflareDnsZone resource:

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: my-origin-rule
spec:
  zoneId:
    valueFrom:
      name: my-dns-zone
  rulesetKind: zone
  phase: http_request_origin
  name: "Route API traffic"
  rules:
    - ref: "api-origin"
      expression: 'http.request.uri.path starts_with "/api"'
      action: route
      description: "Route API calls to backend"
      actionParameters:
        origin:
          host: "api-backend.internal.example.com"
          port: 8443
        sni:
          value: "api-backend.internal.example.com"
```

## WAF Custom Rule — Block by IP

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: block-bad-actors
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_firewall_custom
  name: "Security rules"
  rules:
    - ref: "block-known-attacker"
      expression: 'ip.src in {192.0.2.0/24 198.51.100.0/24}'
      action: block
      description: "Block known malicious IP ranges"
      actionParameters:
        response:
          statusCode: 403
          content: "Access denied"
          contentType: "text/plain"
    - ref: "challenge-high-threat"
      expression: 'cf.threat_score > 50'
      action: managed_challenge
      description: "Challenge high-threat-score requests"
```

## WAF Managed Ruleset — Execute OWASP Core Ruleset

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: managed-waf
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_firewall_managed
  name: "Managed WAF rules"
  rules:
    - ref: "owasp-core"
      expression: "true"
      action: execute
      description: "Execute Cloudflare Managed Ruleset"
      actionParameters:
        id: "efb7b8c949ac4650a09736fc376e9aee"
        overrides:
          categories:
            - category: "xss"
              action: "block"
              enabled: true
            - category: "sqli"
              action: "block"
              enabled: true
          rules:
            - id: "<specific-rule-id>"
              action: "log"
              enabled: true
```

## Cache Rule — Override Caching for Static Assets

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: custom-cache-rules
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_cache_settings
  name: "Custom cache settings"
  rules:
    - ref: "cache-static-assets"
      expression: 'http.request.uri.path starts_with "/assets" or http.request.uri.path starts_with "/_next/static"'
      action: set_cache_settings
      description: "Aggressively cache static assets"
      actionParameters:
        cache: true
        edgeTtl:
          mode: "override_origin"
          defaultTtl: 86400
          statusCodeTtls:
            - statusCode: 404
              value: 60
        browserTtl:
          mode: "override_origin"
          defaultTtl: 3600
    - ref: "bypass-api-cache"
      expression: 'http.request.uri.path starts_with "/api"'
      action: set_cache_settings
      description: "Never cache API responses"
      actionParameters:
        cache: false
```

## Redirect Rule — Dynamic Redirects

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: redirects
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_dynamic_redirect
  name: "URL redirects"
  rules:
    - ref: "old-docs-redirect"
      expression: 'http.request.uri.path starts_with "/old-docs"'
      action: redirect
      description: "Redirect old docs to new location"
      actionParameters:
        fromValue:
          targetUrl:
            expression: 'concat("https://docs.example.com", http.request.uri.path)'
          statusCode: 301
          preserveQueryString: true
```

## URL Rewrite Rule — Request Transform

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: url-rewrites
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_transform
  name: "URL rewrites"
  rules:
    - ref: "rewrite-api-path"
      expression: 'http.request.uri.path starts_with "/v2/api"'
      action: rewrite
      description: "Strip /v2 prefix from API paths"
      actionParameters:
        uri:
          path:
            expression: 'regex_replace(http.request.uri.path, "^/v2", "")'
```

## Response Header Transform

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: response-headers
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_response_headers_transform
  name: "Security response headers"
  rules:
    - ref: "security-headers"
      expression: "true"
      action: rewrite
      description: "Add security headers to all responses"
      actionParameters:
        headers:
          X-Content-Type-Options:
            operation: "set"
            value: "nosniff"
          X-Frame-Options:
            operation: "set"
            value: "DENY"
          Strict-Transport-Security:
            operation: "set"
            value: "max-age=31536000; includeSubDomains"
```

## Skip Rule — Bypass WAF for Trusted Path

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareRuleset
metadata:
  name: waf-exceptions
spec:
  zoneId:
    value: "<cloudflare-zone-id>"
  rulesetKind: zone
  phase: http_request_firewall_custom
  name: "WAF exceptions"
  rules:
    - ref: "skip-waf-for-health"
      expression: 'http.request.uri.path eq "/health" and ip.src in {10.0.0.0/8}'
      action: skip
      description: "Skip WAF for internal health checks"
      actionParameters:
        phases:
          - "http_request_firewall_managed"
        products:
          - "waf"
```

## Pulumi (Go)

```go
_, err := cloudflare.NewRuleset(ctx, "origin-rule", &cloudflare.RulesetArgs{
    ZoneId: pulumi.String(zoneId),
    Kind:   pulumi.String("zone"),
    Phase:  pulumi.String("http_request_origin"),
    Name:   pulumi.String("Route app traffic"),
    Rules: cloudflare.RulesetRuleArray{
        &cloudflare.RulesetRuleArgs{
            Ref:        pulumi.String("route-to-k8s"),
            Expression: pulumi.String(`not http.request.uri.path starts_with "/docs"`),
            Action:     pulumi.String("route"),
            ActionParameters: &cloudflare.RulesetRuleActionParametersArgs{
                HostHeader: pulumi.String("planton.ai"),
                Origin: &cloudflare.RulesetRuleActionParametersOriginArgs{
                    Host: pulumi.String("k8s-lb.example.com"),
                    Port: pulumi.Int(443),
                },
            },
        },
    },
})
```

## Terraform

```hcl
resource "cloudflare_ruleset" "origin_rule" {
  zone_id = var.zone_id
  kind    = "zone"
  phase   = "http_request_origin"
  name    = "Route app traffic"

  rules {
    ref         = "route-to-k8s"
    expression  = "not http.request.uri.path starts_with \"/docs\""
    action      = "route"
    description = "Route non-marketing paths to K8s"

    action_parameters {
      host_header = "planton.ai"
      origin {
        host = "k8s-lb.example.com"
        port = 443
      }
    }
  }
}
```
