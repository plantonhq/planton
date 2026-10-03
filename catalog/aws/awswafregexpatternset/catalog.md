# AWS WAF Regex Pattern Set

Deploys a WAFv2 regex pattern set — a named, reusable collection of regular expressions that web ACL rules match request components against: URI paths, headers, query strings, bodies. Pattern sets centralize the expressions many rules share (scanner probes, banned paths, known-bad user agents) with independent ownership: AppSec maintains the patterns, application teams reference them, and updating the set once propagates to every referencing rule immediately, with no web ACL redeploy. Its `regex_pattern_set_arn` output is what every web ACL `regex_pattern_set_reference` statement binds via ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **WAFv2 Regex Pattern Set** -- the named expression collection in the chosen scope (REGIONAL or CLOUDFRONT). The set name comes from `metadata.name`; scope is create-time immutable, and the expressions update in place

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **The expressions to hold** -- at least one is required (AWS rejects an empty pattern set, unlike an IP set), each up to 200 characters. WAF's engine runs a PCRE subset: no backreferences (`\1`), no lookaround (`(?=…)`, `(?<=…)`) — such patterns fail at AWS create time.
- **No pre-existing resources required** -- the set is a leaf: it references nothing and matches nothing until a web ACL rule binds it.

## Deploy

### Console

Open the deployment store, find **AWS WAF Regex Pattern Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the scope choice pins the region automatically for CloudFront, and the expressions step teaches the PCRE-subset limits in place. Start from the **Scanner Probe Patterns** preset in the [Presets](#presets) tab for the classic block-list shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsWafRegexPatternSet
metadata:
  name: scanner-probes
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  scope: REGIONAL
  regularExpressions:
    - '^/(wp-admin|wp-login|xmlrpc)\.php'
    - '^/\.env'
  description: Known scanner probe paths - maintained by AppSec
```

```shell
planton apply -f waf-regex-pattern-set.yaml
```

This publishes the probe list; pair it with a web ACL block rule whose `regex_pattern_set_reference` statement inspects the URI path. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a pattern set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope decides the set's universe** -- WAF keeps REGIONAL resources (protecting ALBs, API Gateway, AppSync, Cognito, App Runner, Verified Access) and CLOUDFRONT resources (protecting distributions) strictly separate. A web ACL can only reference sets of its own scope, and scope is create-time immutable. CloudFront-scoped sets live in `us-east-1` — the WAF global region.

**ANY-match semantics** -- a referencing rule picks ONE request component and matches when ANY expression in the set matches it. Sets are OR-lists by construction; long alternations split across entries for free.

**The PCRE subset is real** -- no backreferences, no lookaround. AWS rejects such patterns at create time with a generic error; rewrite with alternation and explicit character classes. Expressions cap at 200 characters each, and the default quota is 10 expressions per set (adjustable via Service Quotas).

**Set vs inline regex** -- the web ACL also offers a one-off `regex_match` statement. Reach for a pattern set when the same expressions back multiple rules or need independent ownership; keep a single-rule, single-use regex inline.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The set is a leaf — it references no other Infra Components; web ACLs reference it, never the reverse.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `regex_pattern_set_arn` | Amazon Resource Name of the pattern set | AwsWafWebAcl `regex_pattern_set_reference` rule statements |
| `regex_pattern_set_id` | AWS-assigned set ID (UUID) | Direct WAFv2 API calls together with name and scope |
| `regex_pattern_set_name` | The set name as created in AWS | WAF console URLs and CLI commands |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scanner-probe block-list** -- expressions catching WordPress, phpMyAdmin, and dotfile probes on the URI path, referenced by a block rule — the classic first pattern set in any estate. Start from the **Scanner Probe Patterns** preset.

**Internal path gate** -- expressions matching admin and internal route prefixes, referenced by a rule that blocks (or CAPTCHA-challenges) requests from outside the office allow-list. Start from the **Internal Admin Path Patterns** preset.

## Works With

- [**AWS WAF Web ACL**](/infra-catalog/aws-waf-web-acl) -- references this set through `regex_pattern_set_reference` rule statements; the statement picks the request component and text transformations
- [**AWS WAF IP Set**](/infra-catalog/aws-waf-ip-set) -- the sibling reusable-collection kind for source-IP matching instead of pattern matching
- [**AWS ALB**](/infra-catalog/aws-alb) -- the most common REGIONAL association target of the web ACLs that consume this set
- [**AWS CloudFront**](/infra-catalog/aws-cloud-front) -- the association target of CLOUDFRONT-scoped web ACLs (its `webAclArn` binds the web ACL)
