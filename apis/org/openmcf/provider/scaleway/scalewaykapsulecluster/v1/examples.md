# ScalewayKapsuleCluster Examples

## 1. Minimal Development Cluster

A single-node cluster for development and testing. Non-HA, no autoscaling, smallest instance type.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: dev-cluster
spec:
  region: fr-par
  kubernetesVersion: "1.32"
  cni: cilium
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  deleteAdditionalResources: true
  defaultNodePool:
    nodeType: DEV1-M
    size: 1
```

**Estimated cost**: ~15 EUR/month (1x DEV1-M node, mutualized control plane is free).

## 2. Production Cluster with Autoscaling and Auto-Upgrade

A production-ready cluster with autoscaling, autohealing, auto-upgrade, and nodes isolated from the public internet.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: prod-cluster
  org: mycompany
  env: production
spec:
  region: fr-par
  kubernetesVersion: "1.32"
  cni: cilium
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  deleteAdditionalResources: false
  autoUpgrade:
    enable: true
    maintenanceWindowStartHour: 3
    maintenanceWindowDay: sunday
  autoscalerConfig:
    scaleDownDelayAfterAdd: "15m"
    scaleDownUnneededTime: "15m"
    scaleDownUtilizationThreshold: 0.5
    expander: least-waste
    maxGracefulTerminationSec: 600
  featureGates:
    - GracefulNodeShutdown
  admissionPlugins:
    - AlwaysPullImages
  defaultNodePool:
    name: system
    nodeType: PRO2-S
    size: 3
    autoScale: true
    minSize: 3
    maxSize: 10
    autohealing: true
    publicIpDisabled: true
    upgradePolicy:
      maxSurge: 1
      maxUnavailable: 0
```

**Estimated cost**: ~120-400 EUR/month (3-10x PRO2-S nodes depending on autoscaler).

## 3. Infra Chart Composition with ValueFrom References

Shows how the cluster is composed with upstream resources in an infra chart template. The `privateNetworkId` is wired via `valueFrom` to create a dependency edge.

```yaml
# VPC (Layer 0)
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayVpc
metadata:
  name: "{{ values.env }}-vpc"
spec:
  region: "{{ values.region }}"
  enableRouting: true

# Private Network (Layer 1)
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayPrivateNetwork
metadata:
  name: "{{ values.env }}-network"
spec:
  region: "{{ values.region }}"
  vpcId:
    valueFrom:
      kind: ScalewayVpc
      name: "{{ values.env }}-vpc"
      fieldPath: status.outputs.vpc_id

# Kapsule Cluster (Layer 2)
---
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: "{{ values.env }}-{{ values.cluster_name }}"
spec:
  region: "{{ values.region }}"
  kubernetesVersion: "{{ values.kubernetes_version }}"
  cni: cilium
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: "{{ values.env }}-network"
      fieldPath: status.outputs.private_network_id
  deleteAdditionalResources: true
  {% if values.auto_upgrade | bool %}
  autoUpgrade:
    enable: true
    maintenanceWindowStartHour: 3
    maintenanceWindowDay: sunday
  {% endif %}
  defaultNodePool:
    nodeType: "{{ values.node_type }}"
    size: {{ values.node_count }}
    autoScale: {{ values.auto_scale }}
    {% if values.auto_scale | bool %}
    minSize: {{ values.min_nodes }}
    maxSize: {{ values.max_nodes }}
    {% endif %}
    autohealing: true
    publicIpDisabled: true
```

## 4. Dedicated Control Plane Cluster

For production workloads requiring isolated API server resources and SLA guarantees.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsuleCluster
metadata:
  name: enterprise-cluster
  org: mycompany
  env: production
spec:
  region: nl-ams
  kubernetesVersion: "1.32"
  cni: cilium
  type: kapsule-dedicated-8
  privateNetworkId:
    value: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  description: "Enterprise production cluster with dedicated control plane"
  deleteAdditionalResources: false
  autoUpgrade:
    enable: true
    maintenanceWindowStartHour: 2
    maintenanceWindowDay: saturday
  autoscalerConfig:
    expander: least-waste
    balanceSimilarNodeGroups: true
  podCidr: "10.200.0.0/14"
  serviceCidr: "10.204.0.0/20"
  defaultNodePool:
    name: general
    nodeType: GP1-S
    size: 5
    autoScale: true
    minSize: 3
    maxSize: 8
    autohealing: true
    publicIpDisabled: true
    rootVolumeSizeInGb: 100
    upgradePolicy:
      maxSurge: 2
      maxUnavailable: 0
```

**Estimated cost**: ~500+ EUR/month (dedicated control plane + 3-8x GP1-S nodes).

## Deployment

```bash
# Preview changes
openmcf pulumi preview --manifest cluster.yaml

# Apply
openmcf pulumi up --manifest cluster.yaml --yes

# Get kubeconfig
openmcf stack-outputs --manifest cluster.yaml | jq -r '.kubeconfig' > kubeconfig.yaml
export KUBECONFIG=kubeconfig.yaml
kubectl get nodes
```
