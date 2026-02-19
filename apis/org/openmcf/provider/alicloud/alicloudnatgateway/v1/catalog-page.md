# Alibaba Cloud NAT Gateway

Deploys an Alibaba Cloud Enhanced NAT Gateway with bundled EIP association and SNAT entries. The component provisions all three resources as a single atomic unit, enabling private VSwitch traffic to reach the internet through a managed NAT service.

## What Gets Created

When you deploy an AlicloudNatGateway resource, OpenMCF provisions:

- **NAT Gateway** -- an `alicloud_nat_gateway` resource placed in the specified VPC and VSwitch
- **EIP Association** -- an `alicloud_eip_association` binding the provided Elastic IP to the NAT Gateway
- **SNAT Entries** -- one `alicloud_snat_entry` per entry in `snatEntries`, mapping private VSwitch or CIDR traffic to the EIP

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **An Alibaba Cloud VPC** -- the NAT Gateway must belong to a VPC (create one with AlicloudVpc)
- **A VSwitch** -- the Enhanced NAT Gateway requires placement in a VSwitch (create with AlicloudVswitch)
- **An Elastic IP** -- the NAT Gateway needs an EIP for outbound traffic (create with AlicloudEipAddress)

## Quick Start

Create a file `nat-gateway.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: my-nat
spec:
  region: cn-hangzhou
  vpcId:
    valueFrom:
      name: my-vpc
  vswitchId:
    valueFrom:
      name: my-nat-vswitch
  natGatewayName: my-nat-gateway
  eipId:
    valueFrom:
      name: my-eip
  snatEntries:
    - sourceVswitchId:
        valueFrom:
          name: my-app-vswitch
      snatEntryName: app-zone-a
```

Deploy:

```shell
openmcf apply -f nat-gateway.yaml
```

This creates an Enhanced NAT Gateway with one SNAT entry, enabling the specified VSwitch's traffic to reach the internet through the associated EIP.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`, `cn-shanghai`) | Required; non-empty |
| `vpcId` | StringValueOrRef | VPC ID for the NAT Gateway. Can reference AlicloudVpc via `valueFrom`. | Required |
| `vswitchId` | StringValueOrRef | VSwitch ID for Enhanced NAT Gateway placement. Can reference AlicloudVswitch via `valueFrom`. | Required |
| `natGatewayName` | string | NAT Gateway name | Required; 2-128 characters |
| `eipId` | StringValueOrRef | EIP allocation ID to associate with the NAT Gateway. Can reference AlicloudEipAddress via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | | Human-readable description |
| `natType` | string | `Enhanced` | NAT Gateway type: `Enhanced` or `Normal` |
| `paymentType` | string | `PayAsYouGo` | Billing method: `PayAsYouGo` or `Subscription` |
| `internetChargeType` | string | `PayByLcu` | Metering: `PayByLcu` (capacity units) or `PayBySpec` (fixed tier) |
| `specification` | string | | Fixed tier when `internetChargeType` is `PayBySpec`: `Small`, `Middle`, `Large`, `XLarge.1` |
| `deletionProtection` | bool | `false` | Prevent accidental deletion |
| `tags` | map | | Key-value tags for the NAT Gateway |
| `snatEntries` | list | | SNAT entries for outbound internet access (see below) |

### SNAT Entry Fields

| Field | Type | Description |
|-------|------|-------------|
| `sourceVswitchId` | StringValueOrRef | VSwitch to NAT. Can reference AlicloudVswitch via `valueFrom`. Mutually exclusive with `sourceCidr`. |
| `sourceCidr` | string | CIDR block to NAT (e.g., `10.0.1.0/24`). Mutually exclusive with `sourceVswitchId`. |
| `snatEntryName` | string | Name for this SNAT entry (2-128 characters) |

## Examples

### Minimal NAT Gateway with Single SNAT

The simplest NAT configuration: one gateway, one EIP, one SNAT entry.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: simple-nat
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-abc123
  vswitchId:
    value: vsw-nat-zone
  natGatewayName: simple-nat
  eipId:
    value: eip-abc123
  snatEntries:
    - sourceVswitchId:
        value: vsw-app-zone
      snatEntryName: app-internet-access
```

### Multi-AZ NAT with CIDR-based SNAT

Production NAT Gateway with deletion protection, VSwitch-based and CIDR-based SNAT entries, and foreign key references.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: prod-nat
  org: my-org
  env: production
spec:
  region: cn-shanghai
  vpcId:
    valueFrom:
      name: prod-vpc
  vswitchId:
    valueFrom:
      name: nat-vswitch
  natGatewayName: prod-nat-gateway
  description: Production NAT for multi-AZ outbound traffic
  deletionProtection: true
  tags:
    team: infrastructure
    cost-center: networking
  eipId:
    valueFrom:
      name: prod-nat-eip
  snatEntries:
    - sourceVswitchId:
        valueFrom:
          name: app-vswitch-a
      snatEntryName: zone-a-apps
    - sourceVswitchId:
        valueFrom:
          name: app-vswitch-b
      snatEntryName: zone-b-apps
    - sourceCidr: "10.0.100.0/24"
      snatEntryName: management-subnet
```

### Fixed-Spec NAT Gateway

NAT Gateway with fixed specification tier for predictable performance billing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudNatGateway
metadata:
  name: fixed-spec-nat
spec:
  region: cn-hangzhou
  vpcId:
    value: vpc-fixed
  vswitchId:
    value: vsw-fixed
  natGatewayName: fixed-spec-nat
  internetChargeType: PayBySpec
  specification: Large
  eipId:
    value: eip-fixed
  snatEntries:
    - sourceVswitchId:
        value: vsw-workload
      snatEntryName: workload-nat
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `nat_gateway_id` | string | NAT Gateway resource ID (e.g., `ngw-xxxxx`) |
| `nat_gateway_name` | string | NAT Gateway name as created |
| `snat_table_id` | string | SNAT table ID, for adding SNAT entries outside OpenMCF |
| `forward_table_id` | string | Forward (DNAT) table ID, for adding DNAT entries outside OpenMCF |

## Related Components

- **AlicloudVpc** -- VPC that the NAT Gateway belongs to
- **AlicloudVswitch** -- VSwitch for NAT Gateway placement and SNAT source
- **AlicloudEipAddress** -- Elastic IP associated with the NAT Gateway
- **AlicloudSecurityGroup** -- Network security rules for instances using NAT
- **AlicloudAckManagedCluster** -- Kubernetes cluster nodes often use NAT for outbound access
