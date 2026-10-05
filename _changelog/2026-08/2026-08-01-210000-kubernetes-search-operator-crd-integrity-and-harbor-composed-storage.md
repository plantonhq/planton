# Kubernetes search-operator CRD lifecycle hardened and Harbor's production-composed storage proven live

## What changed

- **The OpenSearch and Solr operator modules now re-adopt their
  retained CRDs and fail loudly when their staged CRD set is
  incomplete — re-proven live on both engines.** Both kinds keep their
  CRDs on uninstall by design (destroying an operator must never
  cascade-delete OpenSearchCluster/SolrCloud resources and their
  data), which means every REINSTALL finds the CRDs already on the
  cluster. On Pulumi, a plain create fails AlreadyExists against those
  retained objects at the pinned provider version; the CRDs now ride
  the dedicated upsert provider seam (server-side apply adopts, scoped
  to the CRDs only — the release keeps ordinary create-conflict
  semantics), and the Solr Terraform module gained the matching
  `force_conflicts` so SSA adoption never fails on a previous
  install's field manager. Both engines also gained exact-count
  fail-loud guards (ten OpenSearch CRDs, four Solr CRD documents): a
  module deployed without its staged `../crds` files previously
  planned ZERO CRDs silently and ran against whatever CRDs happened to
  exist. All four scenario-engine lane pairs re-ran green with blind
  import round-trips — the Terraform plans now stage the full CRD
  sets, and every retained CRD was adopted in place (creation
  timestamps unchanged across eight installs).

- **Harbor's production-composed shape is now live-proven: S3 object
  storage on an in-cluster SeaweedFS, an external Valkey cache, an
  external PostgreSQL, and a TWO-replica registry sharing the object
  store.** A new composed-storage E2E scenario chains four fixtures
  (the CNPG operator → PostgreSQL, an S3-identity Secret, SeaweedFS,
  Valkey) and runs the full registry product proof through the
  composed stack on both engines: admin login from the generated
  Secret, a PRIVATE project, OCI push/pull digest-verified through the
  scaled registry, the 401 API auth gate, and the anonymous pull of
  the private artifact rejected by the registry itself. The
  multi-replica registry is legal exactly because the backend is
  object storage (the filesystem-RWO validation fence) — this lane
  proves that production scaling shape end to end.

- **The declare-on-both-sides composition recipes are now live-proven,
  and the conditional credential-Secret imports ran for the first
  time.** The SeaweedFS fixture hands identity ownership to a
  KubernetesSecret carrying the chart's `seaweedfs_s3_config`
  identities document (mirroring the chart-generated Secret's exact
  key shape, so every consumer of the generated contract — the S3
  verifier included — works unchanged), and the Harbor scenario
  declares the same admin key pair on its storage arm; the Valkey
  fixture declares a known password for its default ACL user and the
  cache arm declares the same value. The blind Terraform round-trip
  therefore imported the module-materialized `redis_auth` and
  `storage_auth` Secrets live — previously offline-validated only.

- **The import-map live-proven ledger is complete again.** Twenty-four
  rows were reconciled from lane records into
  `pkg/iac/importmap/README.md`: the data-platform kinds
  (Valkey, the Percona operators, MySQL, MongoDB), the Kafka ecosystem
  (Connect, Connector, MirrorMaker2, Karapace, KafkaUI), the search
  wave (OpenSearch, Solr, Neo4j), the analytics pair
  (AltinityOperator, ClickHouse), the object/vector pair (SeaweedFS,
  Qdrant), the messaging pair (RabbitMQ operator + cluster), SigNoz,
  and fresh same-day rows for the two search operators' re-proven
  round-trips. The Harbor row now records the composed-storage lane's
  live credential imports.

- **Harbor's E2E budget reflects the composed lane honestly**: the
  kind's timeout rose 30 → 40 minutes (a four-fixture chain deploys
  and tears down per lane), which raises the Tier-2 matrix ceilings
  through the generator's max-times-count formula.

## Why

A kind marked proven must be proven for the path a customer
actually walks: reinstalling an operator over its kept CRDs is the
normal day-2 path, and a registry backed by object storage with an
external cache is Harbor's production posture. Both are now evidence,
not inference.

## Impact

- Reinstalling KubernetesOpenSearchOperator or KubernetesSolrOperator
  over retained CRDs now succeeds on both engines; a partial module
  deployment fails at plan/preview with an instruction instead of
  silently running against stale CRDs.
- The Harbor catalog page's production-composed preset shape is
  live-proven end to end, including the credential handoffs.
- No API changes; no spec changes; existing manifests are unaffected.
