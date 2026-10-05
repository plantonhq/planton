# GCP GKE Node Pool Depth Rebuild

**Date**: July 4, 2026
**Type**: Enhancement
**Components**: API Definitions, GCP Provider, IaC Modules, E2E Framework, Presets, Documentation

## Summary

Deep-rebuilt the `GcpGkeNodePool` kind from a 12-field skeleton to the released
google-provider floor (~40 spec surfaces): full node configuration (GPUs with
GKE-managed drivers, Spot capacity, shielded/confidential VMs, CMEK boot disks, local
SSDs, kubelet and Linux tuning), per-zone and total autoscaling with scale-to-zero,
surge and blue-green upgrade strategies, compact placement, and queued provisioning.
Both IaC engines were rewritten to one contract, fixing a real cross-engine parity
break in how the pool located its parent cluster. Proven live on both engines with
full cluster prerequisite chains, zero orphans.

## What Changed

### Spec redesign (`spec.proto`, contiguous renumbering)

- **Parenting by reference**: `cluster_name` resolves from the cluster's `name` output
  and the new `location` field from its `location` output — the names GKE actually
  assigned. The old `cluster_project_id` (which resolved from the cluster's
  `spec.project_id` and was empty for ambient-project clusters) became `project_id` →
  `GcpProject` with the ambient-project fallback.
- **Sizing**: the fixed-count XOR autoscaling oneof gained the provider's total-limit
  arm (`total_min_nodes`/`total_max_nodes`), scale-to-zero, `location_policy`, and
  `initial_node_count` for autoscaled pools.
- **Node config**: machine shape, boot disk (incl. hyperdisk types), image types,
  `service_account` as a `StringValueOrRef` → `GcpServiceAccount`, OAuth scopes,
  Kubernetes labels + taints, `spot` and `preemptible` as distinct fields, GPU
  accelerators (driver install, MIG partitioning, time-sharing/MPS), shielded and
  confidential VMs, `boot_disk_kms_key` → `GcpKmsKey`, local SSD surfaces, image
  streaming, gVNIC + NCCL Fast Socket, workload metadata mode, reservation affinity,
  secondary boot disks, kubelet tuning, Linux sysctls/cgroup/hugepages, logging
  variant, flex-start, and max run duration.
- **Lifecycle**: `management` with positive booleans (`auto_repair`/`auto_upgrade`,
  default true — replacing the negative `disable_*` pair), surge/blue-green
  `upgrade_settings` with rollout pacing, `placement_policy`, queued provisioning,
  and pool-level `network_config` (dedicated pod ranges, private-nodes override,
  TIER_1 egress).
- **Pre-deploy coherence CEL**: per-zone XOR total autoscaling limits (+ min≤max both
  arms), initial-count-requires-autoscaling, blue-green-settings/strategy pairing,
  batch percentage XOR count, spot XOR preemptible, fast-socket-requires-gVNIC,
  specific-reservation key pairing, create-pod-range-requires-name, and enum-valued
  strings validated against the released provider's accepted values.
- 67-case spec test (from 1 case).

### Both engines, one contract

- **Parity break fixed**: the Terraform module previously ignored the spec's cluster
  location entirely and discovered the cluster through a `google_container_cluster`
  data source with a wildcard location, while Pulumi used the spec field. Both engines
  now consume the same resolved references.
- **Pulumi spot/preemptible conflation fixed**: one flag used to set both provider
  fields; they are distinct capacity models.
- Hardcoded opinions removed on both sides: `upgrade_settings{max_surge=2}`, a fixed
  3-scope `oauth_scopes` list, the `gke-<cluster>` network tag, forced metadata.
- Converter-contract `variables.tf` (plain-string refs), `~> 6.0` float (was hardcoded
  `6.19.0`), `container.googleapis.com` enablement with `disable_on_destroy=false`,
  ambient-project fallback, platform labels merged over user resource labels,
  `disable-legacy-endpoints=true` enforced beneath user metadata, autoscaler-owned
  `node_count` ignored by both engines, canonical `iac/hack/manifest.yaml` (legacy
  `hack/` copy deleted), stale `binary:` comment dropped from Pulumi.yaml.
- Outputs extended (+`location`, `version`); effective min/max computed identically
  from the sizing mode; `pkg/outputs` conformance case added.

### Registry, presets, docs

- Registry `prerequisites: [GcpGkeCluster]`.
- 3 presets: on-demand autoscaling (rewritten), spot cost-optimized (rewritten with
  the taint-fence pattern), gpu-accelerated (new — L4 inference pool with GKE-managed
  drivers and image streaming).
- README, catalog page, research doc, and both IaC READMEs rewritten with the 90/10
  coverage table and the evidence-based recorded-skips ledger (main-only fields
  verified against a released 6.50.0 schema dump).
- Site catalog copies refreshed via the docs copy script (also caught up stale copies
  for previously rebuilt GCP kinds).

### E2E

- New `verify/gke_node_pool.go` (container API `nodePools.get` on the fully qualified
  `node_pool_id`; RUNNING + name + instance-group + autoscaling-posture assertions).
- `gcpgkecluster` gained its published `e2e/prerequisite.yaml` (minimal zonal cluster
  — its first outing as a prerequisite); the node pool carries a consumer-scoped
  subnetwork profile so the cluster chain resolves identically for both consumers.
- Two scenarios: `minimal` (fixed count) and `autoscaling-spot` (scale-to-zero, Spot,
  taints, kubelet hardening, surge settings, `initial_node_count`).
- **Live dual-engine proof on `planton-e2e`: 4/4 scenario-runs green** — Pulumi
  19m22s/18m16s, Terraform 19m03s/21m00s, each deploying and destroying the full
  VPC → subnetwork → zonal-cluster prerequisite chain. Zero orphans.

### Parity doctrine

- `pkg/iac/MODULE_PARITY.md` gained a "Reference resolution source" checklist
  dimension: never substitute a provider data-source lookup on one engine for a spec
  field the other engine reads — the class of defect this rebuild fixed.

## Validation

- Offline: `make protos` ×2 + gazelle; 67/67 spec tests; release-equivalent Pulumi
  build; `tofu validate` + offline `planton tofu plan` through the real tfvars
  converter (plan inspected field-by-field); `secret-coverage --check`;
  `validate-refs --check`; `validate-outputs` 8/8 on both module dirs; new
  `pkg/outputs` conformance case; all presets/hack/scenario/prerequisite manifests
  through `planton validate`; `make build-go`; framework tests green.
- Live: 4/4 dual-engine scenario-runs green, zero orphans.
- Audit: Fully Complete — PARITY ✅ (`docs/audit/2026-07-04-162249.md`).
