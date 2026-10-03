# GCP GKE Cluster Depth Rebuild — Full Control-Plane Surface, Autopilot, and Live Dual-Engine Proof

**Date**: July 4, 2026
**Type**: Enhancement
**Components**: GCP Provider, API Definitions, IAC Modules, E2E Framework, Testing Framework

## Summary

Deep-rebuilt `GcpGkeCluster` (enum 607) from a 14-field private-cluster wrapper to the full released-provider floor of `google_container_cluster` — 56 top-level spec fields spanning VPC-native networking, private topology, control-plane access, upgrades and maintenance, node auto-provisioning, security, observability, addons, fleet registration, and Autopilot mode — with both IaC engines at 100% behavioral parity and live dual-engine E2E proof (minimal zonal + private regional scenarios, both `ip_allocation` arms). Two silent safety-default weakenings are gone, a phantom required field is removed, and a latent E2E-framework backend bug affecting every Terraform scenario with dependencies was found live and fixed.

## Problem Statement / Motivation

The GKE cluster kind was the largest depth gap in the GCP catalog: the provider resource exposes ~92 top-level keys; the spec modeled 14. An advanced organization could not reach maintenance windows, master authorized networks, node auto-provisioning, CMEK, Binary Authorization, per-kind observability, or Autopilot at all.

### Pain Points

- **Silently weakened safety defaults**: both modules hardcoded `deletion_protection=false` (the provider defaults TRUE — one accidental destroy away from losing a cluster) and `enable_private_endpoint=false`.
- **A phantom required field**: `router_nat_name` forced every cluster manifest to reference a `GcpRouterNat` that NEITHER engine consumed — GCP's cluster API has no NAT input.
- **A guaranteed-divergence foreign key**: `GcpGkeNodePool` resolved its parent cluster's name from `metadata.name` while the actual GCP name came from the required `spec.cluster_name` — any difference silently composed a node pool against a nonexistent cluster.
- **Stale module contract**: hardcoded `google = "6.19.0"` pin, `object({value})` ref typing, a release channel typed as a NUMBER (the tfvars converter emits enum names — explicit-channel manifests failed at plan), and a legacy `hack/` manifest location.
- **Region validation rejected real GCP regions**: location patterns accepted only single-digit region numbers, rejecting `europe-west12` and `me-central2` — a defect that turned out to exist in six GCP kinds.

## Solution / What's New

### Spec (released 6.50.0 floor, schema-dump verified)

- **Core**: optional `cluster_name` defaulting to `metadata.name`, location (regional/zonal), description, node locations, resource labels, `deletion_protection` defaulting TRUE (matching GCP), fleet registration.
- **Networking**: always VPC-native; `ip_allocation` models named subnetwork ranges XOR GKE-created CIDR blocks (per-range CEL exclusivity), dual-stack, additional pod ranges, overprovision toggle; Dataplane V2 (`datapath_provider`) plus its dependent features (FQDN/Cilium policy, dataplane observability, in-transit pod encryption); Cloud DNS, Gateway API, L4 ILB subsetting, intranode visibility, SNAT, service external IPs, egress tier.
- **Private topology**: private nodes, private-only endpoint, peering-based (`/28`) XOR PSC-based control planes, master global access, master authorized networks, control-plane DNS/IP endpoint posture.
- **Upgrades**: release channels (incl. EXTENDED), min master version, daily XOR recurring maintenance windows, up to 20 scoped exclusions.
- **Autoscaling**: node auto-provisioning with mandatory resource limits (an unbounded NAP is an unbounded bill), profiles, full auto-provisioning defaults (SA/KMS refs, disks, shielding, auto-repair/upgrade); VPA; HPA profile.
- **Security**: Workload Identity (default on), shielded nodes, CMEK etcd encryption (`GcpKmsKey` ref), Binary Authorization, Security Posture, authenticator groups, confidential nodes, anonymous-auth hardening, identity service, mesh certificates, Secret Manager CSI.
- **Observability**: per-kind logging/monitoring, managed Prometheus + auto-monitoring, Pub/Sub lifecycle notifications (`GcpPubSubTopic` ref), cost allocation, BigQuery usage export (`GcpBigQueryDataset` ref).
- **Addons**: HTTP LB, HPA, PD/Filestore/GCS-Fuse CSI, Backup for GKE, NodeLocal DNSCache, Config Connector, Stateful HA, Ray operator.
- **Autopilot as a mode, not a kind**: same provider resource, one immutable flag; the API's conflict set (NAP, max-pods, intranode visibility, Calico, shielded-nodes flag, dns-cache/stateful-ha addons, `allow_net_admin` on Standard) is enforced by seven message-level CEL rules before any cloud call.

