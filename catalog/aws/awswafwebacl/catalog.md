# AWS WAF Web ACL

Deploys a WAFv2 Web ACL — the ordered rule set that decides which requests reach your application. Rules evaluate managed rule groups, rate limits, geo and ASN matches, IP-set and regex-pattern-set references, and match statements (byte / SQLi / XSS / size / label), with AND / OR / NOT composition for multi-condition logic. REGIONAL scope protects ALB, API Gateway, AppSync, Cognito, and App Runner; CLOUDFRONT scope protects CloudFront distributions and must live in `us-east-1`. Optional logging, custom block responses, CAPTCHA/challenge immunity, body-inspection ceilings, and field-masking data protection round out a production posture. Its `web_acl_arn` output is what CloudFront and load balancers associate via ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **WAFv2 Web ACL** — the named access-control list with its default action, visibility metrics, and sampled-request posture
- **Rules** — an ordered priority list evaluated top-down; each rule carries either a match `action` or a group `override_action`, plus optional per-rule CAPTCHA/challenge immunity and custom responses
- **Custom response bodies** — reusable HTML / plain-text / JSON templates referenced by block actions (the ACL's own response namespace)
- **Logging configuration** — created only when `logging` is set; ships request logs to CloudWatch Logs, S3, or Kinesis Firehose (destination name must start with `aws-waf-logs-`)
- **Association and data-protection policies** — optional body-inspection size ceilings per protected-resource type, and field-masking rules that apply to every WAF output
- **AWS tags** — organization, environment, resource kind, and resource ID applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Region selection** — CLOUDFRONT-scoped web ACLs must be created in `us-east-1`. REGIONAL web ACLs live in the same region as the protected resource.
- **A logging destination** (optional) — a CloudWatch Log Group, S3 bucket, or Kinesis Firehose delivery stream whose name starts with `aws-waf-logs-` (AWS enforces the prefix).
- **IP sets and regex pattern sets** (optional) — referenceable Planton kinds (`AwsWafIpSet`, `AwsWafRegexPatternSet`) or literal ARNs in the same region and scope as the web ACL.

## Deploy

### Console

Open the deployment store, find **AWS WAF Web ACL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the six-step authoring flow — scope & posture first, then the response-body namespace (so rule block actions can pick keys), then the protection rules, then token/challenge, inspection/data protection, and logging. Start from the **Managed Rules Basic** preset in the [Presets](#presets) tab to pre-populate the two most common AWS managed groups.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafWebAcl
metadata:
  name: api-protection
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  scope: REGIONAL
  defaultAction:
    type: allow
  rules:
    - name: aws-managed-common
      priority: 1
      overrideAction: none
      statement:
        managedRuleGroup:
          name: AWSManagedRulesCommonRuleSet
          vendorName: AWS
```

```shell
planton apply -f waf-web-acl.yaml
```

This creates a REGIONAL web ACL with the AWS Common Rule Set enforced and all other traffic allowed. An Infra Job tracks the provisioning in real time.

### InfraChart

When the web ACL deploys alongside its IP sets and pattern sets in one chart, wire the rule references via ValueFromRef:

```yaml
spec:
  region: us-west-2
  scope: REGIONAL
  defaultAction:
    type: block
  rules:
    - name: allow-office
      priority: 0
      action: allow
      statement:
        ipSetReference:
          arn:
            valueFrom:
              kind: AwsWafIpSet
              name: office-allowlist
              fieldPath: status.outputs.ip_set_arn
```

The InfraPipeline resolves the dependency graph, deploys the sets first, then provisions the web ACL with the resolved ARNs. Downstream, a CloudFront distribution or load balancer binds this ACL's own `web_acl_arn` output the same way.

## Key Configuration

These are the most important decisions when configuring a WAF Web ACL. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** — `REGIONAL` protects in-region resources; `CLOUDFRONT` protects distributions and pins the region to `us-east-1`. Scope is create-only: changing it replaces the ACL and every association.

**Default action** — `allow` when rules block known threats; `block` when rules allow known-good traffic. A default-block ACL that ships with no allow rules is an outage — the wizard surfaces that callout before you continue.

**Protection rules** — managed groups, rate limits, geo/ASN matches, IP-set and regex-set references, match statements, and one level of AND/OR/NOT composition. Group rules take `overrideAction` (`none` / `count`); match rules take `action` (`allow` / `block` / `count` / `captcha` / `challenge`). Deeper nesting is preserved on CLI-authored statements and editable via the custom JSON escape hatch.

**Custom response bodies** — declare named bodies before the rules that reference them. Block actions pick a body key from the declared namespace so a dangling key cannot be typed.

**Logging** — destination may be CloudWatch Logs, S3, or Firehose; the name must start with `aws-waf-logs-`, and AWS allows exactly one destination per web ACL. Redact Authorization and Cookie headers in production (URI path, query string, and method can be redacted too). A logging filter keeps or drops records by the action WAF applied or by request labels — keeping only BLOCK/COUNT records is the standard cost control for high-traffic ACLs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsWafIpSet** | `rules[].statement.ipSetReference.arn` | `status.outputs.ip_set_arn` |
| **AwsWafRegexPatternSet** | `rules[].statement.regexPatternSetReference.arn` | `status.outputs.regex_pattern_set_arn` |
| **CloudWatch Log Group / S3 / Firehose** | `logging.destinationArn` | literal ARN or explicit-kind reference (three destination types, so no default kind) |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `capacity` | Total WCUs consumed by all rules (5,000 per-ACL default budget) | Capacity planning before adding rules |
| `web_acl_arn` | Amazon Resource Name of the Web ACL | CloudFront `webAclArn`, ALB / API Gateway association |
| `web_acl_id` | Unique identifier of the Web ACL | AWS console navigation, API calls |
| `web_acl_name` | Name of the Web ACL | CloudWatch metric filtering, inventory |
| `application_integration_url` | SDK integration endpoint | AWS WAF JavaScript SDK for CAPTCHA / challenge |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Managed rules basic** — A permissive web ACL with the AWS Common Rule Set. Suitable for getting started on public-facing applications. Start from the **Managed Rules Basic** preset.

**Rate limiting with managed rules** — Combines managed groups with a rate-based rule. Suitable for APIs and login pages. Start from the **Rate Limiting with Managed Rules** preset.

**Production web app** — Multiple managed groups (common rules, known bad inputs, IP reputation, anonymous IP), rate limiting, and logging. Start from the **Production Web Application** preset.

**Cost-controlled logging** — Full protection with a logging filter that keeps only BLOCK and COUNT records and redacts Authorization and Cookie headers — the standard way to keep WAF log volume affordable on high-traffic ACLs. Start from the **Cost-Controlled Logging** preset.

## Works With

- [**AWS WAF IP Set**](/infra-catalog/aws-waf-ip-set) — allow / deny lists referenced by `ipSetReference` rules
- [**AWS WAF Regex Pattern Set**](/infra-catalog/aws-waf-regex-pattern-set) — reusable regex catalogs for path and header matching
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) — associates via `webAclArn`
- [**AWS ALB**](/infra-catalog/aws-alb) / API Gateway / App Runner — associate the ARN after deploy
