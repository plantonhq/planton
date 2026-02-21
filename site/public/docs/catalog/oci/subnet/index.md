---
title: "Subnet"
description: "Subnet deployment documentation"
icon: "package"
order: 100
componentName: "ocisubnet"
---

# OCI Subnet

Deploys an Oracle Cloud Infrastructure subnet within a Virtual Cloud Network (VCN) with support for both regional and availability-domain-specific placement. The subnet controls public and private network access through VNIC IP assignment rules and optional internet ingress blocking. When route rules are provided, a dedicated route table is created and attached to the subnet automatically.

## What Gets Created

When you deploy an OciSubnet resource, OpenMCF provisions:

- **Subnet** — an `oci_core_subnet` resource in the specified compartment and VCN with the given CIDR block, optional DNS label, and configurable public/private access controls. The subnet is regional by default (spans all availability domains in the region) unless `availabilityDomain` is set.
- **Route Table** — created only when `routeRules` are provided. A dedicated `oci_core_route_table` named `{displayName}-rt` is created in the same compartment and VCN, then automatically associated with the subnet. When `routeTableId` is provided instead, that existing route table is used. When neither is provided, the subnet inherits the VCN's default route table.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the subnet will be created — either a literal value or a reference to an OciCompartment resource
- **A VCN OCID** for the parent Virtual Cloud Network — either a literal value or a reference to an OciVcn resource
- **A CIDR block** within one of the VCN's CIDR ranges that does not overlap with existing subnets in the same VCN
- **Gateway OCIDs** if configuring inline route rules — Internet Gateway, NAT Gateway, or Service Gateway IDs from the parent OciVcn resource

## Quick Start

Create a file `subnet.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: my-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciSubnet.my-subnet
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.1.0/24"
```

Deploy:

```shell
openmcf apply -f subnet.yaml
```

This creates a regional subnet with the 10.0.1.0/24 CIDR block using the VCN's default route table. By default, public IP assignment is allowed on VNICs and internet ingress is not blocked. The subnet ID, domain name, virtual router IP, virtual router MAC, and associated route table ID are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the subnet will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `vcnId` | `StringValueOrRef` | OCID of the parent VCN that this subnet belongs to. Can reference an OciVcn resource via `valueFrom`. | Required |
| `cidrBlock` | `string` | IPv4 CIDR block for the subnet (e.g., `"10.0.1.0/24"`). Must fall within one of the VCN's CIDR blocks and must not overlap with other subnets in the same VCN. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `dnsLabel` | `string` | — | DNS label for the subnet. Combined with the VCN DNS label to form the subnet domain: `<dnsLabel>.<vcnDnsLabel>.oraclevcn.com`. Must be alphanumeric, start with a letter, and be at most 15 characters. |
| `availabilityDomain` | `string` | — | Availability domain name (e.g., `"Iocq:US-ASHBURN-AD-1"`). When omitted, the subnet is regional and spans all ADs. When set, the subnet is scoped to a single AD. |
| `prohibitPublicIpOnVnic` | `bool` | `false` | When `true`, VNICs in this subnet cannot have public IP addresses. This is the primary control for making a subnet private. |
| `prohibitInternetIngress` | `bool` | `false` | When `true`, blocks all inbound internet traffic to VNICs, even if a security rule or NSG would otherwise allow it. |
| `dhcpOptionsId` | `StringValueOrRef` | VCN default | OCID of custom DHCP options to use instead of the VCN's default DHCP options. |
| `routeTableId` | `StringValueOrRef` | VCN default | OCID of an existing route table to associate with this subnet. Mutually exclusive with `routeRules`. |
| `securityListIds` | `StringValueOrRef[]` | VCN default | Security list OCIDs to associate with this subnet. Maximum 5 security lists per subnet. |
| `ipv6CidrBlock` | `string` | — | IPv6 CIDR block for dual-stack subnets (e.g., `"2001:0db8:0123:1111::/64"`). Only valid when the parent VCN has IPv6 enabled. |
| `routeRules` | `RouteRule[]` | — | Route rules for an inline custom route table owned by this subnet. When provided, a dedicated route table is created and associated with the subnet. Mutually exclusive with `routeTableId`. |

