# Kubernetes ingress and metrics rebuilt: ingress-nginx at full depth, metrics-server forged, HPA scaling proven live

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetesingressnginx rebuilt; kubernetesmetricsserver forged), `catalogkind` (KubernetesMetricsServer = 857), `aa_e2e/verify` (ingress-nginx + metrics-server install verifiers, HPA behavioral verifier, manifest helpers), `kuberneteshorizontalpodautoscaler` (behavioral-scaling scenario + CPU-burner fixture), `e2e` + Makefile Tier-1, `pkg/outputs`, `pkg/iac/importmap` (two proven maps + ledger), site catalog, `_rules/catalog-kind` (forge + update lessons)

## What changed

The cluster's HTTP(S) entry point and its resource-metrics pipeline — the
two foundation addons that make traffic routing and autoscaling real —
built to full configuration depth with dual-engine parity and live
kind-cluster E2E, including the program's first live proof that a
HorizontalPodAutoscaler actually scales.

### KubernetesIngressNginx (rebuilt)

- **The Terraform module installed a chart that does not exist.** It pinned
  `kubernetes-ingress-nginx` while the repository serves the chart as
  `ingress-nginx` (the Pulumi module was correct) — every Terraform deploy
  of this kind failed at install. Fixed; both engines now carry the chart
  identity with a parity comment naming the hazard.
- **Multi-instance per cluster is first-class**: the Helm release and chart
  fullname derive from `metadata.name`, each instance owns a distinct
  IngressClass (controller identifier auto-derived as
  `k8s.io/<class-name>`), and leader election isolates per instance. The
  public + internal traffic split is two resources, not a contortion.
- **No provider oneof, no workload identity — by design.** The controller
  never calls cloud APIs; the host cloud shapes the load balancer entirely
  through the controller Service's annotations. The old coupled GKE/EKS/AKS
  arms are gone; the typed Service surface (type, annotations, external
  traffic policy, source ranges, LB class, node ports, the chart's internal
  dual-LB Service) plus documented per-cloud annotation recipes replace
  them.
- **Typed chart surface at full depth**: IngressClass identity (name,
  cluster-default, controller value, watch-without-class), replicas vs
  chart-managed HPA autoscaling (PDB auto-guards at >1 replica), NGINX
  ConfigMap tuning in upstream's own key vocabulary, snippet-annotation
  gate (off per upstream's post-CVE default), cluster-wide default TLS
  certificate (foreign key to a Certificate's secret output — the
  cert-manager seam, rendered as `--default-ssl-certificate`), default
  backend, admission-webhook tuning, metrics + ServiceMonitor, TCP/UDP
  service exposure, DaemonSet/hostNetwork/hostPort bare-metal postures
  (host_network XOR host_ports enforced), scheduling, image registry, and
  the `helm_values` escape hatch merged last with Helm `-f` semantics on
  both engines.
- **Outputs are the composition handles**: ingress class name (what
  KubernetesIngress references), controller Service names, and the LB
  address pair (IP and hostname — providers populate one or the other),
  read back gated on the load_balancer service type. On clusters without a
  cloud LB controller a LoadBalancer install fails loudly at the readiness
  wait instead of leaving a silently Pending entry point; kind lanes run
  node_port by design.
- Pin: chart 4.15.1 (controller v1.15.1), verified against the chart-repo
  index.

### KubernetesMetricsServer (new, enum 857)

- Installs metrics-server — the kubelet-scraping pipeline behind
  `kubectl top` and every Resource-type HPA — from the official chart
  (3.13.1, app 0.8.1, verified against the chart-repo index). One
  installation per cluster (the `v1beta1.metrics.k8s.io` APIService is a
  singleton); the release name is fixed accordingly.
- **The one knob that matters is first-class**: `kubelet_insecure_tls` for
  clusters whose kubelets serve self-signed certificates (kind, k3s,
  kubeadm, on-prem). Both engines wait on readiness, and the `/readyz`
  probe passes only after the first successful kubelet scrape — a wrong
  TLS posture fails the deploy loudly instead of surfacing later as HPAs
  that never scale.
- Full typed surface: kubelet scrape flags (the module owns the chart's
  `defaultArgs` and re-renders it with typed substitutions so the pod spec
  stays canonical), APIService registration (create / skip-TLS-verify /
  CA bundle), serving-certificate provisioning (self-signed, helm,
  cert-manager with an existing Issuer/ClusterIssuer reference,
  existing-secret), HA replicas + PodDisruptionBudget, host networking,
  own-telemetry metrics + ServiceMonitor, scheduling, image override, and
  the `helm_values` escape hatch.

### HPA behavioral scaling — proven live

- The HorizontalPodAutoscaler kind gains a behavioral-scaling scenario:
  metrics-server installs as a fixture (the scenario-level prerequisites
  annotation), a CPU-burning target Deployment deploys as an
  extra-instance fixture, and the verifier asserts ScalingActive AND an
  actual scale-up (desiredReplicas above min) — real metric flow through
  kubelet → metrics-server → metrics API → HPA controller, on both
  engines. The metrics-server verifier itself checks all the way to
  `kubectl top nodes` returning values.

## Validation

- Spec tests green for both kinds (every CEL contract rejection-locked);
  per-kind and release-entrypoint builds; `make build-go`; secret-coverage
  and refcheck suites; outputs conformance (+2 cases); import-map
  conformance; kind map; e2e matrix; site catalog regenerated.
- Offline `tofu` plan proofs: full-surface AND optionals-absent per kind;
  rendered chart values spot-checked for numeric/boolean type fidelity.
- Live on the kind cluster, BOTH engines, full six-phase runner:
  ingress-nginx minimal / tuned-full / multi-instance (the multi-instance
  scenario deploys a second controller against a live sibling), 3×2 lanes;
  metrics-server minimal / full-surface, 2×2 lanes; HPA all four scenarios
  including behavioral-scaling, ×2 engines. Blind import round-trips green
  for both new maps (ingress-nginx ×3 scenarios, metrics-server ×2). Zero
  orphaned namespaces or releases at session end.
- Deferred with reasons: real cloud-LB provisioning per annotation recipe
  (needs a cloud cluster; the LB arm is offline-proven).

## Workflow lessons folded into the rules

- Forge rule: kind-cluster lanes must serialize when one lane's kind
  is another's fixture or a cluster singleton; a lane killed mid-install
  orphans a `pending-install` Helm release that blocks later installs of
  the same name — sweep before re-running.
- Update rule: a chained per-arm ternary selecting between differently
  shaped objects is the same plan-time type-unification failure as the
  two-branch form — render one object literal with per-arm gated keys and
  null-prune it.
