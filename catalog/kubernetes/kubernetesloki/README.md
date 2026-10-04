# Kubernetes Loki

## When NOT to Use This

**One resource is ONE Grafana Loki install** — a log-aggregation backend
that indexes log *labels* and stores compressed chunks in object storage.

Not the right kind when:

- **You need something to SHIP the logs** — Loki stores logs, it does not
  collect them. Deploy a `KubernetesOtelCollector` in daemonset mode (its
  cluster-logs pipeline) pointed at this component's exported
  `gateway_endpoint`, or push from any OTLP/Loki-HTTP client.
- **You want to READ logs in a UI** — that is Grafana. Point a
  `KubernetesGrafana` datasource of type `loki` at the same
  `gateway_endpoint`.
- **You expect a public endpoint out of the box** — everything is
  ClusterIP behind the nginx gateway; exposure composes from first-class
  kinds (`KubernetesIngress`, the Gateway API kinds) over the exported
  handles.

## Deployment modes

- **Monolithic** (default) — every Loki target in one StatefulSet. Right
  for single-node clusters, dev, and small production volumes. A single
  replica may run on a filesystem volume; more than one replica requires
  object storage.
- **Simple-scalable** — write/read/backend tiers that scale
  independently. REQUIRES an object-storage backend.

The chart's microservices ("Distributed") mode and its transitional
migration modes are deliberately not modeled — a deployment that needs
per-component microservices deserves a dedicated operations posture.

## Storage

`filesystem` (a PersistentVolume, single monolithic replica only) or an
object store — `s3` (including S3-compatible endpoints like an in-cluster
`KubernetesSeaweedFs`), `r2` (a Cloudflare R2 bucket by reference, with the
key pair in the module's own `<name>-r2-credentials` Secret), `gcs`, or
`azure`. Credentials are always
references to existing Secrets and ride environment-variable expansion —
they never land in the rendered Loki config. Leaving credentials empty
uses the pod's ambient cloud identity (IRSA / GKE WI / AKS federated
token).

## Schema and retention

Loki normally makes every user hand-author a `schema_config`; this
component derives it (TSDB, schema v13). `retention_period` enables the
compactor's deletion (a multiple of 24h). `schema_from_date` exists only
for importing an existing cluster.

## Tenancy

Single-tenant by default — no `X-Scope-OrgID` header needed, so a composed
Grafana datasource and collector pipeline work with one line of wiring.
Enable `multi_tenancy` for gateway-enforced basic-auth tenants (bcrypt
password hashes; the plaintext never appears).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
