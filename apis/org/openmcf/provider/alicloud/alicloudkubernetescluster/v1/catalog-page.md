# AlicloudKubernetesCluster

Deploys an Alibaba Cloud ACK Managed Kubernetes cluster with configurable CNI networking (Flannel or Terway), cluster addons, control plane logging, RRSA for pod-level IAM, and optional maintenance windows with automatic version upgrades. Worker nodes are managed separately through AlicloudKubernetesNodePool.

## What Gets Created

When you deploy an AlicloudKubernetesCluster resource, OpenMCF provisions:

- **ACK Managed Kubernetes Cluster** — an `alicloud_cs_managed_kubernetes` resource with a fully managed control plane (etcd, API server, controller manager, scheduler)
- **Cluster Addons** — network CNI (Flannel or Terway), storage CSI drivers, and optional monitoring, logging, and ingress addons installed at creation time
- **API Server SLB** — a public-facing SLB for the Kubernetes API server (when `slbInternetEnabled` is true)
- **NAT Gateway** — auto-created for cluster VPC internet access (when `newNatGateway` is true)
- **RRSA OIDC Provider** — RAM OIDC provider for pod-level IAM federation (when `enableRrsa` is true)

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **At least two VSwitches** in different Availability Zones within the same VPC
- **Non-overlapping CIDR ranges** for the VPC, pod network, and service network
- **A NAT gateway** if setting `newNatGateway` to `false` (nodes in private VSwitches need outbound internet access)
- **An SLS log project** if configuring control plane logging with an external project

## Quick Start

Create a file `ack-cluster.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudKubernetesCluster.my-cluster
spec:
  region: cn-hangzhou
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  podCidr: "172.20.0.0/16"
  serviceCidr: "172.21.0.0/20"
  addons:
    - name: flannel
    - name: csi-plugin
    - name: csi-provisioner
```

Deploy:

```shell
openmcf apply -f ack-cluster.yaml
```

This creates a standard-tier ACK cluster with Flannel CNI across two Availability Zones.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region for the cluster (e.g., `cn-hangzhou`, `us-west-1`). Must match the region of the specified VSwitches. | Required; non-empty |
| `vswitchIds` | `StringValueOrRef[]` | VSwitch IDs for control plane and default worker node placement. Use VSwitches in distinct AZs for high availability. | 1–5 items required |
| `vswitchIds[].value` | `string` | Direct VSwitch ID value. | — |
| `vswitchIds[].valueFrom` | `object` | Foreign key reference to an AlicloudVswitch resource. | Default kind: `AlicloudVswitch`, field: `status.outputs.vswitch_id` |
| `serviceCidr` | `string` | CIDR block for Kubernetes ClusterIP services. Must not overlap the VPC CIDR, pod CIDR, or node CIDR. Immutable after creation. | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Cluster name. 1–63 characters; must start with a letter or digit. |
| `version` | `string` | Latest stable | Kubernetes version (e.g., `"1.28"`, `"1.30"`). |
| `clusterSpec` | `string` | `ack.standard` | Cluster tier. `ack.standard` (free) or `ack.pro.small` (paid, enhanced SLA). |
| `clusterDomain` | `string` | `cluster.local` | Kubernetes service discovery domain. Immutable after creation. |
| `podCidr` | `string` | — | Pod network CIDR for Flannel CNI. Required when using the `flannel` addon. Immutable. |
| `podVswitchIds` | `StringValueOrRef[]` | `[]` | VSwitch IDs for pod ENI allocation with Terway CNI. Required when using `terway-eniip`. |
| `proxyMode` | `string` | `ipvs` | kube-proxy mode: `ipvs` or `iptables`. Immutable after creation. |
| `nodeCidrMask` | `int` | `24` | Per-node pod CIDR mask size. Range: 24–28. A `/24` gives ~253 pods per node. Immutable. |
| `newNatGateway` | `bool` | `true` | Whether ACK auto-creates a NAT gateway. Set to `false` when managing your own AlicloudNatGateway. |
| `slbInternetEnabled` | `bool` | `true` | Whether to create a public SLB for the Kubernetes API server. |
| `securityGroupId` | `StringValueOrRef` | Auto-created | Security group for cluster nodes. Can reference an AlicloudSecurityGroup via `valueFrom`. |
| `isEnterpriseSecurityGroup` | `bool` | `false` | Auto-create an advanced security group (65,536 rules, 100,000 ENIs). Conflicts with `securityGroupId`. |
| `enableRrsa` | `bool` | `false` | Enable RRSA for pod-level IAM via OIDC federation. Cannot be disabled once enabled. |
| `deletionProtection` | `bool` | `false` | Prevent accidental cluster deletion via the API. |
| `encryptionProviderKey` | `StringValueOrRef` | — | KMS key ID for encrypting Kubernetes Secrets at rest. Immutable. Can reference AlicloudKmsKey. |
| `customSan` | `string` | — | Additional SANs for the API server TLS certificate (comma-separated IPs or domains). |
| `addons` | `AlicloudKubernetesAddon[]` | `[]` | Addons to install at creation time. See addon fields below. |
| `addons[].name` | `string` | — | Addon identifier (e.g., `flannel`, `terway-eniip`, `csi-plugin`). Required. |
| `addons[].config` | `string` | — | JSON-encoded addon configuration. |
| `addons[].version` | `string` | Default for K8s version | Addon version override. |
| `addons[].disabled` | `bool` | `false` | Register but do not install the addon. |
| `logging` | `object` | — | Control plane and audit logging configuration. |
| `logging.controlPlaneLogProject` | `StringValueOrRef` | Auto-created | SLS project for control plane logs. Can reference AlicloudLogProject. |
| `logging.controlPlaneLogTtl` | `string` | `"30"` | Log retention in days. |
| `logging.controlPlaneLogComponents` | `string[]` | `[]` | Components to log: `apiserver`, `kcm`, `scheduler`, `ccm`, `controlplane-events`, `alb`, `coredns`. |
| `logging.auditLogEnabled` | `bool` | `false` | Enable Kubernetes audit logging. |
| `logging.auditLogSlsProject` | `string` | Same as control plane | Separate SLS project for audit logs. |
| `maintenanceWindow` | `object` | — | Maintenance window for controlled patching. |
| `maintenanceWindow.enable` | `bool` | — | Whether the window is active. |
| `maintenanceWindow.maintenanceTime` | `string` | — | Start time in RFC 3339 format. |
| `maintenanceWindow.duration` | `string` | — | Window duration (e.g., `"3h"`). |
| `maintenanceWindow.weeklyPeriod` | `string` | `Thursday` | Day(s) of the week (e.g., `"Monday,Thursday"`). |
| `autoUpgrade` | `object` | — | Automatic version upgrade policy. Requires a maintenance window. |
| `autoUpgrade.enabled` | `bool` | — | Whether auto-upgrade is active. |
| `autoUpgrade.channel` | `string` | `patch` | Upgrade channel: `patch`, `stable`, or `rapid`. |
| `tags` | `map<string, string>` | `{}` | Tags applied to the cluster and auto-created resources. |
| `resourceGroupId` | `string` | Default group | Alibaba Cloud resource group for organizational grouping. |
| `timezone` | `string` | — | IANA timezone for cluster nodes (e.g., `"Asia/Shanghai"`). |

