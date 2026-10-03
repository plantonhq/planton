# Kubernetes Karpenter (engine + NodePool + EC2NodeClass), Cluster Autoscaler, and Velero at full depth: five new kinds, a live backup/restore DR proof, and the first OCI-served typed Helm kind

**Date**: 2026-07-23
**Scope**: `apis/dev/planton/shared/catalogkind` (five new kinds, 863–867), `apis/dev/planton/provider/kubernetes` (kuberneteskarpenter, kuberneteskarpenternodepool, kuberneteskarpenterec2nodeclass, kubernetesclusterautoscaler, kubernetesvelero forged), `pkg/kubernetes/kubernetestypes` (karpenter CRD generation set: NodePool + EC2NodeClass typed Pulumi SDK), `aa_e2e/verify` (Velero DR + ClusterAutoscaler verifiers, Karpenter registrations), `e2e` (ten new entrypoints), `pkg/outputs` (+5 conformance cases), `pkg/iac/importmap` (+2 proven round-trips ledgered), site catalog, `_rules/catalog-kind` (update rule: OCI chart pinning/wiring; forge rule: chart-rendered sandbox fixtures; spec-validate flow: negated-all() CEL guard)

## What changed

Five new catalog kinds, each at full configuration depth with
dual-engine parity; live kind-cluster E2E on both engines where the kind
cluster can honestly prove it, with clone-verified deferred profiles where
it cannot:

### KubernetesKarpenter (863)

- The Karpenter node-provisioning engine from the official OCI-served Helm
  charts (`oci://public.ecr.aws/karpenter/{karpenter,karpenter-crd}`,
  default 1.14.0) — the catalog's first typed Helm kind on an OCI registry.
  The served charts were verified through the ECR registry API because the
  repo-tag vendored Chart.yaml lags the served chart (1.13.0 vs 1.14.0);
  both engines encode the OCI wiring asymmetry with explanatory comments
  (Terraform: repository + bare chart name, joined by the provider; Pulumi:
  the fully joined chart reference with no repository opts).
- TWO fixed-name releases: `karpenter-crd` (upstream's supported mechanism
  for upgradable CRDs — Helm never upgrades CRDs bundled in the main chart)
  with `helm.sh/resource-policy: keep` stamped by default so uninstall
  never cascade-deletes the cluster's NodePools/NodeClaims, and the
  `karpenter` controller release (skip_crds unconditionally; atomic; 600s).
- Typed surface: cluster identity (name required by the chart's own
  `required` guard, endpoint, EKS control-plane discovery, CA bundle), the
  AWS arm in a cloud oneof (IRSA role, SQS interruption queue, isolated
  VPC, reserved ENIs, zonal shift, VM memory overhead), controller sizing +
  log level, pod-batching windows, scheduler preference/minValues policies,
  all six feature gates, controller scheduling (incl. host_network for
  custom-CNI clusters), ServiceMonitor, helm_values escape hatch.
- Type fidelity proven in offline plans on both engines: `reservedENIs`
  renders as the chart's STRING, `vmMemoryOverheadPercent` as a NUMBER.
- Kind-cluster E2E is honestly deferred with the reason in the profile:
  the controller startup resolves its region from IMDS, loads AWS
  credentials, and refreshes EC2 instance-type data, and each failure is
  fatal (verified in the pinned controller source) — the live lane rides
  the batched EKS real-cluster set.

### KubernetesKarpenterNodePool (864)

- Faithful cluster-scoped projection of the karpenter.sh/v1 NodePool CRD:
  the NodeClaim template (scheduling requirements with Karpenter's extended
  operators and alpha minValues, taints/startup taints, node lifetime and
  drain ceiling), disruption policy (consolidation modes, budgets with
  schedule/duration windows and reasons), resource limits, pool weight, and
  the alpha static-capacity mode — with the CRD's own CEL rules mirrored
  (restricted label domains, requirement operator contracts, budget
  pairings, static-mode exclusions) so mistakes surface at validate time.
- `node_class_ref.name` is a real foreign key to
  KubernetesKarpenterEc2NodeClass — infra charts wire the fleet chain with
  valueFrom and get true dependency edges.
- Terraform module generated as the kubectl_manifest projection; Pulumi on
  the typed crd2pulumi SDK (new karpenter generation set in
  kubernetestypes).

### KubernetesKarpenterEc2NodeClass (865)

- Faithful cluster-scoped projection of the karpenter.k8s.aws/v1
  EC2NodeClass CRD — the AWS machine template: AMI selector terms (alias /
  id / name+owner / SSM parameter / tags, with the CRD's mutual-exclusion
  rules), subnet and security-group discovery terms, role XOR
  instance-profile, EBS block-device mappings, capacity-reservation
  selectors, connection tracking, kubelet configuration (all twelve
  upstream-supported fields with signal-key restrictions and pairing
  rules), IMDS posture (secure CRD defaults preserved by presence-aware
  rendering), network-interface/EFA layout, placement groups, restricted
  tag keys, and user data.
