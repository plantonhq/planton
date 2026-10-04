# AzureEventgridNamespaceTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureEventgridNamespaceTopicSpec** defines a topic INSIDE an Azure
Event Grid namespace (AzureEventgridNamespace) -- one named
CloudEvents stream. A namespace holds many topics with independent
lifecycles: publishers and teams create and delete their own topics
against the shared namespace without touching it or each other,
exactly like consumer groups on an Event Hub.

Azure pins two properties the provider does not expose: the event
schema is always CloudEvents v1.0 and the publisher type is always
"Custom" (your applications publish; Azure services publish to
SYSTEM topics instead). Both engines send exactly those values.

## Example

```yaml
# Example for docs and offline validation: one CloudEvents stream
# inside a namespace, with the retention window set explicitly.
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridNamespaceTopic
metadata:
  name: test-eventgrid-namespace-topic
  id: test-eventgrid-namespace-topic
  org: test-org
  env: test
spec:
  namespaceId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventGrid/namespaces/acme-events-hub
  name: orders
  eventRetentionInDays: 3
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespaceId` | `string \| valueFrom` | yes |  | AzureEventgridNamespace (`status.outputs.namespace_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.eventRetentionInDays` | `int32` |  | `7` |  |

## Field Details

### spec.namespaceId

`string | valueFrom` · required

The Event Grid namespace the topic is created in, by ARM ID.
Reference an AzureEventgridNamespace output or pass a literal ID.

**ForceNew**: changing this destroys and recreates the topic.

- references: AzureEventgridNamespace (`status.outputs.namespace_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridNamespace, name: <that resource's name>, fieldPath: status.outputs.namespace_id}} -- a bare string does not parse

### spec.name

`string` · required

The topic's name -- 3-50 characters; letters, numbers, and
hyphens. Unique within the namespace only (topics have no public
hostname of their own).

**ForceNew**: changing this destroys and recreates the topic.

- rule: Namespace topic names must be 3-50 characters of letters, numbers, and hyphens
- rule: {"required":true}

### spec.eventRetentionInDays

`int32` · optional (explicit presence)

How long published events are retained for delivery, in days,
1-7. This is the topic's ONLY updatable property. Defaults to 7
(the provider's default) -- the platform sends the value
explicitly.

- default: `7`
- rule: {"int32":{"lte":7,"gte":1}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridNamespaceTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace_topic_id` | `string` | The namespace topic's Azure Resource Manager ID ({namespace_id}/topics/{name}). |
| `status.outputs.namespace_topic_name` | `string` | The namespace topic's name. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespaceId` | AzureEventgridNamespace | `status.outputs.namespace_id` |

## See Also

- [Overview](../README.md)