## Examples

### Development Cluster with Flannel

A minimal cluster for development and testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: dev-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudKubernetesCluster.dev-cluster
spec:
  region: cn-hangzhou
  vswitchIds:
    - value: vsw-aaa111
    - value: vsw-bbb222
  podCidr: "172.20.0.0/16"
  serviceCidr: "172.21.0.0/20"
  addons:
    - name: flannel
    - name: csi-plugin
    - name: csi-provisioner
```

### Terway Cluster with Pod-Level IAM

A staging cluster using ENI-based networking with RRSA and control plane logging.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: staging-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AlicloudKubernetesCluster.staging-cluster
spec:
  region: cn-shanghai
  clusterSpec: ack.pro.small
  version: "1.30"
  vswitchIds:
    - value: vsw-node-a
    - value: vsw-node-b
  podVswitchIds:
    - value: vsw-pod-a
    - value: vsw-pod-b
  serviceCidr: "172.21.0.0/20"
  enableRrsa: true
  newNatGateway: false
  addons:
    - name: terway-eniip
    - name: csi-plugin
    - name: csi-provisioner
    - name: arms-prometheus
    - name: metrics-server
  logging:
    controlPlaneLogProject:
      value: staging-logs
    controlPlaneLogComponents:
      - apiserver
      - kcm
      - scheduler
    auditLogEnabled: true
  tags:
    team: platform
    env: staging
```

### Production Cluster with Full Security and Lifecycle

