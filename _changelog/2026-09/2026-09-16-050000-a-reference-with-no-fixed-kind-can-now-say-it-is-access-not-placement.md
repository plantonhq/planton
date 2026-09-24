# A reference with no fixed kind can now say it is access, not placement

## What changed

- **The containment gate accepts `containment_exempt` on a kind-less reference.** A kind-less reference is a `StringValueOrRef` with no `default_kind` -- an event target, a forwarding address, a monitored resource, a private-link target -- whose kind varies per manifest and is stated in `valueFrom`. Until now `TestContainmentExemptTargetsContainerKinds` rejected the exemption on such a field as inert, so the catalog had no way to say that a rule forwarding to another event bus, or a private endpoint reaching a storage account, does not live inside what it names. The gate still rejects an exemption whose `default_kind` names a non-container kind; that case really is inert.
- **The containment-decision registry lists kind-less exemptions under their own form.** `shared/cloudresourcekind/testdata/containment_decisions.txt` gains lines of the shape `exempt <field> -> *`: a reference whose author declared it access, whichever container kind a manifest names. A kind-less reference without the exemption is not listed -- it may or may not reach a container, and the registry records authored verdicts, never possibilities.
- **A permanent fixture.** `TestCloudResourceGenericSpec.kindless_access_ref` (the test kind's live `v1alpha2`) carries the shape, so the registry holds one line of the new form even if every production field is later re-typed, and downstream consumers that must skip such an edge when they resolve nesting can pin the contract against a kind that never changes.
- **Nineteen production references declare access.** AWS: an EventBridge rule target's `arn`, a pipe's `source`, `enrichment`, and `target`, a Scheduler schedule target's `arn` -- what a rule delivers to, what a pipe reads and writes, what a schedule invokes; none is where the rule, pipe, or schedule lives. Azure: a private endpoint's `private_connection_resource_id` (the endpoint lives in its subnet and reaches the service), a flow log's `target_resource_id`, a Data Factory managed private endpoint's `target_resource_id`, a Service Bus queue's and subscription's `forward_to` and `forward_dead_lettered_messages_to`, a Search service shared private link's `target_resource_id`, a Machine Learning workspace outbound rule's `service_resource_id`, a diagnostic setting's `target_resource_id`, a metric alert's `scopes`, an autoscale setting's `target_resource_id` and `metric_resource_id`, and an Event Grid system topic's `source_resource_id`. Each field's comment says why.

## Why

Containment on a diagram is resolved by the kind of the node a manifest actually names, not by the field's default kind. So a kind-less reference into a container nests exactly like a typed one -- and most of what these nineteen fields can name is a container: an event bus, a storage account, a Key Vault, a topic, a cluster. A monitor drawn inside the vault it watches, an endpoint drawn inside the account it reaches, a queue drawn inside the topic it feeds: each is a true sentence about the wrong subject. The exemption is the one word the author has for "access"; refusing it on the fields that needed it most left the catalog unable to tell the truth about them.

## How to check

```bash
go test ./shared/cloudresourcekind/ -run 'TestContainmentDecisions|TestContainmentExemptTargetsContainerKinds'   # green; the golden carries twenty "-> *" lines
grep -c -- '-> \*$' shared/cloudresourcekind/testdata/containment_decisions.txt                                   # 20
grep -n containment_exempt catalog/aws/awseventbridgerule/v1alpha1/spec.proto catalog/azure/azureprivateendpoint/v1alpha1/spec.proto
```
