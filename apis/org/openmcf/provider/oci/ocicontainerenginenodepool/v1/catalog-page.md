# OCI Container Engine Node Pool

Deploys an Oracle Cloud Infrastructure Container Engine for Kubernetes (OKE) node pool — a managed set of worker nodes attached to an OKE cluster. Supports flex and fixed compute shapes, placement across multiple availability domains and fault domains, preemptible (spot) instances, VCN-native pod networking with per-pod NSGs, boot volume encryption via KMS, and rolling upgrade strategies. Worker nodes are managed independently of the cluster control plane, which is provisioned via OciContainerEngineCluster.

## What Gets Created

When you deploy an OciContainerEngineNodePool resource, OpenMCF provisions:

- **OKE Node Pool** — an `oci_containerengine_node_pool` resource in the specified compartment, attached to the target OKE cluster. The node pool manages a set of compute instances running as Kubernetes worker nodes with the specified shape, OS image, and placement configuration. OKE distributes nodes across the configured availability domains and fault domains. Standard OpenMCF freeform tags are applied to both the node pool and its node config for resource tracking.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the node pool will be created — literal value or reference to an OciCompartment resource
- **An OKE cluster OCID** to attach the node pool to — literal value or reference to an OciContainerEngineCluster resource
- **A compute shape name** for the worker nodes (e.g., `VM.Standard.E4.Flex`) — run `oci compute shape list` to see available shapes in your compartment
- **At least one availability domain name** (e.g., `Uocm:PHX-AD-1`) — run `oci iam availability-domain list` to see domains in your region
- **A subnet OCID** in each availability domain for node placement — literal value or reference to an OciSubnet resource
- **Pod subnets** if the cluster uses VCN-native pod networking (`oci_vcn_ip_native` CNI) — pod subnets must have enough IPs for `maxPodsPerNode * nodeCount`

## Quick Start

