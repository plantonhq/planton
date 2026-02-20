# OCI Container Engine Node Pool Examples

This document provides practical examples for deploying Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) node pools using the OpenMCF API. Each example demonstrates a different use case, progressing from a minimal development pool to a fully configured production pool with encryption, VCN-native networking, and rolling upgrade strategies.

## Table of Contents

- [Example 1: Minimal Development Pool](#example-1-minimal-development-pool)
- [Example 2: Production Multi-AD Pool with VCN-Native CNI](#example-2-production-multi-ad-pool-with-vcn-native-cni)
- [Example 3: GPU Node Pool with Labels](#example-3-gpu-node-pool-with-labels)
- [Example 4: Preemptible Batch Processing Pool](#example-4-preemptible-batch-processing-pool)
- [Example 5: ARM-Based Cost-Optimized Pool](#example-5-arm-based-cost-optimized-pool)
- [Example 6: Pool with Cycling and Eviction Settings](#example-6-pool-with-cycling-and-eviction-settings)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Example 1: Minimal Development Pool

**Use Case:** A development or testing node pool with the minimum required configuration. Uses an E4 Flex shape with modest resources in a single availability domain.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (2 OCPUs, 32 GB)
- **Nodes:** 3
- **Placement:** Single AD
- **CNI:** Inherits from cluster (no explicit pod networking config)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: dev-pool
  org: my-org
  env: dev
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciContainerEngineNodePool.dev-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  clusterId:
    value: "ocid1.cluster.oc1.iad.example"
  nodeShape: "VM.Standard.E4.Flex"
  nodeShapeConfig:
    ocpus: 2
    memoryInGbs: 32
  nodeConfigDetails:
    size: 3
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example"
```

**Deploy with OpenMCF CLI:**

```shell
openmcf apply -f dev-pool.yaml
```

**What happens:**
- A 3-node pool is created with E4 Flex instances (2 OCPUs, 32 GB RAM each) in a single availability domain.
- OKE launches the compute instances, installs the kubelet and kube-proxy, and joins the nodes to the cluster.
- The Kubernetes version is inherited from the cluster — no explicit version is set.
- The node pool ID and Kubernetes version are exported as stack outputs.
- Since no `nodeSourceDetails` is specified, OKE uses the default Oracle Linux image for the cluster's Kubernetes version.

---

## Example 2: Production Multi-AD Pool with VCN-Native CNI

**Use Case:** A production node pool distributed across three availability domains with VCN-native pod networking, NSGs on nodes and pods, KMS boot volume encryption, in-transit encryption, and all infrastructure references using `valueFrom` for declarative composition.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (4 OCPUs, 64 GB)
- **Nodes:** 9 (3 per AD)
- **Placement:** Three ADs with regional subnet
- **CNI:** VCN-native (pods get VCN IPs)
- **Encryption:** KMS boot volume + in-transit
- **Image:** Custom image with 100 GB boot volume

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: prod-pool
  org: acme-corp
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: oke-platform
    pulumi.openmcf.org/stack.name: prod.OciContainerEngineNodePool.prod-pool
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  clusterId:
    valueFrom:
      kind: OciContainerEngineCluster
      name: prod-cluster
      fieldPath: status.outputs.clusterId
  nodeShape: "VM.Standard.E4.Flex"
  nodeShapeConfig:
    ocpus: 4
    memoryInGbs: 64
  kubernetesVersion: "v1.28.2"
  nodeSourceDetails:
    imageId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 100
  sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQ..."
  nodeConfigDetails:
    size: 9
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          valueFrom:
            kind: OciSubnet
            name: worker-subnet
            fieldPath: status.outputs.subnetId
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          valueFrom:
            kind: OciSubnet
            name: worker-subnet
            fieldPath: status.outputs.subnetId
      - availabilityDomain: "Uocm:PHX-AD-3"
        subnetId:
          valueFrom:
            kind: OciSubnet
            name: worker-subnet
            fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: worker-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    kmsKeyId:
      value: "ocid1.key.oc1.iad.example"
    isPvEncryptionInTransitEnabled: true
    podNetworkOptionDetails:
      cniType: oci_vcn_ip_native
      maxPodsPerNode: 31
      podNsgIds:
        - valueFrom:
            kind: OciSecurityGroup
            name: pod-nsg
            fieldPath: status.outputs.networkSecurityGroupId
      podSubnetIds:
        - valueFrom:
            kind: OciSubnet
            name: pod-subnet
            fieldPath: status.outputs.subnetId
  initialNodeLabels:
    - key: "workload-type"
      value: "general"
    - key: "environment"
      value: "production"
  nodeEvictionSettings:
    evictionGraceDuration: "PT30M"
    isForceDeleteAfterGraceDuration: true
  nodePoolCyclingDetails:
    isNodeCyclingEnabled: true
    maximumSurge: "1"
    maximumUnavailable: "0"
```

**What happens:**
- 9 nodes are distributed evenly across 3 availability domains (3 per AD) for high availability.
- Each node runs E4 Flex with 4 OCPUs and 64 GB RAM, using a custom image with a 100 GB boot volume.
- VCN-native pod networking assigns VCN IP addresses to each pod. Pod VNICs are protected by a dedicated pod NSG, and pods are allocated IPs from a dedicated pod subnet.
- Boot volumes are encrypted with a customer-managed KMS key. In-transit encryption protects data between each instance and its paravirtualized volumes.
- Worker node VNICs are protected by a worker NSG that controls inbound/outbound traffic.
- Node pool cycling is enabled with maximum surge of 1 (create one new node before draining an old one) and maximum unavailable of 0 (never reduce capacity during upgrades).
- Eviction settings give pods 30 minutes to drain gracefully. If pods cannot be evicted in time, the instance is force-deleted to avoid blocking the upgrade.

---

## Example 3: GPU Node Pool with Labels

**Use Case:** A GPU-accelerated node pool for ML training and inference workloads. Kubernetes labels enable workload scheduling via `nodeSelector`, ensuring only GPU-appropriate pods land on these nodes.

**Configuration:**
- **Shape:** VM.GPU.A10.1 (1 NVIDIA A10 GPU, 15 OCPUs, 240 GB)
- **Nodes:** 4
- **Placement:** Two ADs
- **Image:** GPU-optimized custom image with 200 GB boot volume

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: gpu-pool
  org: acme-corp
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: oke-platform
    pulumi.openmcf.org/stack.name: prod.OciContainerEngineNodePool.gpu-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  clusterId:
    value: "ocid1.cluster.oc1.iad.example"
  nodeShape: "VM.GPU.A10.1"
  nodeSourceDetails:
    imageId: "ocid1.image.oc1.iad.examplegpuimage"
    bootVolumeSizeInGbs: 200
  nodeConfigDetails:
    size: 4
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example"
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example2"
  initialNodeLabels:
    - key: "accelerator"
      value: "nvidia-a10"
    - key: "workload-type"
      value: "gpu"
  nodeMetadata:
    user_data: "IyEvYmluL2Jhc2gKZWNobyAnR1BVIG5vZGUgcmVhZHkn"
```

**What happens:**
- 4 GPU nodes (NVIDIA A10) are created across 2 availability domains (2 per AD).
- GPU shapes are fixed (not flex), so `nodeShapeConfig` is not needed — the shape provides 15 OCPUs and 240 GB RAM per node.
- A custom GPU-optimized OS image is used with a 200 GB boot volume to accommodate NVIDIA drivers and CUDA libraries.
- Kubernetes labels `accelerator=nvidia-a10` and `workload-type=gpu` are applied to each node. ML workloads use `nodeSelector` to target these nodes:

```yaml
nodeSelector:
  accelerator: nvidia-a10
```

- The `nodeMetadata` field provides cloud-init user data (base64-encoded) that runs during instance launch — used here for GPU driver initialization.

---

## Example 4: Preemptible Batch Processing Pool

**Use Case:** A cost-optimized node pool using preemptible (spot) instances for batch processing, CI/CD pipelines, or other fault-tolerant workloads. Preemptible instances cost 60-90% less than on-demand but can be reclaimed by OCI.

**Configuration:**
- **Shape:** VM.Standard.E4.Flex (2 OCPUs, 16 GB)
- **Nodes:** 10
- **Placement:** Three ADs, all preemptible
- **Preemption:** Boot volumes not preserved (cost savings)

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: batch-pool
  org: acme-corp
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: oke-platform
    pulumi.openmcf.org/stack.name: prod.OciContainerEngineNodePool.batch-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  clusterId:
    value: "ocid1.cluster.oc1.iad.example"
  nodeShape: "VM.Standard.E4.Flex"
  nodeShapeConfig:
    ocpus: 2
    memoryInGbs: 16
  nodeConfigDetails:
    size: 10
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example"
        preemptibleNodeConfig:
          isPreserveBootVolume: false
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example2"
        preemptibleNodeConfig:
          isPreserveBootVolume: false
      - availabilityDomain: "Uocm:PHX-AD-3"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example3"
        preemptibleNodeConfig:
          isPreserveBootVolume: false
  initialNodeLabels:
    - key: "workload-type"
      value: "batch"
    - key: "instance-lifecycle"
      value: "preemptible"
  nodeEvictionSettings:
    evictionGraceDuration: "PT5M"
    isForceDeleteAfterGraceDuration: true
```

**What happens:**
- 10 preemptible nodes are distributed across 3 availability domains. Spreading across ADs reduces the risk of simultaneous reclamation.
- Boot volumes are not preserved on termination (`isPreserveBootVolume: false`), reducing storage costs for ephemeral batch workers.
- Kubernetes labels identify these nodes as preemptible, enabling scheduling policies that place only fault-tolerant workloads on them.
- Short eviction grace duration (5 minutes) with force delete ensures rapid node recycling during scale-down or upgrades. Batch workloads should implement checkpoint/retry logic.

**Scheduling workloads to preemptible nodes:**

```yaml
nodeSelector:
  instance-lifecycle: preemptible
tolerations:
  - key: "node.kubernetes.io/unreachable"
    operator: "Exists"
    effect: "NoExecute"
    tolerationSeconds: 30
```

---

## Example 5: ARM-Based Cost-Optimized Pool

**Use Case:** An Ampere A1 Flex (ARM64) node pool for cost-optimized workloads. A1 Flex instances provide approximately 50% cost savings compared to x86 E4 Flex at equivalent performance for ARM-compatible workloads, with capacity reservations guaranteeing availability.

**Configuration:**
- **Shape:** VM.Standard.A1.Flex (4 OCPUs, 24 GB)
- **Nodes:** 6
- **Placement:** Two ADs with capacity reservations and fault domain constraints

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: arm-pool
  org: acme-corp
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: oke-platform
    pulumi.openmcf.org/stack.name: prod.OciContainerEngineNodePool.arm-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  clusterId:
    value: "ocid1.cluster.oc1.iad.example"
  nodeShape: "VM.Standard.A1.Flex"
  nodeShapeConfig:
    ocpus: 4
    memoryInGbs: 24
  nodeConfigDetails:
    size: 6
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example"
        capacityReservationId:
          value: "ocid1.capacityreservation.oc1.iad.example"
        faultDomains:
          - "FAULT-DOMAIN-1"
          - "FAULT-DOMAIN-2"
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example2"
        capacityReservationId:
          value: "ocid1.capacityreservation.oc1.iad.example2"
        faultDomains:
          - "FAULT-DOMAIN-1"
          - "FAULT-DOMAIN-2"
  initialNodeLabels:
    - key: "arch"
      value: "arm64"
    - key: "cost-profile"
      value: "optimized"
```

**What happens:**
- 6 ARM64 nodes are created across 2 ADs (3 per AD), each with 4 OCPUs and 24 GB RAM.
- Capacity reservations guarantee compute availability — nodes are not subject to capacity shortages that can affect on-demand provisioning.
- Fault domain constraints limit nodes to FD-1 and FD-2 within each AD. This is useful when FD-3 has known capacity issues or when aligning with compliance requirements.
- The `arch=arm64` label enables multi-arch scheduling. Container images must be built for `linux/arm64`:

```yaml
nodeSelector:
  arch: arm64
```

---

## Example 6: Pool with Cycling and Eviction Settings

**Use Case:** A node pool configured for controlled rolling upgrades with specific eviction behavior. Demonstrates the full cycling and eviction configuration for environments where zero-downtime upgrades are critical.

**Configuration:**
- **Cycling:** Max surge 25%, max unavailable 0
- **Eviction:** 45-minute grace, force action after grace

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: controlled-upgrade-pool
  org: acme-corp
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: oke-platform
    pulumi.openmcf.org/stack.name: prod.OciContainerEngineNodePool.controlled-upgrade-pool
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  clusterId:
    value: "ocid1.cluster.oc1.iad.example"
  nodeShape: "VM.Standard.E4.Flex"
  nodeShapeConfig:
    ocpus: 4
    memoryInGbs: 64
  kubernetesVersion: "v1.28.2"
  nodeConfigDetails:
    size: 12
    placementConfigs:
      - availabilityDomain: "Uocm:PHX-AD-1"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example"
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example2"
      - availabilityDomain: "Uocm:PHX-AD-3"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example3"
  nodeEvictionSettings:
    evictionGraceDuration: "PT45M"
    isForceActionAfterGraceDuration: true
    isForceDeleteAfterGraceDuration: false
  nodePoolCyclingDetails:
    isNodeCyclingEnabled: true
    maximumSurge: "25%"
    maximumUnavailable: "0"
```

**What happens:**
- During a Kubernetes version upgrade (e.g., changing `kubernetesVersion` from `v1.28.2` to `v1.29.1`), OKE replaces nodes using the cycling strategy:
  1. Up to 25% of the pool size (3 nodes out of 12) can be created as surge capacity simultaneously.
  2. Maximum unavailable is 0 — the pool never drops below 12 healthy nodes during the upgrade.
  3. For each node being replaced: OKE cordons the node, then drains pods with a 45-minute grace period.
  4. If pods cannot be evicted within 45 minutes, `isForceActionAfterGraceDuration: true` means OKE proceeds with the node action (cordon + mark for replacement).
  5. `isForceDeleteAfterGraceDuration: false` means the compute instance is NOT force-deleted — OKE waits for pods to terminate. This protects stateful workloads that need orderly shutdown.
- The combination of 25% surge with 0 unavailable ensures the cluster always has full capacity during upgrades, at the cost of temporary additional compute.

**Cycling strategies compared:**

| Strategy | maximumSurge | maximumUnavailable | Behavior |
|----------|-------------|-------------------|----------|
| Zero-downtime (this example) | `"25%"` | `"0"` | Create new nodes first, then drain old ones. Full capacity maintained. Higher temporary cost. |
| Balanced | `"1"` | `"1"` | Create one new node, drain one old node in parallel. Slight capacity reduction. Moderate speed. |
| Fast replacement | `"50%"` | `"25%"` | Aggressive cycling — half the pool can surge, quarter can be unavailable. Fastest upgrade, highest risk. |
| Budget-conscious | `"0"` | `"1"` | Drain one node at a time, then create replacement. Slowest but no surge cost. |

---

## Common Operations

### Get Node Pool Status

After deploying a node pool, check the node pool ID and Kubernetes version from stack outputs:

```shell
# Pulumi
pulumi stack output node_pool_id
pulumi stack output kubernetes_version

# Terraform
terraform output node_pool_id
terraform output kubernetes_version
```

### List Nodes in the Pool

Use the OCI CLI to list individual nodes and their status:

```shell
NODE_POOL_ID=$(pulumi stack output node_pool_id)

oci ce node-pool get --node-pool-id "$NODE_POOL_ID" \
  --query 'data.nodes[*].{name:"name",state:"lifecycle-state",ad:"availability-domain"}' \
  --output table
```

### Verify Nodes in Kubernetes

After deployment, verify nodes have joined the cluster:

```shell
# List nodes with labels
kubectl get nodes -l workload-type=general

# Check node details
kubectl describe node <node-name>
```

### Scale the Node Pool

To change the number of nodes, update `nodeConfigDetails.size` in the manifest and re-apply:

```yaml
nodeConfigDetails:
  size: 12  # was 9
```

```shell
openmcf apply -f prod-pool.yaml
```

OKE scales up by launching new instances in the configured availability domains and scales down by draining and terminating instances (respecting eviction settings if configured).

### Upgrade Kubernetes Version

To upgrade worker nodes to a new Kubernetes version:

1. Verify the cluster control plane is already at the target version (or higher).
2. Update `kubernetesVersion` in the manifest:

```yaml
kubernetesVersion: "v1.29.1"  # was v1.28.2
```

3. Re-apply:

```shell
openmcf apply -f prod-pool.yaml
```

OKE replaces nodes using the configured cycling strategy. If no cycling details are configured, OKE uses its default replacement behavior.

### Use Node Pool ID in Downstream Resources

The `node_pool_id` output can be referenced by downstream resources:

```yaml
spec:
  nodePoolId:
    valueFrom:
      kind: OciContainerEngineNodePool
      name: prod-pool
      fieldPath: status.outputs.nodePoolId
```

### Check Available Shapes and Images

```shell
# List available compute shapes
oci compute shape list --compartment-id "$COMPARTMENT_ID" \
  --query 'data[*].shape' --output table

# List available node pool images for a Kubernetes version
oci ce node-pool-options get --node-pool-option-id all \
  --compartment-id "$COMPARTMENT_ID"
```

---

## Best Practices

### Shape Selection

| Workload Type | Recommended Shape | OCPU / Memory | Rationale |
|---------------|------------------|---------------|-----------|
| General purpose | `VM.Standard.E4.Flex` | 2-4 / 16-64 GB | AMD EPYC, good price-performance balance. Flex allows precise sizing. |
| Memory intensive | `VM.Standard.E4.Flex` | 2-4 / 128-256 GB | Same shape, higher memory ratio (up to 64 GB per OCPU). |
| CPU intensive | `VM.Standard.E4.Flex` | 8-16 / 32-64 GB | Higher OCPU count, lower memory ratio. |
| Cost optimized | `VM.Standard.A1.Flex` | 4-8 / 24-48 GB | Ampere Altra ARM64. ~50% cost savings for ARM-compatible workloads. |
| GPU / ML | `VM.GPU.A10.1` or `.2` | Fixed | NVIDIA A10 GPU. Fixed shape — no nodeShapeConfig needed. |
| Batch / ephemeral | `VM.Standard.E4.Flex` + preemptible | 2 / 16 GB | Small instances with preemptible config for max cost savings. |

**Flex shapes require `nodeShapeConfig`.** Without it, OCPUs and memory default to the shape's minimum (typically 1 OCPU, 1 GB). Always set `nodeShapeConfig` for flex shapes.

### Availability Domain Placement

- **Production:** Use all 3 ADs in a region (where available). OKE distributes nodes evenly. Combined with Kubernetes topology spread constraints, this provides infrastructure-level HA.
- **Development:** A single AD is acceptable. Reduces complexity and avoids cross-AD data transfer costs.
- **Regions with fewer ADs:** Some OCI regions have only 1 AD. In these regions, use fault domains within that AD for infrastructure diversity.
- **Regional subnets:** When using a regional subnet, provide the same subnet OCID in each placement config — OCI handles AD-specific routing.

### Node Pool Sizing

- **Start small, scale up.** Begin with a conservative size and use horizontal pod autoscaling (HPA) and cluster autoscaler to right-size dynamically.
- **Account for system overhead.** Each node reserves resources for the kubelet, kube-proxy, and OCI agents. Approximately 10-15% of node resources are unavailable for user pods.
- **Max pods per node (VCN-native).** The `maxPodsPerNode` setting directly affects how many VCN IPs are reserved per node from the pod subnet. A 31-pod limit on a 100-node pool consumes 3,100 pod subnet IPs at peak. Size pod subnets accordingly.
- **Odd numbers for HA.** For pools running stateful workloads with quorum requirements (etcd, Kafka, ZooKeeper), use odd node counts distributed evenly across ADs.

### Version Management

- Pin `kubernetesVersion` explicitly for production pools. Omitting it causes the pool to inherit the cluster version, which means a cluster upgrade immediately triggers worker node replacement.
- Upgrade the cluster control plane first, then upgrade node pools one at a time.
- The node pool version must be within one minor version of the cluster version. OKE rejects version jumps larger than one minor version.
- Test upgrades on a non-production pool first. Create a temporary pool with the new version and validate workloads before upgrading production pools.

### Cycling and Eviction

For production pools with zero-downtime requirements:

| Setting | Recommended Value | Rationale |
|---------|------------------|-----------|
| `isNodeCyclingEnabled` | `true` | Controlled replacement instead of disruptive recreation. |
| `maximumSurge` | `"1"` or `"25%"` | At least one surge node ensures capacity is maintained. 25% for faster upgrades on large pools. |
| `maximumUnavailable` | `"0"` | Never reduce pool capacity during upgrades. |
| `evictionGraceDuration` | `"PT30M"` | 30 minutes gives most workloads time for graceful shutdown. |
| `isForceActionAfterGraceDuration` | `true` | Prevents stuck upgrades when a pod cannot be evicted (e.g., no PDB allows disruption). |
| `isForceDeleteAfterGraceDuration` | `false` | Lets the instance shut down cleanly even after force action. Set to `true` only for stateless workloads. |

### Preemptible Instances

- **Spread across all ADs.** Capacity reclamation is per-AD. Spreading preemptible nodes across ADs reduces the chance of losing all nodes simultaneously.
- **Do not use for stateful workloads.** Preemptible instances can be terminated with minimal warning. Only fault-tolerant, restartable workloads should run on preemptible nodes.
- **Combine with on-demand.** Create a small on-demand pool for critical services and a larger preemptible pool for batch/elastic workloads. Use labels and taints to direct scheduling.
- **Set `isPreserveBootVolume: false`.** Unless you need boot volume forensics, delete boot volumes on termination to avoid accumulating orphaned volumes.

### Tagging and Labels

- Use `initialNodeLabels` to classify node pools by purpose (`workload-type`, `environment`, `arch`, `cost-profile`). This enables Kubernetes scheduling policies without post-creation label management.
- OpenMCF automatically applies freeform tags with `resource_kind`, `resource_id`, `organization`, and `environment`. Additional labels from `metadata.labels` are propagated as freeform tags.
- For cost tracking, set `metadata.org` and `metadata.env` — these become freeform tags on the OCI compute instances managed by the node pool.
