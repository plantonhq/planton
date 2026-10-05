# AWS WAF IP Set

Deploys a WAFv2 IP set — a named, reusable collection of IP addresses and CIDR ranges that web ACL rules match requests against. IP sets are the building block of IP-based filtering: allow-lists (office and VPN egress, partner integrations, health-checker fleets) and deny-lists (known-bad ranges, abusive clients). One set can back many rules across many web ACLs — update the set once and every referencing rule sees the change immediately, with no web ACL redeploy. Its `ip_set_arn` output is what every web ACL `ip_set_reference` statement binds via ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **WAFv2 IP Set** -- the named address collection in the chosen scope (REGIONAL or CLOUDFRONT). The set name comes from `metadata.name`; scope and IP address family are create-time immutable, and the address list itself updates in place

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **The CIDR ranges to hold** -- WAF accepts only CIDR notation, never bare addresses: a single IPv4 host is `192.0.2.44/32`, a single IPv6 host is `2001:db8::1/128`. An empty set is valid (a placeholder that rules can reference before the ranges are known) — it matches nothing.
- **No pre-existing resources required** -- the set is a leaf: it references nothing and filters nothing until a web ACL rule binds it.

## Deploy

### Console

Open the deployment store, find **AWS WAF IP Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the scope choice pins the region automatically for CloudFront, and each CIDR entry is validated as you type. Start from the **Office Allow-List** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafIpSet
metadata:
  name: office-allowlist
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  scope: REGIONAL
  ipAddressVersion: IPV4
  addresses:
    - 203.0.113.0/24
    - 198.51.100.44/32
  description: Corporate office and VPN egress ranges
```

```shell
planton apply -f waf-ip-set.yaml
```

This publishes the allow-list; pair it with a web ACL whose default action is block and an early-priority allow rule referencing this set's ARN. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an IP set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope decides the set's universe** -- WAF keeps REGIONAL resources (protecting ALBs, API Gateway, AppSync, Cognito, App Runner, Verified Access) and CLOUDFRONT resources (protecting distributions) strictly separate. A web ACL can only reference sets of its own scope, and scope is create-time immutable. CloudFront-scoped sets live in `us-east-1` — the WAF global region — regardless of where viewers are.

**One address family per set** -- `ipAddressVersion` is IPV4 or IPV6, forever (create-time immutable, and AWS refuses to delete a set that a rule still references, so a replace while referenced fails). Dual-stack coverage uses two sets — one per family — referenced by two rules or one rule with an OR statement. Enabling IPv6 on a load balancer WITHOUT an IPv6 set on the allow rule means IPv6 clients bypass the list silently.

**The action lives on the rule, not the set** -- the same set can back an allow rule in one web ACL and a block rule in another. The set only answers "does this request's source IP match?"

**Empty on purpose is a real pattern** -- deploy a placeholder set, wire web ACL rules to its ARN, and fill in the ranges when they are known. An empty set matches nothing, so an allow rule over it allows nobody and a block rule blocks nobody.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The set is a leaf — it references no other Infra Components; web ACLs reference it, never the reverse.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ip_set_arn` | Amazon Resource Name of the IP set | AwsWafWebAcl `ip_set_reference` rule statements |
| `ip_set_id` | AWS-assigned set ID (UUID) | Direct WAFv2 API calls together with name and scope |
| `ip_set_name` | The set name as created in AWS | WAF console URLs and CLI commands |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Office allow-list** -- a REGIONAL IPv4 set of corporate and VPN egress ranges, referenced by an early-priority allow rule in a default-block web ACL — the standard shape for gating private APIs and staging environments. Start from the **Office Allow-List** preset.

**Placeholder set** -- an empty set deployed so web ACL rules can bind its ARN before NetEng publishes the real ranges; filling it in later never touches the web ACLs. Start from the **Placeholder IP Set** preset.

## Works With

- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) -- references this set through `ip_set_reference` rule statements; the rule's action (allow, block, count, CAPTCHA) decides what a match means
- [**AWS WAF Regex Pattern Set**](/infra-catalog/aws-waf-regex-pattern-set) -- the sibling reusable-collection kind for pattern matching instead of source-IP matching
- [**AWS ALB**](/infra-catalog/aws-alb) -- the most common REGIONAL association target of the web ACLs that consume this set
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) -- the association target of CLOUDFRONT-scoped web ACLs (its `webAclArn` binds the web ACL)
