# Azure Event Grid Topic

Deploys an Azure Event Grid custom topic -- the HTTPS endpoint an application publishes its own events to, from which any number of event subscriptions fan events out to handlers. Publishers POST to `{name}.{region}.eventgrid.azure.net` authenticated with an access key or Microsoft Entra ID; the input schema every incoming event must match is fixed at creation. A topic is free at rest -- billing is per operation. A topic is a single stream under one endpoint and one pair of keys; for many logical streams behind one endpoint (the multi-tenant pattern), use an Event Grid domain instead.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid topic** -- the publish endpoint with its pair of access keys, the input schema (with envelope mappings for custom-schema topics), network posture (public access plus Allow-only inbound IP rules), key-auth switch, and optional managed identity for secured delivery
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **A resource group** -- reference an AzureResourceGroup Infra Component or pass an existing group's name.
- **A region-wide-unique name** -- the name becomes the topic's public DNS hostname, unique across ALL Azure customers in the region; a taken name fails the deploy with a conflict.
- **EventGrid Data Sender grants** (only before disabling key auth) -- every publisher needs the role on the topic BEFORE `localAuthEnabled: false` deploys, or publishing breaks at cutover.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the resource group, region, input schema, network posture, and identity. Start from the **CloudEvents Topic** or **Locked-Down Topic** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridTopic
metadata:
  name: app-events
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: acme-app-events
  region: eastus
  inputSchema: CloudEventSchemaV1_0
```

```shell
planton apply -f eventgrid-topic.yaml
```

This creates a CloudEvents-input custom topic with public access and key auth on (Azure's defaults) -- publishers POST to the endpoint output with the `aeg-sas-key` header. An Infra Job tracks the provisioning in real time.

### InfraChart

When the topic and its subscriptions deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: acme-secure-events
  region: eastus
  inputSchema: CloudEventSchemaV1_0
  localAuthEnabled: false
  identity:
    type: SYSTEM_ASSIGNED
```

The InfraPipeline resolves the dependency graph, deploys the resource group first, then provisions the topic -- and downstream event subscriptions reference this topic's `topic_id` in the same chart.

## Key Configuration

These are the most important decisions when configuring a custom topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pick inputSchema like a public API version** -- the schema is create-only, and a replaced topic means a new endpoint hostname and repointed publishers. Default to `CloudEventSchemaV1_0` for anything new (the vendor-neutral standard); keep `EventGridSchema` only when existing tooling already speaks it; reserve `CustomEventSchema` for retrofitting a producer whose JSON cannot change, mapping its fields onto the envelope with `inputMappingFields` and `inputMappingDefaultValues` (both create-only too). Subscriptions must deliver a schema the topic can map -- a CloudEvents-input topic cannot deliver `EventGridSchema`.

**The name is a region-wide claim** -- topic names share the region's public DNS namespace with every other Azure customer. Namespace them like a storage account (`acme-app-events`, not `orders`) -- both to avoid conflicts and to keep the hostname self-describing in publisher configs and firewall logs. Name, region, and resource group are all ForceNew.

**Publish auth: keys to start, Entra to finish** -- keys are the fast path and rotate cleanly (two keys: move publishers to the secondary, regenerate the primary, move back). For production, grant publishers the EventGrid Data Sender role and set `localAuthEnabled: false` -- the keys still exist but stop working, and leaked-key risk drops to zero. Flip it only after every publisher is on Entra; the switch is immediate.

**No subscribers means silent drops, not queues** -- a topic does not store events: anything published with no matching subscription is evaluated and discarded. "Deploy topic, verify publish, wire subscriptions later" is a trap in production cutovers -- create at least the dead-letter-backed subscription first, then cut publishers over.

**The IP firewall gates only the public path** -- `inboundIpRules` (up to 128 Allow-only CIDR ranges) apply while `publicNetworkAccessEnabled` is true; disabling public access ignores them and leaves private endpoints as the only publish path. For the common locked-down-but-public shape (CI publishers with known egress), keep public access on and list the egress CIDRs.

**Identity is for delivery, not publishing** -- the topic's managed identity authenticates OUTBOUND delivery (to identity-secured Event Hubs, queues, and dead-letter storage); it has nothing to do with who may publish. This resource supports exactly one flavor at a time -- system-assigned or user-assigned, no combined mode. Enable it when subscriptions will target locked-down destinations, and grant it there before creating those subscriptions.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** (optional, per identity) | `identity.identityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `topic_id` | The topic's ARM ID | An Azure Event Grid Event Subscription's `scope`; an Azure Event Grid Namespace's MQTT `routeTopicId` |
| `endpoint` | The HTTPS publish endpoint (`https://{name}.{region}.eventgrid.azure.net/api/events`) | Publisher application configuration |
| `primary_access_key` | The primary publish key (sensitive; the `aeg-sas-key` header) -- inert while `localAuthEnabled` is false | Publisher authentication |
| `secondary_access_key` | The rotation partner (sensitive): move publishers here, regenerate the primary, move back | Key rotation |
| `identity_principal_id` | The system-assigned identity's principal ID (empty when none is configured) | Role assignments granting the topic identity-based access on delivery targets |

`topic_name` is also exported but only echoes the manifest's `name` back.

## Common Patterns

**CloudEvents topic** -- the vendor-neutral default for new application eventing: CloudEvents 1.0 input, public access and key auth on for the fast integration path. Start from the **CloudEvents Topic** preset.

**Locked-down publish edge** -- Entra-only authentication (keys disabled), an IP allowlist gating the public path, and a system-assigned identity for secured delivery targets; the production shape once every publisher holds the Data Sender role. Start from the **Locked-Down Topic** preset.

**Topic per stream, domain for tenants** -- stamp one free-at-rest topic per event stream; when many per-customer streams should share one endpoint and one key pair, reach for an Event Grid domain instead of multiplying topics.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the topic lives in
- [**Azure Event Grid Event Subscription**](/infra-catalog/azure-eventgrid-event-subscription) -- fans the topic's events out to queues, Functions, webhooks, and hubs
- [**Azure Event Grid Domain**](/infra-catalog/azure-eventgrid-domain) -- the multi-tenant alternative: many streams behind one endpoint
- [**Azure Event Grid Namespace**](/infra-catalog/azure-eventgrid-namespace) -- routes its MQTT broker's messages into this topic via `routeTopicId`
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the user-assigned delivery identity option
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- EventGrid Data Sender grants for Entra publishers and delivery-target grants for the topic's identity
