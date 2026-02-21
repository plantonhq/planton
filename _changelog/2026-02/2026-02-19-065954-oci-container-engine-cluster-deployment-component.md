# OCI Container Engine Cluster (OKE) Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi IaC Module, Terraform IaC Module, Provider Framework

## Summary

Added the OciContainerEngineCluster deployment component (R08, enum 3311) to the OCI provider in OpenMCF. This is the first Phase 2 container resource and wraps the OKE managed Kubernetes control plane (`oci_containerengine_cluster`). The spec covers cluster type selection (basic/enhanced), CNI configuration (VCN-native or flannel), API server endpoint networking, Kubernetes network CIDRs, service load balancer subnet placement, OIDC authentication, image verification policies, and KMS secret encryption. Both Pulumi (Go) and Terraform (HCL) modules are implemented with full feature parity.

## Problem Statement / Motivation

OKE is OCI's managed Kubernetes service and the highest-demand OCI service after networking. Without this component, OpenMCF users cannot provision Kubernetes clusters on OCI, blocking the OKE Environment infra chart and the entire container workload story.

### Pain Points

- No way to provision OKE clusters through OpenMCF
- The OKE Terraform resource has deeply nested options blocks (endpoint config, OIDC auth, service LB config, image policy) requiring significant boilerplate
- CNI selection is wrapped in a list structure in the provider despite being a single-value choice
- Deprecated Kubernetes features (Dashboard add-on, Tiller, PSP) are still present in the provider API, creating confusion for new users

## Solution / What's New

Deployment component wrapping `oci_containerengine_cluster` with the standard OpenMCF KRM pattern. The spec design intentionally omits deprecated Kubernetes features (add_ons, admission_controller_options) and flattens the CNI type from a list to a single enum field for cleaner UX.

### Spec Fields (10 top-level)

- `compartmentId` (StringValueOrRef, required) -- compartment for the cluster
- `vcnId` (StringValueOrRef, required) -- VCN hosting the cluster
- `name` (string, optional) -- falls back to metadata.name
- `kubernetesVersion` (string, required) -- e.g. "v1.28.2"
- `type` (ClusterType enum) -- basic_cluster or enhanced_cluster
- `cniType` (CniType enum) -- flannel_overlay or oci_vcn_ip_native
- `endpointConfig` (EndpointConfig) -- API server subnet, public/private, NSGs
- `options` (ClusterOptions) -- network CIDRs, LB subnets, OIDC, PV/LB tags
- `kmsKeyId` (StringValueOrRef) -- Kubernetes secret encryption at rest
- `imagePolicyConfig` (ImagePolicyConfig) -- container image signature verification

### Nested Messages (7)

- **EndpointConfig** -- subnetId, isPublicIpEnabled, nsgIds
- **ClusterOptions** -- kubernetesNetworkConfig, serviceLbSubnetIds, ipFamilies, serviceLbConfig, persistentVolumeConfig, OIDC config, OIDC discovery toggle
- **KubernetesNetworkConfig** -- podsCidr, servicesCidr
- **ServiceLbConfig** -- backendNsgIds, freeformTags, definedTags
- **PersistentVolumeConfig** -- freeformTags, definedTags
- **OpenIdConnectTokenAuthenticationConfig** -- 11 fields covering inline config and configuration file modes
- **ImagePolicyConfig** -- isPolicyEnabled, keyDetails

### Outputs (5)

- `clusterId` -- OCID of the OKE cluster
- `kubernetesVersion` -- version running on the control plane
- `kubernetesEndpoint` -- API server endpoint URL
- `privateEndpoint` -- private native networking endpoint
- `publicEndpoint` -- public native networking endpoint (empty when private-only)

### Infra-Chart Composability

- `compartmentId` references OciCompartment via StringValueOrRef
- `vcnId` references OciVcn via StringValueOrRef
- `endpointConfig.subnetId` references OciSubnet
- `endpointConfig.nsgIds` references OciSecurityGroup
- `options.serviceLbSubnetIds` references OciSubnet
- `options.serviceLbConfig.backendNsgIds` references OciSecurityGroup
- `clusterId` output will be consumed by OciContainerEngineNodePool (R09)

