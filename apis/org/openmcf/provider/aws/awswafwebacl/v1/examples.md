# AwsWafWebAcl Examples

## 1. Minimal — Allow All with Managed Rules

The simplest useful configuration: allow all traffic, block known threats via managed rules.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: basic-waf
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: aws-common-rules
      priority: 1
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesCommonRuleSet
        vendorName: AWS
```

## 2. Rate Limiting

Block IPs that exceed 2,000 requests in 5 minutes.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: rate-limited-api
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: rate-limit-2k
      priority: 1
      action: block
      rateBased:
        limit: 2000
```

## 3. Geographic Blocking

Block traffic from specific countries.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: geo-restricted
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: block-high-risk-countries
      priority: 1
      action: block
      geoMatch:
        countryCodes:
          - RU
          - CN
          - KP
```

## 4. IP Blocklist

Block known malicious IPs using a WAFv2 IP Set.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: ip-filtered
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: block-bad-ips
      priority: 1
      action: block
      ipSetReference:
        arn: arn:aws:wafv2:us-east-1:123456789012:regional/ipset/bad-actors/abc123
```

## 5. Custom Statement — SQL Injection Detection

Use the `customStatement` escape hatch for statement types not modeled as first-class messages.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: sqli-protected
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: block-sqli-body
      priority: 1
      action: block
      customStatement:
        SqliMatchStatement:
          FieldToMatch:
            Body: {}
          TextTransformations:
            - Priority: 0
              Type: URL_DECODE
            - Priority: 1
              Type: HTML_ENTITY_DECODE
```

## 6. Managed Rules with Tuning

Override specific noisy rules within a managed group to `count` for monitoring.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: tuned-waf
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: aws-common-tuned
      priority: 1
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesCommonRuleSet
        vendorName: AWS
        version: Version_1.0
        ruleActionOverrides:
          - name: SizeRestrictions_BODY
            action: count
          - name: GenericLFI_BODY
            action: count
```

## 7. Production-Ready with Logging

Comprehensive protection with rate limiting, managed rules, geo blocking, custom error responses, and logging.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: production-waf
spec:
  scope: REGIONAL
  description: Production Web ACL for public API
  defaultAction:
    type: allow
  rules:
    - name: rate-limit
      priority: 1
      action: block
      rateBased:
        limit: 3000
        evaluationWindowSec: 300
    - name: block-high-risk-geo
      priority: 2
      action: block
      geoMatch:
        countryCodes: [RU, CN, KP, IR]
    - name: aws-ip-reputation
      priority: 10
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesAmazonIpReputationList
        vendorName: AWS
    - name: aws-common-rules
      priority: 20
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesCommonRuleSet
        vendorName: AWS
    - name: aws-sqli-rules
      priority: 30
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesSQLiRuleSet
        vendorName: AWS
    - name: aws-known-bad-inputs
      priority: 40
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesKnownBadInputsRuleSet
        vendorName: AWS
  customResponseBodies:
    - key: rate-limited-json
      content: '{"error":"rate_limited","retryAfter":300}'
      contentType: APPLICATION_JSON
  logging:
    destinationArn:
      value: arn:aws:logs:us-east-1:123456789012:log-group:aws-waf-logs-production
    redactedHeaderNames:
      - authorization
      - cookie
      - x-api-key
    redactQueryString: true
```
