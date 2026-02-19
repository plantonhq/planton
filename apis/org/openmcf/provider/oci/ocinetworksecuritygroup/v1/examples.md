# OCI Network Security Group Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Network Security Groups using the OpenMCF API. Each example demonstrates different security postures from empty NSGs to production-grade micro-segmented configurations with NSG-to-NSG rules.

## Table of Contents

- [Example 1: Minimal NSG](#example-1-minimal-nsg)
- [Example 2: Web Tier NSG](#example-2-web-tier-nsg)
- [Example 3: Private Backend NSG](#example-3-private-backend-nsg)
- [Example 4: Multi-Protocol NSG](#example-4-multi-protocol-nsg)
- [Example 5: Micro-Segmented NSG with Foreign Key References](#example-5-micro-segmented-nsg-with-foreign-key-references)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal NSG

**Use Case:** Create an NSG with no rules — suitable for testing NSG lifecycle management or as a placeholder that will have rules added later through a policy update.

**Configuration:**
- **Rules:** None
- **Effect:** Blocks all traffic (OCI NSGs have no default rules)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: empty-nsg
  org: my-org
  env: dev
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
```

**Deploy with OpenMCF CLI:**

```bash
openmcf apply -f empty-nsg.yaml
```

**What happens:**
- An NSG named `empty-nsg` is created with no security rules.
- All traffic to and from VNICs attached to this NSG is blocked.
- Standard OpenMCF freeform tags are applied automatically.
- The NSG OCID is exported as `network_security_group_id`.

---

## Example 2: Web Tier NSG

**Use Case:** Internet-facing resources such as load balancers, web servers, and API gateways. Allows HTTPS, HTTP, and ICMP Path MTU Discovery inbound. All outbound traffic is permitted.

**Configuration:**
- **Ingress:** HTTPS (443), HTTP (80), ICMP type 3 code 4
- **Egress:** All protocols
- **Rules used:** 4 of 120

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: web-tier-nsg
  org: my-org
  env: staging
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
    - source: "0.0.0.0/0"
      sourceType: cidr_block
      protocol: icmp
      description: "Path MTU Discovery"
      icmpOptions:
        type: 3
        code: 4
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

**Deploy:**

```bash
openmcf apply -f web-tier-nsg.yaml
```

**What happens:**
- An NSG with 4 rules (3 ingress + 1 egress) is created.
- HTTPS and HTTP are allowed from any IPv4 source.
- ICMP type 3 code 4 (Path MTU Discovery) prevents silent TCP stalls for connections that exceed the path MTU.
- All outbound traffic is permitted, allowing the web tier to reach backend services, external APIs, and OCI services.

---

## Example 3: Private Backend NSG

**Use Case:** Backend resources that should only accept traffic from within the VCN. Suitable for databases, application servers, internal microservices, and OKE worker nodes.

**Configuration:**
- **Ingress:** All protocols from VCN CIDR
- **Egress:** All protocols to anywhere
- **Rules used:** 2 of 120

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: backend-nsg
  org: acme-corp
  env: prod
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..aaaaaaaaprodcompartment"
  vcnId:
    value: "ocid1.vcn.oc1.iad.aaaaaaaaprodvcn"
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

**What happens:**
- Only traffic originating from within the VCN (10.0.0.0/16) is allowed inbound.
- No internet traffic can reach resources attached to this NSG, even if those resources have public IPs.
- All outbound traffic is permitted for OS patching, image pulls, DNS resolution, and API calls.
- Adjust the source CIDR to match your VCN's address range if it differs from `10.0.0.0/16`.

---

## Example 4: Multi-Protocol NSG

**Use Case:** An NSG demonstrating all protocol types and options. Combines TCP, UDP, ICMP, and "all" rules to show the full capability of the rule model.

**Configuration:**
- **Ingress:** SSH, HTTPS range, DNS (UDP), ICMP echo, all from management CIDR
- **Egress:** All outbound
- **Rules used:** 6 of 120

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: multi-protocol-nsg
  org: acme-corp
  env: prod
  labels:
    team: platform
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vcnId:
    value: "ocid1.vcn.oc1.iad.example"
  displayName: "Multi-Protocol NSG"
  ingressRules:
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: tcp
      description: "SSH from VCN"
      tcpOptions:
        destinationPortRange:
          min: 22
          max: 22
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: tcp
      description: "Application ports"
      tcpOptions:
        destinationPortRange:
          min: 8080
          max: 8443
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: udp
      description: "DNS queries from VCN"
      udpOptions:
        destinationPortRange:
          min: 53
          max: 53
    - source: "10.0.0.0/16"
      sourceType: cidr_block
      protocol: icmp
      description: "Echo requests from VCN"
      icmpOptions:
        type: 8
    - source: "172.16.0.0/12"
      sourceType: cidr_block
      protocol: all
      description: "All traffic from management network"
  egressRules:
    - destination: "0.0.0.0/0"
      destinationType: cidr_block
      protocol: all
      description: "Allow all outbound traffic"
```

**What happens:**
- TCP rules demonstrate both single-port (22) and range (8080–8443) configurations.
- The UDP rule shows DNS port restriction with `udpOptions`.
- The ICMP rule allows Echo Requests (ping) by specifying type 8 without a code — all codes for type 8 are matched.
- The management network rule uses protocol `all` with no port or ICMP constraints.
- Metadata labels (`team: platform`) are applied as OCI freeform tags alongside the standard OpenMCF tags.

---

## Example 5: Micro-Segmented NSG with Foreign Key References

**Use Case:** A production application-tier NSG that uses NSG-to-NSG references for zero-trust security and foreign key references for all dependent resources. Only traffic from the web-tier NSG is allowed on the application port.

**Configuration:**
- **Ingress:** Application port from web-tier NSG, ICMP from VCN
- **Egress:** All outbound, OCI services via service CIDR
- **Compartment:** Referenced from OciCompartment
- **VCN:** Referenced from OciVcn

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkSecurityGroup
metadata:
  name: app-tier-nsg
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
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
  displayName: "App Tier NSG"
  ingressRules:
    - source: "ocid1.networksecuritygroup.oc1.iad.web-tier-nsg-ocid"
      sourceType: network_security_group
      protocol: tcp
      description: "Application traffic from web tier NSG"
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
    - destination: "all-iad-services-in-oracle-services-network"
      destinationType: service_cidr_block
      protocol: tcp
      description: "OCI services via Service Gateway"
      tcpOptions:
        destinationPortRange:
          min: 443
          max: 443
```

**What happens:**
- The `compartmentId` and `vcnId` are resolved from previously deployed OpenMCF resources instead of hardcoded OCIDs.
- The first ingress rule allows TCP port 8443 only from VNICs attached to the web-tier NSG. This is NSG-to-NSG micro-segmentation — if the web tier scales or changes subnets, the rule still applies based on NSG membership, not IP addresses.
- The egress rules demonstrate both CIDR-based (all outbound) and service-CIDR-based (OCI services) targeting.
- The `service_cidr_block` destination type directs traffic to OCI services (Object Storage, Container Registry, etc.) for routing through the Service Gateway.
- Freeform tags include `team: platform` and `cost-center: infrastructure` from the metadata labels.

---

## Common Operations

### Get NSG ID After Deployment

```bash
# Pulumi
pulumi stack output network_security_group_id

# Terraform
terraform output network_security_group_id
```

### Use NSG ID in a Downstream Resource

The `network_security_group_id` output is the primary cross-resource reference. Use it with `StringValueOrRef` in downstream resources that support NSG associations:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciComputeInstance
metadata:
  name: app-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  nsgIds:
    - valueFrom:
        kind: OciNetworkSecurityGroup
        name: backend-nsg
        fieldPath: status.outputs.networkSecurityGroupId
```

### Destroy an NSG

An NSG must not be attached to any VNICs before deletion. Detach or delete all associated compute instances, load balancers, and other resources first.

```bash
openmcf destroy -f nsg.yaml
```

---

## Best Practices

### Prefer NSGs Over Security Lists

OCI offers two firewall mechanisms: security lists (per-subnet) and NSGs (per-VNIC). For new deployments, prefer NSGs because they are:

- More granular — attached to specific VNICs, not entire subnets
- Composable — a VNIC can have up to 5 NSGs, combining rules from different security domains
- Aligned with OCI's recommendation for new architectures

Use security lists only when you need subnet-wide rules that apply to all VNICs regardless of their NSG membership.

### Use NSG-to-NSG References for Micro-Segmentation

CIDR-based rules break when subnets change, instances scale, or IPs are recycled. NSG-to-NSG references (`sourceType: network_security_group`) provide stable, identity-based security that survives infrastructure changes:

```yaml
ingressRules:
  - source: "ocid1.networksecuritygroup.oc1.iad.web-tier-nsg"
    sourceType: network_security_group
    protocol: tcp
    tcpOptions:
      destinationPortRange:
        min: 8443
        max: 8443
```

### Budget Your 120-Rule Limit

OCI enforces 120 rules per NSG (ingress + egress combined). Plan accordingly:

- A typical web tier uses 3–5 rules
- A typical backend uses 2–4 rules
- Complex multi-service NSGs can use 20–40 rules
- If you approach 120, split into multiple NSGs (a VNIC supports up to 5 NSGs)

### Always Include an Egress Rule

OCI NSGs have no default rules. An NSG with no egress rules blocks all outbound traffic, which breaks OS patching, DNS resolution, container image pulls, and OCI service API calls. Unless you are intentionally isolating a resource, include at least:

```yaml
egressRules:
  - destination: "0.0.0.0/0"
    destinationType: cidr_block
    protocol: all
    description: "Allow all outbound traffic"
```

### Include ICMP Path MTU Discovery

For any NSG attached to resources behind load balancers or involved in cross-subnet communication, include an ICMP Path MTU Discovery rule. Without it, TCP connections can silently stall when packets exceed the path MTU:

```yaml
ingressRules:
  - source: "10.0.0.0/16"
    sourceType: cidr_block
    protocol: icmp
    description: "Path MTU Discovery"
    icmpOptions:
      type: 3
      code: 4
```

### Write Descriptions for Every Rule

Rule descriptions appear in the OCI Console and in audit logs. Descriptive rules make security reviews, incident investigations, and compliance audits significantly easier. Compare:

- Good: `"Allow HTTPS from web tier for API traffic"`
- Bad: `"tcp 443"`

### Prefer Stateful Rules (the Default)

Stateful rules (the default, `stateless: false`) automatically allow return traffic. This is correct for the vast majority of workloads. Use stateless rules only when:

- You need maximum throughput and can manage explicit bidirectional rules
- Compliance requirements mandate explicit return-traffic rules
- You are implementing a DMZ with strict symmetric filtering

### Tag for Cost and Compliance

Metadata labels are applied as OCI freeform tags. Use consistent labels across all resources:

```yaml
metadata:
  org: acme-corp
  env: prod
  labels:
    team: platform
    cost-center: infrastructure
```
