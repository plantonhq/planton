# Azure Firewall Policy

Deploys an Azure Firewall Policy — the reusable rule-and-inspection document Azure Firewall instances enforce. The policy carries everything about WHAT the firewall does (threat intelligence posture, DNS proxying, TLS inspection, intrusion detection, SNAT behavior) while the firewall instance carries WHERE it runs — so security policy is authored once, by the security team, and enforced on every firewall that attaches it, across regions. Rules themselves live in separate AzureFirewallPolicyRuleCollectionGroup children, deployable on their own schedules; policies also support inheritance, where a child policy extends a security-team baseline via `basePolicyId`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Firewall Policy** -- the policy document with its tier, threat-intelligence mode, and every configured block (allowlist, DNS, IDPS, identity, TLS certificate, analytics, explicit proxy, SNAT ranges)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically and merged with the user tags

The sku and threat-intelligence mode are always sent explicitly (Standard/Alert when unspecified) — Azure's own defaults, made deterministic so both IaC engines produce identical payloads. Rules and firewalls are NOT created here: rule collection groups nest under this policy and firewalls attach it, each with its own lifecycle.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the policy will be created — conventionally the security team's, since one policy governs many application firewalls.
- **The tier decision, made deliberately**: the sku is fixed at creation AND must match the tier of every firewall that attaches the policy — Premium features (TLS inspection, IDPS) require choosing PREMIUM now, not later.
- **For TLS inspection**: an AzureKeyVaultCertificate holding your intermediate CA, and an AzureUserAssignedIdentity with "Key Vault Secrets User" on that vault.

## Deploy

### Console

Open the deployment store, find **Azure Firewall Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Egress Baseline** preset in the [Presets](#presets) tab for the classic hub-spoke posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFirewallPolicy
metadata:
  name: egress-baseline
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "rg-security-hub"
  name: egress-baseline
  sku: STANDARD
  threatIntelligenceMode: DENY
  dns:
    proxyEnabled: true
  tags:
    cost-center: network-security
```

```shell
planton apply -f firewall-policy.yaml
```

This creates a STANDARD policy with threat intelligence in deny and the DNS proxy on, ready for rule collection groups to nest under it and firewalls to attach it. An Infra Job tracks the provisioning in real time. The DNS proxy makes FQDN-based network rules deterministic — point spoke DNS at the attached firewall's private IP.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the policy to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: security-hub
      fieldPath: status.outputs.resource_group_name
  tlsCertificate:
    keyVaultSecretId:
      valueFrom:
        kind: AzureKeyVaultCertificate
        name: egress-ca
        fieldPath: status.outputs.versionless_secret_id
    name: egress-ca
```

The InfraPipeline resolves the dependency graph, deploys the resource group (and, for Premium inspection, the Key Vault certificate and identity) first, then provisions the policy — and the rule collection groups and firewalls that use it reference this policy's `firewall_policy_id`.

## Key Configuration

These are the most important decisions when configuring a Firewall Policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tier (`sku`)** -- BASIC (SMB-scale, alert-only threat intelligence), STANDARD (the production default: full rule engine, threat intelligence deny, DNS proxy), PREMIUM (adds TLS inspection, IDPS, URL filtering, web categories). Fixed at creation and matched by every attached firewall — the whole fleet moves together. Unspecified deploys STANDARD.

**Threat intelligence** -- Microsoft's known-malicious feed, evaluated BEFORE your rules. DENY is the production posture; exempt false positives surgically via `threatIntelligenceAllowlist` (at least one entry when the block is present) instead of lowering the mode.

**DNS proxy** -- required for FQDN-based NETWORK rules to resolve deterministically: the firewall and its clients must see the same answers, and pointing client DNS at the firewall's proxy guarantees it.

**TLS inspection (Premium)** -- the intermediate CA's Key Vault SECRET reference (versionless by default, so CA renewals flow through automatically) plus a user-assigned identity that can read it. Deploy the CA chain to workload trust stores before enabling on live traffic.

**Inheritance (`basePolicyId`)** -- the enterprise layering pattern: a security-team baseline extended by per-application child policies. Base and child must share a region and a tier (validated at apply time).

**SNAT private ranges** -- when set, the list REPLACES Azure's RFC 1918 default; `autoLearnPrivateRangesEnabled` lets Azure learn the boundary from route tables instead (Azure records only "Enabled" — disabling happens by omission).

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureFirewallPolicy** (base) | `basePolicyId` | `status.outputs.firewall_policy_id` |
| **AzureKeyVaultCertificate** | `tlsCertificate.keyVaultSecretId` | `status.outputs.versionless_secret_id` |
| **AzureUserAssignedIdentity** | `identity.userAssignedIdentityIds[]` | `status.outputs.identity_id` |
| **AzureLogAnalyticsWorkspace** | `insights.defaultLogAnalyticsWorkspaceId` | `status.outputs.workspace_id` |
| **AzureIpGroup** | `intrusionDetection.trafficBypass[].sourceIpGroups[]` / `destinationIpGroups[]` | `status.outputs.ip_group_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `firewall_policy_id` | Azure Resource Manager ID of the policy | AzureFirewall attachment (`firewallPolicyId`), AzureFirewallPolicyRuleCollectionGroup nesting, child policies' `basePolicyId` |
| `identity_principal_id` | The system-assigned principal (empty without a system identity) | Role assignments — e.g. Key Vault secret read for TLS inspection |

`firewall_policy_name` is also exported for tooling that addresses the policy by name; it has no ValueFromRef consumers.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard egress baseline** -- threat intelligence in DENY with the DNS proxy on: the parent policy of the enterprise pattern. Start from the **Standard Egress Baseline** preset.

**Premium TLS inspection** -- decrypt-inspect-re-encrypt with your CA from Key Vault, plus IDPS in deny: the regulated-environment posture. Start from the **Premium TLS Inspection + IDPS** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the policy is created
- [**Azure Firewall**](/infra-catalog/azure-firewall) -- the enforcement instance that attaches this policy (`firewallPolicyId`)
- [**Azure Firewall Policy Rule Collection Group**](/infra-catalog/azure-firewall-policy-rule-collection-group) -- the ordered rule documents that nest under this policy
- [**Azure IP Group**](/infra-catalog/azure-ip-group) -- named address sets the policy's IDPS bypasses (and its rule groups) reference
- [**Azure Key Vault Certificate**](/infra-catalog/azure-key-vault-certificate) -- holds the intermediate CA TLS inspection signs with
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the identity that reads the CA from Key Vault
- [**Azure Log Analytics Workspace**](/infra-catalog/azure-log-analytics-workspace) -- receives policy analytics
