# Azure Event Hub Schema Group

Creates a schema group in an Event Hubs namespace's schema registry -- a named collection of event schemas with one serialization format and one compatibility (evolution) policy. The registry lets producers and consumers exchange compact, schema-referencing payloads instead of embedding schemas in every event: serializers register and resolve schemas against the group at runtime, through the Azure SDKs or the registry's Kafka-compatible surface. The compatibility policy is what makes schema evolution SAFE -- it controls which changes the registry accepts as new versions. Every field is fixed at creation: Azure exposes no mutable properties on a schema group.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Schema Group** -- in the referenced namespace's schema registry, with your chosen evolution policy and serialization format

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An AzureEventHubNamespace** on STANDARD or higher -- the schema registry requires it, and Azure rejects schema groups on BASIC namespaces at apply time. Reference the namespace's `namespace_id` output via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **Azure Event Hub Schema Group**, and click **Deploy**. The creation wizard walks you through the namespace attachment and the group's constitution -- the evolution policy and serialization format, with the recommended BACKWARD + Avro pairing preselected. Start from the **Backward-Compatible Avro Group** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventHubSchemaGroup
metadata:
  name: telemetry-schemas
  org: acme-corp
  env: prod
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
  schemaGroupName: telemetry-schemas
  schemaCompatibility: BACKWARD
  schemaType: AVRO
```

```shell
planton apply -f schema-group.yaml
```

This creates a BACKWARD-compatible Avro schema group named `telemetry-schemas` in the `telemetry-hubs` namespace's registry. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the group to a namespace deployed in the same InfraPipeline:

```yaml
spec:
  namespaceId:
    valueFrom:
      kind: AzureEventHubNamespace
      name: telemetry-hubs
      fieldPath: status.outputs.namespace_id
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the group in its registry.

## Key Configuration

These are the most important decisions when configuring a schema group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The evolution policy** -- `schemaCompatibility` decides which changes the registry accepts as new schema versions, and therefore which side of the stream upgrades first. BACKWARD (new schemas read old data -- upgrade consumers first) is the standard choice for analytics pipelines, because upgraded readers can replay the whole retained stream. FORWARD (old schemas read new data) suits producer-led rollouts. NONE disables checking entirely -- safe only when producers and consumers deploy together.

**The serialization format** -- `schemaType` applies to every schema in the group (a group never mixes formats). AVRO is the registry's first-class format: compact binary payloads with rich evolution semantics; JSON covers JSON-Schema payloads.

**The name** -- `schemaGroupName` is the registry coordinate serializers are configured with. Name it after the event DOMAIN (telemetry, orders, payments), not any one application.

**Nothing edits in place** -- every field is fixed at creation: Azure exposes no mutable properties on a schema group, so any change replaces it and DROPS the schemas registered inside. Treat the group as append-only infrastructure: to change the policy, create the new group, re-register its schemas, cut serializers over, then retire the old one.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventHubNamespace** | `namespaceId` | `status.outputs.namespace_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `schema_group_id` | Azure Resource Manager ID of the group | Governance tooling and ARM-level references |
| `schema_group_name` | The group's name within the registry | What producer and consumer serializers are configured with (and the Kafka schema-registry group id) |

Access to the registry is keyless: grant Entra data-plane roles (Schema Registry Reader / Schema Registry Contributor) on the namespace -- the group itself carries no secrets.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Analytics pipeline vocabulary** -- one BACKWARD Avro group per event domain: consumers upgrade first and replay the full retention window under the new schema. Start from the **Backward-Compatible Avro Group** preset.

**Producer-led rollout** -- a FORWARD group where producers add fields ahead of lagging consumers -- pick it when the write side owns the deploy train.

## Works With

- [**Azure Event Hub Namespace**](/infra-catalog/azure-event-hub-namespace) -- the STANDARD+ namespace whose registry holds the group
- [**Azure Event Hub**](/infra-catalog/azure-event-hub) -- the streams whose events conform to the group's schemas
- [**Azure Event Hub Consumer Group**](/infra-catalog/azure-event-hub-consumer-group) -- the readers whose serializers resolve schemas from the group
- [**Azure Role Assignment**](/infra-catalog/azure-role-assignment) -- keyless Schema Registry Reader/Contributor grants on the namespace
