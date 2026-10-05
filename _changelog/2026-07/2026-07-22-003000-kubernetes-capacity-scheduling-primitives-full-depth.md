# Kubernetes capacity and scheduling primitives at full configuration depth: six new kinds forged

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetespersistentvolumeclaim, kubernetesstorageclass, kubernetesresourcequota, kubernetespriorityclass, kubernetespoddisruptionbudget, kuberneteshorizontalpodautoscaler — all new), `apis/dev/planton/shared/catalogkind`, `aa_import`, `aa_e2e/verify`, `e2e`, `pkg/outputs`, `pkg/iac/importmap`, Makefile E2E tiers, site catalog, `_rules/catalog-kind/forge`

## What changed

Six new Kubernetes building-block kinds covering the cluster's capacity and
scheduling controls, each designed field-by-field from the upstream core API
types and shipped with both IaC engines, presets, docs, catalog pages, import
recipes, and dual-engine E2E:

### KubernetesPersistentVolumeClaim (new)

The durable-disk primitive: access modes (workload-template vocabulary),
storage request/limit, `storage_class_name` as a foreign key to
KubernetesStorageClass, volume mode (Filesystem/Block), static binding
(`volume_name` + selector), and typed data sources (clone a claim / restore a
VolumeSnapshot). Two upstream nuances modeled honestly:

- **Empty vs absent StorageClass**: Kubernetes distinguishes
  `storageClassName: ""` (bind only pre-provisioned volumes) from an unset
  class (cluster default applies). A single string cannot carry that, so the
  spec has an explicit `disable_dynamic_provisioning` bool, CEL-guarded
  against combining with a named class.
- **Neither engine waits for Bound**: under a WaitForFirstConsumer class a
  claim is correctly Pending until a pod consumes it. Pulumi sets skipAwait;
  Terraform sets `wait_until_bound = false` (declared config-only in the
  import catalog at authoring time).

One documented parity exception: the terraform kubernetes provider cannot
express PVC data sources; the Terraform module fails the plan loudly via a
precondition (negative-proofed offline) instead of silently provisioning an
empty volume where restored data was requested.

### KubernetesStorageClass (new)

The cluster's storage menu: immutable provisioner + parameters
(provisioner-vocabulary map — deliberately upstream-shaped), reclaim policy,
volume binding mode, expansion, mount options, allowed topologies, and a
first-class `is_default_class` bool that renders upstream's
`storageclass.kubernetes.io/is-default-class` annotation. Both engines
replace (delete-before-create) on immutable-field changes. One documented
parity exception: the terraform provider models allowed_topologies as a
single selector term (max_items = 1); multiple OR'd terms fail the TF plan
loudly via precondition.

### KubernetesResourceQuota (new)

Namespace resource governance as ONE kind managing TWO objects: the
ResourceQuota (hard caps, scopes, scope selectors — with CEL mirroring live
API rules: conflicting scope pairs, Exists-only operators on pod-behavior
scopes, best_effort-caps-only-pods) plus an optional companion LimitRange
(`limit_defaults`: per-type max/min/defaults/ratio) sharing the quota's name.
They answer one governance question — and pairing them is what keeps a
compute quota livable, since a quota on requests/limits makes the API reject
pods that omit them. KubernetesNamespace's T-shirt resource profiles remain
the simple path for Planton-created namespaces; this kind is the
full-fidelity instrument.

### KubernetesPriorityClass (new)

The workload importance ladder: value (CEL-capped at the user-definable
1,000,000,000 ceiling; `system-` name prefix rejected), global default flag,
description, preemption policy (preempt_lower_priority default / never for
queue-jumping batch tiers). Value is immutable; both engines replace
delete-before-create.

### KubernetesPodDisruptionBudget (new)

The voluntary-disruption floor: required selector (empty-but-present selects
all pods — required because a null selector in policy/v1 protects nothing),
int-or-percent `min_available`/`max_unavailable` with the exactly-one CEL
rule and the same string vocabulary as the workload kinds' built-in blocks,
and the unhealthy-pod eviction policy. One documented parity exception: the
terraform provider cannot express `unhealthyPodEvictionPolicy`; requesting
`always_allow` fails the TF plan via precondition.

### KubernetesHorizontalPodAutoscaler (new)

The complete autoscaling/v2 surface: scale target as a foreign key
(default KubernetesDeployment via its exported `deployment_name` output;
DaemonSet targets CEL-rejected), min/max replicas, all five metric source
families (resource, container_resource, pods, object, external) with the
three target value forms (utilization / raw_value / average_value —
`raw_value` maps to the API's "Value"; generated code cannot use the bare
word "value" as an enum constant), and full per-direction scaling behavior
(stabilization windows, select policies, velocity policies). CEL enforces
the source-matches-type and value-form-matches-target-type contracts the
controller otherwise fails silently. Feature-gated scale-to-zero and
configurable tolerance are deliberately unmodeled until they graduate.

Both new standalone HPA/PDB kinds document the boundary with the workload
kinds' built-in availability blocks: built-in for a Planton workload's own
pods, standalone for operator-managed/non-Planton pods and the advanced
surface — never both on one target.

## Registry and satellites

- Enum entries 815–820 in the Kubernetes building-blocks sub-band; kind map
  regenerated.
- Import catalog: three new resource-type rows (PVC with `wait_until_bound`
  config-only, StorageClass, PriorityClass — quota/LimitRange/HPA/PDB rows
  already existed from the workload satellites); import maps for all six
  kinds; blind re-import round-trip proven live for ALL SIX (15
  scenario-engine lanes), recorded in the importmap README ledger.
- Outputs conformance cases for all six kinds; six new verifier cases
  (existence-based, with the PVC/HPA behavioral caveats documented at the
  case site); Tier-1 E2E targets and test entrypoints extended; e2e matrix
  and site catalog regenerated (six new public catalog pages).
- 20 presets across the six kinds, all machine-validated.

## Validation

- Six spec-test suites green (protovalidate; live-API rules mirrored from
  upstream validation source).
- Per-kind builds + release-equivalent Pulumi entrypoint builds green;
  repo-wide `make build-go` green; secret-coverage, validate-refs,
  import-map conformance, outputs conformance green.
- Offline tofu proofs per kind with optional blocks present AND absent,
  plus three negative proofs (each Terraform precondition fires).
- Dual-engine E2E green on a kind cluster for all six kinds — full
  create → verify-outputs → verify-resources → destroy → verify-clean —
  including two composed foreign-key scenarios resolved live:
  StorageClass → PVC, and Deployment → HPA scale target.
- Import round-trip green for all six kinds (terraform lane).
- Deferred honestly (not runnable on kind): HPA behavioral scaling (no
  metrics-server), PriorityClass preemption under pressure, PVC
  clone/snapshot data-source provisioning (no snapshot-capable CSI);
  object-shape and API acceptance are verified live.

## Forge-rule freshness fix

The forge rule's E2E test-name discovery note claimed Kubernetes kinds
"still fall back to the legacy prefix table"; discovery has been
registry-driven for all providers with only a small verified-deviation
override map. The sentence now describes the actual mechanism.
