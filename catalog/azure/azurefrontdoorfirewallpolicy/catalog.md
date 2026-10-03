# Azure Front Door Firewall Policy

Deploys a Web Application Firewall (WAF) policy for Azure Front Door -- the rule set Microsoft's edge enforces on HTTP traffic before requests ever reach an origin. Custom match and rate-limit rules evaluate first by ascending priority, Microsoft's managed rule sets (PREMIUM) evaluate next with set/group/rule-scoped tuning, and policy settings decide the enforcement mode, body inspection, block-response customization, and access-log scrubbing. The policy is GLOBAL and resource-group-scoped, and it enforces nothing on its own: an Azure Front Door Security Policy associates it with a profile's domains, and that association turns the WAF on. This is the FRONT DOOR policy type -- the regional WAF that Application Gateways attach is a different resource with a different rule vocabulary.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door WAF Policy** -- a global `Microsoft.Network/frontDoorWebApplicationFirewallPolicies` resource in the resource group, carrying every layer below as inline configuration
- **Custom rules** -- your IP/geo allowlists, header exceptions, and per-client rate limits, evaluated first by ascending priority
- **Managed rules** (PREMIUM only) -- Microsoft_DefaultRuleSet and Bot Manager with set-wide exclusions and per-group/per-rule overrides
- **Policy settings** -- DETECTION/PREVENTION mode, request-body inspection, block-response status code and body, redirect URL, and log-scrubbing rules
- **Azure Tags** -- resource metadata tags (organization, environment, resource ID) merged with user tags; a user tag with the same key wins

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **A resource group** for the policy (`resourceGroup`). The policy itself is global -- the group scopes RBAC and lifecycle.
- **A PREMIUM Front Door profile** (only for managed rules or the JS-challenge/CAPTCHA actions) -- the sku pairing is enforced when the policy is associated, so the tier decision spans the profile and its policies together.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Firewall Policy**, and click **Deploy**. The wizard walks you through placement, the tier and enforcement mode, inspection and block-response behavior, the custom-rules builder, the PREMIUM-gated managed-rules tuning tree, and log scrubbing. Start from the **Standard Rate Limit and IP Denylist** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorFirewallPolicy
metadata:
  name: edge-waf
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "acme-security-rg"
  policyName: acmeedgewaf
  mode: DETECTION
  customRules:
    - name: LoginRateLimit
      ruleType: RATE_LIMIT_RULE
      rateLimitThreshold: 20
      action: BLOCK
      matchConditions:
        - matchVariable: REQUEST_URI
          operator: BEGINS_WITH
          matchValues:
            - /login
