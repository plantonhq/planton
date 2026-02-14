# Affinity Server Group

This preset creates a server group with the affinity policy. Instances placed in this group are scheduled on the same physical hypervisor, minimizing network latency between them. Use this when low-latency inter-instance communication is more important than fault isolation.

## When to Use

- Tightly coupled applications that communicate heavily over the network (e.g., HPC, distributed computing)
- Batch processing pipelines where data locality reduces transfer time
- Dev/test environments where co-location reduces networking overhead

## Key Configuration Choices

- **Affinity** (`policy: affinity`) -- strict scheduling constraint; all instances land on the same hypervisor
- **Immutable** -- all fields are ForceNew; changing the policy requires recreating the group

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`.

## Related Presets

- **01-anti-affinity** -- Use instead when fault isolation is more important than low-latency communication
