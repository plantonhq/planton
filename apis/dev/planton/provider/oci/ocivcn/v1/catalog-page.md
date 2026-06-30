# OCI VCN

Deploys an Oracle Cloud Infrastructure Virtual Cloud Network (VCN) with optional Internet, NAT, and Service Gateways. The VCN supports multiple CIDR blocks, IPv6, and DNS hostnames. Each gateway is controlled by a boolean toggle and only provisioned when enabled.

## What Gets Created

When you deploy an OciVcn resource, Planton provisions:

- **Virtual Cloud Network** — an `oci_core_vcn` resource in the specified compartment with one or more CIDR blocks, optional DNS label, and optional IPv6 prefix. OCI automatically creates a default route table, default security list, and default DHCP options alongside the VCN.
- **Internet Gateway** — created only when `isInternetGatewayEnabled` is `true`. Provides direct inbound and outbound internet access for resources in public subnets.
- **NAT Gateway** — created only when `isNatGatewayEnabled` is `true`. Allows resources in private subnets to initiate outbound internet connections without accepting inbound traffic.
- **Service Gateway** — created only when `isServiceGatewayEnabled` is `true`. Provides private access to OCI services (Object Storage, etc.) without traffic leaving the Oracle backbone network. Automatically configured for all services in the Oracle Services Network.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the VCN and gateways will be created — either a literal value or a reference to an OciCompartment resource
- **A CIDR plan** — at least one IPv4 CIDR block between /16 and /30

## Quick Start

Create a file `vcn.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVcn
metadata:
  name: my-vcn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciVcn.my-vcn
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
```

Deploy:

```shell
planton apply -f vcn.yaml
```

This creates a VCN with a single 10.0.0.0/16 CIDR block and no gateways. The VCN ID, default route table, default security list, and default DHCP options are exported as stack outputs for use by downstream resources such as OciSubnet.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the VCN and its gateways will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `cidrBlocks` | `string[]` | IPv4 CIDR blocks for the VCN. OCI supports multiple non-overlapping CIDRs per VCN (e.g., `["10.0.0.0/16", "172.16.0.0/16"]`). Each block must be between /16 and /30. | Minimum 1 item required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name for the VCN shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `dnsLabel` | `string` | — | DNS label for the VCN. When set, the VCN domain becomes `<dnsLabel>.oraclevcn.com` and enables DNS hostnames within the VCN. Must be alphanumeric, start with a letter, and be at most 15 characters. |
| `isIpv6Enabled` | `bool` | `false` | When `true`, allocates an Oracle-assigned /56 IPv6 GUA prefix for the VCN. |
| `isInternetGatewayEnabled` | `bool` | `false` | When `true`, creates an Internet Gateway attached to the VCN. Required for resources that need direct inbound and outbound internet access. |
| `isNatGatewayEnabled` | `bool` | `false` | When `true`, creates a NAT Gateway attached to the VCN. Allows private resources to initiate outbound internet connections without exposing them to inbound traffic. |
| `isServiceGatewayEnabled` | `bool` | `false` | When `true`, creates a Service Gateway attached to the VCN. Provides private access to OCI services without traffic leaving the Oracle network. Automatically configured for all services in the Oracle Services Network. |

## Examples

### Minimal VCN

A VCN with a single CIDR block and no gateways — suitable for development or isolated workloads:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVcn
metadata:
  name: dev-vcn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciVcn.dev-vcn
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
```

### VCN with Internet and NAT Gateways

A VCN for workloads that need both public-facing and private subnets. The Internet Gateway serves public subnets; the NAT Gateway gives private subnets outbound internet access:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVcn
metadata:
  name: web-vcn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OciVcn.web-vcn
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
  displayName: "Web Tier VCN"
  dnsLabel: "webvcn"
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
```

### Full-Featured Production VCN

All gateways enabled, IPv6 for dual-stack workloads, multiple CIDRs for address segmentation, and DNS resolution for hostname-based communication:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVcn
metadata:
  name: prod-vcn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVcn.prod-vcn
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  cidrBlocks:
    - "10.0.0.0/16"
    - "172.16.0.0/16"
  displayName: "Production VCN"
  dnsLabel: "prodvcn"
  isIpv6Enabled: true
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
  isServiceGatewayEnabled: true
```

### Using Foreign Key References

Reference an Planton-managed compartment instead of hardcoding the OCID:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVcn
metadata:
  name: ref-vcn
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVcn.ref-vcn
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  cidrBlocks:
    - "10.0.0.0/16"
  isInternetGatewayEnabled: true
  isNatGatewayEnabled: true
  isServiceGatewayEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vcn_id` | `string` | OCID of the created VCN |
| `default_route_table_id` | `string` | OCID of the default route table created with the VCN |
| `default_security_list_id` | `string` | OCID of the default security list created with the VCN |
| `default_dhcp_options_id` | `string` | OCID of the default DHCP options created with the VCN |
| `internet_gateway_id` | `string` | OCID of the Internet Gateway. Empty when `isInternetGatewayEnabled` is `false`. |
| `nat_gateway_id` | `string` | OCID of the NAT Gateway. Empty when `isNatGatewayEnabled` is `false`. |
| `service_gateway_id` | `string` | OCID of the Service Gateway. Empty when `isServiceGatewayEnabled` is `false`. |

## Related Components

- [OciSubnet](/docs/catalog/oci/ocisubnet) — creates subnets within this VCN, with route table attachments for gateway routing
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — manages network security rules for resources attached to this VCN
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
