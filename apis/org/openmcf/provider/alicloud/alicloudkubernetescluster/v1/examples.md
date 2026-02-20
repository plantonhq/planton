# AlicloudKubernetesCluster Examples

## Minimal Cluster with Flannel

A basic development cluster using Flannel overlay networking.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: dev-cluster
spec:
  region: cn-hangzhou
  vswitchIds:
    - vsw-aaa111
    - vsw-bbb222
  podCidr: "172.20.0.0/16"
  serviceCidr: "172.21.0.0/20"
  addons:
    - name: flannel
    - name: csi-plugin
    - name: csi-provisioner
```

## Terway Cluster with RRSA

A cluster using Terway ENI-based networking with RRSA enabled for pod IAM.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: staging-cluster
spec:
  region: cn-shanghai
  clusterSpec: ack.pro.small
  version: "1.30"
  vswitchIds:
    - vsw-node-a
    - vsw-node-b
  podVswitchIds:
    - vsw-pod-a
    - vsw-pod-b
  serviceCidr: "172.21.0.0/20"
  enableRrsa: true
  newNatGateway: false
  addons:
    - name: terway-eniip
    - name: csi-plugin
    - name: csi-provisioner
    - name: logtail-ds
      config: '{"IngressDashboardEnabled":"true","sls_project_name":"staging-logs"}'
    - name: arms-prometheus
    - name: metrics-server
  logging:
    controlPlaneLogProject: staging-logs
    controlPlaneLogComponents:
      - apiserver
      - kcm
      - scheduler
    auditLogEnabled: true
  tags:
    team: platform
    env: staging
```

## Production Cluster (Full Configuration)

A production-grade cluster with professional SLA, Terway networking, Secrets
encryption, maintenance windows, and auto-upgrade.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudKubernetesCluster
metadata:
  name: production-cluster
  org: acme-corp
  env: production
spec:
  region: cn-hangzhou
  name: acme-prod-ack
  clusterSpec: ack.pro.small
  version: "1.30"
  clusterDomain: cluster.local

  vswitchIds:
    - vsw-node-a
    - vsw-node-b
    - vsw-node-c
  podVswitchIds:
    - vsw-pod-a
    - vsw-pod-b
    - vsw-pod-c
  serviceCidr: "172.21.0.0/20"
  proxyMode: ipvs
  nodeCidrMask: 26
  newNatGateway: false
  slbInternetEnabled: true

  securityGroupId: sg-prod-ack
  enableRrsa: true
  deletionProtection: true
  encryptionProviderKey: kms-prod-key-id
  customSan: "10.0.0.1,api.acme.internal"

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
    controlPlaneLogProject: acme-prod-logs
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
