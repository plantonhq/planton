# OCI Network Security Group

Deploys an Oracle Cloud Infrastructure Network Security Group (NSG) with inline ingress and egress security rules. Rules are split by direction so that source and destination fields are never ambiguous. Each NSG supports up to 120 rules total and attaches to individual VNICs for per-resource traffic control.

## What Gets Created

When you deploy an OciNetworkSecurityGroup resource, OpenMCF provisions:

- **Network Security Group** — an `oci_core_network_security_group` resource in the specified compartment and VCN with a display name and freeform tags.
- **Security Rules** — one `oci_core_network_security_group_security_rule` for each entry in `ingressRules` and `egressRules`. Each rule specifies the direction, protocol, source or destination, and optional port or ICMP constraints.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the NSG will be created — either a literal value or a reference to an OciCompartment resource
- **A VCN OCID** that the NSG belongs to — either a literal value or a reference to an OciVcn resource

## Quick Start

Create a file `nsg.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: my-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciNetworkSecurityGroup.my-nsg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
```

Deploy:

```shell
openmcf apply -f nsg.yaml
```

This creates an empty NSG with no security rules in the specified VCN. An empty NSG blocks all traffic — add `ingressRules` and `egressRules` to define traffic policies. The NSG ID is exported as a stack output for use by downstream resources such as OciComputeInstance and OciContainerEngineCluster.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the NSG will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `vcnId` | `StringValueOrRef` | OCID of the VCN that this NSG belongs to. Changing this forces recreation. Can reference an OciVcn resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. Falls back to `metadata.name` if not provided. |
| `ingressRules` | `IngressRule[]` | — | Inbound security rules. Each rule defines traffic allowed TO resources associated with this NSG. |
| `egressRules` | `EgressRule[]` | — | Outbound security rules. Each rule defines traffic allowed FROM resources associated with this NSG. |

OCI enforces a maximum of 120 security rules per NSG (ingress + egress combined). The proto schema validates this constraint at submission time.

### Ingress and Egress Rules

Ingress rules and egress rules share the same structure. Ingress rules use `source` and `sourceType` to identify where traffic originates; egress rules use `destination` and `destinationType` to identify where traffic goes. All other fields are identical.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `source` / `destination` | `string` | A CIDR block (e.g., `"10.0.0.0/16"`), a service CIDR label (e.g., `"all-iad-services-in-oracle-services-network"`), or the OCID of another NSG. | Required, min length 1 |
| `sourceType` / `destinationType` | `TargetType` | Interpretation of the source/destination field. | Defaults to `cidr_block` |
| `protocol` | `Protocol` | Transport protocol for this rule. | Required |
| `description` | `string` | Human-readable description. | — |
| `stateless` | `bool` | When `true`, the rule is stateless (return traffic must be explicitly allowed). When `false` (the default), the rule is stateful. | Default `false` |
| `tcpOptions` | `TcpOptions` | TCP port constraints. | Only valid when protocol is `tcp` |
| `udpOptions` | `UdpOptions` | UDP port constraints. | Only valid when protocol is `udp` |
| `icmpOptions` | `IcmpOptions` | ICMP type/code constraints. | Only valid when protocol is `icmp` or `icmpv6` |

### Protocol Values

| Value | Description |
|-------|-------------|
| `all` | All protocols — no port or ICMP options needed |
| `tcp` | TCP — use `tcpOptions` to constrain ports |
| `udp` | UDP — use `udpOptions` to constrain ports |
| `icmp` | ICMP — use `icmpOptions` to constrain type and code |
| `icmpv6` | ICMPv6 — use `icmpOptions` to constrain type and code |

### Target Type Values

| Value | Description |
|-------|-------------|
| `cidr_block` | An IPv4 or IPv6 CIDR block (e.g., `"0.0.0.0/0"`, `"10.0.0.0/16"`) |
| `service_cidr_block` | An OCI service CIDR label (e.g., `"all-iad-services-in-oracle-services-network"`) |
| `network_security_group` | The OCID of another NSG — enables NSG-to-NSG micro-segmentation |

