# OCI Container Engine Node Pool (OKE) Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi IaC Module, Terraform IaC Module, Provider Framework

## Summary

Added the OciContainerEngineNodePool deployment component (R09, enum 3312) to the OCI provider in OpenMCF. This wraps the OKE managed node pool (`oci_containerengine_node_pool`) -- the worker node layer of Kubernetes clusters. The spec covers compute shape and flex shape configuration, node placement across availability domains and fault domains, VCN-native pod networking, preemptible (spot) instances, node eviction controls, rolling update strategy, and boot volume encryption. Both Pulumi (Go) and Terraform (HCL) modules are implemented with full feature parity.

## Problem Statement / Motivation

With OciContainerEngineCluster (R08) providing the Kubernetes control plane, there is no way to provision worker nodes through OpenMCF. Node pools are where the actual compute resources live -- without them, an OKE cluster has no capacity to run workloads. This component is the second half of the OKE story and a prerequisite for the OKE Environment infra chart.

### Pain Points

- No way to provision OKE worker nodes through OpenMCF
- The Terraform node pool resource has overlapping deprecated and modern config paths (`subnet_ids` vs `node_config_details`, `node_image_name` vs `node_source_details`) creating confusion
- `source_type` in `node_source_details` only accepts "IMAGE" but is required, adding noise
- Preemptible node configuration requires a nested `preemption_action.type = "TERMINATE"` (the only valid value), which is unnecessary boilerplate
- Pod networking options require understanding OCI-specific CNI type naming conventions

## Solution / What's New

Deployment component wrapping `oci_containerengine_node_pool` with the standard OpenMCF KRM pattern. The spec design omits all deprecated paths and simplifies single-value constants:

- **source_type hardcoded**: Only "IMAGE" is valid, so IaC modules hardcode it -- users provide just `imageId` and optional `bootVolumeSizeInGbs`
- **Preemptible config simplified**: The `preemption_action.type` is always "TERMINATE", so the spec exposes only `isPreserveBootVolume` -- the IaC modules inject the TERMINATE action
- **node_config_details required**: Since deprecated `subnet_ids`/`quantity_per_subnet` are omitted, `nodeConfigDetails` is the only path and is required
- **kubernetes_version optional**: When omitted, inherits from the parent cluster (standard OKE behavior)

### Spec Fields (13 top-level)

- `compartmentId` (StringValueOrRef, required) -- compartment for the node pool
- `clusterId` (StringValueOrRef, required) -- parent OKE cluster
- `name` (string, optional) -- falls back to metadata.name
- `kubernetesVersion` (string, optional) -- inherits from cluster when omitted
- `nodeShape` (string, required) -- compute shape (e.g., "VM.Standard.E4.Flex")
- `nodeShapeConfig` (NodeShapeConfig) -- OCPUs and memory for flex shapes
- `nodeSourceDetails` (NodeSourceDetails) -- OS image and boot volume size
- `nodeConfigDetails` (NodeConfigDetails, required) -- placement, size, NSGs, pod networking, encryption
- `sshPublicKey` (string) -- debug SSH access to nodes
- `initialNodeLabels` (repeated NodeLabel) -- Kubernetes labels for scheduling
- `nodeMetadata` (map) -- cloud-init key/value pairs
- `nodeEvictionSettings` (NodeEvictionSettings) -- graceful drain controls
- `nodePoolCyclingDetails` (NodePoolCyclingDetails) -- rolling update strategy

### Nested Messages (8)

- **NodeShapeConfig** -- ocpus, memoryInGbs (float for flex shape sizing)
- **NodeSourceDetails** -- imageId, bootVolumeSizeInGbs (source_type hardcoded to IMAGE)
- **NodeConfigDetails** -- placementConfigs, size, nsgIds, kmsKeyId, isPvEncryptionInTransitEnabled, podNetworkOptionDetails
- **PlacementConfig** -- availabilityDomain, subnetId, faultDomains, capacityReservationId, preemptibleNodeConfig
- **PreemptibleNodeConfig** -- isPreserveBootVolume (preemption action hardcoded to TERMINATE)
- **PodNetworkOptionDetails** -- cniType, maxPodsPerNode, podNsgIds, podSubnetIds
- **NodeLabel** -- key, value
- **NodeEvictionSettings** -- evictionGraceDuration (ISO 8601), isForceActionAfterGraceDuration, isForceDeleteAfterGraceDuration
- **NodePoolCyclingDetails** -- isNodeCyclingEnabled, maximumSurge, maximumUnavailable

### Outputs (2)

- `nodePoolId` -- OCID of the node pool
- `kubernetesVersion` -- version running on the nodes

### Infra-Chart Composability

