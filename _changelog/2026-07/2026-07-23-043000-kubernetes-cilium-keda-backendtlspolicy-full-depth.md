# Kubernetes Cilium, KEDA, and BackendTLSPolicy at full depth: three new kinds, a CNI-capable E2E cluster profile, and live NetworkPolicy enforcement proofs

**Date**: 2026-07-23
**Scope**: `apis/dev/planton/shared/catalogkind` (three new kinds; Gateway API family band extended to 850, Istio family renumbered 851–859, MetricsServer 860), `apis/dev/planton/provider/kubernetes` (kubernetescilium, kuberneteskeda, kubernetesbackendtlspolicy forged; kubernetesnetworkpolicy behavioral scenario), `pkg/kubernetes/kubernetestypes` (BackendTLSPolicy added to the gateway-api generation set), `aa_e2e` (cluster profiles) + `aa_e2e/verify` (three new verifiers + NetworkPolicy behavioral verifier), `e2e` + `e2e/framework/runner` (per-scenario cluster routing), Makefile tiers, `pkg/outputs`, `pkg/iac/importmap`, site catalog, `_rules/catalog-kind` (forge + spec-proto flow)

## What changed

Three new catalog kinds, each at full configuration depth with
dual-engine parity, live kind-cluster E2E on both engines, and blind
state-import round-trips proven:

### KubernetesCilium (861)

- The eBPF CNI/network-security/observability engine from the official chart
  (`cilium` @ https://helm.cilium.io, default 1.19.6 — the newest stable
  line). Typed surface: cluster identity, IPAM modes (cluster-pool tuning,
  kubernetes, cloud modes), routing (tunnel/native + protocol + native-CIDR
  knobs), kube-proxy replacement with the API-server address pair it
  requires, **CNI chaining** (aws-cni/flannel/generic-veth/portmap, chaining
  target, the exclusive=false requirement enforced by CEL) so Cilium can run
  ON TOP of an existing CNI instead of replacing it, cloud datapath arms
  (AWS ENI, AKS BYOCNI, GKE — mutually exclusive by CEL), the full Hubble
  stack (relay/UI/metric families/ServiceMonitor with dependency CEL),
  transparent encryption (WireGuard/IPsec + node encryption), policy
  enforcement mode, Cilium's Gateway API implementation (CEL-gated on
  kube-proxy replacement — the operator otherwise disables it with only a
  log warning), bandwidth manager + BBR, operator/agent sizing, telemetry,
  and the helm_values escape hatch merged last.
- Fixed release name `cilium` (the dataplane is a cluster singleton); no
  fullnameOverride (the chart's workload names are fixed); 600s atomic
  install (the agent must roll out on every node while nodes transition
  NotReady→Ready).
- Chart-values type traps encoded in both engines: `kubeProxyReplacement`
  and `k8sServicePort` are STRINGS in the chart's values contract.

### KubernetesKeda (862)

- Kubernetes Event-Driven Autoscaling from the official chart (`keda` @
  https://kedacore.github.io/charts, default 2.20.1). Typed surface: CRD
  lifecycle with keep-on-uninstall implemented via the
  `helm.sh/resource-policy: keep` annotation (the chart has no keep knob and
  a plain uninstall cascade-deletes every ScaledObject in the cluster),
  watch-namespace fencing, per-kind sizing (operator / metrics API
  server / admission webhooks incl. failure policy), ambient pod identity
  for scalers (AWS IRSA, Azure Workload Identity, GCP Workload Identity —
  each with completeness CEL), internal-TLS certificates (operator
  self-generation or cert-manager with an issuer FK), scaler HTTP timeout,
  scheduling, telemetry, helm_values.
- Fixed release name `keda` (the external.metrics.k8s.io APIService is a
  cluster singleton). Outputs export the fixed `keda-operator` service
  account name — the subject keyless cloud bindings are written against.
- The chart's asymmetric value layout (`operator.replicaCount` vs
  `resources.operator`; the singular `resources.metricServer` key) is
  encoded and commented in both engines.

### KubernetesBackendTlsPolicy (850)

- The Gateway API v1 BackendTLSPolicy — gateway-to-backend TLS origination
  and verification — as a faithful projection kind on the route-kind
  precedent: kubectl_manifest on Terraform (plannable before the CRDs
  exist), the typed SDK on Pulumi. Target refs are real foreign keys to
  KubernetesService (section_name narrows to a named port); CA-bundle refs
  are foreign keys to KubernetesConfigMap; the trust-anchor arms
  (ca_certificate_refs XOR well_known_ca_certificates) and the SAN
  type/value pairing rules mirror the CRD's own CEL exactly; the same-target
  section-name presence/uniqueness rules are enforced at validate time.
- The Gateway API family enum band extends to 850 so the policy sits
  family-adjacent; the Istio family shifts to 851–859 and MetricsServer to
  860 (zero-adoption mechanical renumber; every functional surface is
  name-keyed).
- The typed SDK generation set now includes the BackendTLSPolicy CRD.

### E2E cluster profiles (framework)

- A scenario manifest can now opt into a dedicated persistent kind cluster
  via the `planton.dev/e2e-cluster-profile` annotation. The `cilium-cni`
  profile provisions a single-node cluster with the default CNI disabled —
  the posture a real Cilium cluster runs and the only honest way to prove a
  CNI kind. The profile cluster is created lazily, waits on API-server
  readiness instead of kind's node gate (a CNI-less cluster's nodes are
  NotReady BY DESIGN until the CNI installs), and the test entrypoints
  activate the selected cluster's kubeconfig per scenario (scenarios run
  serially in-process, so the routing is race-free by construction).
  Profiles are ignored in the external-cluster lane.

