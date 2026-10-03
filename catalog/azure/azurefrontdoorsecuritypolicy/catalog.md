# Azure Front Door Security Policy

Deploys a security policy inside an Azure Front Door (Standard/Premium) profile -- the association that attaches a Front Door WAF policy (AzureFrontDoorFirewallPolicy) to the hostnames the profile serves. A WAF policy enforces nothing until a security policy associates it; this kind is the enforcement seam. The `domainIds` list names the protected hostnames, and each entry is either an endpoint's ARM ID (the generated `*.azurefd.net` hostname) or a custom domain's ARM ID -- Azure accepts both interchangeably, and one list can mix them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Security Policy** -- a named child of the profile binding one WAF policy to a domain list
- **Live enforcement** -- from the moment it deploys, the WAF's rules run on every request to the associated domains

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An Azure Front Door profile** with at least one endpoint or validated custom domain.
- **A Front Door WAF policy of the SAME sku** as the profile -- Azure rejects a mismatched pairing at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Security Policy**, and click **Deploy**. The wizard walks you through the profile and name, then the enforcement scope: the WAF policy and the protected-domains list, whose per-row toggle switches between endpoint and custom-domain targets. Start from the **Endpoint WAF Association** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorSecurityPolicy
metadata:
  name: waf-attach
  org: acme-corp
  env: prod
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  securityPolicyName: waf-attach
  firewallPolicyId:
    valueFrom:
      kind: AzureFrontDoorFirewallPolicy
      name: edge-waf
      fieldPath: status.outputs.firewall_policy_id
  domainIds:
    - valueFrom:
        kind: AzureFrontDoorEndpoint
        name: web-endpoint
        fieldPath: status.outputs.endpoint_id
```

```shell
planton apply -f front-door-security-policy.yaml
```

This turns the WAF on for the endpoint's generated hostname; custom domains join the same list as they validate, and the list updates in place. An Infra Job tracks the provisioning in real time.

### InfraChart

In an InfraChart, wire the full enforcement scope through `valueFrom`:

```yaml
spec:
  profileId:
    valueFrom:
      kind: AzureFrontDoorProfile
      name: cdn-profile
      fieldPath: status.outputs.profile_id
  securityPolicyName: waf-attach
  firewallPolicyId:
    valueFrom:
      kind: AzureFrontDoorFirewallPolicy
      name: edge-waf
      fieldPath: status.outputs.firewall_policy_id
  domainIds:
    - valueFrom:
        kind: AzureFrontDoorEndpoint
        name: web-endpoint
        fieldPath: status.outputs.endpoint_id
    - valueFrom:
        kind: AzureFrontDoorCustomDomain
        name: www-example-com
        fieldPath: status.outputs.custom_domain_id
```

The InfraPipeline resolves the dependency graph, provisioning the profile, the WAF policy, the endpoint, and the domain before the association that binds them.

## Key Configuration

These are the most important decisions when configuring a Front Door security policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Security policy name** -- unique within the profile; letters, digits, and hyphens. ForceNew: a rename replaces the association with a brief enforcement gap while the old association tears down.

**Firewall policy** -- the WAF to enforce. ForceNew: swapping it replaces the association (fast, metadata-only). The WAF policy's sku must match the profile's sku -- Azure rejects a mismatched pairing at deploy time.

**Domain list** -- at least 1, at most 500 (100 on a STANDARD profile -- the cap rides the profile's sku, checked at deploy time). Updatable in place: adding a domain extends protection without touching the others.

**Path scope is not yours to choose** -- the WAF applies to ALL paths (`/*`) on every associated domain; Azure's security policies accept no other pattern. Scope enforcement by choosing WHICH domains to associate, not which paths.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorProfile** | `profileId` | `status.outputs.profile_id` |
| **AzureFrontDoorFirewallPolicy** | `firewallPolicyId` | `status.outputs.firewall_policy_id` |
| **AzureFrontDoorEndpoint** | `domainIds[]` | `status.outputs.endpoint_id` |
| **AzureFrontDoorCustomDomain** | `domainIds[]` | `status.outputs.custom_domain_id` |

### What This Kind Provides

Nothing composes on a security policy -- it is itself the association, the terminal node of the WAF wiring. `status.outputs` carries `security_policy_id` and `security_policy_name` for operational addressing (diagnostics, RBAC scoping, ARM reads), not for downstream ValueFromRef wiring.

## Common Patterns

**Turn the WAF on for the default hostname** -- the second half of every Front Door WAF rollout: deploy the AzureFrontDoorFirewallPolicy, then associate it with the endpoint's `endpoint_id` so the generated `*.azurefd.net` hostname is protected from day one. Without this association, the WAF policy sits idle. Start from the **Endpoint WAF Association** preset.

**Production association with custom domains** -- once real domains serve traffic, list them alongside the endpoint in the same `domainIds` -- endpoint IDs and custom-domain IDs mix freely. The list updates in place, so each new custom domain joins without touching the others. A pending domain can be associated, but the WAF only sees traffic once DNS validation passes and the CNAME points at Front Door. Start from the **Custom Domain WAF Association** preset.

**Different rules for different hostnames** -- one security policy binds exactly one WAF policy to its domain list. To enforce different rule sets (say, a stricter API policy and a lenient marketing-site policy), create one security policy per WAF policy and split the domains between them -- never try to path-scope a single association.

## Works With

- [**Azure Front Door Firewall Policy**](/infra-catalog/azure-front-door-firewall-policy) -- the WAF rule set this association enforces, referenced by `firewallPolicyId`; its sku must match the profile's
- [**Azure Front Door Profile**](/infra-catalog/azure-front-door-profile) -- the parent container whose sku also caps the domain list (100 on STANDARD, 500 on PREMIUM)
- [**Azure Front Door Endpoint**](/infra-catalog/azure-front-door-endpoint) -- a protectable target: its `endpoint_id` covers the generated `*.azurefd.net` hostname
- [**Azure Front Door Custom Domain**](/infra-catalog/azure-front-door-custom-domain) -- a protectable target: its `custom_domain_id` covers the branded hostname
