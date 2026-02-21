# OCI Network Firewall

Deploys an Oracle Cloud Infrastructure Network Firewall with an inline firewall policy. The firewall appliance is deployed into a subnet and inspects traffic according to security rules that match on source/destination IP addresses, TCP/UDP ports, and URL patterns. The policy and all sub-resources (address lists, services, service lists, URL lists, security rules) are declared in a single manifest.

## What Gets Created

When you deploy an OciNetworkFirewall resource, OpenMCF provisions:

- **Firewall Policy** — a `networkfirewall.NetworkFirewallPolicy` resource in the specified compartment. Created first as the container for all policy sub-resources.
- **Address Lists** — one `networkfirewall.NetworkFirewallPolicyAddressList` per entry in `policy.addressLists`. Each list contains IP CIDRs or FQDNs referenced by security rules.
- **Services** — one `networkfirewall.NetworkFirewallPolicyService` per entry in `policy.services`. Each service defines TCP or UDP port ranges.
- **Service Lists** — one `networkfirewall.NetworkFirewallPolicyServiceList` per entry in `policy.serviceLists`. Each list groups services for reuse across rules.
- **URL Lists** — one `networkfirewall.NetworkFirewallPolicyUrlList` per entry in `policy.urlLists`. Each list contains URL patterns for L7 HTTP(S) inspection.
- **Security Rules** — one `networkfirewall.NetworkFirewallPolicySecurityRule` per entry in `policy.securityRules`. Rules are evaluated in list order (priority derived from position). Each rule references address lists, services, and URL lists by name.
- **Network Firewall** — a `networkfirewall.NetworkFirewall` appliance in the specified subnet, associated with the policy. Created last, after all policy sub-resources.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the firewall and policy will be created — either a literal value or a reference to an OciCompartment resource
- **A subnet OCID** where the firewall appliance will be deployed — either a literal value or a reference to an OciSubnet resource
- **At least one security rule** defined in the policy

## Quick Start

Create a file `network-firewall.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: my-firewall
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciNetworkFirewall.my-firewall
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  policy:
    addressLists:
      - name: "internal-networks"
        type: ip
        addresses:
          - "10.0.0.0/8"
          - "172.16.0.0/12"
    securityRules:
      - name: "allow-internal"
        action: allow
        condition:
          sourceAddresses:
            - "internal-networks"
```

Deploy:

```shell
openmcf apply -f network-firewall.yaml
```

This creates a network firewall with a policy that allows traffic from internal RFC 1918 networks. The firewall OCID, IPv4 address, and policy OCID are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the firewall and policy will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `subnetId` | `StringValueOrRef` | OCID of the subnet where the firewall appliance will be deployed. ForceNew. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `policy` | `Policy` | Inline firewall policy with security rules and supporting objects. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the firewall. |
| `ipv4Address` | `string` | auto-assigned | Static IPv4 address for the firewall. ForceNew. |
| `ipv6Address` | `string` | auto-assigned | Static IPv6 address for the firewall (if subnet supports IPv6). ForceNew. |
| `availabilityDomain` | `string` | auto-selected | Availability domain for firewall placement. ForceNew. |
| `networkSecurityGroupIds` | `StringValueOrRef[]` | — | NSGs applied to the firewall. Can reference OciSecurityGroup resources via `valueFrom`. |
| `natConfiguration` | `NatConfiguration` | — | NAT configuration for egress traffic inspection. |
| `shape` | `string` | OCI default | Firewall shape determining throughput capacity. |

### NatConfiguration

| Field | Type | Description |
|-------|------|-------------|
| `mustEnablePrivateNat` | `bool` | When true, the firewall uses a private IP for NAT instead of a public IP. |

### Policy

| Field | Type | Description |
|-------|------|-------------|
| `displayName` | `string` | Display name for the policy. Defaults to `"{firewall}-policy"`. |
| `description` | `string` | Description of the policy. |
| `addressLists` | `AddressList[]` | IP/FQDN address lists referenced by security rules. |
| `services` | `Service[]` | TCP/UDP port definitions referenced by security rules. |
| `serviceLists` | `ServiceList[]` | Groups of services for reuse across rules. |
| `urlLists` | `UrlList[]` | URL pattern lists for L7 HTTP(S) inspection. |
| `securityRules` | `SecurityRule[]` | Security rules evaluated in list order. |

### AddressList

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name within the policy. ForceNew. Referenced by security rules. | Min length 1 |
| `type` | `enum` | `ip` for IP addresses/CIDRs, `fqdn` for domain names. | Required |
| `addresses` | `string[]` | IP CIDRs (e.g., `"10.0.0.0/8"`) or FQDNs (e.g., `"example.com"`). | Min 1 item |
| `description` | `string` | Optional description. | — |

### Service

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name within the policy. ForceNew. Referenced by security rules and service lists. | Min length 1 |
| `type` | `enum` | `tcp_service` or `udp_service`. | Required |
| `portRanges` | `PortRange[]` | Port ranges for this service. | Min 1 item |
| `description` | `string` | Optional description. | — |

### PortRange

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `minimumPort` | `int32` | Start of the port range (inclusive). | 1-65535 |
| `maximumPort` | `int32` | End of the port range (inclusive). When omitted, equals `minimumPort` (single port). | >= minimumPort |

