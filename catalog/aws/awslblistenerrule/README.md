# Overview

The AwsLbListenerRule API resource provisions an ALB listener rule: a
condition-action pair evaluated in priority order before the listener's
default action, routing matching requests to a target group, a redirect, a
fixed response, or through an authentication step.

## Why We Created This API Resource

The rule is the unit of per-service routing. A shared HTTPS listener stays
untouched while each service deploys its own rule -- "host `api.example.com`
forwards to the api target group", "path `/admin/*` requires OIDC login
first" -- and removes it when the service goes away. Modeling the rule as its
own component lets you:

- **Deploy routing with the service**: a service's manifest set carries its
  target group and its rule; nothing else in the environment changes.
- **Share one load balancer across many services**: host- and path-based
  rules multiplex a single ALB and certificate, instead of one load balancer
  per service.
- **Scope cross-cutting behavior to routes**: authentication, redirects, and
  canary weights apply to exactly the requests a rule matches.

## Key Features

### Conditions

- **Six criteria**: host header, path pattern, HTTP header, HTTP method,
  query string, and source IP -- each supporting multiple values (OR within a
  condition), with up to five conditions ANDed together per rule.
- **Wildcards and regex**: host, path, and header conditions take wildcard
  patterns (`*`, `?`) and regular expressions. Regex matching is an opt-in
  load balancer attribute the Terraform provider does not expose -- enable
  it on the ALB via the AWS console or CLI before regex rules can match.

### Actions

- **The full ALB action set**: forward (up to five weighted target groups
  with optional group stickiness), redirect, fixed-response, and
  authentication steps (`authenticate-cognito`, `authenticate-oidc`,
  `jwt-validation`) chained before the terminal action.
- **Per-route canaries**: weighted forwards shift traffic between two target
  groups for just this route, without touching any other rule.

### Ordering and Rewrites

- **Explicit priorities**: 1-50000, unique per listener, lower evaluates
  first; omit to append after the current highest.
- **Transforms**: host-header rewrite and URL rewrite (regex
  find-and-replace) applied to matching requests before the action runs.

## Benefits

- **Composability**: the rule references its listener and target groups
  through `valueFrom`, so the routing graph shows exactly which hosts and
  paths reach which backends.
- **Independent lifecycles**: rules come and go with services; the listener
  and load balancer never churn.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `rule_arn`: ARN of the rule (the handle audit tooling and imports reference)
- `priority`: the priority AWS assigned -- meaningful when the spec left it unset

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
