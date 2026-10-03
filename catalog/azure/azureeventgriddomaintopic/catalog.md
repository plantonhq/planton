# Azure Event Grid Domain Topic

Deploys one named event stream (domain topic) inside an Azure Event Grid domain -- the per-tenant mailbox of the multi-tenant pattern. Publishers address it by stamping the topic's name into events sent to the DOMAIN's endpoint (a domain topic has no endpoint or keys of its own), and subscribers attach event subscriptions to the topic's own ARM ID. Declaring topics explicitly is the governance posture: tenants join and leave as reviewable IaC without touching the domain or each other.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Grid domain topic** -- a named stream under the domain (`{domain_id}/topics/{name}`), the scope event subscriptions attach to

The topic is pure addressing: it carries no endpoint, no keys, and no configuration beyond its name, and it is free at rest -- operations are billed on the domain.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Event Grid domain** to hold the topic -- reference an AzureEventgridDomain's `domain_id` output or provide the domain's literal ARM ID (`domainId`).
- **The domain's lifecycle flags, for the full governance posture** -- a declared topic persists regardless of the domain's `autoCreate`/`autoDelete` flags, but only setting both false on the domain stops Azure from materializing UNDECLARED topics the moment someone creates a subscription.

## Deploy

### Console

Open the deployment store, find **Azure Event Grid Domain Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Tenant Stream** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridDomainTopic
metadata:
  name: customer-fabrikam-stream
  org: acme-corp
  env: prod
spec:
  domainId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/acme-prod-rg/providers/Microsoft.EventGrid/domains/acme-tenant-events
  name: customer-fabrikam
```

```shell
planton apply -f eventgrid-domain-topic.yaml
```

This declares the `customer-fabrikam` stream inside the `acme-tenant-events` domain -- publishers stamp exactly that name into the event's topic field. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the topic to its domain:

```yaml
spec:
  domainId:
    valueFrom:
      kind: AzureEventgridDomain
      name: tenant-events
      fieldPath: status.outputs.domain_id
  name: customer-fabrikam
```

The InfraPipeline resolves the dependency graph, deploys the domain first, then declares the topic under it -- and the reverse-dependency destroy order (topics before domain) produces clean teardowns naturally.

## Key Configuration

These are the most important decisions when configuring a domain topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is the publisher contract** -- publishers stamp this exact value (`name`, 3-128 letters, numbers, and hyphens, unique within the domain) into every event's topic field; it is API surface, not decoration. Name it after the stable identity it carries (`customer-fabrikam`, `orders`), never after infrastructure or environments -- the domain already scopes those. The name is create-only: renaming replaces the topic and drops its subscriptions, so treat a rename like an API version change.

**The domain reference, wired or literal** -- `domainId` defaults to referencing an AzureEventgridDomain's `domain_id` output, so the domain and its topics compose in one manifest set with only the resource name needed. It is ForceNew: moving a topic to a different domain destroys and recreates it.

**Pin topics where tenant identity is real** -- declared topics turn "which streams exist" into reviewable IaC, the right posture when topics map to customers, billing entities, or data-isolation boundaries. Pair this kind with the domain's lifecycle flags set false; otherwise Azure happily materializes undeclared topics the moment someone creates a subscription, and your inventory lies.

**Nothing else to configure** -- publishing auth, network posture, and input schema all live on the domain; the topic inherits everything. If you find yourself wanting per-stream keys or a per-stream firewall, you wanted a standalone Azure Event Grid Topic instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureEventgridDomain** | `domainId` | `status.outputs.domain_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain_topic_id` | The topic's Azure Resource Manager ID (`{domain_id}/topics/{name}`) | The scope an AzureEventgridEventSubscription attaches to -- the tenant's delivery wiring |
| `domain_topic_name` | The topic's name | The value publishers stamp into the event's topic field |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tenant onboarding as one chart move** -- a topic with no subscriptions drops every event silently (Event Grid stores nothing at the topic), so declare the domain topic AND its dead-letter-backed subscription(s) in the same set, references flowing from `domain_topic_id`. Start from the **Tenant Stream** preset.

**Pinning an auto-managed domain's streams** -- declaring a topic pins it regardless of the domain's lifecycle flags, so you can convert an auto-managed domain to governed streams tenant by tenant before flipping the domain's flags.

**Offboarding as teardown** -- destroying a tenant's topic instantly stops delivery to that tenant's handlers (its subscriptions go with it) while every sibling stream continues untouched; that independence is exactly why the topic is its own resource. Delivery failures against a deleted topic in the domain's metrics mean a publisher is still stamping its name.

## Works With

- [**Azure Event Grid Domain**](/infra-catalog/azure-eventgrid-domain) -- the shared endpoint and key pair this topic lives under; provides `domain_id`
- [**Azure Event Grid Event Subscription**](/infra-catalog/azure-eventgrid-event-subscription) -- attaches to this topic's `domain_topic_id` to deliver the tenant's events to queues, Functions, webhooks, and hubs
- [**Azure Event Grid Topic**](/infra-catalog/azure-eventgrid-topic) -- the standalone alternative when a stream needs its own endpoint, keys, and firewall