Create a file `node-pool.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: general-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciContainerEngineNodePool.general-pool
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

Deploy:

```shell
openmcf apply -f node-pool.yaml
```

This creates a 3-node pool using E4 Flex instances with 2 OCPUs and 32 GB of memory each, placed in a single availability domain. OKE launches the compute instances, installs the Kubernetes kubelet, and joins the nodes to the cluster. The node pool ID and Kubernetes version are exported as stack outputs. Add more entries to `placementConfigs` to distribute nodes across multiple availability domains.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the node pool will be created. Changing this after creation forces node pool recreation. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `clusterId` | `StringValueOrRef` | OCID of the OKE cluster to which this node pool is attached. Changing this after creation forces node pool recreation. Can reference an OciContainerEngineCluster resource via `valueFrom`. | Required |
| `nodeShape` | `string` | Compute shape for all nodes in this pool (e.g., `VM.Standard.E4.Flex`, `VM.Standard.A1.Flex`, `VM.GPU.A10.1`). For flex shapes, also set `nodeShapeConfig` to specify OCPUs and memory. | Minimum 1 character |
| `nodeConfigDetails` | `NodeConfigDetails` | Node placement, sizing, networking, and encryption configuration. See [nodeConfigDetails fields](#nodeconfigdetails-fields). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. |
| `kubernetesVersion` | `string` | Cluster version | Kubernetes version for the nodes. When omitted, inherits the cluster's Kubernetes version. Set explicitly to pin a specific version or perform a rolling version upgrade independently of the control plane. Example: `v1.28.2`. |
| `nodeShapeConfig` | `NodeShapeConfig` | — | CPU and memory configuration for flex shapes. Required when `nodeShape` is a flex shape (e.g., `VM.Standard.E4.Flex`). Ignored for fixed shapes. See [nodeShapeConfig fields](#nodeshapeconfig-fields). |
| `nodeSourceDetails` | `NodeSourceDetails` | OKE default | OS image and boot volume configuration. When omitted, OKE uses the default Oracle Linux image for the cluster's Kubernetes version. See [nodeSourceDetails fields](#nodesourcedetails-fields). |
| `sshPublicKey` | `string` | — | SSH public key installed on each node for debug access. The corresponding private key allows SSH to nodes via their private IP (or public IP if the subnet allows it). |
| `initialNodeLabels` | `NodeLabel[]` | — | Kubernetes labels applied to each node after it joins the cluster. Commonly used for scheduling constraints (`nodeSelector`, affinity rules). See [nodeLabel fields](#nodelabel-fields). |
| `nodeMetadata` | `map<string, string>` | — | Key/value pairs added to each underlying OCI compute instance at launch. Used for cloud-init user data and instance metadata configuration. |
| `nodeEvictionSettings` | `NodeEvictionSettings` | — | Controls graceful node eviction behavior during node pool operations (scale-down, version upgrades, shape changes). See [nodeEvictionSettings fields](#nodeevictionsettings-fields). |
| `nodePoolCyclingDetails` | `NodePoolCyclingDetails` | — | Rolling update strategy for node pool operations. Controls how many nodes can be replaced simultaneously during upgrades. See [nodePoolCyclingDetails fields](#nodepoolcyclingdetails-fields). |

### nodeShapeConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `ocpus` | `float` | Number of OCPUs allocated to each node. Example: `2.0` for a 2-OCPU flex instance. |
| `memoryInGbs` | `float` | Memory in gigabytes allocated to each node. Example: `32.0` for 32 GB of RAM. |

### nodeSourceDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `imageId` | `string` | OCID of the OCI platform image or custom image for the node OS. Use `oci ce node-pool-options get` to list available images for a given Kubernetes version. | Minimum 1 character |
| `bootVolumeSizeInGbs` | `int64` | Boot volume size in gigabytes. Minimum 50 GB. When omitted, uses the image's default boot volume size (typically 50 GB). | Optional |

### nodeConfigDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `placementConfigs` | `PlacementConfig[]` | Placement configurations determining which availability domains and subnets receive nodes. Provide one entry per AD for regional subnets, or one entry per AD-specific subnet. See [placementConfig fields](#placementconfig-fields). | Minimum 1 item |
| `size` | `int32` | Desired number of nodes in this pool. OKE distributes nodes across the placement configs as evenly as possible. | Greater than 0 |
| `nsgIds` | `StringValueOrRef[]` | OCIDs of network security groups applied to the node VNICs. Can reference OciSecurityGroup resources via `valueFrom`. | Optional |
| `kmsKeyId` | `StringValueOrRef` | OCID of the KMS key for encrypting boot volumes at rest. | Optional |
| `isPvEncryptionInTransitEnabled` | `bool` | Whether to enable in-transit encryption for the data volume's paravirtualized attachment. Applies to both boot and block volumes. | Optional |
| `podNetworkOptionDetails` | `PodNetworkOptionDetails` | Pod networking configuration. Required when the cluster uses OCI VCN-native pod networking (`oci_vcn_ip_native` CNI). See [podNetworkOptionDetails fields](#podnetworkoptiondetails-fields). | Optional |

### placementConfig Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `availabilityDomain` | `string` | Availability domain name where nodes will be launched. Example: `Uocm:PHX-AD-1`. | Minimum 1 character |
| `subnetId` | `StringValueOrRef` | OCID of the subnet in which to place nodes in this AD. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `faultDomains` | `string[]` | Fault domains within the AD to constrain node placement. When omitted, OKE distributes nodes across all fault domains. Example: `["FAULT-DOMAIN-1", "FAULT-DOMAIN-2"]`. | Optional |
| `capacityReservationId` | `StringValueOrRef` | OCID of a compute capacity reservation to use for nodes in this AD. | Optional |
| `preemptibleNodeConfig` | `PreemptibleNodeConfig` | Preemptible node configuration. When set, nodes in this placement use preemptible (spot) instances that can be reclaimed by OCI. Suitable for fault-tolerant and batch workloads. See [preemptibleNodeConfig fields](#preemptiblenodeconfig-fields). | Optional |

### preemptibleNodeConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `isPreserveBootVolume` | `bool` | Whether to preserve the boot volume when the preemptible instance is terminated. Defaults to `false` (boot volume is deleted). |

### podNetworkOptionDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `cniType` | `enum` | CNI plugin type. Must match the cluster's CNI configuration. Values: `flannel_overlay`, `oci_vcn_ip_native`. | Required (cannot be unspecified) |
| `maxPodsPerNode` | `int32` | Maximum number of pods per node. Limited by the number of VNICs attachable to the node shape. Only applicable for `oci_vcn_ip_native`. | Optional |
| `podNsgIds` | `StringValueOrRef[]` | OCIDs of NSGs applied to pod VNICs. Only applicable for `oci_vcn_ip_native`. Can reference OciSecurityGroup resources via `valueFrom`. | Optional |
| `podSubnetIds` | `StringValueOrRef[]` | OCIDs of subnets for pod IP allocation. Only applicable for `oci_vcn_ip_native`. Can be the same as or different from the node subnets. Can reference OciSubnet resources via `valueFrom`. | Optional |

### nodeLabel Fields

| Field | Type | Description |
|-------|------|-------------|
| `key` | `string` | Kubernetes label key. |
| `value` | `string` | Kubernetes label value. |

### nodeEvictionSettings Fields

| Field | Type | Description |
|-------|------|-------------|
| `evictionGraceDuration` | `string` | Maximum time OKE will attempt to evict pods before giving up. ISO 8601 duration format. Default: `PT60M`. Range: `PT0M` to `PT60M`. `PT0M` means delete the node immediately without cordon and drain. |
| `isForceActionAfterGraceDuration` | `bool` | Whether to proceed with the node action if not all pods can be evicted within the grace period. |
| `isForceDeleteAfterGraceDuration` | `bool` | Whether to delete the underlying compute instance if pods cannot be fully evicted within the grace period. |

### nodePoolCyclingDetails Fields

| Field | Type | Description |
|-------|------|-------------|
| `isNodeCyclingEnabled` | `bool` | Whether node cycling is enabled for this pool. |
| `maximumSurge` | `string` | Maximum additional nodes that can be temporarily created during cycling. Accepts an integer (`"1"`) or percentage (`"25%"`). Default: `"1"`. |
| `maximumUnavailable` | `string` | Maximum nodes that can be unavailable during cycling. Accepts an integer (`"0"`) or percentage (`"25%"`). Default: `"0"`. |

## Examples

### Minimal Node Pool

A basic node pool with a single placement config — the simplest path to running worker nodes:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: dev-pool
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

### Production Multi-AD Node Pool with VCN-Native CNI

A production node pool distributed across three availability domains with VCN-native pod networking, NSGs on nodes and pods, KMS boot volume encryption, and in-transit encryption. All infrastructure references use `valueFrom` for declarative composition:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: prod-pool
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
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

### GPU Node Pool with Preemptible Instances

A GPU-accelerated node pool for ML/AI workloads using preemptible instances for cost savings. Kubernetes labels enable workload scheduling via `nodeSelector`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: gpu-pool
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
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
        preemptibleNodeConfig:
          isPreserveBootVolume: false
      - availabilityDomain: "Uocm:PHX-AD-2"
        subnetId:
          value: "ocid1.subnet.oc1.iad.example2"
        preemptibleNodeConfig:
          isPreserveBootVolume: false
  initialNodeLabels:
    - key: "accelerator"
      value: "nvidia-a10"
    - key: "workload-type"
      value: "gpu"
  nodeMetadata:
    user_data: "IyEvYmluL2Jhc2gKZWNobyAnR1BVIG5vZGUgcmVhZHkn"
```

### ARM-Based Cost-Optimized Pool with Capacity Reservation

An Ampere A1 Flex node pool for cost-optimized workloads with a capacity reservation guarantee and fault domain constraints:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciContainerEngineNodePool
metadata:
  name: arm-pool
  org: acme
  env: prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
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

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `node_pool_id` | `string` | OCID of the OKE node pool. |
| `kubernetes_version` | `string` | Kubernetes version running on the nodes in this pool. Matches the cluster version when not explicitly overridden. |

## Related Components

- [OciContainerEngineCluster](/docs/catalog/oci/ocicontainerenginecluster) — provides the cluster referenced by `clusterId` via `valueFrom`; every node pool must be attached to exactly one cluster
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides subnets for node placement (`placementConfigs[].subnetId`) and pod IP allocation (`podNetworkOptionDetails.podSubnetIds`) via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — manages network security rules for node VNICs (`nodeConfigDetails.nsgIds`) and pod VNICs (`podNetworkOptionDetails.podNsgIds`) via `valueFrom`
