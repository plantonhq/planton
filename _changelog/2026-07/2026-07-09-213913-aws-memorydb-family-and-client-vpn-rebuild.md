# AWS MemoryDB Family and Client VPN: RBAC Decomposition and the Full Endpoint Surface

**Date**: July 9, 2026
**Type**: Feature
**Components**: API Definitions, Terraform Modules, Pulumi Modules, E2E Harness

## Summary

MemoryDB's authentication model becomes representable in the resource graph: **`AwsMemorydbUser`** (enum 373) and **`AwsMemorydbAcl`** (enum 374) are forged as first-class kinds, and **`AwsMemorydbCluster`** is rebuilt so its `acl_name` is a required reference into that chain instead of a bare string defaulting to `open-access`. **`AwsClientVpn`** is rebuilt (breaking) to the full provider surface: typed authentication options (certificate, directory, federated SAML — up to two arms), the embedded shadow security group retired, session/banner/route-enforcement/self-service-portal/dual-stack surfaces modeled, a transit-gateway attachment arm, and the folded authorization-rule and route satellites restructured as typed blocks. All four kinds ship first-ever E2E artifacts; the offline gate is fully green on both engines.

## What Was Built

### The MemoryDB RBAC pair — `AwsMemorydbUser` (373) + `AwsMemorydbAcl` (374)

- **`AwsMemorydbUser`**: the user name derives from `metadata.name` (MemoryDB's user name IS the resource identity — create-time immutable), a Redis `access_string`, and a typed `authentication_mode` — `password` (1–2 sensitive passwords, 16–128 chars, dual passwords being the zero-downtime rotation shape) or `iam` (no credential material). Modeled from `internal/service/memorydb/` directly, not copied from the ElastiCache sibling: MemoryDB has no ElastiCache-style `user_id`/`user_name` split, no `no_password_required` mode, and no engine dial on the user.
- **`AwsMemorydbAcl`**: the ACL name derives from `metadata.name`; membership is `user_names` as a `repeated StringValueOrRef` → `AwsMemorydbUser.status.outputs.user_name` (membership folds onto the ACL — the ElastiCache user-group-association precedent; never a glue kind).
- The FK chain composes end to end: cluster → ACL → users, all by reference with `default_kind` wiring.

### `AwsMemorydbCluster` — rebuilt to the full provider surface

- **`acl_name` is now a required `StringValueOrRef`** → `AwsMemorydbAcl.status.outputs.acl_name` (was an optional bare string silently defaulting to unauthenticated `open-access`).
- **New surface**: `network_type` + `ip_discovery` (IPv6/dual-stack), `multi_region_cluster_name` (the join field for AWS's multi-region clusters — the global-cluster join precedent), and external `subnet_group_name` / `parameter_group_name` arms XOR the folded `subnet_ids`/`parameters` lists (the settled data-family shape, CEL-enforced).
- Identity converged on `metadata.name` in both engines; `snapshot_retention_limit` always sent (explicit 0 disables snapshots identically on both engines); the tls-disabled↔open-access coupling documented in spec comments (AWS enforces it at create; a message CEL cannot dereference a `StringValueOrRef` sub-field).
- Terraform floor lifted to `>= 6.34.0` (`ip_discovery`/`network_type` land there).

### `AwsClientVpn` — rebuilt (breaking) to the complete endpoint surface

- **Typed `authentication_options`** (up to two arms, all ForceNew): `certificate-authentication` (client root-CA chain ref), `directory-service-authentication` (Active Directory ID), `federated-authentication` (SAML provider ARN + optional self-service portal SAML ARN). The invented single-arm `authentication_type` enum is retired.
- **Embedded shadow security group RETIRED** — the module created an SG from `allowed_cidr_blocks`; the catalog-wide zero-embedded-SG posture now genuinely holds. `security_group_ids` is the attach shape.
- **Full endpoint surface**: `split_tunnel`, `session_timeout_hours` + `disconnect_on_session_timeout`, `self_service_portal_enabled`, `client_connect_options` (Lambda authorizer), `client_login_banner`, `client_route_enforcement_enabled`, `connection_log` (CloudWatch group/stream refs), and `endpoint_ip_address_type` / `traffic_ip_address_type` (IPv4/IPv6/dual-stack).
- **Transit-gateway arm**: `transit_gateway_configuration` attaches the endpoint to a TGW instead of VPC subnets (XOR the VPC fields, CEL-gated); `transit_gateway_attachment_id` exported in every arm (empty for VPC-attached endpoints, matching Terraform).
- **Folded satellites restructured**: `subnet_ids` (network associations, refs), typed `authorization_rules` (group-scoped or all-users grants), typed `routes` (ordered after associations in both engines). The invented port↔protocol CEL with no provider basis is removed. A zero-association endpoint is valid (a pre-staged front door).
- New outputs: `client_vpn_endpoint_arn`, `self_service_portal_url`, `transit_gateway_attachment_id`. Terraform floor `>= 6.11.0`.

### Cross-engine parity defects fixed (found in review, all three verified)

- Pulumi omitted `snapshot_retention_limit` when 0 while Terraform sent the explicit 0 — retention now always sent.
- Output-set divergence: `subnet_group_name`/`parameter_group_name` (cluster) and `transit_gateway_attachment_id` (Client VPN) exported in every arm, exactly once, matching Terraform.
- The ACL Pulumi module could send a phantom `""` member for an unresolved user reference — empty values now skipped.

## E2E

- Two new verifiers registered: `memorydb.go` (DescribeUsers/DescribeACLs/DescribeClusters keyed on names, state-aware — deleting counts as absent) and `client_vpn.go` (DescribeClientVpnEndpoints keyed on the endpoint ID); the memorydb SDK module added.
- Registry prerequisites: cluster → `[AwsSubnet]`; Client VPN → `[AwsCertManagerCert]` (a new imported self-signed ACM fixture serves as both server certificate and client CA chain — the first `awscertmanagercert` prerequisite).
- Six scenarios with documented arm exclusions, eight entrypoints, four conformance cases in `pkg/outputs`, hack manifests + `iac/pulumi/iac-input.yaml` for all four kinds.
- **Live lanes deferred by owner decision** (recorded in all four profiles): user/ACL dual-engine create/delete cycles were exercised against the live account (verified via CloudTrail) but a clean observed completion is pending; the Client VPN and cluster lanes carry provisioning-cost deferrals (~25–45 minutes per engine lane). The interrupted run's orphans — two Client VPN endpoints with associations, two prerequisite VPCs with subnets, two ACM fixture certs — were swept; the account verified clean (zero endpoints, zero non-default MemoryDB resources, zero fixture VPCs/certs).

## Validation

- Full offline gate: stub regen, spec/CEL suites for all four kinds, targeted Go builds + all four Pulumi entrypoints (release-equivalent) + the harness build + the Bazel repo build (`make build-go`), foreign-key guard, secret-coverage gate (user passwords sensitive), drift guard (all four enrolled) + outputs conformance, `tofu init`/`validate` + full-surface `plan` proofs for all four modules (user/ACL render 1 resource each; cluster renders 3 — cluster + folded groups; Client VPN renders 5 — endpoint, association, two authorization rules, route), all presets/scenarios/fixtures/hack manifests CLI-validated (22 manifests), site catalog mirror regenerated, scaffolding-leakage grep clean.
- Terraform-module rule uplift: the offline `tofu plan` proof procedure recorded (render tfvars via `planton tofu load-tfvars`, plan with `-refresh=false`; the AWS provider validates credentials via STS at plan time, so a real credential chain is required — with the SSO-token pitfall and the `aws configure export-credentials` escape hatch).
