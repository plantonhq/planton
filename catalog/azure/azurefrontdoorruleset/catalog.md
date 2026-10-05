# Azure Front Door Rule Set

Deploys a rule set inside an Azure Front Door (Standard/Premium) profile -- the delivery-rules engine that runs at Microsoft's edge. A rule set is an ordered collection of rules; each rule pairs match conditions (up to 10 per rule, ANDed together, across 19 condition types) with actions (header edits, a redirect or rewrite, a route-configuration override; 1 to 5 per rule). The set runs against no traffic on its own: Azure Front Door Route resources attach it via `ruleSetIds`, and its rules then evaluate on every request those routes match -- one set is commonly shared by many routes, which is exactly why the policy is its own first-class resource.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Rule Set** -- a named child of the profile; the set itself carries no properties, only its rules
- **Delivery rules** -- one ARM child resource per rule, keyed by the rule's name, each with conditions (path, headers, cookies, device type, scheme, geo/IP, TLS protocol, and more), actions (request/response header changes, redirect XOR rewrite, route-configuration override), an evaluation order, and CONTINUE/STOP behavior

ARM does not support tags on rule sets, so no Azure tags are applied here.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door Profile** the rule set nests under (`profileId`). Rule sets and the routes that attach them must live in the same profile.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Rule Set**, and click **Deploy**. The wizard walks you through the profile and name, then the rules builder: pick a condition type (each offers only its legal operators), pair it with actions, and let the live counters hold Azure's 10-condition and 5-action caps. Start from the **Security Headers Policy** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorRuleSet
metadata:
  name: security-headers
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  ruleSetName: securityHeaders
  rules:
    - name: addSecurityHeaders
      order: 1
      actions:
        responseHeaders:
          - headerAction: OVERWRITE
            headerName: Strict-Transport-Security
            value: "max-age=31536000; includeSubDomains"
          - headerAction: OVERWRITE
            headerName: X-Content-Type-Options
            value: nosniff
```

```shell
planton apply -f front-door-rule-set.yaml
```

This creates a rule set with one condition-less rule -- a rule with no conditions matches every request its routes serve, the deliberate shape for policy-for-everything rules like security headers -- ready to attach to routes via the `rule_set_id` output. An Infra Job tracks the provisioning in real time.

### InfraChart

A header-steered canary policy wired to an origin group in the same InfraPipeline:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  ruleSetName: canarySteering
  rules:
    - name: toCanary
      order: 5
      behaviorOnMatch: STOP
      conditions:
        requestHeader:
          - headerName: X-Canary
            operator: EQUAL
            matchValues:
              - "true"
      actions:
        routeConfigurationOverride:
          originGroupId:
            valueFrom:
              kind: AzureFrontDoorOriginGroup
              name: canary-origins
              fieldPath: status.outputs.origin_group_id
          forwardingProtocol: MATCH_REQUEST
          cacheBehavior: DISABLED
```

The InfraPipeline resolves the dependency graph, deploys the profile and the canary origin group first, then provisions the rule set with the resolved ARM IDs.

## Key Configuration

These are the most important decisions when configuring a rule set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Rule set name** -- `ruleSetName` is 1-60 characters, letter-first, letters and numbers only (no hyphens). ForceNew: renaming replaces the set AND every rule inside it, stranding every attached route until it repoints.

**Rule identity and order** -- rule names (1-260 characters, letter-first, no hyphens) must be unique within the set: each rule is its own ARM resource keyed by name, so duplicates would silently target one resource, and renaming a rule replaces it. Evaluation is ascending `order` -- leave gaps (10, 20, 30) so later inserts need no renumbering. Everything except the names updates IN PLACE: tuning never replaces the set.

**Match behavior** -- `behaviorOnMatch: CONTINUE` (the default) lets later rules still run; STOP ends evaluation. Use STOP for terminal actions like redirects, where accumulating further actions would be surprising. A rule with no conditions applies to everything.

**Conditions** -- up to 10 per rule across the 19 typed condition groups; conditions AND together, the up-to-25 values within one condition OR together. Each type accepts only its own operators: URL path is the only type with WILDCARD, remote address takes GEO_MATCH/IP_MATCH, and equality-only types (method, scheme, HTTP version, device, TLS protocol) carry no free-form operator choice.

**Actions** -- 1 to 5 per rule, and EACH header action counts toward the cap. A redirect and a rewrite cannot coexist on one rule: a redirect answers the client directly while a rewrite forwards to the origin, so the two contradict (spec-enforced).

**Route-configuration override** -- steers matched requests to a different origin group (paired with a `forwardingProtocol` -- the spec requires the pair) and/or overrides caching. The override always makes an explicit cache decision -- there is no leave-it-alone value -- and forced cache durations use `d.HH:MM:SS` (days 1-365) or `HH:MM:SS` format.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |
| **AzureFrontDoorOriginGroup** (route overrides) | `rules[].actions.routeConfigurationOverride.originGroupId` | `status.outputs.origin_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_set_id` | ARM resource ID of the rule set | AzureFrontDoorRoute's `ruleSetIds` -- the attachment that puts the policy on traffic |
| `rule_set_name` | The set's name within its profile | Operator tooling, portal cross-reference |

The rules inside the set export no individual IDs on purpose: nothing references a rule -- routes attach the whole set.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Baseline security headers** -- one condition-less rule stamping HSTS and nosniff on every response and DELETING the backend's technology fingerprint (X-Powered-By). OVERWRITE beats APPEND here: the edge is authoritative, and overwriting prevents duplicates when a backend also sets the headers. Attach the set to every route so the posture holds regardless of which backend answered. Start from the **Security Headers Policy** preset.

**HTTPS upgrade with tiered caching** -- a permanent HTTPS redirect with STOP (a redirected request should not accumulate further actions), 7-day OVERRIDE_ALWAYS caching for fingerprinted assets with tracking parameters collapsed out of the cache key, and caching DISABLED on API paths. Use HONOR_ORIGIN instead of OVERRIDE_ALWAYS when the backend already sends precise Cache-Control. Start from the **HTTPS Upgrade + Tiered Caching** preset.

**Path rewrite plus cookie canary** -- a rewrite decouples the public URL surface from the backend's real path layout (the client never sees it), and a cookie condition steers flagged requests to a canary origin group -- an edge-level canary that needs no route or DNS change, with caching disabled on the canary so its responses never reach non-canary users. Start from the **Path Rewrite + Cookie Canary** preset.

## Works With

- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container the set nests under
- [**Azure Front Door Route**](/infra-catalog/azure-front-door-route) -- attaches this set to traffic via `ruleSetIds`; the route side owns the attachment
- [**Azure Front Door Origin Group**](/infra-catalog/azure-front-door-origin-group) -- the target of a route-configuration override, the canary steering gesture