### TCP and UDP Options

TCP and UDP options constrain traffic by port range. When omitted, all ports are allowed for the specified protocol.

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `tcpOptions.destinationPortRange.min` | `int32` | Minimum destination port | 1–65535 |
| `tcpOptions.destinationPortRange.max` | `int32` | Maximum destination port | 1–65535, must be >= min |
| `tcpOptions.sourcePortRange.min` | `int32` | Minimum source port | 1–65535 |
| `tcpOptions.sourcePortRange.max` | `int32` | Maximum source port | 1–65535, must be >= min |

UDP options follow the same structure: `udpOptions.destinationPortRange` and `udpOptions.sourcePortRange`.

Set `min` equal to `max` to specify a single port (e.g., `min: 443, max: 443`).

### ICMP Options

| Field | Type | Description |
|-------|------|-------------|
| `icmpOptions.type` | `int32` | ICMP message type (e.g., 3 for "Destination Unreachable", 8 for "Echo Request") |
| `icmpOptions.code` | `int32` | ICMP message code. When omitted, all codes for the given type are matched. Code 0 is a valid value and is distinct from "not set". |

## Examples

### Web Tier NSG

An NSG for internet-facing resources such as load balancers and web servers. Allows HTTPS and HTTP inbound from anywhere, ICMP Path MTU Discovery from the VCN, and all outbound traffic:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: web-tier-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciNetworkSecurityGroup.web-tier-nsg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  displayName: "Web Tier NSG"
  ingressRules:
    - source: "0.0.0.0/0"
      sourceType: cidr_block
      protocol: tcp
      description: "Allow HTTPS from anywhere"
      tcpOptions:
        destinationPortRange:
          min: 443
          max: 443
    - source: "0.0.0.0/0"
      sourceType: cidr_block
      protocol: tcp
      description: "Allow HTTP from anywhere"
      tcpOptions:
        destinationPortRange:
          min: 80
          max: 80
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: icmp
      description: "Path MTU Discovery from VCN"
      icmpOptions:
        type: 3
        code: 4
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

### Private Backend NSG

An NSG for resources that should only accept traffic from within the VCN. Suitable for databases, application servers, and OKE worker nodes:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: backend-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkSecurityGroup.backend-nsg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  displayName: "Private Backend NSG"
  ingressRules:
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: all
      description: "Allow all traffic from within the VCN"
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

### Micro-Segmented NSG

An NSG that restricts ingress to traffic from a specific NSG rather than a CIDR block. This enables zero-trust micro-segmentation where only resources attached to the web-tier NSG can reach the application tier:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: app-tier-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkSecurityGroup.app-tier-nsg
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  displayName: "App Tier NSG"
  ingressRules:
    - source: "ocid1.networksecuritygroup.oc1.iad.web-tier-nsg-ocid"
      sourceType: network_security_group
      protocol: tcp
      description: "Allow HTTPS from web tier NSG"
      tcpOptions:
        destinationPortRange:
          min: 8443
          max: 8443
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: icmp
      description: "Path MTU Discovery from VCN"
      icmpOptions:
        type: 3
        code: 4
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding OCIDs. The compartment and VCN are resolved from deployed resources:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: ref-nsg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkSecurityGroup.ref-nsg
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
  ingressRules:
    - source: "0.0.0.0/0"
      sourceType: cidr_block
      protocol: tcp
      description: "Allow HTTPS from anywhere"
      tcpOptions:
        destinationPortRange:
          min: 443
          max: 443
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `network_security_group_id` | `string` | OCID of the created network security group |

## Related Components

- [OciVcn](/docs/catalog/oci/ocivcn) — provides the parent VCN referenced by `vcnId`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — creates subnets within the same VCN where NSG-attached resources are placed
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciComputeInstance](/docs/catalog/oci/ocicomputeinstance) — attaches VNICs to this NSG for per-instance traffic control
- [OciContainerEngineCluster](/docs/catalog/oci/ocicontainerenginecluster) — references NSGs for OKE API endpoint and node pool security