### Behavioral proofs (live, both engines)

- **NetworkPolicy enforcement**: on the cilium-cni profile cluster, an
  ingress-deny policy provably BLOCKS a client pod's request to the selected
  backend (curl times out against the datapath drop) and traffic flows again
  after the policy is destroyed — the enforce-and-release cycle that kind's
  default CNI can never demonstrate.
- **Cilium install**: agent DaemonSet fully rolled out, operator Available,
  and all nodes NotReady→Ready — the transition IS the proof Cilium became
  the cluster's CNI.
- **KEDA scaling**: the behavioral verifier applies a deterministic
  cron-trigger ScaledObject against a plain target-Deployment fixture after
  the install assertions pass, proves KEDA drives a real scale-up (1 → 2
  ready replicas), and deletes what it applied. The driver CR is
  verifier-owned because scenario fixtures deploy before the kind under
  test — a fixture-borne ScaledObject would precede the CRDs KEDA itself
  installs.

### Import round-trips

- Blind re-import proven for all three kinds (Cilium on the profile cluster;
  KEDA incl. a re-import with the live scale-target fixture present;
  BackendTLSPolicy's 4-part namespaced composed IDs across all three
  scenarios). Ledgered in `pkg/iac/importmap/README.md`.

### Faithful-projection lesson institutionalized

- The BackendTLSPolicy full-surface Terraform lane caught live that protojson
  derives `wellKnownCaCertificates` from the field name while the CRD key is
  `wellKnownCACertificates` (capital CA) — the API server rejects the
  miscased key as undeclared, and only on the engine that renders manifests
  through protojson. Fixed by pinning `json_name` in the spec; the
  acronym-casing rule is now timeless guidance in the spec-proto flow rule.

### Workflow rules

- Forge rule: Kubernetes cluster-profile mechanics (when a kind changes
  what the cluster itself is), single-node-topology sizing for profile
  clusters, and the verifier-owned-driver-CR pattern for behavioral proofs
  whose CR is served by the kind under test.
- Spec-proto flow rule: the `json_name` acronym-casing contract for CRD
  projection kinds.

## Validation

- Spec suites green for all three kinds (every CEL contract
  rejection-locked); per-kind + release-entrypoint builds; `make build-go`;
  secret-coverage, validate-refs, importmap + outputs conformance (three new
  outputs cases); kind map; e2e matrix; full-surface AND optionals-absent
  tofu plan proofs ×3 with type-fidelity spot-checks (`kubeProxyReplacement`
  renders as the chart's string contract; `group: ""` survives to the
  rendered CR); 10 presets + 13 scenario/fixture manifests CLI-validated.
- Live on kind, BOTH engines, full six-phase runner: Cilium 2×2 on the
  profile cluster; NetworkPolicy 5×2 incl. the behavioral-deny proof; KEDA
  2×2 incl. the behavioral scale proof; BackendTLSPolicy 3×2 incl. the
  composed Service-FK scenario. Blind import round-trips ×3 kinds. Zero
  orphans on both clusters.
- Not verified (recorded): Cilium kube-proxy replacement, encryption, cloud
  datapath arms, and CNI chaining live (need purpose-built or real-cloud
  clusters — offline plan proofs cover the rendering; batched real-cluster
  rows); KEDA cloud-identity arms live (same class); BackendTLSPolicy
  behavioral honoring (no in-catalog Gateway implementation programs it —
  istiod 1.30 supports it only in the experimental agentgateway controller,
  Cilium 1.19 lists it as unsupported).
