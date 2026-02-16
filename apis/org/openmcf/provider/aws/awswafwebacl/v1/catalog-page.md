# AWS WAF Web ACL

Deploys an AWS WAFv2 Web Access Control List with ordered rules for managed rule groups, rate limiting, geographic filtering, and IP-based access control. Includes optional request logging with field redaction. Rules are evaluated by priority; the first match takes action.

## What Gets Created

When you deploy an AwsWafWebAcl resource, OpenMCF provisions:

- **WAFv2 Web ACL** — an `aws_wafv2_web_acl` resource with the specified scope (REGIONAL or CLOUDFRONT), default action, visibility config, and rules passed via `rule_json`
- **Rules** — an ordered set of rules constructed from managed rule group references, rate-based limits, geographic match conditions, IP set references, and custom statement escape hatches
- **Custom Response Bodies** — reusable response templates referenced by block actions
- **Logging Configuration** — an `aws_wafv2_web_acl_logging_configuration` resource created only when `logging` is provided, with destination ARN and optional field redaction

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **Appropriate IAM permissions** for `wafv2:*` operations
- **us-east-1 provider region** if using CLOUDFRONT scope
- **WAFv2 IP Sets** created separately if using `ipSetReference` rules
- **A logging destination** named starting with `aws-waf-logs-` if enabling logging (CloudWatch Logs log group, S3 bucket, or Kinesis Firehose delivery stream)

## Quick Start

Create a file `waf.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: my-web-acl
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsWafWebAcl.my-web-acl
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

Deploy:

```shell
openmcf apply -f waf.yaml
```

This creates a REGIONAL Web ACL that allows all traffic by default and blocks known threats via the AWS Common Rule Set.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `scope` | `string` | Where the Web ACL can be used: `REGIONAL` or `CLOUDFRONT` | Must be one of the two valid values |
| `defaultAction.type` | `string` | Baseline action when no rule matches: `allow` or `block` | Must be `allow` or `block` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description. Max 256 characters. |
| `rules` | `object[]` | `[]` | Ordered rule set evaluated by priority. |
| `visibilityConfig.cloudwatchMetricsEnabled` | `bool` | `true` | Enable CloudWatch metrics for the Web ACL. |
| `visibilityConfig.sampledRequestsEnabled` | `bool` | `true` | Enable request sampling. |
| `visibilityConfig.metricName` | `string` | resource name | CloudWatch metric name. |
| `customResponseBodies` | `object[]` | `[]` | Reusable response body templates referenced by block actions. |
| `tokenDomains` | `string[]` | `[]` | Domains for CAPTCHA/Challenge token validation. |
| `logging.destinationArn` | `string` | — | ARN of CloudWatch Logs, S3, or Firehose destination. Supports `valueFrom`. |
| `logging.redactedHeaderNames` | `string[]` | `[]` | HTTP headers to redact from logs. |
| `logging.redactUriPath` | `bool` | `false` | Redact URI path from logs. |
| `logging.redactQueryString` | `bool` | `false` | Redact query string from logs. |

### Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Unique rule name (1-128 characters). |
| `priority` | `int` | Evaluation order (lower numbers first). |
| `action` | `string` | For custom rules: `allow`, `block`, `count`, `captcha`, or `challenge`. |
| `overrideAction` | `string` | For managed rule groups: `count` or `none`. |
| `managedRuleGroup.name` | `string` | Managed rule group name (e.g., `AWSManagedRulesCommonRuleSet`). |
| `managedRuleGroup.vendorName` | `string` | Vendor name (e.g., `AWS`). |
| `managedRuleGroup.version` | `string` | Pin to a specific version. |
| `managedRuleGroup.ruleActionOverrides` | `object[]` | Override actions for specific rules within the group. |
| `managedRuleGroup.scopeDownStatement` | `object` | Struct narrowing which requests the group evaluates. |
| `rateBased.limit` | `int` | Max requests per evaluation window (10–2,000,000,000). |
| `rateBased.evaluationWindowSec` | `int` | Window in seconds: 60, 120, 300, or 600. |
| `rateBased.aggregateKeyType` | `string` | `IP`, `FORWARDED_IP`, `CONSTANT`, or `CUSTOM_KEYS`. |
| `geoMatch.countryCodes` | `string[]` | ISO 3166-1 alpha-2 country codes. |
| `ipSetReference.arn` | `string` | ARN of a WAFv2 IP Set. |
| `customStatement` | `object` | Raw AWS WAFv2 JSON statement (escape hatch). |

## Examples

### Managed Rules with Rate Limiting

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: api-protection
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsWafWebAcl.api-protection
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: rate-limit
      priority: 1
      action: block
      rateBased:
        limit: 2000
    - name: aws-common-rules
      priority: 10
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesCommonRuleSet
        vendorName: AWS
    - name: aws-sqli-rules
      priority: 20
      overrideAction: none
      managedRuleGroup:
        name: AWSManagedRulesSQLiRuleSet
        vendorName: AWS
```

### Geographic Blocking with Custom Error Response

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: geo-restricted
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsWafWebAcl.geo-restricted
spec:
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: block-embargoed
      priority: 1
      action: block
      geoMatch:
        countryCodes: [RU, CN, KP, IR]
      customResponse:
        responseCode: 403
        customResponseBodyKey: geo-blocked
  customResponseBodies:
    - key: geo-blocked
      content: '{"error":"geo_restricted","message":"Service not available in your region"}'
      contentType: APPLICATION_JSON
```

### Production Web ACL with Logging

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsWafWebAcl
metadata:
  name: prod-waf
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsWafWebAcl.prod-waf
spec:
  scope: REGIONAL
  description: Production Web ACL
  defaultAction:
    type: allow
  rules:
    - name: rate-limit
      priority: 1
      action: block
      rateBased:
        limit: 3000
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
  logging:
    destinationArn:
      value: arn:aws:logs:us-east-1:123456789012:log-group:aws-waf-logs-prod
    redactedHeaderNames:
      - authorization
      - cookie
    redactQueryString: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `web_acl_arn` | `string` | ARN of the Web ACL, used to associate with ALB, API Gateway, CloudFront, and other protected resources |
| `web_acl_id` | `string` | Unique identifier of the Web ACL |
| `web_acl_name` | `string` | Name of the Web ACL |
| `capacity` | `int` | Web ACL Capacity Units consumed (maximum 5,000 per Web ACL) |

## Related Components

- [AwsAlb](/docs/catalog/aws/alb) — commonly protected by a WAF Web ACL via association
- [AwsHttpApiGateway](/docs/catalog/aws/http-api-gateway) — protectable via WAF association
- [AwsCloudFront](/docs/catalog/aws/cloudfront) — protected by CLOUDFRONT-scoped Web ACLs
- [AwsCognitoUserPool](/docs/catalog/aws/cognito-user-pool) — protectable via WAF association