- `compartmentId` references OciCompartment via StringValueOrRef
- `clusterId` references OciContainerEngineCluster via StringValueOrRef
- `nodeConfigDetails.placementConfigs[].subnetId` references OciSubnet
- `nodeConfigDetails.nsgIds` references OciSecurityGroup
- `nodeConfigDetails.podNetworkOptionDetails.podNsgIds` references OciSecurityGroup
- `nodeConfigDetails.podNetworkOptionDetails.podSubnetIds` references OciSubnet

## Implementation Details

### Files Created

**Proto API** (`apis/org/openmcf/provider/oci/ocicontainerenginenodepool/v1/`):
- `spec.proto` -- 13 top-level fields, 8 embedded messages, 1 enum, buf-validate rules
- `api.proto` -- KRM wiring with api_version/kind const validation
- `stack_input.proto` -- IaC module input (target + provider config)
- `stack_outputs.proto` -- 2 deployment outputs
- `spec_test.go` -- 36 Ginkgo/Gomega validation tests (22 valid, 14 invalid scenarios)

**Pulumi Module** (`iac/pulumi/`):
- `module/main.go` -- Entry point with provider setup
- `module/locals.go` -- Display name fallback, freeform tag assembly
- `module/node_pool.go` -- Node pool creation with 8 builder functions
- `module/outputs.go` -- Output constant definitions
- `main.go` -- Pulumi entrypoint

**Terraform Module** (`iac/tf/`):
- `provider.tf` -- OCI provider >= 5.0
- `variables.tf` -- Full spec type definition with all nested optional objects
- `locals.tf` -- Display name, freeform tags, CNI type map, NSG/subnet ID extraction
- `main.tf` -- oci_containerengine_node_pool with dynamic blocks for shape config, source details, config details (with nested placement and pod network), eviction settings, cycling details, and initial labels
- `outputs.tf` -- 2 outputs matching Pulumi

**Kind Registration**:
- Added `OciContainerEngineNodePool = 3312` to cloud_resource_kind.proto under `// --- Containers ---` section
- Regenerated kind_map_gen.go

### Design Decisions

**kubernetes_version made optional** -- The R09 plan stub listed this as required, but both providers treat it as optional (inherits from cluster). Making it optional is better UX: users pin versions at the cluster level and only override at the node pool level for deliberate rolling upgrades.

**source_type hardcoded to IMAGE** -- The TF provider's ValidateFunc only accepts "IMAGE". Exposing a single-value field adds noise and confusion. The spec contains only `imageId` and `bootVolumeSizeInGbs`; IaC modules inject `source_type = "IMAGE"`.

**Preemptible config simplified** -- The provider requires `preemption_action.type = "TERMINATE"` (the only valid value). Rather than exposing a single-value enum, the spec represents preemptible config as a presence message with optional `isPreserveBootVolume`. The IaC modules hardcode the TERMINATE action.

**node_config_details required** -- Since deprecated alternatives (`subnet_ids`, `quantity_per_subnet`) are omitted, this is the only configuration path and should be required with `min_items = 1` on placement_configs and `gt = 0` on size.

**float for shape config** -- OCPUs and memory use proto `float` (Go float32). Values like 2.0 and 64.0 are perfectly representable. The Pulumi module converts `float32` to `float64` (lossless for these value ranges).

### Validation Results

- `go build` -- clean
- `go vet` -- clean
- 36/36 spec tests passed
- `terraform validate` -- success
- kind_map_gen.go regenerated and compiles clean

## Benefits

- Full OKE worker node provisioning in a single KRM manifest
- Cleaner YAML UX than raw Terraform (no deprecated fields, no boilerplate constants)
- Flex shape support with explicit OCPU and memory configuration
- Preemptible (spot) instances for cost-optimized and batch workloads
- VCN-native pod networking with pod NSGs and pod subnets
- Rolling update controls (surge/unavailable) for safe node pool upgrades
- 36 validation tests ensure spec correctness before deployment

## Impact

- Completes the OKE cluster + node pool pair (R08 + R09)
- Enables the OKE Environment infra chart (the highest-priority OCI infra chart)
- Last prerequisite before R10 OciContainerInstance (completes Phase 2: Compute and Containers)

## Related Work

- R08 OciContainerEngineCluster -- parent cluster that this node pool attaches to
- R01 OciVcn, R02 OciSubnet, R03 OciSecurityGroup -- networking components referenced by placement configs and pod networking
- R04 OciCompartment -- compartment for organizational isolation
- DD03: OKE Cluster/NodePool Split -- design decision for separate components
- R10 OciContainerInstance (next in queue) -- serverless container offering

---

**Status**: Production Ready
