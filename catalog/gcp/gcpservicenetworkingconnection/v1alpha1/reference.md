# GcpServiceNetworkingConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpServiceNetworkingConnectionSpec defines a private services access
connection — the VPC peering between one of your networks and a service
producer's network (for Google managed services, the
servicenetworking.googleapis.com producer). This peering is what lets
Cloud SQL, AlloyDB, Memorystore (PRIVATE_SERVICE_ACCESS mode), and
Filestore hand out private IPs from ranges you reserved inside your VPC.

The connection composes two things that already exist as first-class
resources: the VPC network being peered, and one or more INTERNAL
VPC_PEERING address ranges (GcpGlobalAddress) reserved for the producer
to carve service subnets from. One connection per (network, service) pair
is the API's cardinality — GCP rejects a second connection for the same
pair, so widening capacity later means adding ranges to THIS resource,
not creating another connection.

The connection has no name of its own in GCP: it is addressed by the
network + service pair, and the created peering appears on the VPC as
a peering named after the service (e.g. servicenetworking-googleapis-com).

Teardown ordering matters: GCP refuses to delete the connection while any
service producer still holds subnets in the reserved ranges — destroy the
Cloud SQL / AlloyDB / Memorystore instances that use private IP BEFORE
destroying this connection.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceNetworkingConnection
metadata:
  name: my-sample-psa-connection
spec:
  # GCP project used to enable the required APIs.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # The VPC network to peer with the service producer — a name or full
  # self-link (a GcpVpcNetwork reference resolves to the self-link).
  network:
    value: projects/my-gcp-project-123/global/networks/my-vpc

  # Names of INTERNAL VPC_PEERING global address ranges reserved for the
  # producer. At least one is required; append more to grow capacity.
  reservedPeeringRanges:
    - value: my-vpc-psa-range

  # DELETE keeps destroys real (destroy producer instances first);
  # PREVENT belongs under production database fleets.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.service` | `string` |  | `servicenetworking.googleapis.com` |  |
| `spec.reservedPeeringRanges` | `[]string \| valueFrom` | yes |  | GcpGlobalAddress (`status.outputs.name`) |
| `spec.updateOnCreationFail` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project used to enable the required APIs
(servicenetworking.googleapis.com and compute.googleapis.com) before the
connection is created. Can be a literal project ID or a reference to a
GcpProject resource. If omitted, the provider's default project is used.
The connection itself is addressed by the network (GCP derives the
owning project from it) — set this only when the network's project
differs from the provider's default project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.network

`string | valueFrom` · required

The VPC network to peer with the service producer. Accepts a network
name or full self-link URL; a reference resolves to the GcpVpcNetwork's
self-link. Immutable: changing it destroys and recreates the connection
(and with it, every producer resource's private connectivity).

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.service

`string`

The service producer to peer with (default
servicenetworking.googleapis.com — the producer behind Cloud SQL,
AlloyDB, Memorystore, and Filestore private IP). Third-party producers
publish their own service names. Immutable: one connection exists per
(network, service) pair.

- default: `servicenetworking.googleapis.com`
- rule: service must be a DNS-style service name such as servicenetworking.googleapis.com

### spec.reservedPeeringRanges

`[]string | valueFrom` · required

Named INTERNAL VPC_PEERING address ranges (GcpGlobalAddress resources,
referenced by NAME — not self-link or CIDR) reserved for the producer to
allocate service subnets from. At least one range is required. Mutable:
append ranges here when the producer runs out of space — GCP keeps
already-provisioned producer subnets even when the list changes, so
growth is additive and safe. Size ranges generously up front (a /16 is
the common default): producers cannot use ranges that are too fragmented.

- references: GcpGlobalAddress (`status.outputs.name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalAddress, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.updateOnCreationFail

`bool`

Recovery lever for a known API wrinkle: when a connection for this
(network, service) pair already exists outside of Planton's management
(e.g. created by gcloud or a console flow), the create call fails with
"Cannot modify allocated ranges". Setting this to true converts that
failure into an in-place update of the existing connection's reserved
ranges, adopting it instead of erroring. Leave false unless you are
deliberately taking over a pre-existing connection.

### spec.deletionPolicy

`string`

Deletion policy for the connection — what happens when this resource
is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the peering is deleted (GCP refuses while any service
               producer still holds subnets in the reserved ranges —
               destroy the Cloud SQL / AlloyDB / Memorystore instances
               first; see the teardown note above)
  "PREVENT" -- destroy FAILS; protects the private connectivity every
               producer instance on this network depends on
  "ABANDON" -- the connection is removed from management but keeps
               serving private IPs in GCP — the historical workaround
               for stuck teardowns, at the cost of an unmanaged peering

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpServiceNetworkingConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.peering` | `string` | Name of the VPC peering GCP created on the network for this connection (e.g. servicenetworking-googleapis-com). This is the peering an operator looks for on the network's peerings list when auditing private services access, and the handle for peering-level route settings. |
| `status.outputs.network` | `string` | Self-link of the peered VPC network as the connection resolved it — confirms which network the producer is attached to without chasing the reference chain. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.reservedPeeringRanges` | GcpGlobalAddress | `status.outputs.name` |

## See Also

- [Overview](../README.md)
