# GcpServerlessVpcConnector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpServerlessVpcConnectorSpec defines a Serverless VPC Access connector
(`google_vpc_access_connector`) — the managed bridge that lets serverless
workloads (Cloud Functions, Cloud Run, App Engine) send traffic into a VPC
network. The connector runs a small fleet of forwarding instances inside
the VPC; serverless egress configured to use it reaches private IPs
(Cloud SQL private IP, Memorystore, internal load balancers) as if the
workload lived in the network.

A connector is regional and attaches to exactly ONE placement:
  - network + ip_cidr_range — the connector carves a dedicated /28 out of
    the network's address space (the range must not overlap any existing
    subnet or route), or
  - subnet — the connector occupies an EXISTING /28 subnetwork created
    for it (required for Shared VPC, where the range lives in the host
    project; also the only mode that supports choosing the subnet's
    project).
Exactly one of the two modes must be set.

Consumers attach by full resource name (projects/*/locations/*/
connectors/*) — the `self_link` stack output. One connector serves many
functions/services in its region; it is shared infrastructure, not
per-workload.

## Example

```yaml
# Development manifest for GcpServerlessVpcConnector — exercises network
# placement with explicit scaling for offline plan verification.
#
# Usage: planton tofu plan --manifest catalog/gcp/gcpserverlessvpcconnector/e2e/manifest.yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServerlessVpcConnector
metadata:
  name: hack-vpc-connector
  id: gcpvpcconn-hack-001
  org: planton-dev
  env: dev
spec:
  region: us-central1
  connectorName: hack-egress
  network:
    value: hack-vpc
  ipCidrRange: 10.8.0.0/28
  machineType: e2-micro
  minInstances: 2
  maxInstances: 4

  # DELETE keeps destroys real; PREVENT belongs on the region's
  # production serverless egress path.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.connectorName` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.ipCidrRange` | `string` |  |  |  |
| `spec.subnet` | `GcpServerlessVpcConnectorSubnet` |  |  |  |
| `spec.subnet.name` | `string \| valueFrom` | yes |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.subnet.projectId` | `string` |  |  |  |
| `spec.machineType` | `string` |  | `e2-micro` |  |
| `spec.minInstances` | `int32` |  |  |  |
| `spec.maxInstances` | `int32` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the connector is created in. Accepts a literal project
ID or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region the connector lives in, e.g. "us-central1". Immutable. Serverless
workloads can only use a connector in their own region.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.connectorName

`string`

Name of the connector in GCP. Immutable. If not specified, defaults to
metadata.name. Maximum 25 characters (a GCP limit stricter than most
resource names): lowercase letters, digits, and hyphens, starting with
a letter and ending with a letter or digit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"25","pattern":"^[a-z]([a-z0-9-]{0,23}[a-z0-9])?$"}}

### spec.network

`string | valueFrom`

VPC network to attach to, for network placement. Accepts a literal
network name or a reference to a GcpVpcNetwork resource. Requires
ip_cidr_range; mutually exclusive with subnet. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.ipCidrRange

`string`

Unused /28 range the connector's instances occupy, in CIDR notation
(e.g. "10.8.0.0/28") — network placement only. GCP requires exactly a
/28; the range must not overlap any existing subnet, peered range, or
route in the network. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/28$"}}

### spec.subnet

`GcpServerlessVpcConnectorSubnet`

Existing /28 subnetwork the connector occupies, for subnet placement.
The required mode on Shared VPC (the subnet lives in the host project).
Mutually exclusive with network/ip_cidr_range. Immutable.

### spec.subnet.name

`string | valueFrom` · required

Name of the /28 subnetwork the connector occupies (the short name, not
a path — GCP resolves it in the connector's region). The subnetwork
must be dedicated to this connector: exactly /28, no other workloads.
Accepts a literal name or a reference to a GcpSubnetwork resource.
Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.subnet.projectId

`string`

Project that owns the subnetwork. Only needed on Shared VPC, where the
subnet lives in the host project; defaults to the connector's project.

### spec.machineType

`string`

Machine type of the connector's forwarding instances. Larger types
carry more throughput per instance (f1-micro ~100 Mbps, e2-micro
~200 Mbps, e2-standard-4 ~1 Gbps class). Mutable in place — a
fleet-wide capacity lever that needs no replacement.

- default: `e2-micro`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","f1-micro","e2-micro","e2-standard-4"]}}

### spec.minInstances

`int32` · optional (explicit presence)

Minimum instances kept running (2–9; GCP defaults to 2). Must be
strictly lower than max_instances. The connector never scales below
this floor — and note the fleet NEVER scales in on its own: after a
burst it stays at the high-water mark until the instances are manually
reduced. Decreasing this value forces the connector to be REPLACED
(brief egress outage); increasing it applies in place.

- rule: {"int32":{"lte":9,"gte":2}}

### spec.maxInstances

`int32` · optional (explicit presence)

Maximum instances the connector scales to (3–10; GCP defaults to 10).
Must be strictly higher than min_instances. Sizes the throughput
ceiling: instances × per-instance throughput for the machine type.
Decreasing this value forces the connector to be REPLACED (brief
egress outage); increasing it applies in place.

- rule: {"int32":{"lte":10,"gte":3}}

### spec.deletionPolicy

`string`

Deletion policy for the connector — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the connector is deleted (GCP tears down the forwarding
               fleet, ~3-5 minutes; serverless egress through it stops)
  "PREVENT" -- destroy FAILS; protects the private-egress path every
               function and service in the region may depend on
  "ABANDON" -- the connector is removed from management but keeps
               forwarding traffic in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `placement.network_xor_subnet`: choose exactly one placement: network (with ipCidrRange) to carve a new /28, or subnet to use an existing /28 subnetwork
- `placement.network_requires_cidr`: network placement carves a dedicated range out of the VPC — set ipCidrRange to an unused /28 (e.g. 10.8.0.0/28)
- `placement.cidr_requires_network`: ipCidrRange only applies to network placement — set network, or use subnet placement instead
- `scaling.min_lt_max`: minInstances must be strictly lower than maxInstances — the gap is the connector's room to scale

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpServerlessVpcConnector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Short name of the connector as created in GCP. |
| `status.outputs.self_link` | `string` | Fully qualified resource name (projects/{project}/locations/{region}/connectors/{name}) — the handle serverless workloads (Cloud Functions, Cloud Run, App Engine) set as their VPC connector. |
| `status.outputs.state` | `string` | State of the connector (READY, CREATING, DELETING, ERROR, UPDATING). |
| `status.outputs.region` | `string` | Region the connector lives in (plain region name). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.subnet.name` | GcpSubnetwork | `status.outputs.subnetwork_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudFunction | `spec.serviceConfig.vpcConnector` | `status.outputs.self_link` |
| GcpCloudRun | `spec.vpcAccess.connector` | `status.outputs.self_link` |
| GcpCloudRunJob | `spec.template.vpcAccess.connector` | `status.outputs.self_link` |
| GcpCloudRunWorkerPool | `spec.vpcAccess.connector` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
