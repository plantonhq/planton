# AWS LB Listener Rule

Deploys an ALB listener rule — a condition-action pair that routes matching requests, evaluated in priority order before the listener's default action. The rule is the unit of per-service routing: a shared HTTPS listener stays untouched while each service deploys its own rule ("host api.example.com forwards to the api group", "path /admin/* requires OIDC login first") and removes it when the service goes away. Rules are an ALB concept — NLB listeners route purely by port and protocol.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Listener Rule** -- on the referenced listener, with its priority, condition blocks, action chain, and optional host/URL transforms

The listener, target groups, and any Cognito user pools are separate components referenced by ARN.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Listener** -- an AwsLbListener referenced by its `listener_arn` output.
- **Target Groups** -- AwsLbTargetGroup resources for the forward actions, referenced by their `target_group_arn` outputs.

### AWS Account

- **ELB permissions** -- the credentials used by the Provider Connection must have `elasticloadbalancing:CreateRule`, `ModifyRule`, `SetRulePriorities`, and `DeleteRule`.
- **Priority uniqueness** -- explicit priorities must be unique per listener; plan a spacing convention (100, 200, 300) for rules that shadow each other.
- **HTTPS for auth actions** -- the authentication actions (Cognito, OIDC, JWT) require the parent listener to be HTTPS.

## Deploy

### Console

Open the deployment store, find **AWS LB Listener Rule**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Path-Based Routing** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLbListenerRule
metadata:
  name: api-route
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  listenerArn:
    valueFrom:
      kind: AwsLbListener
      name: https-listener
      fieldPath: status.outputs.listener_arn
  priority: 100
  conditions:
    - pathPattern:
        values:
          - /api/*
  actions:
    - type: forward
      forward:
        targetGroups:
          - arn:
              valueFrom:
                kind: AwsLbTargetGroup
                name: api-servers
                fieldPath: status.outputs.target_group_arn
```

```shell
planton apply -f rule.yaml
```

This routes every /api/* request on the shared HTTPS listener to the api-servers group at priority 100. An Infra Job tracks the provisioning in real time.

### InfraChart

When the rule deploys alongside its listener and target group in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  listenerArn:
    valueFrom:
      kind: AwsLbListener
      name: https-listener
      fieldPath: status.outputs.listener_arn
  priority: 100
  conditions:
    - pathPattern:
        values:
          - /api/*
  actions:
    - type: forward
      forward:
        targetGroups:
          - arn:
              valueFrom:
                kind: AwsLbTargetGroup
                name: api-servers
                fieldPath: status.outputs.target_group_arn
```

The InfraPipeline resolves the dependency graph, deploys the listener and target group first, then attaches the rule.

## Key Configuration

These are the most important decisions when configuring a rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The listener attachment is a one-way door** -- moving a rule replaces it. Priority, conditions, actions, and transforms all update in place.

**Priority is the evaluation order** -- 1–50000, unique per listener, lower first, first match wins. Priority 0 lets AWS append after the current highest — safe for non-overlapping routes, but rules that shadow each other (/api/* alongside /*) need explicit priorities.

**Conditions AND across blocks, OR within** -- 1–5 blocks per rule, each matching exactly one thing (host, path, header, method, query string, or source IP); a request must satisfy all blocks, and the values inside one block are alternatives. Wildcards (* and ?) cover most needs; regex patterns require enabling regex matching on the load balancer — an attribute the Terraform provider does not expose, so enable it via the AWS console or CLI.

**Actions mirror the listener's model** -- authentication actions (Cognito, OIDC, JWT validation) may front the terminal action, and every chain ends in exactly one forward, redirect, or fixed response. Weighted forwards enable route-scoped canaries. The OIDC client secret is a managed secret reference, resolved just-in-time.

**Transforms rewrite before routing** -- an optional host-header rewrite and/or URL rewrite (regex find-and-replace with capture groups), for backends that expect a different shape than clients send — the classic /api/v1/* prefix strip.

**Source IP matches the connecting address** -- not X-Forwarded-For; behind CloudFront the connecting address is CloudFront's, not the user's.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLbListener** | `listenerArn` | `status.outputs.listener_arn` |
| **AwsLbTargetGroup** | `actions[].forward.targetGroups[].arn` | `status.outputs.target_group_arn` |
| **AwsCognitoUserPool** | `actions[].authenticateCognito.userPoolArn` / `.userPoolDomain` | `status.outputs.user_pool_arn` / `status.outputs.user_pool_domain` |
| **AwsCognitoUserPoolClient** | `actions[].authenticateCognito.userPoolClientId` | `status.outputs.client_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_arn` | ARN of the rule | Auditing and the AWS CLI |
| `priority` | The priority AWS actually assigned | Verifying auto-appended rules landed where expected |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Path-based routing** -- /api/* to the api group on a shared listener. Start from the **Path-Based Routing** preset.

**Host-based routing** -- one rule per service hostname: the microservice front door. Start from the **Host-Based Routing** preset.

**Canary rollout** -- a weighted forward (90/10) scoped to one route. Start from the **Canary Weighted Forward** preset.

## Works With

- [**AWS LB Listener**](/infra-catalog/aws-lb-listener) -- the listener this rule attaches to, referenced by `listenerArn`.
- [**AWS LB Target Group**](/infra-catalog/aws-lb-target-group) -- the forward destinations, referenced per action.
- [**AWS ALB**](/infra-catalog/aws-alb) -- the load balancer at the top of the chain (rules are an ALB concept).
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) -- the user pool behind authenticate-cognito actions.
