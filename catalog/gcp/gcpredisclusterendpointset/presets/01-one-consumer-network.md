# One Consumer Network

## Use Case

Reach a Memorystore for Redis Cluster from a VPC its own connectivity automation cannot place endpoints in -- a consumer VPC in another project, or a network without a service connection policy. The cluster publishes service attachments; the consumer builds one reserved address and one forwarding rule per attachment; this set registers them.

## When to Use

- A cluster shared across projects or VPCs
- A consumer network where a `GcpServiceConnectionPolicy` is not wanted
- Full control over the endpoint addresses clients connect to

## What This Creates

- A registration on `orders-cache` of one consumer network's connections: the discovery and primary forwarding rules, each with its reserved address, in `consumer-vpc`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `cluster` | `orders-cache` | Your cluster, created without `pscConfigs`. |
| `endpoints[].connections` | discovery + primary | Add a reader connection when the cluster has replicas (`reader_service_attachment`); add an `endpoints[]` entry per additional consumer VPC. |
| `endpoints[].connections[].projectId` | the rule's project | Set only when the forwarding rule lives in a different project than the cluster. |
| `deletionPolicy` | `DELETE` | `PREVENT` for a cluster other VPCs depend on. |

Each connection's five fields are outputs of the blocks that built it -- the same forwarding rule referenced twice (`self_link` and `psc_connection_id`), the address, the network, and the cluster's attachment handle. Removing a connection here breaks its forwarding rule; delete both in one change.
