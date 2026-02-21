# OCI Compute Instance Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi IaC Module, Terraform IaC Module, Provider Framework

## Summary

Added the OciComputeInstance deployment component (R07, enum 3310) to the OCI provider in OpenMCF. This is the first Phase 2 (Compute and Containers) resource and the most complex OCI component to date, with 18 spec fields, 9 nested messages, and comprehensive coverage of OCI compute instance features including flex shapes, VNIC networking, cloud-init metadata, agent configuration, preemptible instances, launch options, and platform security settings. Both Pulumi (Go) and Terraform (HCL) modules are implemented with full feature parity.

## Problem Statement / Motivation

Compute instances are the fundamental workload primitive on OCI. Without this component, OpenMCF users cannot provision virtual machines or bare metal hosts, blocking the entire Compute and Containers phase (R07-R10) and three of the five planned infra charts (OKE Environment, Compute Environment, Serverless Stack).

### Pain Points

- No way to launch OCI compute instances through OpenMCF
- OCI compute instances have 50+ provider fields across 11 nested blocks -- raw Terraform/Pulumi requires significant boilerplate
- Flex shapes (the modern default) require separate shape_config blocks with OCPUs and memory, adding complexity
- VNIC networking configuration requires careful coordination with OciSubnet and OciSecurityGroup
- Preemptible instances, platform security features, and agent configuration are enterprise-critical but rarely configured correctly from scratch

## Solution / What's New

Comprehensive deployment component wrapping `oci_core_instance` with the standard OpenMCF KRM pattern. The spec design covers the full mainstream API surface while deliberately excluding niche features (HPC clustering, PXE boot, licensing configs) that add complexity without broad value.

### Spec Fields (18 top-level)

- `compartmentId` (StringValueOrRef, required) -- compartment for the instance
- `availabilityDomain` (string, required) -- AD placement
- `shape` (string, required) -- hardware profile (e.g., "VM.Standard.E4.Flex")
- `displayName` (string, optional) -- falls back to metadata.name
- `shapeConfig` (ShapeConfig) -- OCPUs, memory, baseline utilization, NVMe drives
- `sourceDetails` (SourceDetails, required) -- image OCID or boot volume with size/VPUs/KMS
- `createVnicDetails` (CreateVnicDetails, required) -- subnet, NSGs, public IP, hostname, private IP
- `metadata` (map) -- SSH keys (ssh_authorized_keys) and cloud-init (user_data)
- `faultDomain` (string) -- HA distribution across fault domains
- `isPvEncryptionInTransitEnabled` (optional bool) -- paravirtualized encryption
- `agentConfig` (AgentConfig) -- Oracle Cloud Agent plugins with per-plugin overrides
- `availabilityConfig` (AvailabilityConfig) -- live migration and recovery action
- `launchOptions` (LaunchOptions) -- boot volume type, network type, firmware
- `instanceOptions` (InstanceOptions) -- IMDSv2 legacy endpoint toggle
- `preemptibleInstanceConfig` (PreemptibleInstanceConfig) -- spot-like pricing
- `capacityReservationId` (StringValueOrRef) -- capacity reservation binding
- `dedicatedVmHostId` (StringValueOrRef) -- physical isolation
- `platformConfig` (PlatformConfig) -- secure boot, TPM, memory encryption, NUMA, SMT

### Outputs

- `instanceId` -- OCID of the compute instance
- `privateIp` -- private IP of the primary VNIC
- `publicIp` -- public IP (empty if not assigned)
- `bootVolumeId` -- OCID of the boot volume
- `availabilityDomain` -- AD where the instance was placed

### Infra-Chart Composability

- `createVnicDetails.subnetId` references OciSubnet via StringValueOrRef
- `createVnicDetails.nsgIds` references OciSecurityGroup (max 5) via repeated StringValueOrRef
- `compartmentId` references OciCompartment via StringValueOrRef
- `sourceDetails.kmsKeyId` uses StringValueOrRef (default_kind will be added when OciKmsKey R25 is implemented)

## Implementation Details

### Files Created

**Proto API** (`apis/org/openmcf/provider/oci/ocicomputeinstance/v1/`):
- `spec.proto` -- 18 fields, 9 embedded messages, 6 enums, buf-validate rules
- `api.proto` -- KRM wiring with api_version/kind const validation
- `stack_input.proto` -- IaC module input (target + provider config)
- `stack_outputs.proto` -- 5 deployment outputs
- `spec_test.go` -- 32 Ginkgo/Gomega validation tests (17 valid, 15 invalid scenarios)

**Pulumi Module** (`iac/pulumi/`):
- `module/main.go` -- Entry point with provider setup
- `module/locals.go` -- Display name fallback, freeform tag assembly
- `module/instance.go` -- Core instance creation with 9 builder functions for nested configs
- `module/outputs.go` -- Output constant definitions

**Terraform Module** (`iac/tf/`):
- `provider.tf` -- OCI provider >= 5.0
- `variables.tf` -- Full spec type definition with all nested optional objects
- `locals.tf` -- Display name, freeform tags, enum mapping tables (source type, firmware, recovery action, platform type)
- `main.tf` -- oci_core_instance resource with 8 dynamic blocks
- `outputs.tf` -- 5 outputs matching Pulumi

**Kind Registration**:
- Added `OciComputeInstance = 3310` to cloud_resource_kind.proto under new `// --- Compute ---` section
- Regenerated kind_map_gen.go

### Design Decisions

**LaunchOptions uses strings for boot_volume_type and network_type** -- Both BootVolumeType and NetworkType enums contain "VFIO" and "PARAVIRTUALIZED" values. Protobuf's C++ scoping rules prevent duplicate enum values within the same enclosing message. Using strings for these rarely-set fields avoids the collision while maintaining full API coverage.

**assignPublicIp is optional bool (not string)** -- The OCI API accepts "true"/"false" strings for this field, but the proto uses `optional bool` for better user experience. The tri-state semantics (unset = subnet default, true = assign, false = don't assign) map naturally to proto3 optional. The Pulumi module converts to string and the TF module uses `tostring()`.

**PlatformConfig includes full field set** -- All 11 platform config fields are included even though most are bare-metal-only. The OCI API validates which fields apply to which platform type, so the proto doesn't need to encode this complex validation matrix. Fields are documented with their applicability (VM vs BM).

### Validation Results

- `go build` -- clean
- `go vet` -- clean
- 32/32 spec tests passed
- `terraform validate` -- success

## Benefits

- Full mainstream compute instance feature coverage in a single KRM manifest
- Infra-chart composable via StringValueOrRef for compartment, subnet, NSGs, and KMS key
- 32 validation tests ensure spec correctness before deployment
- Enterprise features (preemptible, platform security, capacity reservation, dedicated hosts) available without Terraform boilerplate

## Impact

- Unblocks Phase 2 of OCI provider expansion (R08-R10: OKE Cluster, Node Pool, Container Instance)
- Required for 3 of 5 planned infra charts (OKE Environment, Compute Environment, Serverless Stack)
- Most complex OCI component to date -- establishes patterns for handling deeply nested provider APIs

## Related Work

- R01-R06 (Phase 1: Foundation) -- networking and identity components this instance references
- R08 OciContainerEngineCluster (next in queue) -- depends on same VCN/Subnet pattern
- R23 OciBlockVolume (Phase 5) -- will handle volume attachment separately from launch-time config

---

**Status**: Production Ready
