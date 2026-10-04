# AzureEventgridDomainTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureEventgridDomainTopicSpec** defines one named event stream
inside an Azure Event Grid domain -- the per-tenant mailbox of the
multi-tenant pattern. Publishers address it by naming the topic in
events sent to the DOMAIN's endpoint (a domain topic has no endpoint
of its own), and subscribers attach event subscriptions to it.

The domain topic is a first-class resource because it is
many-per-domain with its own lifecycle: each topic typically belongs
to one tenant, and tenants join and leave without touching the
domain or each other -- exactly like AzureEventHubConsumerGroup on a
shared Event Hub. Declare topics explicitly (with the domain's
auto_create/auto_delete set false) for the governance posture where
topics exist only by decision; leave the domain on its auto-managed
defaults when subscriptions should materialize topics on demand.

## Example

```yaml
# Deep-shape example for docs and offline validation: one tenant's
# stream inside a domain. References are literal ARM ids so the
# manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureEventgridDomainTopic
metadata:
  name: test-eventgrid-domain-topic
  id: test-eventgrid-domain-topic
  org: test-org
  env: test
spec:
  domainId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.EventGrid/domains/test-org-tenant-events
  # The name publishers stamp into every event's topic field.
  name: customer-fabrikam
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.domainId` | `string \| valueFrom` | yes |  | AzureEventgridDomain (`status.outputs.domain_id`) |
| `spec.name` | `string` | yes |  |  |

## Field Details

### spec.domainId

`string | valueFrom` · required

The Event Grid domain the topic lives in, by ARM resource ID.
Takes the domain's full ARM ID; defaults to referencing an
AzureEventgridDomain's domain_id output so the domain and its
topics compose in one manifest set.

**ForceNew**: changing this destroys and recreates the topic.

- references: AzureEventgridDomain (`status.outputs.domain_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureEventgridDomain, name: <that resource's name>, fieldPath: status.outputs.domain_id}} -- a bare string does not parse

### spec.name

`string` · required

The topic's name, unique within the domain -- 3-128 characters;
letters, numbers, and hyphens. Publishers put exactly this name in
the event's topic field; name it after the tenant or stream it
carries ("customer-fabrikam", "orders").

**ForceNew**: changing this destroys and recreates the topic (a
brief gap for its subscriptions, nothing else -- the domain and
sibling topics are untouched).

- rule: Domain topic names must be 3-128 characters of letters, numbers, and hyphens
- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureEventgridDomainTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.domain_topic_id` | `string` | The domain topic's Azure Resource Manager ID ({domain_id}/topics/{name}) -- the scope event subscriptions attach to. |
| `status.outputs.domain_topic_name` | `string` | The topic's name -- the value publishers put in the event's topic field. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.domainId` | AzureEventgridDomain | `status.outputs.domain_id` |

## See Also

- [Overview](../README.md)
