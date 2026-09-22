# GcpNetworkEndpointGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpNetworkEndpointGroupSpec builds a network endpoint group (NEG): a
named set of IP:port endpoints a backend service points at instead of an
instance group. One kind, two scopes, selected by `zone`:

  ZONAL (zone set) -- endpoints inside a VPC network: VMs by instance
  (GCE_VM_IP_PORT for Application and proxy load balancers, GCE_VM_IP
  for passthrough Network Load Balancers), on-premises or other-cloud
  addresses reached over VPN or Interconnect (NON_GCP_PRIVATE_IP_PORT,
  hybrid connectivity), or internet endpoints (INTERNET_IP_PORT,
  INTERNET_FQDN_PORT) for a regional external Application Load Balancer.
  Zonal endpoints are written as one set through Google's bulk endpoint
  operation, so the manifest's list IS the group's membership.

  GLOBAL (zone empty) -- an internet NEG for a global external
  Application Load Balancer: INTERNET_IP_PORT or INTERNET_FQDN_PORT
  endpoints outside Google Cloud (an origin on another cloud or a CDN),
  each written as its own endpoint resource.

Serverless (Cloud Run, Cloud Functions, App Engine), Private Service
Connect, and REGIONAL internet groups are a different resource family and
live in GcpRegionNetworkEndpointGroup. Nearly everything here is
immutable: the name, scope, network, subnetwork, type, default port, and
description recreate the group when changed; only the endpoint list
changes in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkEndpointGroup
metadata:
  name: web-neg
spec:
  projectId:
    value: my-gcp-project
  negName: web-neg
  description: Zonal group of web VMs behind the regional Application Load Balancer
  # A zone name builds a ZONAL group inside the VPC; leave zone empty for
  # a global internet group.
  zone: us-central1-a
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
  subnetwork:
    valueFrom:
      kind: GcpSubnetwork
      name: web-subnet
      fieldPath: status.outputs.subnetwork_self_link
  # GCE_VM_IP_PORT is the default: VM IP and port, the backend of
  # Application and proxy load balancers.
  networkEndpointType: GCE_VM_IP_PORT
  defaultPort: 8080
  # The list IS the membership: written as one set, changed in place.
  endpoints:
    - instance:
        valueFrom:
          kind: GcpComputeInstance
          name: web-1
          fieldPath: status.outputs.instance_name
      ipAddress: 10.0.1.5
    - instance:
        valueFrom:
          kind: GcpComputeInstance
          name: web-2
          fieldPath: status.outputs.instance_name
      ipAddress: 10.0.1.6
      port: 8081
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.negName` | `string` |  |  |  |
| `spec.zone` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.networkEndpointType` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.defaultPort` | `int32` |  |  |  |
| `spec.endpoints` | `[]GcpNetworkEndpoint` |  |  |  |
| `spec.endpoints[].instance` | `string \| valueFrom` |  |  | GcpComputeInstance (`status.outputs.instance_name`) |
| `spec.endpoints[].ipAddress` | `string` |  |  |  |
| `spec.endpoints[].fqdn` | `string` |  |  |  |
| `spec.endpoints[].port` | `int32` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the group. A literal project ID or a
reference to a GcpProject. If omitted, the provider's default project
is used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.negName

`string`

Name of the group in GCP. 1-63 characters, lowercase letters, digits,
and hyphens, starting with a letter and not ending with a hyphen.
Defaults to metadata.name. Immutable.

- rule: neg_name must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and not ending with a hyphen

### spec.zone

`string`

The scope selector. Empty builds a GLOBAL internet network endpoint
group (INTERNET_IP_PORT or INTERNET_FQDN_PORT, for a global external
Application Load Balancer); a zone name such as us-central1-a builds a
ZONAL group in that zone (VM, hybrid, or internet endpoints, for
regional and passthrough load balancers). A zonal group needs a
network; a global one has none. Immutable: a group cannot move
between scopes or zones.

- rule: zone must be a valid GCP zone name such as us-central1-a, or empty for a global internet network endpoint group

### spec.description

`string`

Human-readable description of the group. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.networkEndpointType

`string`

The kind of endpoint every member is. Zonal: GCE_VM_IP_PORT (default;
VM IP and port -- Application and proxy load balancers), GCE_VM_IP
(VM IP only -- passthrough Network Load Balancers), NON_GCP_PRIVATE_IP_PORT
(hybrid: on-premises or other-cloud addresses reached over VPN or
Interconnect; only backend services with the EXTERNAL, EXTERNAL_MANAGED,
INTERNAL_MANAGED, or INTERNAL_SELF_MANAGED scheme in RATE or CONNECTION
mode), INTERNET_IP_PORT, INTERNET_FQDN_PORT (internet endpoints for a
regional external ALB), GCE_VM_IP_DEDICATED_BACKEND. Global:
INTERNET_IP_PORT or INTERNET_FQDN_PORT (required). Immutable.

- rule: network_endpoint_type must be one of GCE_VM_IP, GCE_VM_IP_PORT, NON_GCP_PRIVATE_IP_PORT, INTERNET_IP_PORT, INTERNET_FQDN_PORT, GCE_VM_IP_DEDICATED_BACKEND, or empty

### spec.network

`string | valueFrom`

The VPC network every endpoint belongs to (zonal groups; required
there): a network self-link or a reference to a GcpVpcNetwork.
Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom`

