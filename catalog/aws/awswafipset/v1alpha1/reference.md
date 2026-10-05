# AwsWafIpSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsWafIpSetSpec defines an AWS WAFv2 IP set — a named, reusable collection
of IP addresses and CIDR ranges that web ACL rules match requests against.
IP sets are the building block for allow-lists (office/VPN ranges, partner
integrations, health-checker fleets) and deny-lists (known-bad ranges,
abusive clients), and one set can be referenced by many rules across many
web ACLs — update the set once and every referencing rule sees the change
immediately, with no web ACL redeploy.

A web ACL references the set through an ip_set_reference statement using
the set's ARN (exported as the ip_set_arn output). The action —
allow, block, count, CAPTCHA — lives on the referencing RULE, not on the
set, so the same set can back an allow rule in one web ACL and a block
rule in another.

scope and ip_address_version are create-time immutable (ForceNew):
changing either replaces the set (and AWS refuses to delete a set that is
still referenced by a rule, so a replace while referenced fails — detach
first). The address list itself updates in place.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafIpSet
metadata:
  name: office-allowlist
  org: acme
  env: dev
spec:
  region: us-west-2
  scope: REGIONAL
  ipAddressVersion: IPV4
  addresses:
    - 203.0.113.0/24
    - 198.51.100.44/32
  description: Corporate office egress ranges
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.scope` | `string` | yes |  |  |
| `spec.ipAddressVersion` | `string` | yes |  |  |
| `spec.addresses` | `[]string` |  |  |  |
| `spec.description` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the IP set is created. Must match the scope of the
web ACLs that will reference it: a REGIONAL set lives in the same region
as the resources it protects (ALBs, API Gateway stages, Cognito user
pools), while a CLOUDFRONT set must be created in us-east-1 (the WAF
global region).
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.scope

`string` · required

Whether the set is usable by REGIONAL web ACLs (protecting ALBs, API
Gateway, AppSync, Cognito, App Runner, Verified Access) or CLOUDFRONT
web ACLs (protecting CloudFront distributions). Create-time immutable
(ForceNew). A web ACL can only reference sets of its own scope.

- rule: {"required":true,"string":{"in":["REGIONAL","CLOUDFRONT"]}}

### spec.ipAddressVersion

`string` · required

The IP address family this set holds: "IPV4" or "IPV6". Create-time
immutable (ForceNew). A set holds exactly one family — deployments that
match both families use two sets (one per family) referenced by two
rules or one rule with an OR statement.

- rule: {"required":true,"string":{"in":["IPV4","IPV6"]}}

### spec.addresses

`[]string`

The addresses in the set, in CIDR notation — AWS accepts ONLY CIDR
ranges, never bare addresses: a single IPv4 host is "192.0.2.44/32"
(not "192.0.2.44"), a single IPv6 host is "2001:db8::1/128". IPv4
supports /1 through /32; IPv6 supports /1 through /128. Up to 10,000
entries. An empty list is valid (a placeholder set that rules can
reference before the ranges are known) — an empty set matches nothing.

- rule: {"repeated":{"maxItems":"10000","items":{"cel":[{"id":"address_is_cidr","message":"each address must be in CIDR notation — write a single IPv4 host as x.x.x.x/32 and a single IPv6 host as ::x/128","expression":"this.matches('^.+/[0-9]{1,3}$')"}]}}}

### spec.description

`string`

Description of what the set represents and who maintains it. AWS
restricts the character set: letters, digits, whitespace, and
_ + = : # @ / - , . only (notably NO parentheses), 3-256 characters —
WAF rejects anything else at create time, so the constraint is enforced
here where the failure is immediate and readable.

- rule: description may only contain letters, digits, whitespace, and _+=:#@/-,. (no parentheses), and must be at least 3 characters when set
- rule: {"string":{"maxLen":"256"}}

## Validation Rules

- `cloudfront_scope_requires_us_east_1`: CloudFront-scoped WAF resources live in the global (us-east-1) region — set region to 'us-east-1' when scope is 'CLOUDFRONT'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsWafIpSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ip_set_arn` | `string` | The Amazon Resource Name of the IP set (arn:aws:wafv2:<region>:<account>:<scope>/ipset/<name>/<id>). The identifier web ACL rules reference. |
| `status.outputs.ip_set_id` | `string` | The AWS-assigned IP set ID (a UUID). Used together with the name and scope when addressing the set through the WAFv2 API directly. |
| `status.outputs.ip_set_name` | `string` | The IP set name as created in AWS (derived from metadata.name). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsWafWebAcl | `spec.rules[].statement.ipSetReference.arn` | `status.outputs.ip_set_arn` |

## See Also

- [Overview](../README.md)
