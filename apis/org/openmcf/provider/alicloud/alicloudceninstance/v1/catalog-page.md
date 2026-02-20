# Alibaba Cloud CEN Instance

Deploys an Alibaba Cloud Cloud Enterprise Network (CEN) instance with bundled child-instance attachments for private multi-VPC, multi-region, and hybrid connectivity. CEN is a global resource — a single instance can connect networks across any Alibaba Cloud region.

## What Gets Created

When you deploy an AlicloudCenInstance resource, OpenMCF provisions:

- **CEN Instance** — an `alicloud_cen_instance` resource serving as the global networking hub with optional CIDR overlap protection and resource group assignment
- **CEN Instance Attachments** — one `alicloud_cen_instance_attachment` per entry in `spec.attachments[]`, connecting a VPC, VBR (Virtual Border Router), or CCN (Cloud Connect Network) to the CEN hub

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **At least one VPC** (or VBR/CCN) to attach to the CEN instance
- **VPC IDs and their regions** for each network to attach — the VPCs can be in any Alibaba Cloud region
- **Non-overlapping CIDR blocks** across attached VPCs (unless `protectionLevel` is set to `REDUCED`)

## Quick Start

Create a file `cen.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCenInstance
metadata:
  name: my-cen
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudCenInstance.my-cen
spec:
  region: cn-hangzhou
  cenInstanceName: my-cen
  attachments:
    - childInstanceId:
        value: vpc-abc123
      childInstanceRegionId: cn-hangzhou
```

Deploy:

```shell
openmcf apply -f cen.yaml
```

This creates a CEN instance and attaches one VPC in cn-hangzhou. Additional VPCs in any region can be added to the `attachments` list.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for API routing. CEN is global, so this does not restrict attachment regions. | Required; non-empty |
| `cenInstanceName` | `string` | CEN instance name. | Required; 2-128 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Human-readable description of the CEN instance. |
| `protectionLevel` | `string` | strict (empty) | CIDR overlap protection. Set to `REDUCED` to allow overlapping CIDR blocks between attached networks (routing controlled by route maps). Leave empty for strict mode that rejects overlaps. |
| `resourceGroupId` | `string` | — | Alibaba Cloud resource group ID for organizational access control. |
| `tags` | `map(string)` | — | Tags to apply to the CEN instance. |
| `attachments` | `list` | `[]` | Child-instance attachments. See attachment fields below. |

### Attachment Fields (`attachments[]`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `childInstanceId` | `StringValueOrRef` | (required) | ID of the child instance to attach (VPC ID, VBR ID, or CCN ID). Can reference an `AlicloudVpc` resource via `valueFrom`. ForceNew. |
| `childInstanceType` | `string` | `VPC` | Type of child instance: `VPC`, `VBR`, or `CCN`. ForceNew. |
| `childInstanceRegionId` | `string` | (required) | Region where the child instance resides (e.g., `cn-hangzhou`, `us-west-1`). ForceNew. |

## Examples

### Same-Region Multi-VPC

Connect multiple VPCs in the same region for private inter-VPC communication without VPC peering:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCenInstance
metadata:
  name: intra-region-cen
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: networking
    pulumi.openmcf.org/stack.name: dev.AlicloudCenInstance.intra-region-cen
spec:
  region: cn-hangzhou
  cenInstanceName: intra-region-backbone
  description: Connects production and shared-services VPCs
  attachments:
    - childInstanceId:
        value: vpc-production
      childInstanceRegionId: cn-hangzhou
    - childInstanceId:
        value: vpc-shared-services
      childInstanceRegionId: cn-hangzhou
```

### Cross-Region Global Backbone

Connect VPCs across multiple regions with REDUCED protection for overlapping CIDRs:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCenInstance
metadata:
  name: global-cen
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: networking
    pulumi.openmcf.org/stack.name: prod.AlicloudCenInstance.global-cen
spec:
  region: cn-hangzhou
  cenInstanceName: global-backbone
  description: Multi-region backbone connecting China and international regions
  protectionLevel: REDUCED
  resourceGroupId: rg-networking
  tags:
    team: platform
    purpose: global-connectivity
  attachments:
    - childInstanceId:
        value: vpc-hangzhou
      childInstanceRegionId: cn-hangzhou
    - childInstanceId:
        value: vpc-shanghai
      childInstanceRegionId: cn-shanghai
    - childInstanceId:
        value: vpc-singapore
      childInstanceRegionId: ap-southeast-1
```

### Managed VPC References with valueFrom

Connect VPCs managed as OpenMCF resources, automatically resolving VPC IDs from their stack outputs:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudCenInstance
metadata:
  name: managed-cen
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: networking
    pulumi.openmcf.org/stack.name: prod.AlicloudCenInstance.managed-cen
spec:
  region: cn-hangzhou
  cenInstanceName: managed-backbone
  description: CEN connecting OpenMCF-managed VPCs
  attachments:
    - childInstanceId:
        valueFrom:
          name: prod-vpc-hangzhou
      childInstanceRegionId: cn-hangzhou
    - childInstanceId:
        valueFrom:
          name: prod-vpc-shanghai
      childInstanceRegionId: cn-shanghai
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cen_id` | `string` | CEN instance ID assigned by Alibaba Cloud (e.g., `cen-xxxxx`) |
| `cen_instance_name` | `string` | CEN instance name as configured in the spec |

## Related Components

- [AlicloudVpc](/docs/catalog/alicloud/alicloudvpc) — provides VPCs to attach to the CEN
- [AlicloudVpnGateway](/docs/catalog/alicloud/alicloudvpngateway) — alternative point-to-point VPN connectivity
- [AlicloudVswitch](/docs/catalog/alicloud/alicloudvswitch) — subnets within attached VPCs