The subnetwork every endpoint belongs to (zonal groups; optional): a
subnetwork self-link or a reference to a GcpSubnetwork in the group's
region. VM endpoints' IPs must fall inside it. Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.defaultPort

`int32` · optional (explicit presence)

The port an endpoint listens on when its own port is unset. Unset
sends nothing (Google records none). Not for GCE_VM_IP groups.
Immutable.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.endpoints

`[]GcpNetworkEndpoint`

The group's members. The list is the whole membership: an endpoint
removed here is detached from the group in place, an endpoint added is
attached. Leave empty to create the group and attach endpoints later
(an autoscaler or another controller may own membership).

- rule: an endpoint names either an ip_address or an fqdn (INTERNET_FQDN_PORT groups), never both

### spec.endpoints[].instance

`string | valueFrom`

The VM the endpoint belongs to: a reference to a GcpComputeInstance or
its instance name in the group's zone. Zonal GCE_VM_IP and
GCE_VM_IP_PORT groups only.

- references: GcpComputeInstance (`status.outputs.instance_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeInstance, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.endpoints[].ipAddress

`string`

The endpoint's IP address. For VM endpoints a primary or alias IP of
the instance in the group's subnetwork (unset means the instance's
primary internal IP); for hybrid and internet endpoints the reachable
address, required.

- rule: ip_address must be an IPv4 or IPv6 address

### spec.endpoints[].fqdn

`string`

Fully qualified domain name of an INTERNET_FQDN_PORT endpoint
(resolved by Google at connection time).

### spec.endpoints[].port

`int32` · optional (explicit presence)

Port the endpoint listens on. Unset falls back to the group's
default_port; GCE_VM_IP groups carry no port at all.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.deletionPolicy

`string`

What destroy does to the group and its endpoints:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the group is deleted (GCP refuses while a backend
               service still uses it)
  "PREVENT" -- destroy FAILS
  "ABANDON" -- the group leaves management but keeps serving

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `zonal_type_valid`: a zonal network endpoint group's network_endpoint_type must be one of GCE_VM_IP, GCE_VM_IP_PORT, NON_GCP_PRIVATE_IP_PORT, INTERNET_IP_PORT, INTERNET_FQDN_PORT, GCE_VM_IP_DEDICATED_BACKEND (or empty for GCE_VM_IP_PORT)
- `global_type_valid`: a global (internet) network endpoint group's network_endpoint_type must be INTERNET_IP_PORT or INTERNET_FQDN_PORT; set zone for VM, hybrid, or dedicated-backend groups
- `network_required_zonal`: network is required for a zonal network endpoint group (the VPC every endpoint belongs to)
- `network_global_only_absent`: network and subnetwork apply only to a zonal network endpoint group; a global internet group has no VPC
- `vm_endpoints_name_an_instance`: every endpoint of a GCE_VM_IP or GCE_VM_IP_PORT group names its instance
- `non_vm_endpoints_carry_no_instance`: instance applies only to GCE_VM_IP and GCE_VM_IP_PORT endpoints; hybrid and internet endpoints are addresses
- `gce_vm_ip_endpoints_carry_no_port`: GCE_VM_IP endpoints carry no port (the passthrough load balancer forwards every port); remove port or use GCE_VM_IP_PORT
- `fqdn_only_in_fqdn_groups`: fqdn endpoints belong to an INTERNET_FQDN_PORT group; other groups address endpoints by ip_address
- `fqdn_groups_name_fqdns`: every endpoint of an INTERNET_FQDN_PORT group names an fqdn
- `address_endpoints_name_an_ip`: every endpoint of a NON_GCP_PRIVATE_IP_PORT or INTERNET_IP_PORT group names an ip_address

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpNetworkEndpointGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the group: the value a GcpBackendService backend names. A zonal link carries zones/{zone}, a global one says global. |
| `status.outputs.neg_name` | `string` | Name of the group as it exists in GCP -- the resolved value (metadata.name when the spec left neg_name empty). |
| `status.outputs.neg_id` | `string` | Google's unique identifier for the group. |
| `status.outputs.zone` | `string` | Zone of a zonal group; empty for a global one, so a consumer can tell the scope from the outputs alone. |
| `status.outputs.size` | `string` | Number of endpoints declared in the group -- the manifest's list is the whole membership on both scopes. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.endpoints[].instance` | GcpComputeInstance | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
