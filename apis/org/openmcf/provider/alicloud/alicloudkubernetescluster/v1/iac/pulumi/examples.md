# AliCloudKubernetesCluster Pulumi Examples

## CLI

```bash
# Deploy using the OpenMCF CLI
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Preview changes
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

---

## Minimal Cluster with Flannel

A basic development cluster using Flannel overlay networking with default settings.

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

**Key Points:**
- Uses Flannel CNI with a `/16` pod CIDR
- Two VSwitches in distinct AZs for high availability
- Storage CSI drivers installed for persistent volume support
- All other settings use defaults: `ack.standard`, `ipvs` proxy mode, NAT gateway auto-created

---

## Terway Cluster with RRSA and Logging

A staging cluster using Terway ENI-based networking with RRSA enabled for pod-level IAM.

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
    - name: logtail-ds
      config: '{"IngressDashboardEnabled":"true","sls_project_name":"staging-logs"}'
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

**Key Points:**
- Professional tier (`ack.pro.small`) for enhanced SLA
- Terway CNI with dedicated pod VSwitches (separate from node VSwitches)
- RRSA enabled for pod-level IAM via OIDC federation
- NAT gateway managed externally (`newNatGateway: false`)
- Control plane logging and audit logging to SLS

---

## Production Cluster (Full Configuration)

A production-grade cluster with Secrets encryption, maintenance windows, and auto-upgrade.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudKubernetesCluster
metadata:
  name: production-cluster
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

**Key Points:**
- Three-AZ deployment with Terway for maximum availability
- KMS encryption for Kubernetes Secrets at rest
- Deletion protection prevents accidental cluster destruction
- Maintenance window restricts patching to Wednesday 3 AM
- Auto-upgrade on `patch` channel for security fixes
- Custom SAN for internal API server access
- 90-day log retention with dedicated audit log project

---

**Next Steps:**

- See [README.md](./README.md) for CLI flows and debugging instructions
- See [overview.md](./overview.md) for module architecture details