### Route Rules

When `routeRules` are provided, a dedicated route table is created and associated with the subnet. Each rule defines a routing entry:

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `routeRules[].destination` | `string` | Target IP range in CIDR notation (e.g., `"0.0.0.0/0"`) or a service CIDR label (e.g., `"all-iad-services-in-oracle-services-network"`). | Required |
| `routeRules[].destinationType` | `enum` | Whether `destination` is a CIDR block or a service CIDR label. Values: `cidr_block`, `service_cidr_block`. | — |
| `routeRules[].networkEntityId` | `StringValueOrRef` | OCID of the network entity to route matching traffic to (Internet Gateway, NAT Gateway, DRG, Service Gateway, or Local Peering Gateway). | Required |
| `routeRules[].description` | `string` | Human-readable description for this rule. | — |

## Examples

### Private Subnet

A private subnet where VNICs cannot have public IPs and inbound internet traffic is blocked — suitable for databases, application backends, and OKE worker nodes:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: private-app
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciSubnet.private-app
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.10.0/24"
  dnsLabel: "privapp"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
```

### Public Subnet with DNS

A public subnet with DNS resolution enabled — suitable for load balancers, bastion hosts, and services that need direct internet-facing access:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: public-lb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciSubnet.public-lb
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.0.0/24"
  displayName: "Load Balancer Subnet"
  dnsLabel: "publb"
```

### Private Subnet with Inline Route Rules

A private subnet with a dedicated route table that routes internet traffic through the VCN's NAT Gateway and OCI service traffic through the Service Gateway:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: app-tier
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciSubnet.app-tier
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  cidrBlock: "10.0.20.0/24"
  displayName: "Application Tier"
  dnsLabel: "apptier"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
  routeRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      networkEntityId:
        value: "ocid1.natgateway.oc1.iad.example"
      description: "Internet traffic via NAT Gateway"
    - destination: "all-iad-services-in-oracle-services-network"
      destinationType: service_cidr_block
      networkEntityId:
        value: "ocid1.servicegateway.oc1.iad.example"
      description: "OCI services via Service Gateway"
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding OCIDs. The compartment, VCN, and gateway IDs are resolved from other deployed resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciSubnet
metadata:
  name: ref-subnet
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciSubnet.ref-subnet
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  vcnId:
    valueFrom:
      kind: OciVcn
      name: prod-vcn
      fieldPath: status.outputs.vcnId
  cidrBlock: "10.0.30.0/24"
  prohibitPublicIpOnVnic: true
  prohibitInternetIngress: true
  routeRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      networkEntityId:
        valueFrom:
          kind: OciVcn
          name: prod-vcn
          fieldPath: status.outputs.natGatewayId
      description: "Internet traffic via NAT Gateway"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subnet_id` | `string` | OCID of the created subnet |
| `subnet_domain_name` | `string` | Fully qualified domain name of the subnet (e.g., `"subnet1.vcn1.oraclevcn.com"`). Set when `dnsLabel` is configured. |
| `virtual_router_ip` | `string` | IP address of the virtual router in this subnet |
| `virtual_router_mac` | `string` | MAC address of the virtual router in this subnet |
| `route_table_id` | `string` | OCID of the route table associated with this subnet — the inline route table (when `routeRules` are provided), the externally referenced route table (when `routeTableId` is set), or the VCN's default route table |

## Related Components

- [OciVcn](/docs/catalog/oci/vcn) — provides the parent VCN, gateways (Internet, NAT, Service), and default route table referenced by this subnet
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages stateful security rules for VNICs attached to resources in this subnet
- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciComputeInstance](/docs/catalog/oci/compute-instance) — deploys VM instances with VNICs attached to this subnet
- [OciContainerEngineCluster](/docs/catalog/oci/container-engine-cluster) — uses subnets for OKE API endpoint and node pool placement
