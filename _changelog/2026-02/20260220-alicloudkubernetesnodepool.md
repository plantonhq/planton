# AlicloudKubernetesNodePool Component Added

**Date**: 2026-02-20
**Component**: AlicloudKubernetesNodePool
**Enum**: 3092
**ID Prefix**: acknp

## Summary

Added the AlicloudKubernetesNodePool deployment component -- an ACK Kubernetes node pool with auto-scaling, spot instance support, managed lifecycle, and flexible disk configuration.

This component wraps a single provider resource (`alicloud_cs_kubernetes_node_pool` / `cs.NodePool`). It manages worker nodes for an AlicloudKubernetesCluster (R24).

## What Was Created

### API Definition
- `apis/org/openmcf/provider/alicloud/alicloudkubernetesnodepool/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AlicloudKubernetesNodePool = 3092` in `CloudResourceKind` enum
- 7 proto messages: spec, system_disk, data_disk, taint, scaling_config, management, spot_price_limit, plus the API/status wrappers

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider and `cs.NewNodePool` with all spec fields mapped. Converts map<string,string> labels to NodePoolLabelArray. Exports node_pool_id and scaling_group_id.
- **Terraform** (HCL): Single `alicloud_cs_kubernetes_node_pool` resource with dynamic blocks for data_disks, labels (from map), taints, scaling_config, management, spot_price_limit

### Tests
- Ginkgo/Gomega spec validation tests: 40 specs covering valid inputs (minimal, auto-scaling, spot, management, labels/taints, PrePaid billing, full production config), invalid inputs (missing required fields, invalid enum values, out-of-range values, invalid effects)

### Documentation
- README.md with component overview and directory structure
- catalog-page.md with user-facing documentation
- examples.md with 3 YAML examples (minimal fixed-size, auto-scaling, production spot)
- docs/README.md with provider research notes

## Design Decisions

- **Renamed from AlicloudAckNodePool**: T02 originally named this AlicloudAckNodePool. Renamed to AlicloudKubernetesNodePool per established naming pattern (matches DigitalOceanKubernetesNodePool, CivoKubernetesNodePool) and the parent cluster's README/spec.proto references.
- **region field added**: Every other Alibaba Cloud component has a region field on the spec for provider setup. Included for consistency even though the node pool conceptually inherits region from the cluster.
- **Labels as map<string,string>**: Provider uses repeated key/value objects, but map is cleaner for end users. IaC modules convert to the provider's format.
- **desired_size as int32**: The Terraform schema oddly uses string type for this field. Proto uses int32 which is semantically correct; the Pulumi module converts to string for the provider.
- **80/20 field coverage**: ~30 fields covering all common production use cases. Excluded kubelet_configuration (30+ sub-fields), instance_patterns, TEE, eflo, and other niche features for v2.
- **Only non-deprecated fields**: Uses node_pool_name (not name), security_group_ids (not security_group_id), image_type (not platform), desired_size (not node_count).

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (40/40 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