- Acronym-cased CRD keys (`associatePublicIPAddress`, `kmsKeyID`,
  `snapshotID`, `clusterDNS`, `cpuCFSQuota`, `imageGC*ThresholdPercent`,
  `httpProtocolIPv6`, `ownerID`) are pinned with json_name and verified
  against both the CRD schema and the generated SDK.

### KubernetesClusterAutoscaler (866)

- The Kubernetes Cluster Autoscaler from the official chart
  (`cluster-autoscaler` @ https://kubernetes.github.io/autoscaler, default
  chart 9.59.0 / app 1.35.0), with a desirability-curated provider oneof:
  AWS (tag auto-discovery XOR static ASGs; IRSA XOR declared keys), Azure
  VMSS (workload identity XOR managed identity XOR service principal), GCE
  MIG name-prefixes (+ GKE Workload Identity annotation), Cluster API (all
  five connection modes with the kubeconfig-secret contract), Civo, and
  the KWOK simulation arm (sandbox/test-only — the arm that makes a
  cloud-free kind lane possible at all). The docs are explicit that GKE and
  AKS ship a MANAGED autoscaler configured on the cluster kinds — this
  kind is for EKS/self-managed postures.
- Typed scaling block for the flags every installation tunes (expander
  chain, node-group balancing, scan interval, provision timeout, the full
  scale-down set) + the chart's own `extraArgs` map contract for the long
  tail, user entries winning on collision.
- Live E2E green on both engines on the kwok arm — install plus the
  reconcile loop's own heartbeat (the `cluster-autoscaler-status`
  ConfigMap) — with the blind import round-trip proven. The chart is
  self-contained for kwok (it renders the provider ConfigMaps itself);
  the deployment/service-account names embed the provider arm via the
  chart's `<release>-<cloudProvider>-<chartName>` fullname, which the
  verifier resolves through the release's instance label.

### KubernetesVelero (867)

- Velero cluster backup / disaster recovery from the official chart
  (`velero` @ https://vmware-tanzu.github.io/helm-charts, default chart
  12.1.0 / app 1.18.1). The default BackupStorageLocation and its provider
  plugin are a typed backend oneof: S3 AND any S3-compatible store (MinIO,
  Ceph RGW, Spaces — endpoint URL + path-style + optional CA), GCS, and
  Azure Blob; per-arm plugin init-containers default to the official
  images at the version paired with Velero 1.18 (v1.14.2) and stay
  overridable. Keyless postures first-class per arm (IRSA / GCP Workload
  Identity / Azure Workload Identity with the client-id annotation + use
  label); declared credentials render the plugin-documented `cloud` file
  formats byte-for-byte (verified against the pinned plugin READMEs) into
  the chart's Secret, secret-by-default.
- Volume data travels either path: CSI snapshots (features EnableCSI +
  optional snapshot data-mover) or file-system backup (node-agent
  DaemonSet, kopia) — with typed Schedule entries (cron, TTL, scope
  filters, per-schedule overrides), server tuning, and the DR-safety
  posture documented: the chart's crds/-directory CRDs survive uninstall
  by Helm's own contract, so backup records outlive the kind
  (cleanup_on_uninstall exists but is loudly destructive).
- Live E2E green on both engines including the program's first FULL DR
  behavioral proof: against an in-cluster MinIO fixture, the verifier
  backs up a live namespace (Backup → Completed), DELETES the namespace,
  restores it (Restore → Completed), and asserts the marker data came
  back intact — with verifier-owned Backup/Restore CRs (their CRDs are
  installed by the kind under test) and blind import round-trips on
  both scenarios.

## E2E and verification

- Velero: 2 scenarios × 2 engines green incl. the DR proof; import
  round-trips proven. ClusterAutoscaler: kwok lane × 2 engines green +
  round-trip. Karpenter trio: deferred profiles with clone-verified
  reasons; all six entrypoints skip cleanly on kind and activate when the
  EKS batch flips the profiles. Zero orphans on both persistent clusters.
- Offline: 271 spec tests across the five kinds; ten tofu plan proofs
  (full-surface + minimal per kind) with type-fidelity spot-checks;
  secret-coverage, validate-refs, outputs conformance (+5 cases),
  import-map conformance, `make build-go`, kind map, e2e matrix, site
  catalog regen — all green.

## Workflow rules sharpened (timeless)

- Update rule: OCI-served charts have no index.yaml — verify served
  versions through the registry tags API and read the served chart itself;
  plus the two-engine OCI wiring asymmetry and where its exemplar lives.
- Forge rule: before authoring a fixture for a sandbox arm, check whether
  the chart renders the sandbox's supporting objects itself — pre-creating
  one fails the install with Helm's release-ownership error, and the
  failure class is live-only.
- Spec-validate flow rule: a negated `all()` CEL rule over an optional
  repeated field is vacuously violated on the empty list — guard with
  `size() == 0 ||` unless the list has `min_items >= 1`, and always test
  the omitted-list case.