A production-grade cluster with Secrets encryption, maintenance windows, auto-upgrade, and deletion protection.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: prod-cluster
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.AlicloudKubernetesCluster.prod-cluster
spec:
  region: cn-hangzhou
  name: acme-prod-ack
  clusterSpec: ack.pro.small
  version: "1.30"
  vswitchIds:
    - value: vsw-node-a
    - value: vsw-node-b
    - value: vsw-node-c
  podVswitchIds:
    - value: vsw-pod-a
    - value: vsw-pod-b
    - value: vsw-pod-c
  serviceCidr: "172.21.0.0/20"
  proxyMode: ipvs
  nodeCidrMask: 26
  newNatGateway: false
  slbInternetEnabled: true
  securityGroupId:
    value: sg-prod-ack
  enableRrsa: true
  deletionProtection: true
  encryptionProviderKey:
    value: kms-prod-key-id
  addons:
    - name: terway-eniip
    - name: csi-plugin
    - name: csi-provisioner
    - name: logtail-ds
      config: '{"IngressDashboardEnabled":"true","sls_project_name":"acme-prod-logs"}'
    - name: arms-prometheus
    - name: metrics-server
    - name: ack-node-problem-detector
  logging:
    controlPlaneLogProject:
      value: acme-prod-logs
    controlPlaneLogTtl: "90"
    controlPlaneLogComponents:
      - apiserver
      - kcm
      - scheduler
      - ccm
      - controlplane-events
    auditLogEnabled: true
    auditLogSlsProject: acme-prod-audit
  maintenanceWindow:
    enable: true
    maintenanceTime: "2026-03-01T03:00:00+08:00"
    duration: "3h"
    weeklyPeriod: Wednesday
  autoUpgrade:
    enabled: true
    channel: patch
  tags:
    team: platform
    cost-center: infra-001
  resourceGroupId: rg-acme-prod
  timezone: Asia/Shanghai
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: ref-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AlicloudKubernetesCluster.ref-cluster
spec:
  region: cn-hangzhou
  vswitchIds:
    - valueFrom:
        kind: AlicloudVswitch
        name: node-vsw-a
        field: status.outputs.vswitch_id
    - valueFrom:
        kind: AlicloudVswitch
        name: node-vsw-b
        field: status.outputs.vswitch_id
  podVswitchIds:
    - valueFrom:
        kind: AlicloudVswitch
        name: pod-vsw-a
        field: status.outputs.vswitch_id
    - valueFrom:
        kind: AlicloudVswitch
        name: pod-vsw-b
        field: status.outputs.vswitch_id
  serviceCidr: "172.21.0.0/20"
  securityGroupId:
    valueFrom:
      kind: AlicloudSecurityGroup
      name: ack-sg
      field: status.outputs.security_group_id
  encryptionProviderKey:
    valueFrom:
      kind: AlicloudKmsKey
      name: ack-kms-key
      field: status.outputs.key_id
  enableRrsa: true
  addons:
    - name: terway-eniip
    - name: csi-plugin
    - name: csi-provisioner
  logging:
    controlPlaneLogProject:
      valueFrom:
        kind: AlicloudLogProject
        name: platform-logs
        field: status.outputs.project_name
    auditLogEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | ACK cluster ID assigned by Alibaba Cloud |
| `cluster_name` | `string` | Cluster name as created |
| `api_server_internet` | `string` | Public API server endpoint (empty when `slbInternetEnabled` is `false`) |
| `api_server_intranet` | `string` | Private (VPC-internal) API server endpoint |
| `vpc_id` | `string` | VPC ID computed from the provided VSwitches |
| `security_group_id` | `string` | Security group ID used by cluster nodes (user-provided or auto-created) |
| `nat_gateway_id` | `string` | NAT gateway ID (empty when `newNatGateway` is `false`) |
| `worker_ram_role_name` | `string` | RAM role name attached to worker nodes |
| `rrsa_oidc_issuer_url` | `string` | RRSA OIDC issuer URL for pod IAM trust policies (empty when `enableRrsa` is `false`) |
| `ram_oidc_provider_name` | `string` | RRSA OIDC provider name in RAM (empty when `enableRrsa` is `false`) |
| `ram_oidc_provider_arn` | `string` | RRSA OIDC provider ARN (empty when `enableRrsa` is `false`) |

## Related Components

- [AlicloudVpc](/docs/catalog/alicloud/alicloudvpc) — provides the VPC for cluster networking
- [AlicloudVswitch](/docs/catalog/alicloud/alicloudvswitch) — provides VSwitches for node and pod placement
- [AlicloudSecurityGroup](/docs/catalog/alicloud/alicloudsecuritygroup) — controls inbound and outbound traffic for cluster nodes
- [AlicloudNatGateway](/docs/catalog/alicloud/alicloudnatgateway) — provides outbound internet access for nodes in private VSwitches
- [AlicloudKmsKey](/docs/catalog/alicloud/alicloudkmskey) — provides the encryption key for Kubernetes Secrets at rest
- [AlicloudLogProject](/docs/catalog/alicloud/alicloudlogproject) — provides the SLS project for control plane and audit logs
- [AlicloudKubernetesNodePool](/docs/catalog/alicloud/alicloudkubernetesnodepool) — manages worker node pools within this cluster
