# Azure Monitor Data Collection Rule Association

Attaches one machine (VM, VM scale set, or Arc-enabled server) to an Azure Monitor data collection rule or data collection endpoint -- the resource that actually puts a machine under monitoring. The association is an extension resource living on the target machine: creating it puts the machine under the rule, destroying it detaches the machine without touching the rule or any other machine's collection. One rule serves any number of machines, and one machine can carry many associations -- several rule bindings plus at most one endpoint binding.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Data collection rule association** -- an extension resource scoped under the target machine's ARM ID (`{target_id}/providers/Microsoft.Insights/dataCollectionRuleAssociations/{name}`), binding the machine to a data collection rule or, in the endpoint form, to a Data Collection Endpoint for configuration access

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **A target machine** -- reference an Azure Virtual Machine's `vm_id` output via valueFrom, or pass the literal ARM ID of a VM, scale set, or Arc-enabled server.
- **A data collection rule** -- reference an Azure Monitor Data Collection Rule's `data_collection_rule_id` output, or provide a Data Collection Endpoint ARM ID for the endpoint form.

### Azure Subscription

- **Exactly one binding per association** -- a rule OR an endpoint, never both; the spec rejects manifests carrying both. A machine joins several rules by carrying several associations.
- **The name is required for rule bindings** and must be left unset for endpoint bindings: Azure mandates the fixed name `configurationAccessEndpoint` there, and both engines apply it automatically.
- **Collection needs the agent** -- the association creates fine on a machine without the Azure Monitor Agent; telemetry starts flowing when the agent runs and discovers the association.
- **The association is free** -- the telemetry the rule collects is billed at its destinations.

## Deploy

### Console

Open the deployment store, find **Azure Monitor Data Collection Rule Association**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target machine reference, and the rule-or-endpoint binding. Start from the **Attach VM to Rule** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMonitorDataCollectionRuleAssociation
metadata:
  name: web-vm-linux-baseline
  org: acme-corp
  env: prod
spec:
  targetResourceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-monitoring/providers/Microsoft.Compute/virtualMachines/web-vm
  name: linux-baseline-assoc
  dataCollectionRuleId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-monitoring/providers/Microsoft.Insights/dataCollectionRules/linux-baseline
```

```shell
planton apply -f dcr-association.yaml
```

This attaches the `web-vm` virtual machine to the `linux-baseline` data collection rule under the association name `linux-baseline-assoc`. An Infra Job tracks the provisioning in real time.

### InfraChart

When the machine and the rule are Infra Components in the same chart, wire both by reference instead of pasting ARM IDs:

```yaml
spec:
  targetResourceId:
    valueFrom:
      kind: AzureVirtualMachine
      name: web-vm
      fieldPath: status.outputs.vm_id
  name: linux-baseline-assoc
  dataCollectionRuleId:
    valueFrom:
      name: linux-baseline
```

The InfraPipeline resolves the dependency graph, provisioning the machine and the rule before the association that binds them.

## Key Configuration

These are the most important decisions when configuring an Azure Monitor Data Collection Rule Association. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Rule binding or endpoint binding -- exactly one.** Set `dataCollectionRuleId` to put the machine under a rule's collection policy, or `dataCollectionEndpointId` to route the agent's configuration access through a Data Collection Endpoint (private-link estates, forced tunneling). Setting both, or neither, is rejected at validation. These are different jobs: a machine in a locked-down network typically carries one endpoint association plus one or more rule associations.

**The target reference is explicit by design.** `targetResourceId` has no default kind because several kinds can be attached -- VMs, VM scale sets, Arc-enabled servers. Reference the resource's `*_id` output with an explicit `kind` and `fieldPath`, or pass a literal ARM ID. Changing the target destroys and recreates the association.

**Name for the reader of a fleet listing.** Association names live under the machine and appear in every "what feeds this rule" listing; duplicate names on one machine collide. A convention like `{rule-shortname}-assoc` tells an operator at a glance what a machine is bound to. The name is ForceNew -- renaming recreates the association (harmless and fast, but a plan diff worth understanding). For endpoint bindings, leave `name` unset: Azure mandates `configurationAccessEndpoint` and the engines apply it.

**Layering rules on one machine is the intended composition.** A machine carrying the fleet baseline rule plus a workload-specific rule is normal; Azure evaluates all of a machine's associations together. Treat "which machines feed which rules" as association inventory -- onboarding is creating an association, offboarding is destroying it, and the rule is untouched either way.

**The agent is a separate concern, on purpose.** Configuration and agent rollout are decoupled: an association that "does nothing" usually means the Azure Monitor Agent is not running on the target, and an agent with no associations collects nothing. Check the machine's extensions first when telemetry is missing.

**Destroy order in charts.** The association dies automatically when its target machine is deleted (extension-resource semantics), but a clean chart teardown destroys associations before the rule -- Azure briefly holds rule deletion while associations reference it. Wiring both sides with valueFrom gives Planton's reverse-dependency destroy order this for free.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|---|---|---|
| Azure Virtual Machine (or scale set / Arc server ARM ID) | `targetResourceId` | `status.outputs.vm_id` |
| Azure Monitor Data Collection Rule | `dataCollectionRuleId` | `status.outputs.data_collection_rule_id` |

`dataCollectionEndpointId` takes a literal Data Collection Endpoint ARM ID -- the endpoint is not yet a catalog kind.

### What This Kind Provides

`status.outputs` carries the association's ARM ID (`data_collection_rule_association_id`) and its name on the target (`data_collection_rule_association_name`). Nothing downstream consumes an association by reference -- it is a leaf resource that binds two other resources together -- so these outputs exist for identification and import rather than composition.

## Common Patterns

**Fleet baseline onboarding** -- One association per machine per rule: every machine that joins the fleet gets an association binding it to the shared baseline rule. Start from the **Attach VM to Rule** preset.

**Machine born monitored** -- Inside a chart that deploys a VM, add the association wired by valueFrom to the chart's VM and the shared rule. The machine and its monitoring share one lifecycle, and tearing the chart down detaches cleanly.

**Private-link configuration access** -- Machines whose outbound path to Azure Monitor's public configuration endpoints is blocked carry one endpoint association (no name, no rule) alongside their rule associations. Start from the **Endpoint Access Association** preset.

## Works With

- [**Azure Virtual Machine**](/infra-catalog/azure-virtual-machine) -- the most common target; reference its `vm_id` output as `targetResourceId`.
- [**Azure Monitor Data Collection Rule**](/infra-catalog/azure-monitor-data-collection-rule) -- the collection policy this association attaches; reference its `data_collection_rule_id` output.
