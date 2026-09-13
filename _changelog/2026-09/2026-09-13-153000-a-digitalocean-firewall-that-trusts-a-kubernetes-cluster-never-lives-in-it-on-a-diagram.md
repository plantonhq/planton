# A DigitalOcean firewall that trusts a Kubernetes cluster never lives in it on a diagram

## What changed

- **The cloud firewall's Kubernetes-cluster rule references are containment-exempt.** `DigitalOceanFirewallInboundRule.source_kubernetes_ids` names the clusters whose traffic a rule admits; `DigitalOceanFirewallOutboundRule.destination_kubernetes_ids` names the clusters the protected Droplets may reach. A DigitalOcean firewall is an account-level rule set applied to Droplets by id or by tag; the cluster in a rule is a trusted source or a permitted destination, never the firewall's home. Until now both references were placement by omission, so a firewall whose one rule admitted traffic from a DOKS cluster would have been drawn inside that cluster.
- **The database firewall's trusted-cluster reference is containment-exempt.** `DigitalOceanDatabaseFirewallSpec.kubernetes_cluster_ids` is the managed database cluster's inbound trusted-sources list; a Kubernetes cluster named there is a client the database admits. The rule set stays with its database cluster on a diagram, with a line to each trusted cluster.
- The containment-decision registry (`shared/cloudresourcekind/testdata/containment_decisions.txt`) moves exactly those three lines from `contained` to `exempt`; nothing else in the registry moved.

## Why

`container_kind` says a kind is a box other resources nest inside, and `containment_exempt` says a reference into such a box is access, not placement. The only references into a DigitalOcean Kubernetes cluster that are genuine placement are a node pool's `cluster` -- a pool is API-addressed under its cluster and cannot exist first. Everything else that names a cluster today is a guard consulting it: the cloud firewall's rules and the database firewall's trusted sources. Those are the two kinds a DigitalOcean operator writes to let a cluster's workloads reach a fleet or a database, and each would have been drawn inside the very cluster it merely admits.

## How to check

```bash
go test ./shared/cloudresourcekind/... -run TestContainmentDecisions   # green; the golden carries the three exempt lines
grep -n containment_exempt catalog/digitalocean/digitaloceanfirewall/v1alpha1/spec.proto catalog/digitalocean/digitaloceandatabasefirewall/v1alpha1/spec.proto
```
