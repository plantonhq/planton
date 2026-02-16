# GCP Vertex AI Notebook Deployment Component

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added the GcpVertexAiNotebook deployment component (R19) to OpenMCF, enabling declarative provisioning of Vertex AI Workbench instances with GPU accelerators, CMEK encryption, private VPC networking, and pre-built or custom container images. This is the 20th GCP resource kind in the expansion project, entering the AI/ML category.

## Problem Statement / Motivation

Data scientists and ML engineers need managed JupyterLab notebook environments for model development and experimentation. Vertex AI Workbench (successor to the deprecated AI Platform Notebooks) provides these as Compute Engine VMs with pre-installed ML frameworks, but provisioning them manually through the console doesn't support infrastructure-as-code workflows, team standardization, or composition with other GCP resources.

### Pain Points

- No declarative YAML-based way to provision Workbench instances in OpenMCF
- GPU configuration, disk encryption, and VPC networking require deep GCP knowledge
- No integration with OpenMCF's foreign key system for composing notebooks with VPCs, service accounts, and KMS keys

## Solution / What's New

### GcpVertexAiNotebook Component

A complete deployment component following the Terraform `google_workbench_instance` / Pulumi `workbench.Instance` resource model (v2 API). The component targets the current Workbench v2 API, not the deprecated Notebooks v1 API.

### Key Design Decisions

1. **Flattened gce_setup** -- The Terraform/Pulumi providers nest all VM configuration under a `gce_setup` wrapper block. We flatten these fields to the spec top level, matching the GcpComputeInstance pattern. The component IS a workbench instance -- the wrapper adds no semantic value.

2. **Singular sub-messages** -- Providers use repeated fields with MaxItems:1 for accelerators, data disks, network interfaces, and service accounts. We use singular messages for clarity (e.g., `accelerator_config` not `accelerator_configs`).

3. **Int32 for disk sizes and core count** -- Providers use strings for these numeric values. We use `int32` for proto-level range validation (10-64000 for disks) and convert to strings in the IaC modules.

4. **Derived disk_encryption** -- Instead of exposing a redundant `disk_encryption` field, we derive CMEK/GMEK from the presence of `kms_key`.

5. **Service account as flat StringValueOrRef** -- Since scopes are always `cloud-platform` (computed, not configurable), we simplified from a sub-message to a direct StringValueOrRef for infra-chart composability.

## Implementation Details

### Proto API (4 proto files, 8 message types)

- `spec.proto` with 19 spec fields across 8 sub-messages
- 6 `StringValueOrRef` fields: project_id (GcpProject), network (GcpVpc), subnet (GcpSubnetwork), service_account (GcpServiceAccount), boot_disk.kms_key (GcpKmsKey), data_disk.kms_key (GcpKmsKey)
- 7 CEL validations: vm_image/container_image mutual exclusion, disk types, accelerator types, nic_type, desired_state, instance_name pattern, zone pattern
- `stack_outputs.proto` with 6 outputs: instance_id, instance_name, proxy_uri, state, creator, create_time

### Pulumi Module (4 Go files)

- `workbench.NewInstance()` with gce_setup block reconstructed from flattened spec fields
- `fmt.Sprintf` conversions for int32 disk sizes and core counts to SDK's string types
- Framework GCP labels applied to the instance
- Single-element arrays for accelerator, network interface, and service account (SDK pattern)

### Terraform Module (6 files)

- Provider `~> 6.0` for Workbench v2 API support
- Dynamic blocks for boot_disk, data_disks, accelerator_configs, network_interfaces, service_accounts, vm_image, container_image, shielded_instance_config
- `tostring()` conversions for disk sizes
- Feature parity with Pulumi implementation

### Validation Tests (45 tests, all passing)

- 25 positive cases: minimal spec, all disk types (8), all accelerator types (10), networking, CMEK, shielded VM, container image, desired_state, full-featured spec
- 20 negative cases: missing required fields, invalid zone format, invalid names, invalid enums, mutual exclusion violations, boundary values

### Documentation

- README.md, examples.md (7 examples), docs/README.md (research, 200+ lines)
- Catalog page following the standard structure (passed audit with zero Critical issues)
- Pulumi overview.md, Pulumi README.md, Terraform README.md

### Presets (3)

- basic-notebook: e2-standard-4, PD_SSD, CPU-only with common-cpu-notebooks image
- gpu-ml-notebook: n1-standard-8 + NVIDIA_TESLA_T4, PD_SSD disks, tf-latest-gpu image
- secure-private-notebook: private VPC, CMEK encryption, Shielded VM, no public IP

## Benefits

- Data scientists can provision managed notebooks through declarative YAML
- Foreign key references enable composition with GcpVpc, GcpServiceAccount, GcpKmsKey in infra charts
- Pre-built presets cover the three most common notebook deployment patterns
- Dual IaC support (Pulumi + Terraform) with full feature parity

## Impact

- Adds the first AI/ML resource kind to OpenMCF's GCP provider
- Enables the planned `gcp-ml-notebook-environment` infra chart (BigQuery + Notebook + GCS + SA + VPC)
- Total GCP resource kinds: 39 (20 new + 19 existing)

## Related Work

- Part of the 20260215.01.sp.gcp-resource-expansion sub-project
- R19 of 23 resources (R01-R18 previously completed)
- Next: R20 GcpVertexAiEndpoint, R21 GcpCloudArmorPolicy

---

**Status**: Production Ready
