# AliCloudKubernetesCluster Component Added

**Date**: 2026-02-20
**Component**: AliCloudKubernetesCluster
**Enum**: 3091
**ID Prefix**: acack

## Summary

Added the AliCloudKubernetesCluster deployment component -- an ACK Managed Kubernetes cluster with dual CNI support, RRSA, control plane logging, and maintenance window configuration.

This component wraps a single provider resource (`alicloud_cs_managed_kubernetes` / `cs.ManagedKubernetes`). Worker nodes are managed separately through AliCloudKubernetesNodePool (R25).

## What Was Created

### API Definition
- `apis/dev/planton/provider/alicloud/alicloudkubernetescluster/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudKubernetesCluster = 3091` in `CloudResourceKind` enum under the Containers category
- 6 proto messages: spec, addon, logging, maintenance window, auto-upgrade, plus the API/status wrappers

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and a single `cs.ManagedKubernetes` resource with all spec fields mapped, plus computed output exports (connections, RRSA metadata)
- **Terraform** (HCL): Single `alicloud_cs_managed_kubernetes` resource with dynamic blocks for addons, maintenance_window, audit_log_config, and operation_policy

### Tests
- Ginkgo/Gomega spec validation tests: 30 specs covering valid inputs (minimal, Flannel, Terway, security config, addons, logging, maintenance, auto-upgrade, full production config), invalid inputs (wrong api_version/kind, missing metadata/spec, empty required fields, out-of-range values, invalid enum values)

### Documentation
- README.md with component overview and directory structure
- examples.md with 3 YAML examples (minimal Flannel, Terway with RRSA, full production)

## Design Decisions

- **Renamed from AliCloudAckManagedCluster**: T02 originally named this AliCloudAckManagedCluster. Renamed to AliCloudKubernetesCluster per user direction, aligning with DigitalOceanKubernetesCluster and CivoKubernetesCluster naming patterns.
- **Dual CNI support**: Both Flannel (`pod_cidr`) and Terway (`pod_vswitch_ids`) are supported. The user selects the CNI via the addons list.
- **No kubeconfig output**: Certificate fields are deprecated since provider v1.248.0. Outputs API server endpoints (internet + intranet) and RRSA metadata instead.
- **service_cidr is required**: No safe default value due to CIDR overlap concerns with VPC and pod networks.
- **enable_rrsa defaults to false**: One-way operation (cannot be disabled). Follows provider default; documentation recommends enabling it.
- **new_nat_gateway defaults to true**: Matches provider default. Users managing their own NAT gateway should set this to false.
- **Logging as sub-message**: Groups control_plane_log_* and audit_log_config into a single `logging` sub-message for cleaner UX.
- **Maintenance + auto-upgrade**: Included for production readiness. Auto-upgrade requires a maintenance window.

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (30/30 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
