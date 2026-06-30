# Managed Rules Basic

This preset creates a Web ACL with two AWS Managed Rule Groups that cover the most common web application threats. It allows all traffic by default and relies on the managed rules to block known bad patterns.

## When to Use

- Quick baseline protection for any web application
- Starting point before adding custom rules
- Development and staging environments
- Applications that need protection without deep customization

## Key Configuration Choices

- **REGIONAL scope** -- protects ALB, API Gateway, AppSync, Cognito, or App Runner
- **Allow default action** -- permissive baseline; managed rules block known threats
- **AWSManagedRulesCommonRuleSet** -- protects against XSS, file inclusion, log injection, and other OWASP Top 10 threats
- **AWSManagedRulesKnownBadInputsRuleSet** -- blocks requests with known bad patterns (Log4j, Java deserialization, command injection)
- **override_action: none** -- uses the rule groups' own block/count actions

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<acl-name>` | Unique name for the Web ACL (lowercase, alphanumeric, hyphens) |

## Common Additions

- Add `AWSManagedRulesSQLiRuleSet` for SQL injection protection
- Add `AWSManagedRulesAmazonIpReputationList` for IP reputation filtering
- Add a rate-based rule to prevent volumetric attacks
- Add logging with a CloudWatch Logs destination
- Use `ruleActionOverrides` to set specific noisy rules to `count` during initial deployment

## Related Presets

- **02-rate-limiting-with-managed-rules** -- adds rate limiting on top of managed rules
- **03-production-web-app** -- full production configuration with 5 managed groups, rate limiting, geo blocking, and logging
