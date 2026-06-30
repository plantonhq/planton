# OpenStack Floating IP Associate

Associates an existing OpenStack Neutron floating IP with a port, providing external connectivity to whatever is attached to the port (typically a compute instance). This is a join resource that connects two independently managed resources -- a floating IP allocated via OpenStackFloatingIp and a port created via OpenStackNetworkPort -- making the association relationship explicit as a visible DAG node in InfraCharts.

## What Gets Created

When you deploy an OpenStackFloatingIpAssociate resource, Planton provisions:

- **Floating IP Association** — an `openstack_networking_floatingip_associate_v2` resource that binds the specified floating IP address to the target port

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **An allocated floating IP** — either the IP address or UUID of an existing floating IP (typically from an OpenStackFloatingIp resource)
- **A port with at least one fixed IP** — the target port must exist and have at least one fixed IP address (typically from an OpenStackNetworkPort resource)

## Quick Start

Create a file `fip-associate.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: my-fip-associate
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: dev.OpenstackFloatingIpAssociate.my-fip-associate
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module
spec:
  floatingIp:
    value: "203.0.113.42"
  portId:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

Deploy:

```shell
planton apply -f fip-associate.yaml
```

This associates the floating IP `203.0.113.42` with the specified port, enabling external connectivity through that floating IP.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `floatingIp` | `StringValueOrRef` | The floating IP address (or UUID) to associate. Can reference an OpenStackFloatingIp resource via `valueFrom` (defaults to `status.outputs.address`). The Terraform provider accepts either an IP address or a floating IP UUID. |
| `portId` | `StringValueOrRef` | The UUID of the port to associate the floating IP with. The port must have at least one fixed IP address. Can reference an OpenStackNetworkPort resource via `valueFrom` (defaults to `status.outputs.port_id`). |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fixedIp` | `string` | — | Specifies which fixed IP address on the port to map the floating IP to. Only relevant when the port has multiple fixed IP addresses. If omitted and the port has a single IP, that IP is used automatically. If omitted and the port has multiple IPs, OpenStack picks the first one. |
| `region` | `string` | provider default | Overrides the region from the provider config for this association. ForceNew: changing the region recreates the association. |

## Examples

### Basic Association with Literal Values

Associates a floating IP with a port using literal IP address and port UUID values. Suitable when both resources are managed outside of Planton or when you already know the values:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: web-fip-associate
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: dev.OpenstackFloatingIpAssociate.web-fip-associate
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module
spec:
  floatingIp:
    value: "203.0.113.42"
  portId:
    value: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
```

### Association with Foreign Key References

Uses `valueFrom` to reference an OpenStackFloatingIp and an OpenStackNetworkPort managed in the same InfraChart. The FK resolver automatically retrieves the floating IP address and port UUID at deploy time. Since `kind` and `fieldPath` default to the annotated values, only `name` is required in each reference:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: app-fip-associate
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: staging.OpenstackFloatingIpAssociate.app-fip-associate
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module
spec:
  floatingIp:
    valueFrom:
      name: app-floating-ip
  portId:
    valueFrom:
      name: app-port
```

### Association with Fixed IP on a Multi-IP Port

When the target port has multiple fixed IP addresses, use `fixedIp` to specify which one the floating IP should map to. Without this field, OpenStack picks the first fixed IP on the port:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: multi-ip-fip-associate
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: prod.OpenstackFloatingIpAssociate.multi-ip-fip-associate
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module
spec:
  floatingIp:
    valueFrom:
      name: prod-floating-ip
  portId:
    valueFrom:
      name: prod-multi-ip-port
  fixedIp: "10.0.1.100"
```

### Association in a Specific Region

Overrides the provider-level region for deployments that span multiple OpenStack regions:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackFloatingIpAssociate
metadata:
  name: region2-fip-associate
  labels:
    planton.dev/provisioner: pulumi
    planton.dev/stack.jobId: prod.OpenstackFloatingIpAssociate.region2-fip-associate
    planton.dev/stack.module.source: github.com/plantonhq/planton//apis/dev/planton/provider/openstack/openstackfloatingipassociate/v1/iac/pulumi/module
spec:
  floatingIp:
    value: "198.51.100.10"
  portId:
    valueFrom:
      name: region2-app-port
  region: RegionTwo
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `id` | `string` | Terraform resource identifier for the association (typically the floating IP address) |
| `floating_ip` | `string` | The floating IP address that was associated |
| `port_id` | `string` | UUID of the port the floating IP was associated with |
| `fixed_ip` | `string` | The fixed IP address on the port that the floating IP maps to (computed by OpenStack if not specified) |
| `region` | `string` | OpenStack region where the association was created |

## Related Components

- [OpenStackFloatingIp](/docs/catalog/openstack/openstackfloatingip) — allocates floating IPs from external networks; use its built-in `portId` field for simple allocation-and-association in a single manifest
- [OpenStackNetworkPort](/docs/catalog/openstack/openstacknetworkport) — creates ports with specific IPs and security groups that floating IPs can be associated with
- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — provides the Layer 2 network on which ports are created
- [OpenStackSubnet](/docs/catalog/openstack/openstacksubnet) — defines IP address ranges and DHCP settings that determine the fixed IPs available on ports
- [OpenStackRouter](/docs/catalog/openstack/openstackrouter) — provides routing between internal networks and external connectivity
