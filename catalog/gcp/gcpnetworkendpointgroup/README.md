# GCP Network Endpoint Group

Builds a network endpoint group (NEG) — a named set of IP:port endpoints a backend service points at instead of an instance group. One kind, two scopes: set `zone` for a ZONAL group inside a VPC (VMs by instance, on-premises or other-cloud addresses reached over VPN or Interconnect, or internet endpoints for a regional external Application Load Balancer); leave it empty for a GLOBAL internet group (an origin outside Google Cloud behind a global external Application Load Balancer). The manifest's endpoint list is the group's whole membership. Serverless, Private Service Connect, and regional internet groups live in `GcpRegionNetworkEndpointGroup`.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions exactly one of:

- **Zonal group** (`zone` set) -- the `compute_network_endpoint_group` in that zone on your network, plus its membership written as ONE set through Google's bulk endpoint operation (`compute_network_endpoints`) when `endpoints` is non-empty
- **Global internet group** (`zone` empty) -- the `compute_global_network_endpoint_group`, plus one `compute_global_network_endpoint` per entry in `endpoints`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/compute.networkAdmin` on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Networks

- **Zonal groups need a network** -- declare it with `GcpVpcNetwork` (and optionally a `GcpSubnetwork` the endpoint IPs fall in); VM endpoints reference `GcpComputeInstance` resources in the same zone.
- **Global groups have no VPC** -- they name internet endpoints only (`INTERNET_IP_PORT` or `INTERNET_FQDN_PORT`).
- **A backend service consumes the group** -- `GcpBackendService.backends[].group` takes this kind's `self_link`; the balancing mode must suit the endpoint type (RATE or CONNECTION for VM and hybrid groups).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpNetworkEndpointGroup
metadata:
  name: web-neg
spec:
  zone: us-central1-a
  network:
    valueFrom:
      kind: GcpVpcNetwork
      name: main-vpc
      fieldPath: status.outputs.network_self_link
  defaultPort: 8080
  endpoints:
    - instance:
        valueFrom:
          kind: GcpComputeInstance
          name: web-1
          fieldPath: status.outputs.instance_name
```

```shell
planton apply -f network-endpoint-group.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `network` | `StringValueOrRef` | Zonal groups only (required there): the VPC every endpoint belongs to. Immutable. |
| `networkEndpointType` | `string` | Global groups only (required there): `INTERNET_IP_PORT` or `INTERNET_FQDN_PORT`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. Immutable. |
| `negName` | `string` | `metadata.name` | Name in GCP. Immutable. |
| `zone` | `string` | — (global) | A zone name builds a zonal group; empty builds a global internet group. Immutable. |
| `description` | `string` | — | Free text. Immutable. |
| `networkEndpointType` | `string` | `GCE_VM_IP_PORT` (zonal) | Zonal: `GCE_VM_IP`, `GCE_VM_IP_PORT`, `NON_GCP_PRIVATE_IP_PORT`, `INTERNET_IP_PORT`, `INTERNET_FQDN_PORT`, `GCE_VM_IP_DEDICATED_BACKEND`. Immutable. |
| `subnetwork` | `StringValueOrRef` | — | Zonal groups: the subnet endpoint IPs fall in. Immutable. |
| `defaultPort` | `int32` | — | Port for endpoints that set none. Not for `GCE_VM_IP`. Immutable. |
| `endpoints` | `[]object` | — | The whole membership: `instance` (`GcpComputeInstance` reference; VM groups), `ipAddress`, `fqdn` (FQDN groups), `port`. Changes in place. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Zonal groups require `network`; global groups reject `network` and `subnetwork`.
- The type list is per scope; `SERVERLESS` and `PRIVATE_SERVICE_CONNECT` belong to `GcpRegionNetworkEndpointGroup`.
- VM groups (`GCE_VM_IP`, `GCE_VM_IP_PORT`): every endpoint names its `instance`; other groups reject `instance`. `GCE_VM_IP` endpoints and groups carry no port.
- `INTERNET_FQDN_PORT` endpoints name an `fqdn` (no other group does); `NON_GCP_PRIVATE_IP_PORT` and `INTERNET_IP_PORT` endpoints name an `ipAddress`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | The group's self link -- a backend service's `backends[].group` |
| `neg_name` | `string` | Name in GCP |
| `neg_id` | `string` | Google's numeric id (zonal groups; empty globally) |
| `zone` | `string` | Zone of a zonal group; empty for a global one |
| `size` | `string` | Number of endpoints declared |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The list is the membership.** An endpoint removed from `endpoints` is detached in place; one added is attached. Leave the list empty to let an autoscaler or another controller own membership.
- **Nearly everything is immutable** -- name, scope, network, subnetwork, type, default port, description all recreate the group; only the endpoint list changes in place.
- **A VM endpoint without `ipAddress`** uses the instance's primary internal IP; without `port`, the group's `defaultPort`.
- **Cost**: the group and its endpoints are free; the load balancer that consumes it bills under its own kinds.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- [GcpBackendService](/docs/catalog/gcp/gcpbackendservice) — consumes the group as a backend
- [GcpRegionNetworkEndpointGroup](/docs/catalog/gcp/gcpregionnetworkendpointgroup) — serverless, PSC, and regional internet groups
- [GcpComputeInstance](/docs/catalog/gcp/gcpcomputeinstance) — the VMs a zonal group names
- [GcpVpcNetwork](/docs/catalog/gcp/gcpvpcnetwork), [GcpSubnetwork](/docs/catalog/gcp/gcpsubnetwork) — the zonal group's network placement

## Additional Resources

- [Network endpoint groups overview](https://cloud.google.com/load-balancing/docs/negs)
- [Zonal NEGs](https://cloud.google.com/load-balancing/docs/negs/zonal-neg-concepts), [Internet NEGs](https://cloud.google.com/load-balancing/docs/negs/internet-neg-concepts), [Hybrid connectivity NEGs](https://cloud.google.com/load-balancing/docs/negs/hybrid-neg-concepts)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
