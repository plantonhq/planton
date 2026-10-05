# Azure Event Grid Namespace

Deploys an Azure Event Grid namespace -- the capacity-scaled hub of the newer Event Grid that hosts CloudEvents namespace topics and an optional MQTT broker behind one set of regional endpoints. Where the classic resources each own one endpoint, a namespace is a shared hub: throughput units set its ceiling, topics are created inside it as their own resources, and MQTT clients connect to it directly when the broker is enabled. Unlike a classic topic, a namespace is not free at rest -- it bills per throughput unit per hour from the moment it exists.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid namespace** -- the hub itself: throughput capacity (Standard SKU, the only value Azure defines, sent by the platform), network posture (public access plus Allow-only inbound IP rules), optional managed identity, and the optional MQTT broker ("topic spaces") configuration
- **Azure Tags** -- Planton-derived metadata tags merged with the manifest's `tags` (user values win on key conflicts)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Azure Subscription

- **A resource group** -- reference an AzureResourceGroup Infra Component or pass an existing group's name; changing it later replaces the namespace and every topic inside it.
- **A same-region CloudEvents custom topic** (only for MQTT routing) -- `topicSpacesConfiguration.routeTopicId` forwards MQTT messages into an AzureEventgridTopic, which must live in the same region and use the CloudEvents schema.
- **Client certificates** (only for MQTT) -- the broker authenticates MQTT clients against certificates; have the fleet's certificate convention decided before onboarding devices.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Namespace**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields: the resource group, region, capacity, network posture, and the MQTT broker block. Start from the **CloudEvents Hub** or **MQTT Broker** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridNamespace
metadata:
  name: events-hub
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: acme-events-hub
  region: eastus
  capacity: 1
```

```shell
planton apply -f eventgrid-namespace.yaml
```

This creates a pure-CloudEvents namespace at the 1-TU floor with public access on and no MQTT broker. An Infra Job tracks the provisioning in real time.

### InfraChart

When the namespace and its MQTT route topic deploy in the same InfraChart, wire them with ValueFromRef:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-events
      fieldPath: status.outputs.resource_group_name
  name: acme-mqtt-broker
  region: eastus
  topicSpacesConfiguration:
    alternativeAuthenticationNameSources:
      - ClientCertificateSubject
    routeTopicId:
      valueFrom:
        kind: AzureEventgridTopic
        name: device-telemetry
        fieldPath: status.outputs.topic_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and custom topic first, then provisions the namespace with the resolved values.

## Key Configuration

These are the most important decisions when configuring a namespace. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Decide the MQTT question before you create** -- `topicSpacesConfiguration` is the namespace's one irreversible choice: the block cannot be added, removed, or changed after create -- the provider replaces the namespace instead, dropping every namespace topic inside it. If there is ANY chance the namespace will serve MQTT clients, create it with the block (an empty block is legal and costs nothing extra); a pure-CloudEvents namespace that later needs MQTT is a migration, not an update.

**A namespace is not free the way a topic is** -- it bills per throughput unit per hour from creation, with one TU as the floor. That changes the deployment pattern: share one namespace across an environment's services (streams are cheap namespace topics) instead of stamping one namespace per service the way classic topics are stamped.

**capacity is shared, and it moves in place** -- throughput units cap ingress and egress for ALL topics and MQTT traffic in the namespace together, so one noisy publisher can starve its neighbors' streams. Watch the throttled-requests metric and raise `capacity` (1-40) in place when it climbs; no downtime, no replacement.

**The classic and namespace worlds do not mix** -- namespace topics speak CloudEvents v1.0 only and are not valid targets for an Azure Event Grid Event Subscription's `scope` arm (Azure models namespace-topic subscriptions as a different resource the Terraform provider does not ship yet). The one supported bridge today is `topicSpacesConfiguration.routeTopicId`, which forwards every MQTT message into a classic custom topic where the full delivery machinery applies.

**Network posture** -- `publicNetworkAccessEnabled` defaults to true; set it false to restrict access to private endpoints only. `inboundIpRules` (up to 128 IPv4 CIDR ranges, Allow-only) narrow public access but are ignored while public access is disabled.

**Client certificates are the MQTT identity model** -- `alternativeAuthenticationNameSources` decides which certificate field carries a client's identity (subject, DNS SAN, URI SAN, IP SAN, or email). Pick ONE convention fleet-wide before onboarding devices -- changing the source later means re-issuing or re-mapping every client. The session dials (`maximumClientSessionsPerAuthenticationName`, `maximumSessionExpiryInHours`) are per-authentication-name; devices sharing a name share that budget.

**Identity for downstream grants** -- `identity` supports system-assigned, user-assigned, or both at once. User-assigned identities can be granted on delivery targets BEFORE the namespace exists; the system-assigned principal only exists after create (its principal ID is an output).

**The name behaves like a DNS label** -- it appears in the namespace's regional hostnames and is ForceNew, as are `resourceGroup` and `region`. Prefix it with your org.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureEventgridTopic** (MQTT routing) | `topicSpacesConfiguration.routeTopicId` | `status.outputs.topic_id` |
| **AzureUserAssignedIdentity** (optional, per identity) | `identity.identityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace_id` | The namespace's ARM ID | An Azure Event Grid Namespace Topic's `namespaceId` reference |
| `identity_principal_id` | The system-assigned identity's principal ID (empty when none is configured) | Role assignments granting the namespace identity-based access on delivery targets |

`namespace_name` is also exported but only echoes the manifest's `name` back.

## Common Patterns

**One hub per environment** -- a pure-CloudEvents namespace shared by an environment's services, each owning a namespace topic inside it; replaces per-service classic topics with a single capacity-managed hub. Start from the **CloudEvents Hub** preset.

**MQTT ingestion bridge** -- the broker enabled with client-certificate authentication and every MQTT message routed into a classic custom topic, where Functions, queues, and webhooks subscribe through the classic delivery machinery. Start from the **MQTT Broker** preset.

**Locked-down hub** -- public access disabled (or narrowed with `inboundIpRules`) for namespaces reachable only from private networks; decide this alongside the consumers' network placement, since rules are ignored once public access is off.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- the group the namespace lives in
- [**Azure Event Grid Namespace Topic**](/infra-catalog/azure-eventgrid-namespace-topic) -- the CloudEvents streams created inside the namespace
- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the classic custom topic MQTT messages route into
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- identities attached for identity-based delivery access
