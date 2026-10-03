# GCP Dataproc Fold + Autoscaling Policy Kind, Cloud Composer Depth + User Workloads Kinds

**Date**: July 7, 2026
**Type**: Feature
**Components**: API Definitions, Terraform Modules, Pulumi Modules, E2E Framework, Kind Registry

## Summary

The GCP analytics family reaches the 90/10 bar. `GcpDataprocVirtualCluster` is
retired and its Dataproc-on-GKE surface folded into `GcpDataprocCluster` as a
mutually exclusive arm — the provider has exactly one cluster resource, and the
catalog now mirrors it with exactly one kind. `GcpDataprocAutoscalingPolicy`
becomes a first-class, shareable kind (one policy governs many clusters by
reference), closing the cluster's plain-string policy-URI composition gap.
`GcpCloudComposerEnvironment` is deep-rebuilt to the Composer 2/3 floor, and
two delivery kinds — `GcpCloudComposerUserWorkloadsSecret` (the catalog's
first `(sensitive)` map) and `GcpCloudComposerUserWorkloadsConfigMap` — give
Airflow DAGs their credentials and configuration as composable nodes.

## What Changed

### GcpDataprocCluster (651) — deep rebuild with the folded GKE arm

- **Two deployment arms, mirroring the API**: `cluster_config` (Compute
  Engine) XOR `virtual_cluster_config` (Dataproc-on-GKE), enforced pre-deploy
  by presence-only CEL, with the API's labels-rejected-on-virtual-clusters
  rule enforced too. The virtual arm composes by reference:
  `gke_cluster_target` → `GcpGkeCluster.cluster_id`, `node_pool_target[]` →
  `GcpGkeNodePool.node_pool_id`, and `spark_history_server_config` is a
  self-kind reference on `cluster_id`.
- **GCE-arm depth to the released 6.x floor**: `cluster_tier`, shielded and
  confidential VMs, reservation affinity (with the SPECIFIC_RESERVATION
  coherence rule), sole-tenant node-group affinity, disk
  `local_ssd_interface`, secondary-worker `instance_flexibility_policy`
  (ranked machine-type selections + the standard/spot capacity mix),
  `security_config` (Kerberos XOR personal-cluster identity mapping — all
  Kerberos secret fields are KMS-encrypted GCS URIs, paths not material),
  `metastore_config` (an un-defaulted reference, ready for a future
  metastore-service kind), `dataproc_metric_config`, `auxiliary_node_groups`
  (dedicated DRIVER capacity), and user `labels`.
- **`autoscaling_policy_uri` is now a real reference** with
  `default_kind = GcpDataprocAutoscalingPolicy`.
- **Output honesty**: the phantom `cluster_uuid` output (hardcoded empty in
  both engines; the provider has no such attribute) is gone; `cluster_id` is
  the fully qualified resource path on both engines (the Pulumi module
  previously exported the short name — an output-parity defect).
- Ambient project, `planton-ai_*` Terraform labels, Dataproc API enablement,
  canonical Pulumi.yaml. 94-case spec test; 4 presets including
  `spark-on-gke`.

### GcpDataprocAutoscalingPolicy (652, `gcpdpasp`) — new kind

Primary/secondary worker bounds and weights plus the YARN algorithm
(scale-up/down factors 0.0–1.0, cooldown, graceful decommission window),
with the API's constraints enforced pre-deploy. The `name` output is the
handle clusters attach; the Dataproc API's `regions|locations` path-segment
equivalence is documented on the output. Full anatomy: 44-case spec test,
3 presets, both engines, E2E scenario + published prerequisite profile.

### GcpCloudComposerEnvironment (680) — deep rebuild

- New floor surfaces: `ip_allocation_policy` (per-range named-range XOR
  CIDR), `master_authorized_networks_config`, `data_retention_config`
  (task-log storage mode + Airflow metadata retention 30–730 days),
  `storage_bucket` (→ GcpGcsBucket), `cloud_data_lineage_integration`,
  `enable_ip_masq_agent`, user `labels`, `ENVIRONMENT_SIZE_EXTRA_LARGE`.
- Defects closed: the Pulumi module had NO Pulumi.yaml (it could never have
  deployed — a silent no-op), the Terraform module stamped bare unprefixed
  label keys (a live label-parity break), no API enablement, no hack
  manifest, stale `object({value})` variable typing, and no ambient-project
  fallback. The `gke_cluster` output is documented honestly: empty on
  Composer 3 (tenant-project GKE), populated on Composer 2 — verified live.
- Composer 3's mandatory workloads service account (holding
  `roles/composer.worker`) is encoded as part of the environment's minimum
  composition (registry prerequisites + spec guidance).

### GcpCloudComposerUserWorkloadsSecret (681) / ConfigMap (682) — new kinds

Kubernetes Secret/ConfigMap delivery into an environment's workloads
namespace, composing against the environment by reference. The Secret's
`data` map is `(sensitive)` — base64 contract validated pre-deploy, secret
in Pulumi state, never in outputs. Full anatomy on both kinds.

### GcpFirewallRule — conformance

Ambient-project fallback on both engines, converter-contract variable typing
(plain strings for resolved references), canonical Pulumi.yaml (the stale
`binary:` option made it undeployable as an E2E dependency), and its first
E2E verifier.

### E2E framework and registry

- Harness gains Dataproc + Composer API clients and six verifiers with
  posture assertions (cluster RUNNING + staging bucket; policy bounds +
  algorithm; environment RUNNING + Airflow UI + DAG bucket; secret/configmap
  existence; firewall payload).
- Registry prerequisites now express real composition requirements found
  live: Dataproc needs a VM identity with `roles/dataproc.worker` and an
  intra-cluster firewall rule on custom VPCs; Composer 3 needs its explicit
  workloads identity. Consumer-scoped prerequisite profiles carry the
  grants.
- `e2e/README.md`: Dataproc/Composer batch budgets and the
  identity-validated-at-create pattern.

## Validation

- Spec tests: 242 cases across the five kinds, all green.
- Both engines build per kind; `tofu validate` + offline `planton tofu plan`
  through the real converter for all six touched modules.
- `secret-coverage --check` (the sensitive map correctly classified),
  `validate-refs --check` (clean after the retirement), `validate-outputs`
  on both module dirs for all five kinds, five new outputs-conformance
  cases, 34 manifests through `planton validate`.
- Live on the test project: autoscaling policy and Composer environment
  green on BOTH engines (full create → verify → destroy; the environment
  runs ~33 minutes per engine). The Dataproc cluster's live batch runs the
  five-node chain (VPC → subnetwork → firewall → identity → policy) —
  see the kind's `e2e/` for scenarios.
