# AliCloudKubernetesCluster Terraform Examples

Below are several examples demonstrating how to deploy ACK managed clusters with the OpenMCF Terraform module.

After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic Flannel Cluster

A minimal development cluster using Flannel overlay networking.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudKubernetesCluster
metadata:
  name: dev-cluster
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

This example:
- Uses Flannel CNI with a `/16` pod CIDR
- Deploys across two AZs via two VSwitches
- Installs storage CSI drivers for PersistentVolume support
- Uses defaults for all optional settings (`ack.standard`, `ipvs`, NAT auto-created)

---

## Terway Cluster with RRSA

A staging cluster with ENI-based networking and pod-level IAM.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudKubernetesCluster
metadata:
  name: staging-cluster
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
    auditLogEnabled: true
```

This example:
- Uses professional tier for enhanced SLA
- Terway CNI with dedicated pod VSwitches
- RRSA enabled for Kubernetes service account → RAM role federation
- External NAT gateway management
- Control plane and audit logging to SLS

---

## Production Cluster with Encryption and Maintenance

A production-grade cluster with all security and lifecycle features enabled.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudKubernetesCluster
metadata:
  name: prod-cluster
  org: acme-corp
  env: production
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

This configuration:
- Three-AZ deployment with Terway for maximum fault tolerance
- KMS encryption for Kubernetes Secrets at rest
- Deletion protection prevents accidental cluster destruction
- Wednesday 3 AM maintenance window for controlled patching
- Auto-upgrade on `patch` channel for security fixes only
- 90-day log retention with separate audit log project

---

## After Deploying

Verify the cluster using the Alibaba Cloud CLI:

```bash
# Get cluster details
aliyun cs DescribeClusterDetail --ClusterId <cluster-id>

# Get kubeconfig
aliyun cs DescribeClusterUserKubeconfig --ClusterId <cluster-id>

# List node pools
aliyun cs DescribeClusterNodePools --ClusterId <cluster-id>
```