### ServiceList

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name within the policy. ForceNew. Referenced by security rules. | Min length 1 |
| `services` | `string[]` | Names of services to include. Each must match a Service defined in the policy. | — |
| `description` | `string` | Optional description. | — |

### UrlList

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name within the policy. ForceNew. Referenced by security rules. | Min length 1 |
| `urls` | `UrlPattern[]` | URL patterns for matching. | Min 1 item |
| `description` | `string` | Optional description. | — |

### UrlPattern

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `pattern` | `string` | URL pattern (e.g., `"*.example.com"`, `"malware.example.com/path"`). | Min length 1 |

### SecurityRule

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `name` | `string` | Unique name within the policy. ForceNew. | Min length 1 |
| `action` | `enum` | `allow`, `drop`, `reject`, or `inspect`. | Required |
| `condition` | `SecurityRuleCondition` | Traffic matching criteria. | Required |
| `inspection` | `enum` | `intrusion_detection` or `intrusion_prevention`. Required when action is `inspect`. | — |
| `description` | `string` | Optional description. | — |

### SecurityRuleCondition

All non-empty fields are AND-ed together. Within each field, values are OR-ed.

| Field | Type | Description |
|-------|------|-------------|
| `sourceAddresses` | `string[]` | Names of address lists matching source IPs. |
| `destinationAddresses` | `string[]` | Names of address lists matching destination IPs. |
| `services` | `string[]` | Names of services or service lists matching traffic ports. |
| `urls` | `string[]` | Names of URL lists matching HTTP(S) request URLs. |

## Examples

### Basic Allow/Deny Firewall

A firewall allowing internal traffic and dropping everything else:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: basic-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciNetworkFirewall.basic-fw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  policy:
    addressLists:
      - name: "internal"
        type: ip
        addresses:
          - "10.0.0.0/8"
          - "172.16.0.0/12"
          - "192.168.0.0/16"
    securityRules:
      - name: "allow-internal"
        action: allow
        condition:
          sourceAddresses:
            - "internal"
      - name: "deny-all"
        action: drop
        condition: {}
```

### Web Application Firewall with Port and URL Rules

A firewall allowing HTTPS traffic to specific destinations and blocking known malicious URLs:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: web-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkFirewall.web-fw
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  subnetId:
    valueFrom:
      kind: OciSubnet
      name: firewall-subnet
      fieldPath: status.outputs.subnetId
  policy:
    addressLists:
      - name: "web-servers"
        type: ip
        addresses:
          - "10.0.1.0/24"
      - name: "blocked-domains"
        type: fqdn
        addresses:
          - "malware.example.com"
          - "phishing.example.com"
    services:
      - name: "https"
        type: tcp_service
        portRanges:
          - minimumPort: 443
            maximumPort: 443
      - name: "http"
        type: tcp_service
        portRanges:
          - minimumPort: 80
            maximumPort: 80
    serviceLists:
      - name: "web-traffic"
        services:
          - "http"
          - "https"
    urlLists:
      - name: "blocked-urls"
        urls:
          - pattern: "*.malware.example.com"
          - pattern: "phishing.example.com/*"
    securityRules:
      - name: "block-malicious-urls"
        action: drop
        condition:
          urls:
            - "blocked-urls"
      - name: "allow-web-to-servers"
        action: allow
        condition:
          destinationAddresses:
            - "web-servers"
          services:
            - "web-traffic"
      - name: "deny-all"
        action: drop
        condition: {}
```

### Firewall with IDS/IPS Inspection

A firewall with intrusion detection on inbound traffic and private NAT:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNetworkFirewall
metadata:
  name: ids-fw
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNetworkFirewall.ids-fw
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  subnetId:
    value: "ocid1.subnet.oc1..example"
  natConfiguration:
    mustEnablePrivateNat: true
  networkSecurityGroupIds:
    - valueFrom:
        kind: OciSecurityGroup
        name: fw-nsg
        fieldPath: status.outputs.networkSecurityGroupId
  policy:
    addressLists:
      - name: "external"
        type: ip
        addresses:
          - "0.0.0.0/0"
      - name: "internal"
        type: ip
        addresses:
          - "10.0.0.0/8"
    services:
      - name: "all-tcp"
        type: tcp_service
        portRanges:
          - minimumPort: 1
            maximumPort: 65535
    securityRules:
      - name: "inspect-inbound"
        action: inspect
        inspection: intrusion_prevention
        condition:
          sourceAddresses:
            - "external"
          destinationAddresses:
            - "internal"
          services:
            - "all-tcp"
      - name: "allow-internal"
        action: allow
        condition:
          sourceAddresses:
            - "internal"
      - name: "deny-all"
        action: drop
        condition: {}
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `firewall_id` | `string` | OCID of the network firewall |
| `ipv4_address` | `string` | IPv4 address of the firewall appliance. Used for configuring route table entries to direct traffic through the firewall. |
| `policy_id` | `string` | OCID of the firewall policy |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides the subnet where the firewall appliance is deployed
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — provides NSGs for the firewall via `valueFrom`
- [OciVcn](/docs/catalog/oci/ocivcn) — the firewall's subnet must belong to a VCN; route table entries direct traffic through the firewall's IPv4 address
