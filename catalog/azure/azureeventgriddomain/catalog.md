# Azure Event Grid Domain

Deploys an Azure Event Grid domain -- one publishing endpoint and one pair of access keys serving many event streams (domain topics), the multi-tenant pattern. A SaaS publishes every customer's events to the same domain endpoint, naming the domain topic per event; each customer subscribes only to their own topic. A domain is free at rest -- billing is per operation -- and its topics follow either an auto-managed lifecycle (Azure creates a topic with its first subscription and deletes it with the last) or a pinned governance posture where every topic is a declared Azure Event Grid Domain Topic resource.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid domain** -- the shared publish endpoint with its domain-topic lifecycle flags, the input schema every incoming event must match (plus custom-schema envelope mappings when used), network posture (public-access dial and IP allowlist), key-auth dial, and optional managed identity
- **Azure Tags** -- Planton-derived metadata tags merged with your `tags` map (user values win on key conflicts)

Azure derives the rest from the domain: the HTTPS publish endpoint at `{name}.{region}.eventgrid.azure.net` and the primary/secondary access-key pair, all surfaced as outputs.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the domain will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef (`resourceGroup`).
- **A region-wide-unique name** -- the domain's `name` becomes a public DNS hostname (`{name}.{region}.eventgrid.azure.net`), unique across ALL Azure customers in the region; a taken name fails the deploy with a conflict. Prefix it with your org, like a storage account name.
- **A user-assigned identity (only for USER_ASSIGNED identity)** -- an AzureUserAssignedIdentity whose grants on delivery targets can be composed before the domain exists (`identity.identityIds`).

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Domain**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Multi-Tenant Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridDomain
metadata:
  name: tenant-events
  org: acme-corp
  env: prod
spec:
  resourceGroup:
    value: acme-prod-rg
  name: acme-tenant-events
  region: eastus
  inputSchema: CloudEventSchemaV1_0
```

```shell
planton apply -f eventgrid-domain.yaml
```

This creates a CloudEvents domain on Azure's auto-managed topic lifecycle: subscriptions materialize their topics, and every stream publishes through the single `acme-tenant-events` endpoint. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the domain to its dependencies:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  identity:
    type: USER_ASSIGNED
    identityIds:
      - valueFrom:
          kind: AzureUserAssignedIdentity
          name: eventgrid-delivery
          fieldPath: status.outputs.identity_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group and identity first, then provisions the domain -- and the domain topics that reference this domain's `domain_id` deploy after it.

## Key Configuration

These are the most important decisions when configuring an Event Grid domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Domain or custom topic** -- choose a domain when the number of STREAMS is the scaling problem (a topic per customer, per device fleet, per team). If you have three well-known streams, three Azure Event Grid Topic resources are simpler: separate keys, separate firewalls, no per-event topic naming. A domain trades per-stream control at the publish edge for one integration serving thousands of streams.

**The name is a public hostname** -- `name` becomes `{name}.{region}.eventgrid.azure.net`, unique across every Azure customer in the region. It is ForceNew: renaming destroys and recreates the domain AND its endpoint hostname, so every publisher must be repointed. So are `region` and `resourceGroup`.

**One input schema rules every topic** -- `inputSchema` is domain-wide and create-only: every tenant's events arrive in the same envelope, so the choice is a day-one decision for the whole tenant base. `EventGridSchema` is the default (sent explicitly by the platform); `CloudEventSchemaV1_0` is the answer for new integrations that span clouds or vendors; `CustomEventSchema` accepts your application's own JSON, mapped onto the envelope with `inputMappingFields` and `inputMappingDefaultValues` -- all of it ForceNew.

**Topic lifecycle is a governance decision** -- the auto-managed defaults (`autoCreateTopicWithFirstSubscription` and `autoDeleteTopicWithLastSubscription` both true) are frictionless: a subscription materializes its topic, the last unsubscribe removes it. Setting both false pins the topics: they exist only as declared Azure Event Grid Domain Topic resources, a subscription against an undeclared topic fails instead of materializing one, and "which streams exist and why" is answerable from code review. The flags update in place, so you can start frictionless and tighten later.

**One key set authorizes every tenant's stream** -- there are no per-topic keys; the domain's key pair publishes to EVERY topic. That is the right shape when the publisher is your own service tier. Never hand a tenant the domain key -- front untrusted publishers with your API, which stamps the topic name server-side. For production, grant the publishing tier EventGrid Data Sender via Entra ID and set `localAuthEnabled: false`: the keys still exist but stop working.

**Network posture** -- `publicNetworkAccessEnabled: false` restricts publishing to private endpoints only; `inboundIpRules` (up to 128 IPv4 CIDR ranges, Allow-only -- Azure defines no deny action on this resource) narrows the public endpoint instead. Rules only take effect while public network access is enabled.

**Managed identity for delivery, not publishing** -- add `identity` when event DELIVERY should authenticate as the domain: delivering to an Event Hub or storage queue with identity-based access, or dead-lettering to a storage account. SYSTEM_ASSIGNED is created with the domain; USER_ASSIGNED brings an AzureUserAssignedIdentity you can grant on targets BEFORE the domain exists. The provider supports exactly one flavor at a time.

**Destroying the domain destroys its topics** -- teardown removes every domain topic and subscription under it, with no safety net. In charts, topics referencing the domain by `valueFrom` get the reverse-dependency destroy order (topics before domain) naturally.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureUserAssignedIdentity** (only for USER_ASSIGNED identity) | `identity.identityIds` | `status.outputs.identity_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain_id` | The domain's Azure Resource Manager ID | AzureEventgridDomainTopic `domainId` -- every pinned topic references it; also the scope for Entra role grants |
| `endpoint` | The HTTPS URL publishers POST events to -- one endpoint for every topic; the event's topic field names the stream | Publisher configuration in application services |
| `primary_access_key` | The publish key (sent as the `aeg-sas-key` header); inert while `localAuthEnabled` is false | Publisher secrets |
| `secondary_access_key` | The rotation partner: move publishers here, regenerate the primary, move back | Zero-downtime key rotation |
| `identity_principal_id` | The system-assigned identity's principal ID (empty when no identity is configured) | Role grants on delivery targets that use identity-based access |

`domain_name` is also exported, echoing the name back for automation.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Multi-tenant SaaS eventing, auto-managed** -- one CloudEvents domain where subscription creation IS tenant onboarding: no topic registry to maintain, at the cost of "which streams exist" being answered by Azure instead of your IaC. Start from the **Multi-Tenant Domain** preset.

**Pinned-topics governance** -- both lifecycle flags off, publishing Entra-only: every tenant stream is an auditable Azure Event Grid Domain Topic resource, and topics never vanish because subscriptions churned. The posture for SaaS products with a real tenant registry or data-isolation boundaries. Start from the **Pinned-Topics Domain** preset.

**Onboard topic and subscriptions together** -- events published to a topic with no subscription are dropped, not queued. In the pinned posture, treat the domain topic and its tenant's subscription as one onboarding unit; consumption isolation lives entirely on the subscription side.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group the domain is created in
- [**Azure Event Grid Domain Topic**](/infra-catalog/azure-eventgrid-domain-topic) -- the declared per-tenant streams; each references this domain's `domain_id`
- [**Azure Event Grid Event Subscription**](/infra-catalog/azure-eventgrid-event-subscription) -- fans a domain topic's events out to queues, Functions, webhooks, and hubs
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- the pre-grantable delivery identity for USER_ASSIGNED domains
- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the single-stream alternative when a handful of well-known streams beats thousands of tenant topics