## Implementation Details

### Files Created

**Proto API** (`apis/org/openmcf/provider/oci/ocicontainerenginecluster/v1/`):
- `spec.proto` -- 10 top-level fields, 7 embedded messages, 3 enums, buf-validate rules
- `api.proto` -- KRM wiring with api_version/kind const validation
- `stack_input.proto` -- IaC module input (target + provider config)
- `stack_outputs.proto` -- 5 deployment outputs
- `spec_test.go` -- 26 Ginkgo/Gomega validation tests (18 valid, 8 invalid scenarios)

**Pulumi Module** (`iac/pulumi/`):
- `module/main.go` -- Entry point with provider setup
- `module/locals.go` -- Display name fallback, freeform tag assembly
- `module/cluster.go` -- Cluster creation with 7 builder functions + ipFamilyString helper
- `module/outputs.go` -- Output constant definitions
- `main.go` -- Pulumi entrypoint

**Terraform Module** (`iac/tf/`):
- `provider.tf` -- OCI provider >= 5.0
- `variables.tf` -- Full spec type definition with all nested optional objects
- `locals.tf` -- Display name, freeform tags, enum mapping tables (cluster type, CNI type, IP family), NSG/subnet ID extraction
- `main.tf` -- oci_containerengine_cluster with dynamic blocks for endpoint_config, options, cluster_pod_network_options, image_policy_config
- `outputs.tf` -- 5 outputs matching Pulumi

**Kind Registration**:
- Added `OciContainerEngineCluster = 3311` to cloud_resource_kind.proto under new `// --- Containers ---` section
- Regenerated kind_map_gen.go

### Design Decisions

**Deprecated features omitted** -- The provider's `add_ons` (Kubernetes Dashboard, Tiller) and `admission_controller_options` (Pod Security Policy) are intentionally skipped. These map to removed Kubernetes features (PSP removed in K8s 1.25, Tiller deprecated since Helm 3). Including them in a new component would create day-one technical debt and steer users toward broken paths.

**CNI type flattened to single enum** -- The provider accepts `cluster_pod_network_options` as a list, but in practice exactly one CNI is selected per cluster. The spec uses a single `cni_type` enum field for a cleaner YAML experience. The IaC modules wrap it back into the list structure the provider expects.

**IP families use custom string conversion** -- The IpFamily enum values are `ipv4`/`ipv6` (lowercase for YAML UX), but the OCI API expects `IPv4`/`IPv6` (mixed case). A dedicated `ipFamilyString()` helper handles this conversion since `strings.ToUpper()` would produce `IPV4` (wrong).

**Endpoints extracted via ApplyT** -- The Pulumi SDK returns endpoints as a `ClusterEndpointArrayOutput` (array of structs). The module uses `ApplyT` with a function that safely indexes `endpoints[0]` and nil-checks each endpoint field, exporting them as individual string outputs.

### Validation Results

- `go build` -- clean
- `go vet` -- clean
- 26/26 spec tests passed
- `terraform validate` -- success
- kind_map_gen.go regenerated and compiles clean

## Benefits

- Full OKE cluster provisioning in a single KRM manifest
- Cleaner YAML UX than raw Terraform (flattened CNI, no deprecated fields)
- Infra-chart composable via StringValueOrRef for compartment, VCN, subnets, and NSGs
- 26 validation tests ensure spec correctness before deployment
- Enterprise features (OIDC, image verification, KMS encryption, dual-stack) available without boilerplate

## Impact

- Unblocks R09 OciContainerEngineNodePool (depends on cluster OCID output)
- Required for the OKE Environment infra chart (the highest-priority OCI infra chart)
- Establishes the container engine pattern for R09 and R10

## Related Work

- R01-R06 (Phase 1: Foundation) -- networking and identity components referenced by the cluster
- R07 OciComputeInstance -- first Phase 2 resource, established complex builder function pattern
- R09 OciContainerEngineNodePool (next in queue) -- will consume clusterId output
- DD03: OKE Cluster/NodePool Split -- design decision for separate components

---

**Status**: Production Ready
