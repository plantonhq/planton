# Kubernetes observability tier live-proven: kube-prometheus-stack, Grafana, Loki and Tempo

## What changed

- **Four kinds proven against a live cluster, both engines** —
  KubernetesKubePrometheusStack, KubernetesGrafana, KubernetesLoki,
  KubernetesTempo. Every behavioral promise ran with verifier-output
  evidence: the stack answered PromQL over live-scraped metrics and
  delivered its always-firing Watchdog alert through Prometheus INTO
  Alertmanager (the full alerting pipeline, proven without waiting for a
  failure); a Grafana dashboard authored through the API survived a
  pod replacement through its PersistentVolume; a log line pushed
  through Loki's gateway was returned by LogQL after the Loki pod was
  killed and replaced; a span pushed over OTLP was retrieved by trace ID
  after the Tempo pod was killed and replaced. The stack's destroy
  asserts the monitoring.coreos.com CRDs SURVIVE uninstall — the
  crds-subchart keep posture is designed behavior, verified rather than
  tolerated. Blind import round-trips proved all four kinds' recipes
  (every resource re-imported blind; the follow-up plan proposed no real
  change). All four entered the green E2E CI matrix.

- **The Loki chunks-cache default footprint, taught where its readers
  live** (verified live): the chart's default chunks cache allocates
  8192MB and requests container memory at 1.2× — a 9830Mi request that
  never schedules on a node with less than ~10Gi allocatable, and
  because the install is atomic the WHOLE release rolls back, reading as
  "Loki won't install." The spec's caching comments, the kind docs,
  and the dev preset now teach the failure mode and the sizing recipe
  (`caching.chunks_cache_memory_mb`; 128–1024MB serves light query
  loads), and the kind-lane scenarios size the caches explicitly.

- **Loki's multi-tenant proof now runs AS a tenant**: with tenants
  declared, the chart guards every gateway route with basic auth and
  derives `X-Scope-OrgID` from the authenticated user — there is no
  unauthenticated path. The E2E verifier authenticates as the first
  declared tenant (plaintext held as a paired constant beside the
  verifier; the scenario's bcrypt hash is generated AND mechanically
  verified against it), so the lane proves htpasswd materialization,
  gateway auth, tenant-header injection, and tenant-scoped storage and
  query end to end. The multi-tenant scenario's original hash never
  matched its claimed plaintext — regenerated with the `htpasswd -vbB`
  verification now required by the forge workflow.

- **Tempo's verifier matches OTLP/protojson IDs in base64**: Tempo's
  trace-query API returns OTLP JSON, where trace/span IDs are protobuf
  bytes fields rendered BASE64 — a hex substring check holds the
  successful response and still reports failure. The verifier converts
  the pushed span ID to base64 for its success check (hex stays in the
  human-facing log lines), and the forge workflow teaches the class for
  every verifier asserting against protojson APIs.

- **Workflow lessons folded into the kind forge rule**: chart
  defaults can be unschedulable by RESOURCE REQUEST, not just topology —
  read every companion workload's default requests (including computed
  factors) at the pin, and treat a request exceeding a small node's
  allocatable as a product caveat (spec comment + dev preset + sized
  lane scenarios); one-way-hash credentials in E2E artifacts are
  verified pairings; protojson bytes IDs render base64.

## Verification

24/24 scenario-engine lanes green across the four kinds (three Loki and
three Tempo lanes re-run to an honest green after their fixes); 12/12
blind import round-trips; repo-wide `make build-go` green; spec tests
green; all edited manifests CLI-validated; zero orphans on the shared
cluster (stock namespaces, zero releases; kept CRDs per designed
posture).
