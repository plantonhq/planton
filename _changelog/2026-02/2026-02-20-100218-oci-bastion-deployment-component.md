# OCI Bastion Deployment Component

**Date**: 2026-02-20
**Type**: New Feature
**Components**: `apis/org/openmcf/provider/oci/ocibastion/v1/`

## Summary

Added the OciBastion deployment component -- OCI's managed SSH gateway that provides secure, time-limited access to resources in private subnets without requiring public IP addresses on target resources. Supports client CIDR allow lists, configurable session TTL, and DNS/SOCKS5 proxy for FQDN-based connections. Fourth and final resource of Phase 6 (Security and Secrets), completing the phase. Resource R27 in the OCI provider expansion.

## Problem Statement / Motivation

OpenMCF's OCI provider had encryption infrastructure (KMS Vault, KMS Key, Vault Secret) for secrets management, but no way to declaratively provision secure access to private resources. OCI Bastion is the managed alternative to self-hosted jump boxes -- it provides auditable, time-limited SSH access to compute instances and other resources in private subnets. Without it, platform teams cannot declaratively manage secure connectivity for debugging, administration, and port forwarding into private infrastructure.

## Solution / What's New

A complete OciBastion deployment component with both Pulumi (Go) and Terraform (HCL) modules.

### Proto API

- **spec.proto**: 6 fields -- compartment_id (StringValueOrRef), target_subnet_id (StringValueOrRef), display_name, client_cidr_block_allow_list (repeated), max_session_ttl_in_seconds (optional int32), is_dns_proxy_enabled (optional bool)
- **buf.validate**: compartment_id required, target_subnet_id required
- **api.proto**: Standard KRM wrapper with const-validated api_version and kind
- **stack_outputs.proto**: 2 outputs (bastion_id, private_endpoint_ip_address)

### Design Decisions

- **bastion_type hardcoded to STANDARD**: Only publicly documented type. Hardcoded in both IaC modules; not exposed in spec. Same pattern as source_type in OciContainerEngineNodePool and content_type in OciVaultSecret.
- **Sessions NOT bundled**: `oci_bastion_session` is an ephemeral operational artifact (short-lived SSH tunnels with TTLs), not infrastructure. Same reasoning as excluding `oci_kms_key_version` from OciKmsKey.
- **is_dns_proxy_enabled as optional bool**: Provider uses "ENABLED"/"DISABLED" strings, but `isDnsProxyEnabled: true` is cleaner YAML UX. Tri-state: nil = omit (provider default), true = "ENABLED", false = "DISABLED".
- **Directory name**: `ocibastion` (per WA02 -- lowercased kind name, not id_prefix `ocibst`).

### Bundled Resources

1. **Bastion** -- the managed SSH gateway with client CIDR controls, session TTL, and DNS proxy support

### Pulumi Module (Go)

5 files across the module package:
- `main.go` -- orchestrator calling bastionResource() with OCI provider setup
- `locals.go` -- Locals struct with freeform tags and display name fallback
- `bastion.go` -- bastion.NewBastion() with conditional client CIDR, max TTL, and DNS proxy status conversion
- `outputs.go` -- 2 output constants (bastion_id, private_endpoint_ip_address)

### Terraform Module (HCL)

5 files:
- `main.tf` -- oci_bastion_bastion.this with client CIDR list, conditional max TTL, dns_proxy_status from locals
- `locals.tf` -- display_name fallback, freeform tags, bool-to-string dns_proxy_status conversion
- `variables.tf`, `outputs.tf`, `provider.tf`

### Validation Tests

16 Ginkgo/Gomega tests (10 valid, 6 invalid scenarios) covering:
- Minimal bastion (compartment_id + target_subnet_id only)
- With display_name, CIDR allow list, max session TTL, DNS proxy enabled/disabled
- StringValueOrRef with literal and valueFrom patterns for both compartment_id and target_subnet_id
- Full configuration with all fields populated
- Empty CIDR list (valid)
- Required field validation (compartment_id, target_subnet_id, metadata, spec)
- Invalid api_version and kind

### Kind Registration

`OciBastion = 3353` under "Security and Secrets" section in CloudResourceKind enum.

### Excluded

- `phone_book_entry` -- not applicable to STANDARD bastions
- `static_jump_host_ip_addresses` -- not applicable to STANDARD bastions
- `security_attributes` -- Oracle ZPR, very low adoption (consistent with OciRedisCluster)
- `defined_tags`, `system_tags` -- managed by platform
- `freeform_tags` -- auto-populated from metadata labels

## Impact

- Adds 1 new CloudResourceKind to the OCI provider (R27 of 37)
- Completes Phase 6 (Security and Secrets) -- all 4 resources done (OciKmsVault, OciKmsKey, OciVaultSecret, OciBastion)
- Enables declarative provisioning of secure access to private infrastructure
- Next: R28 OciFunctionsApplication (enum 3360, Phase 7: Serverless and Functions)
