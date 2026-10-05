# Kubernetes metrics: kube-prometheus-stack + rebuilt standalone Grafana — two kinds at full depth; the Prometheus stub retired

## What changed

- **KubernetesKubePrometheusStack (870, new)** — deploys the
  kube-prometheus-stack, the industry-standard cluster monitoring
  bundle, from the official prometheus-community chart pinned 87.19.1
  (Prometheus Operator v0.92.1). Typed surface: the Prometheus server
  (HA replicas, time+size retention, PVC storage — persistent by
  default against the chart's emptyDir grain, with an explicit
  ephemeral arm; resources, external labels, intervals, scheduling),
  remote write with all four cloud auth arms (basic auth, bearer
  token, AWS SigV4 keyless-via-IRSA or static keys, Azure AD managed
  identity — passwords/tokens/keys ride existing Secrets; basic-auth
  usernames materialize in a module-owned `<name>-remote-write-auth`
  Secret because the Prometheus CRD wants both halves from Secrets),
  the remote-write receiver, the raw scrape-config seam, Alertmanager
  (gossip HA, retention, persistent state, the config document with
  the credentials-via-`_file` teaching), the bundled Grafana
  (default-on, chart-owned lookup-stable admin Secret or an existing
  one, dashboards, persistence), the operator (admission-webhook
  posture incl. the cert-manager arm), both exporters, control-plane
  scrape toggles carrying the managed-cloud truth (EKS/GKE/AKS hide
  the controller-manager/etcd/scheduler; kube-proxy binds localhost on
  several platforms and vanishes under Cilium KPR), curated-rule group
  disables paired with those toggles, registry/pull-secret overrides,
  and `helm_values` merged last. Monitor discovery is deliberately
  CLUSTER-WIDE by default (wider than the chart's release-fenced
  default — what makes every kind's `service_monitor_enabled`
  toggle light up with zero wiring), with `release_managed_only`
  restoring the fence. The CRDs ride the chart's crds SUBCHART
  (install-once, never chart-upgraded, KEPT on uninstall); `skip_crds`
  is the bring-your-own arm and `crd_upgrade_job` the SSA hook for
  operator-crossing upgrades. The modules pin the chart fullname to
  metadata.name and fail loudly past the chart's silent 26-character
  truncation budget.

- **KubernetesGrafana (871, rebuilt from the pre-program shell)** — the
  standalone composition hub on the official grafana chart pinned
  12.8.0 (Grafana 13.1.1) from grafana-community/helm-charts — the
  chart's CURRENT home; the old grafana.github.io index stalled at
  10.5.x, and kube-prometheus-stack's own dependency block is the
  authoritative tell. Typed surface: the three-arm state model
  (ephemeral default / `storage` PVC single-writer / external
  Postgres-MySQL `database` — the HA requirement, with CEL enforcing
  replicas>1 → database and storage ↔ single replica), chart-owned
  lookup-stable admin credentials or an existing Secret, datasources
  provisioned as code with the URL as a foreign key defaulting to a
  KubernetesKubePrometheusStack's Prometheus endpoint, the
  dashboard-discovery sidecar contract (cluster-wide
  `grafana_dashboard: "1"` ConfigMaps), community dashboards by
  grafana.com ID with pinnable revisions, plugins with the Grafana-13
  moved-out-of-core teaching and bundled-plugin shadowing, server
  root_url / anonymous-auth / SMTP, ServiceMonitor toggle, image
  override, scheduling, `helm_values` merged last. Every credential
  (admin, database, datasource basic-auth, SMTP) rides Secrets through
  environment expansion — the chart itself refuses secrets in its
  config ConfigMap and the design honors it. The embedded
  ingress from the pre-program shell is gone: exposure composes from
  first-class kinds over the exported service handle.

- **KubernetesPrometheus retired** — the raw-container Prometheus stub
  (a namespace-only module) is deleted; the operator-based stack is
  the 2026 way to run Prometheus on Kubernetes. Enum 870 reassigned to
  the stack kind; e2e entrypoints, Makefile tiers, kind map,
  containment golden and site catalog swept clean.

- **Forge rule: the chart image-value SHAPE teaching** — charts take
  the image reference either combined (`image.repository`) or SPLIT
  (`image.registry` + `image.repository`, registry defaulting to
  docker.io); a spec air-gap field carrying a full reference must be
  split by the module or every mirror override renders a
  docker.io-prefixed broken reference. The class is identically wrong
  in both engines, so parity review never catches it — only reading
  the pod template at the pin does. Caught in review on the Grafana
  modules, fixed in both engines, and landed as a permanent forge-rule
  teaching.

- **E2E surface (authored and compiled; live proving runs
  separately)**: dedicated verifiers — the stack (operator Deployment,
  operator-reconciled Prometheus/Alertmanager StatefulSets at their
  declared replica counts, bundled Grafana, a LIVE metric-flow proof
  through Prometheus' own API, the Watchdog dead-man's-switch alert
  asserted in BOTH Prometheus and Alertmanager APIs on the alerting
  scenario, and the crds-subchart keep posture asserted on destroy)
  and Grafana (admin-Secret read as the credential-wiring proof,
  /api/health, authenticated API round-trip, datasource provisioning,
  and a dashboard surviving pod replacement through the PVC on the
  persistence scenario) — plus six scenarios, import maps for both
  kinds, and outputs-conformance cases. Profiles ship as
  `pending_proof`; the CI matrix regen honestly DROPS the pre-rebuild
  Grafana green until the rebuilt surface is live-proven.

## Validation

Spec tests for both kinds (every CEL rule accept+reject locked, two
coverage gaps closed in review); offline `tofu` plan and
`pulumi preview` proofs across full-surface AND minimal shapes for all
four modules, re-run after the review fixes with the split image
rendering spot-checked in the plan output; secret-coverage, reference,
containment, import-map and outputs conformance gates;
repo-wide build; e2e-build/e2e-vet; license footers; all presets and
scenario manifests CLI-validated. Three independent line-level reviews
(both module pairs against the chart sources, and the full E2E/satellite
surface) caught and fixed in-session: the split-image-values class
above, a verifier hardcoding Alertmanager readiness to one replica, two
spec-test coverage gaps, two stale pre-rebuild engine READMEs, and
comment/idiom hygiene in both engines' locals. Known pre-existing
failure NOT from this change: the import-map conformance suite fails on
`awsecrrepo` (an AWS catalog row gap) at the prior HEAD as well.