68-case spec test; every provider ConflictsWith/ExactlyOneOf constraint fails at manifest time instead of 20 minutes into a cluster create.

### Composition fixes

- `router_nat_name` removed — NAT ordering belongs to composition relationships, not a phantom spec field.
- `GcpGkeNodePool.cluster_name` FK repointed to the cluster's new `status.outputs.name` (the name GCP actually assigned) — correct even when the cloud name differs from `metadata.name`.
- Outputs extended (extend-only for the node pool and Dataproc consumers): `name`, `location`, `self_link`, `master_version` join endpoint/CA/WI-pool/cluster-id.
- Registry gains `prerequisites: [GcpVpcNetwork, GcpSubnetwork]`.

### Both engines, one contract

Modern converter-contract `variables.tf` (plain-string refs, enum-as-string), `google ~> 6.0` (no beta dependency — every modeled field is GA on the released line), container API enablement with `disable_on_destroy=false`, ambient-project fallback, canonical `iac/hack/manifest.yaml`, richly commented modules (immutability ledger, Autopilot suppression, the Calico enforcement/addon pairing, the WI pool contract). Standard clusters always remove the API-mandated default node pool; Autopilot omits the node-management fields entirely. A parity defect was caught in review: the Pulumi module ignored `spec.resource_labels` while Terraform merged them — both engines now merge user labels beneath the platform attribution labels.

### Cross-kind defect sweep (multi-digit regions)

The `^[a-z]+-[a-z]+[0-9]$` region-pattern class (rejecting `europe-west12`) was fixed in six kinds: GcpGkeCluster (location, node locations, NAP locations), GcpCloudSql, GcpCloudRun, GcpCloudFunction (region + trigger region), GcpVertexAiNotebook, and GcpComputeInstance — with spec tests re-run for all six and the pattern guidance folded into the forge validation rule.

### E2E framework fix (found live)

Dependency prerequisites always deploy via Pulumi — even for Terraform scenarios — but the GCP and AWS test runners set the run-scoped Pulumi file backend only for Pulumi scenarios. Terraform scenarios' dependency stacks silently rode the machine's ambient `pulumi login` backend; when that stale `/tmp` backend vanished mid-run, dependency deploys and destroys failed. Both runners now set the backend unconditionally, and the contract is documented in the E2E README.

## Validation

- **Offline**: `make protos` ×2; 68/68 spec tests (+5 sibling kinds re-run after the pattern sweep); release-equivalent Pulumi build; `tofu validate` + offline `planton tofu plan` through the real tfvars converter with the plan inspected field-by-field; `secret-coverage --check`; `validate-refs --check`; `validate-outputs` 8/8 on BOTH module dirs; new `pkg/outputs` conformance case; every preset/hack/scenario/prerequisite manifest through `planton validate`; `make build-go`; `make reset-gazelle`; framework tests green.
- **Live (project planton-e2e, dual-engine)**: `verify/gke_cluster.go` (container API; RUNNING + name/endpoint/WI-pool posture assertions); minimal zonal scenario (GKE-managed ranges) and private regional scenario (named ranges + `/28` control plane + Dataplane V2 + maintenance window). Pulumi: minimal 18m20s, private 18m56s. Terraform: minimal 27m05s, private green on re-run after the backend fix. Zero orphaned clusters/networks/subnets after the final sweep.
- **Audit**: Fully Complete — **PARITY ✅**, zero PARITY-EXCEPTIONs (`docs/audit/2026-07-04-134500.md`).

## Impact

- GKE on Planton now covers what advanced organizations actually deploy — private production clusters, Autopilot, CMEK/BinAuthz compliance shapes, NAP — with cross-field mistakes caught before any cloud call.
- The `deletion_protection` default now matches GCP: destroying a real cluster requires an explicit, reviewable spec change.
- Every Terraform E2E scenario with prerequisites (all providers) now runs against the run-scoped backend instead of ambient developer state.

## Related Work

- Builds on the VPC-network rename and the subnetwork depth rebuild (network FKs land on the final vocabulary).
- Recorded skips (released-vs-main deltas, beta-only blocks, deliberate exclusions) live in the kind's `docs/README.md`.
- `GcpGkeNodePool` depth rebuild is the natural next step; its FK contract is already corrected.

---

**Status**: ✅ Production Ready
