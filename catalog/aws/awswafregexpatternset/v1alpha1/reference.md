# AwsWafRegexPatternSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsWafRegexPatternSetSpec defines an AWS WAFv2 regex pattern set — a named,
reusable collection of regular expressions that web ACL rules match request
components against (URI paths, headers, query strings, bodies). Pattern
sets centralize regexes that many rules share: update the set once and
every referencing rule sees the change immediately, with no web ACL
redeploy.

A web ACL references the set through a regex_pattern_set_reference
statement using the set's ARN (exported as the regex_pattern_set_arn
output); the statement matches when ANY regex in the set matches the
inspected component. For a one-off regex that no other rule shares, the
web ACL's inline regex_match statement is the simpler choice — reach for a
pattern set when the same expressions back multiple rules or need
independent ownership.

scope is create-time immutable (ForceNew). AWS supports a subset of PCRE —
notably NO backreferences or lookaround assertions; patterns using them
are rejected by AWS at create time.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafRegexPatternSet
metadata:
  name: blocked-path-probes
  org: acme
  env: dev
spec:
  region: us-west-2
  scope: REGIONAL
  regularExpressions:
    - ^/wp-admin/.*
    - ^/\.env$
    - ^/phpmyadmin/.*
  description: Common scanner probe paths
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.scope` | `string` | yes |  |  |
| `spec.regularExpressions` | `[]string` | yes |  |  |
| `spec.description` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the pattern set is created. Must match the scope of
the web ACLs that will reference it: a REGIONAL set lives in the same
region as the resources it protects, while a CLOUDFRONT set must be
created in us-east-1 (the WAF global region).
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.scope

`string` · required

Whether the set is usable by REGIONAL web ACLs (protecting ALBs, API
Gateway, AppSync, Cognito, App Runner, Verified Access) or CLOUDFRONT
web ACLs (protecting CloudFront distributions). Create-time immutable
(ForceNew). A web ACL can only reference sets of its own scope.

- rule: {"required":true,"string":{"in":["REGIONAL","CLOUDFRONT"]}}

### spec.regularExpressions

`[]string` · required

The regular expressions in the set, each up to 200 characters. AWS's
default quota is 10 expressions per set (adjustable via Service Quotas).
AWS supports a subset of PCRE — no backreferences (\1), no lookaround
((?=...)/(?<=...)); such patterns fail at AWS create time. At least one
expression is required — unlike IP sets, AWS rejects an empty pattern
set.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1","maxLen":"200"}}}}

### spec.description

`string`

Description of what the patterns match and who maintains them. AWS
restricts the character set: letters, digits, whitespace, and
_ + = : # @ / - , . only (notably NO parentheses), 3-256 characters —
WAF rejects anything else at create time, so the constraint is enforced
here where the failure is immediate and readable.

- rule: description may only contain letters, digits, whitespace, and _+=:#@/-,. (no parentheses), and must be at least 3 characters when set
- rule: {"string":{"maxLen":"256"}}

## Validation Rules

- `cloudfront_scope_requires_us_east_1`: CloudFront-scoped WAF resources live in the global (us-east-1) region — set region to 'us-east-1' when scope is 'CLOUDFRONT'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsWafRegexPatternSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.regex_pattern_set_arn` | `string` | The Amazon Resource Name of the regex pattern set (arn:aws:wafv2:<region>:<account>:<scope>/regexpatternset/<name>/<id>). The identifier web ACL rules reference. |
| `status.outputs.regex_pattern_set_id` | `string` | The AWS-assigned pattern set ID (a UUID). Used together with the name and scope when addressing the set through the WAFv2 API directly. |
| `status.outputs.regex_pattern_set_name` | `string` | The pattern set name as created in AWS (derived from metadata.name). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsWafWebAcl | `spec.rules[].statement.regexPatternSetReference.arn` | `status.outputs.regex_pattern_set_arn` |

## See Also

- [Overview](../README.md)
