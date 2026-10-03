# Azure Web Application Firewall Policy

Deploys a regional Web Application Firewall (WAF) policy — the rule set an Azure Application Gateway enforces on HTTP traffic. This is the APPLICATION GATEWAY policy type (`Microsoft.Network/ApplicationGatewayWebApplicationFirewallPolicies`); Azure Front Door's WAF is a different ARM resource with a different rule vocabulary. A policy attaches to gateways at three levels — gateway-wide, per HTTP listener, and per URL path rule — so a single org-standard policy governs many gateways while specific routes carry stricter or looser variants (most specific wins); the attached gateway must be on the WAF_v2 SKU. A policy has three layers, evaluated in order: custom rules (your match and rate-limit rules, by ascending priority), then managed rules (Microsoft's curated OWASP / bot-manager sets), governed by policy settings (Prevention vs Detection, body-inspection limits, log scrubbing).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **WAF Policy** -- the policy with its custom rules, managed rule sets and tuning, enforcement settings, and log-scrubbing rules
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

Azure REQUIRES at least one managed rule set — a WAF policy without one is rejected. The gateway attachment is NOT created here: an AzureApplicationGateway references this policy's `policy_id` output.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the policy will be created — conventionally beside the Application Gateways it protects.
- **An Application Gateway on the WAF_v2 SKU** to attach the policy (created separately; the policy can also exist unattached).
- **A tuning window**: new policies run best in Detection mode against real traffic first, then switch to Prevention once false positives are tuned out.

## Deploy

### Console

Open the deployment store, find **Azure Web Application Firewall Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **OWASP 3.2 Baseline** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureWebApplicationFirewallPolicy
metadata:
  name: waf-baseline
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-web"
  policyName: waf-baseline
  managedRules:
    managedRuleSets:
      - type: OWASP
        version: "3.2"
```

```shell
planton apply -f waf-policy.yaml
```

This creates OWASP 3.2 in Prevention mode (Azure's default), blocking SQL injection, XSS, RCE, LFI, and protocol violations out of the box. An Infra Job tracks the provisioning in real time.

### InfraChart

When the resource group, policy, and gateway deploy in the same InfraChart, wire the policy's resource group with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: web-rg
      fieldPath: status.outputs.resource_group_name
  policyName: waf-baseline
  managedRules:
    managedRuleSets:
      - type: OWASP
        version: "3.2"
```

The InfraPipeline resolves the dependency graph, deploys the resource group and policy first, then the Application Gateway that attaches the policy by referencing its `policy_id` output as `firewallPolicyId`.

## Key Configuration

These are the most important decisions when configuring a WAF policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Managed rules (required)** -- OWASP 3.2 is the production standard (several enforcement dials only work with it); the bot manager runs beside it. Tune with per-rule OVERRIDES (remember: listing a rule DISABLES it unless explicitly enabled) and per-part EXCLUSIONS — never disable a whole set for one false positive. Exclusions carry a NARROWER version vocabulary than the rule sets.

**Custom rules** -- evaluated first, by ascending priority. MATCH rules act on every matching request; RATE_LIMIT rules act past a threshold within a window (and cannot ALLOW). Conditions AND together; values within one condition OR; transforms (LOWERCASE + URL_DECODE) catch encoding evasions.

**Enforcement (`policy_settings`)** -- Prevention blocks, Detection only logs (the tuning mode). The body-inspection dials carry Azure defaults; a 0 inspect limit means UNLIMITED. Omit the whole block for Azure's defaults.

**Log scrubbing** -- redact auth headers, PII arguments, and client IPs from the WAF logs before they land in Log Analytics.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_id` | Azure Resource Manager ID of the WAF policy | AzureApplicationGateway `firewallPolicyId` — gateway-wide, per-listener, or per-path-rule (most specific wins) |

The outputs also carry `policy_name` -- gateways attach the policy by ARM ID, so the name has no ValueFromRef consumer.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**OWASP baseline** -- the OWASP 3.2 core rule set in Prevention mode, no custom rules, no overrides: the policy almost every gateway starts from. Start from the **OWASP 3.2 Baseline** preset.

**Rate limit and geo** -- OWASP plus the bot manager, with custom rules for per-client rate limiting and geo restriction. Start from the **Edge Protection: Rate Limits, Geo Fencing, Bot Challenges** preset.

**Detection tuning** -- the same baseline in Detection mode for watching real traffic before enforcing. Start from the **Detection-Mode Tuning with Exclusions and Overrides** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the policy is created
- [**Azure Application Gateway**](/infra-catalog/azure-application-gateway) -- the WAF_v2 gateway that attaches this policy at up to three levels
