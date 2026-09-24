# GcpDatastreamPrivateConnection

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDatastreamPrivateConnectionSpec defines a Datastream private
connection (`google_datastream_private_connection`) -- the network link
that lets Datastream reach databases with no public address. Connection
profiles in the same project and location use it through their
private_connection reference, so one link serves every source in a VPC.

Set exactly ONE of vpc_peering_config or psc_interface_config.

Everything is immutable except labels and deletion_policy: a change
replaces the connection, and every profile using it must be re-pointed.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamPrivateConnection
metadata:
  name: data-vpc
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Data VPC
  vpcPeeringConfig:
    vpc:
      value: projects/my-gcp-project/global/networks/data
    # A free /29 no subnet, peering, or route in the VPC uses.
    subnet: 10.200.0.0/29
  labels:
    team: data
  deletionPolicy: FORCE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.privateConnectionId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.vpcPeeringConfig` | `GcpDatastreamPrivateConnectionVpcPeeringConfig` |  |  |  |
| `spec.vpcPeeringConfig.vpc` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.vpcPeeringConfig.subnet` | `string` | yes |  |  |
| `spec.pscInterfaceConfig` | `GcpDatastreamPrivateConnectionPscInterfaceConfig` |  |  |  |
| `spec.pscInterfaceConfig.networkAttachment` | `string` | yes |  |  |
| `spec.createWithoutValidation` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the private connection lives in: a literal project ID
or a GcpProject reference. If omitted, the provider's default project
is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Datastream region, e.g. "us-central1". Profiles and streams that
use this connection live in the same region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.privateConnectionId

`string`

The private connection's ID. Defaults to metadata.name. Immutable.

### spec.displayName

`string`

The name shown in the console. Defaults to metadata.name. Immutable.

### spec.vpcPeeringConfig

`GcpDatastreamPrivateConnectionVpcPeeringConfig`

Peer Datastream's network with a VPC.

### spec.vpcPeeringConfig.vpc

`string | valueFrom` · required

The VPC Datastream peers with, as projects/{project}/global/networks/{name}
-- a GcpVpcNetwork reference (its network_id output) or a literal in
that form. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.vpcPeeringConfig.subnet

`string` · required

A free /29 range in the VPC's address space for Datastream's side of
the peering, e.g. "10.200.0.0/29" -- it must overlap no subnet, no
other peering, and no range routed into the VPC. Immutable.

- rule: subnet must be a free IPv4 CIDR of /29, e.g. 10.200.0.0/29
- rule: {"required":true}

### spec.pscInterfaceConfig

`GcpDatastreamPrivateConnectionPscInterfaceConfig`

Reach a VPC through a Private Service Connect interface.

### spec.pscInterfaceConfig.networkAttachment

`string` · required

The network attachment, as
projects/{project}/regions/{region}/networkAttachments/{name}, in the
private connection's region. It must accept connections from
Datastream's service project. Immutable.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/regions/[^/]+/networkAttachments/[^/]+$"}}

### spec.createWithoutValidation

`bool`

Skip Google's checks at create (for example, that the /29 is free).
Useful when the network is still being built; a real conflict then
surfaces when a profile first connects. Immutable.

### spec.labels

`map<string, string>`

Labels on the private connection. The platform attribution labels are
added on top and win on a key conflict.

### spec.deletionPolicy

`string`

What happens to the private connection when this resource is
destroyed:
  "" / "FORCE" -- deleted together with the routes Datastream created
                  on it (the provider's default)
  "DELETE"     -- deleted without force; fails while routes remain
  "PREVENT"    -- destroy fails
  "ABANDON"    -- it leaves management and stays in GCP
Delete the profiles that use it first; Google refuses to delete a
connection a profile still references.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON, FORCE

## Validation Rules

- `spec.exactly_one_connectivity`: set exactly one of vpc_peering_config or psc_interface_config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDatastreamPrivateConnection, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/privateConnections/{private_connection_id}. The value a GcpDatastreamConnectionProfile's private_connection takes. |
| `status.outputs.private_connection_id` | `string` | The private connection's ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.vpcPeeringConfig.vpc` | GcpVpcNetwork | `status.outputs.network_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDatastreamConnectionProfile | `spec.privateConnection` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