```

```shell
planton apply -f front-door-firewall-policy.yaml
```

This creates a STANDARD policy in DETECTION mode with one per-client rate-limit rule -- it logs matches without blocking until you attach it to domains through an Azure Front Door Security Policy and flip the mode to PREVENTION. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the policy to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: security-rg
      fieldPath: status.outputs.resource_group_name
  policyName: acmeedgewafpremium
  sku: PREMIUM
  mode: PREVENTION
  managedRules:
    - type: Microsoft_DefaultRuleSet
      version: "2.1"
      action: RULE_SET_BLOCK
    - type: Microsoft_BotManagerRuleSet
      version: "1.1"
      action: RULE_SET_BLOCK
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the policy with the resolved name.

## Key Configuration

These are the most important decisions when configuring a Front Door WAF policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Policy name** -- `policyName` is 1-128 characters, letter-first, letters and numbers ONLY (no hyphens -- the Front Door family outlier). ForceNew: renaming replaces the policy, which detaches it from every security policy referencing it.

**SKU** -- unspecified deploys STANDARD, the right answer unless you need the managed rule sets or the JS-challenge/CAPTCHA actions, which are PREMIUM-only. The sku must match the sku of every Front Door profile the policy gets associated with, it is ForceNew, and Azure refuses a PREMIUM-to-STANDARD downgrade outright -- choose PREMIUM deliberately.

**Mode** -- DETECTION logs matches without acting (the tuning posture for new policies watching real traffic); PREVENTION blocks, redirects, or challenges (the production posture). The flip is an in-place update, so configuring the intended production actions while still in DETECTION means going live changes exactly one field.

**Custom rules** -- up to 100, evaluated by ascending priority (lower runs first); conditions within a rule AND together, values within a condition OR together. RATE_LIMIT_RULE counts per-client matches in a sliding window (`rateLimitDurationInMinutes`, threshold default 10). An ALLOW action skips the remaining custom AND managed rules -- the allowlist gesture. Prefer SOCKET_ADDR over REMOTE_ADDR for IP matching: it is the source IP Front Door actually sees, while REMOTE_ADDR trusts the spoofable X-Forwarded-For.

**Managed rules** (PREMIUM) -- Microsoft_DefaultRuleSet 1.1/2.0/2.1 (2.x uses anomaly scoring: per-rule overrides may only substitute OVERRIDE_ANOMALY_SCORING or OVERRIDE_LOG), Microsoft_BotManagerRuleSet 1.0/1.1 (the only home of OVERRIDE_JS_CHALLENGE), and the legacy DefaultRuleSet 1.0. A rule listed in an override is DISABLED unless `enabled: true` -- the common false-positive tuning gesture. Scoped `exclusions` (skip one cookie or header) are the surgical alternative to disabling a rule.

**Block response** -- `customBlockResponseStatusCode` accepts 200, 403, 405, 406, 429, or 990-999 (403 is Azure's default); `customBlockResponseBody` carries a base64-encoded branded error page. `redirectUrl` is required at deploy time whenever any rule or override uses the REDIRECT action.

**Body inspection** -- `requestBodyCheckEnabled` defaults to true; turning it off blinds the WAF to body-borne attacks (SQL injection in POST forms, JSON payload attacks).

**Log scrubbing** (Azure preview) -- masks auth headers, PII arguments, or client IPs before WAF logs land in storage. SCRUB_REQUEST_IP_ADDRESS and SCRUB_REQUEST_URI accept only the every-key SELECTOR_EQUALS_ANY operator; the keyed collections take a named selector.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `firewall_policy_id` | ARM resource ID of the WAF policy | AzureFrontDoorSecurityPolicy's `firewallPolicyId` -- the association that turns enforcement on |
| `firewall_policy_name` | The policy's name within its resource group | Operator tooling, portal cross-reference |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard-tier custom rules** -- per-client rate limiting and IP denylisting at Microsoft's edge without paying for PREMIUM. The trade: no managed rule sets, so attack-signature coverage is entirely your own rules, and moving to PREMIUM later replaces the policy. Start from the **Standard Rate Limit and IP Denylist** preset.

**Premium managed prevention** -- Microsoft's managed rule sets in blocking mode, the default posture for production workloads on a PREMIUM profile: OWASP-class coverage and bot classification maintained server-side by Microsoft, while you keep only the tuning (exclusions, overrides) in code. Start from the **Premium Managed Rules Prevention** preset.

**Detection-first rollout** -- deploy in DETECTION associated with real domains, review the WAF logs for matches on legitimate traffic, add exclusions or OVERRIDE_LOG overrides for each false positive, then flip `mode` to PREVENTION. The safe first step of every WAF rollout in front of an existing production application. Start from the **Detection-Mode Rollout** preset.

**One policy, many profiles** -- the policy is resource-group-scoped, not profile-nested, so a single tuned policy is commonly shared by many profiles' security policies. The trade: a tuning change lands everywhere at once.

## Works With

- [**Azure Front Door Security Policy**](/infra-catalog/azure-front-door-security-policy) -- attaches this policy to a profile's domains; without one the policy sits idle
- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the delivery container whose sku must match this policy's sku
- [**Azure Front Door Custom Domain**](/infra-catalog/azure-front-door-custom-domain) -- a domain a security policy scopes this WAF to
- [**Azure Front Door Endpoint**](/infra-catalog/azure-front-door-endpoint) -- an endpoint a security policy scopes this WAF to
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the policy lives in
- [**Azure Web Application Firewall Policy**](/infra-catalog/azure-web-application-firewall-policy) -- the DIFFERENT, regional policy type Application Gateways attach; do not confuse the two
