# Azure Monitor Activity Log Alert

Deploys an Azure Monitor Activity Log Alert -- the control-plane watchdog. It fires when a matching entry appears in the Azure Activity Log: the subscription-level record of control-plane operations, service-health incidents, resource-health transitions, policy events, and Advisor recommendations. It is the ONLY way to alert on the events that never show up as metrics -- a VM someone deleted, a region-wide Azure incident, a resource going Unavailable, a policy denial.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Activity Log Alert** -- a `Microsoft.Insights/activityLogAlerts` resource carrying the scopes, the category-driven matching criteria, and the action-group wiring. The definition defaults to the global location (the alert evaluates the subscription-global Activity Log regardless)
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the alert definition will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **An action group** to notify (optional but recommended) -- an alert with no actions records matches but notifies nobody.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Activity Log Alert**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Service-Health Incident Alert** preset in the [Presets](#presets) tab to pre-populate the incident watch every subscription should carry.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorActivityLogAlert
metadata:
  name: vm-delete-watch
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "governance-rg"
  name: vm-delete-watch
  scopes:
    - valueFrom:
        kind: AzureResourceGroup
        name: production-rg
        fieldPath: status.outputs.resource_group_id
  criteria:
    category: ADMINISTRATIVE
    operationName: Microsoft.Compute/virtualMachines/delete
    statuses:
      - Succeeded
  actions:
    - actionGroupId:
        valueFrom:
          kind: AzureMonitorActionGroup
          name: platform-governance
          fieldPath: status.outputs.action_group_id
  description: Pages when a production VM is deleted
```

```shell
planton apply -f activity-alert.yaml
```

This creates a deletion watch: any successful VM delete under the production resource group notifies the governance action group. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the alert to the resource group it watches and the action group it fires:

```yaml
spec:
  scopes:
    - valueFrom:
        kind: AzureResourceGroup
        name: production-rg
        fieldPath: status.outputs.resource_group_id
  actions:
    - actionGroupId:
        valueFrom:
          kind: AzureMonitorActionGroup
          name: platform-governance
          fieldPath: status.outputs.action_group_id
```

The InfraPipeline resolves the dependency graph and deploys the watched scope and the action group before the alert.

## Key Configuration

These are the most important decisions when configuring an activity log alert. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Category first** -- `criteria.category` selects WHICH slice of the Activity Log the alert watches (Administrative, Service Health, Resource Health, Policy, Autoscale, Security, Recommendation); every other filter narrows within it. Empty filters match anything in the category -- scope wide, filter precisely.

**Category-specific filters** -- the Recommendation category matches by category/impact OR by one recommendation type ID (mutually exclusive, Azure's rule); Resource Health matches state TRANSITIONS (current Unavailable + reason Platform-initiated is the "Azure broke it" watch); Service Health matches incident types, affected regions, and services. The caller filter has no meaning for the health categories.

**Scopes** -- a subscription ARM path watches everything in it; a resource-group reference watches one environment; a specific resource ID watches one resource. Events UNDER a scope are evaluated.

**Definition location** -- unspecified means global, the correct choice for virtually every alert; the named regions exist only for data residency of the alert definition itself.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureResourceGroup** | `scopes[]` | `status.outputs.resource_group_id` |
| **AzureMonitorActionGroup** | `actions[].actionGroupId` | `status.outputs.action_group_id` |

### What This Kind Provides

The alert is a leaf in the dependency graph: `status.outputs` carries only the alert's own identifiers (`activity_log_alert_id`, `activity_log_alert_name`) for audit and CLI reference -- no downstream Infra Component consumes them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Service health** -- Azure incident notifications for your regions and services: the alert every subscription should carry. Start from the **Service-Health Incident Alert** preset.

**Resource delete** -- successful deletions under a scope, into the governance channel. Start from the **Administrative Change Alert** preset.

## Works With

- [**Azure Monitor Action Group**](/infra-catalog/azure-monitor-action-group) -- the notification hub the alert fires into
- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- both the definition's home and the classic watch scope
