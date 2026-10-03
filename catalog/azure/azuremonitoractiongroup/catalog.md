# Azure Monitor Action Group

Deploys an Azure Monitor Action Group -- the notification and automation hub alerts fire into. When a metric alert, scheduled query alert, or activity log alert triggers, Azure notifies every receiver in the referenced group: human channels (email, SMS, voice, the Azure mobile app), automation (webhooks, Azure Functions, Logic Apps, Automation runbooks), streaming and ITSM systems (Event Hubs, ServiceNow), and role-based fan-out (every holder of an ARM role). One group typically serves many alert rules -- it is the stable routing node; the alert rules are the volatile edge.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Action Group** -- a `Microsoft.Insights/actionGroups` resource (GLOBAL -- it lives in a resource group but not in a region, so notifications keep flowing during regional outages) carrying the configured receiver lists and the short name shown as the SMS/push sender identity
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically, merged with any user tags (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the action group will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **Receiver-side prerequisites, per channel** -- automation receivers reference things that must already exist: a Function App's trigger URL, a Logic App's callback URL, an Automation runbook's webhook, an Event Hub, or an ITSM Connector configured on a Log Analytics workspace.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Action Group**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **On-Call Team Notifications** preset in the [Presets](#presets) tab to pre-populate a human notification group.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorActionGroup
metadata:
  name: platform-oncall
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: "observability-rg"
  shortName: PltOnCall
  emailReceivers:
    - name: oncall-email
      emailAddress: oncall@example.com
      useCommonAlertSchema: true
  smsReceivers:
    - name: primary-oncall-phone
      countryCode: "1"
      phoneNumber: "5551230000"
```

```shell
planton apply -f action-group.yaml
```

This creates a global action group whose SMS messages arrive signed "PltOnCall"; a group with no receivers at all is also legal -- a routing node declared before its channels exist. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to a resource group deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: observability-rg
      fieldPath: status.outputs.resource_group_name
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the action group -- and alert rules deployed afterwards reference the group's `action_group_id` output the same way.

## Key Configuration

These are the most important decisions when configuring an action group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Short name** -- The required `shortName` (1-12 characters) is the sender identity every SMS and mobile push carries. The person paged at 3 AM sees this, not the resource name -- make it recognizable on a phone screen ("PltOnCall", "DBTeam").

**Receivers by persona** -- Eleven receiver lists, all optional. Human channels (email, SMS, voice, app push) page people -- match the channel to severity, since SMS and voice are rate-limited by Azure per number. Automation hooks (webhook, Azure Function, Logic App, Automation runbook) trigger software; the webhook's optional `aadAuth` block is the keyless posture -- the call authenticates as an Entra application instead of a secret baked into the URL. Event Hub receivers stream the payload to SIEMs; ITSM receivers open work items through an ITSM Connector (the `workspaceId` is Azure's pipe-joined composite `{subscription_id}|{workspace_customer_id}`, not an ARM ID). ARM role receivers notify every holder of a role -- the membership is the distribution list.

**Common alert schema** -- Per-receiver `useCommonAlertSchema` picks the payload shape: one consistent JSON across all alert types (prefer it for anything parsed by software) or the legacy per-alert-type formats.

**The kill switch** -- `enabled: false` silences the whole group: every alert firing into it is swallowed while the alert rules keep evaluating. The maintenance-window tool -- silence one group instead of editing fifteen alert rules.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureFunctionApp** | `azureFunctionReceivers[].functionAppResourceId` | `status.outputs.function_app_id` |
| **AzureRoleDefinition** | `armRoleReceivers[].roleId` | `status.outputs.role_definition_guid` |
| **AzureEventHub** | `eventHubReceivers[].eventHubName` | `status.outputs.event_hub_name` |
| **AzureEventHubNamespace** | `eventHubReceivers[].eventHubNamespace` | `status.outputs.namespace_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `action_group_id` | Azure Resource Manager ID of the action group | The alert seam: AzureMonitorMetricAlert, AzureMonitorScheduledQueryAlert, and AzureMonitorActivityLogAlert actions all reference it |
| `action_group_name` | Name of the action group | Azure CLI references, portal navigation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-call team** -- email plus SMS for the humans, the common alert schema on. The stable routing node every paging alert references. Start from the **On-Call Team Notifications** preset.

**Automation hooks** -- a webhook into incident tooling plus an Azure Function for auto-triage; no human channel at all. Start from the **Automation Hooks (Webhook + Function)** preset.

**Role fan-out** -- notify every subscription Owner about governance events with no address list to maintain. Start from the **ARM Role Fan-Out** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the action group is created
- [**Azure Monitor Metric Alert**](/infra-catalog/azure-monitor-metric-alert) -- fires this group when a platform metric breaches
- [**Azure Monitor Scheduled Query Alert**](/infra-catalog/azure-monitor-scheduled-query-alert) -- fires this group when a KQL query result crosses a threshold
- [**Azure Monitor Activity Log Alert**](/infra-catalog/azure-monitor-activity-log-alert) -- fires this group on control-plane and service-health events
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- receives the streamed alert payload for SIEM pipelines
