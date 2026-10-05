# Kubernetes data stores rebuilt at full depth: Redis renamed to Valkey on the official chart, MySQL forged on Percona XtraDB Cluster, MongoDB rebuilt on the Percona operator — five kinds with live durability proofs on both engines

**Date**: 2026-07-23
**Scope**: `apis/dev/planton/shared/catalogkind` (KubernetesRedis renamed → KubernetesValkey = 902; KubernetesMysql = 904 forged with prerequisite KubernetesPerconaMysqlOperator = 903; KubernetesPerconaMongoOperator = 905 + KubernetesMongodb = 906 renumbered into family adjacency), `apis/dev/planton/provider/kubernetes` (kubernetesvalkey renamed + rebuilt; kubernetesmysql forged; kubernetesmongodb rebuilt end to end; both Percona operator kinds rebuilt), `pkg/kubernetes/kubernetestypes` (stale perconamysql PS-operator set removed; perconamongodb + perconaxtradb generation retired — the CRDs' map-of-nested-objects fields defeat crd2pulumi, so both database modules render their CRs untyped), `aa_e2e/verify` (Valkey, PXC cluster, PSMDB cluster, and both operator install verifiers incl. behavioral durability, failover, and backup proofs), `e2e` (ten entrypoints across the five kinds), `pkg/outputs` (+5 conformance cases), `pkg/iac/importmap` (+5 proven round-trips), `aa_import/catalog.yaml` (ValidatingWebhookConfiguration row), site catalog, `_rules/catalog-kind` (forge + update rules: crd2pulumi map-of-objects limitation; operator-runtime cluster-singleton ownership)

## What changed

### KubernetesValkey (902, renamed from KubernetesRedis)

- The kind now deploys Valkey — the Linux Foundation's Redis-compatible
  store — from the OFFICIAL `valkey` chart (0.11.0 = Valkey 9.1.1 at
  https://valkey.io/valkey-helm/), replacing the retired Bitnami Redis
  path. Release name and chart fullname pin to `metadata.name`.
- Deep typed surface: standalone vs primary/replica replication (read
  Service, write-safety bounds, diskless sync), module-owned `valkey.conf`
  rendering (append-only, RDB save points, maxmemory + eviction policy),
  ACL users secret-by-default (`<name>-auth`, one key per username — the
  chart's usersExistingSecret contract; auth requires the `default` user),
  TLS via existing Secret (cert-manager seam), metrics + ServiceMonitor,
  scheduling, PDB (replication mode), `helm_values` escape hatch.
- Sentinel HA and Cluster mode are NOT modeled — upstream chart
  milestones; documented as deliberate exclusions with the honest
  no-automated-failover story (durability through AOF persistence).

### KubernetesPerconaMysqlOperator (903, rebuilt)

- Rebuilt on the official `pxc-operator` chart 1.20.0 (Percona Operator
  for MySQL based on Percona XtraDB Cluster), release named
  `metadata.name`; full typed surface (watch scoping, reconcile/backup
  concurrency, structured logging, telemetry opt-out, leader election,
  the XtraBackup-sidecar gate, sizing, image override, pull secrets).
- **The operator's cluster-scoped validation webhook is MODULE-OWNED in
  the widened-watch arms.** Upstream registers one fixed-name Fail-closed
  ValidatingWebhookConfiguration at startup and nothing ever removes it
  (Kubernetes never garbage-collects a cluster-scoped dependent of a
  namespaced owner; a later operator refreshes only the CA bundle, never
  the service pointer) — uninstall used to strand a webhook that bricked
  every future PerconaXtraDBCluster admission in the cluster. The modules
  now render the object first (the operator adopts it and stamps its CA
  bundle), delete it with the resource, and ignore the operator-managed
  fields; the Pulumi side adopts pre-existing instances via
  `pulumi.com/patchForce`. The E2E verifier asserts presence (pointing at
  the right namespace) while deployed and absence after destroy.

### KubernetesMysql (904, new)

- One Percona XtraDB Cluster (Galera synchronous replication) declared as
  a `pxc.percona.com/v1 PerconaXtraDBCluster` CR: 3-node quorum default
  (1-node dev requires `unsafe.cluster_size`), HAProxy-or-ProxySQL proxy
  oneof (write port 3306, HAProxy replicas read Service 3307), TLS on by
  default with the cert-manager issuer seam (disabling requires
  `unsafe.tls`), declarative users (`<name>-user-<u>` password Secrets),
  XtraBackup storages (S3/S3-compatible/Azure/PVC) + typed schedules with
  count-based retention + PITR, scheduling/PDB/log-collector/SmartUpdate,
  pause, image-based version selection.
- Terraform renders the CR via `kubectl_manifest` with null-prune locals;
  Pulumi renders the IDENTICAL body via `apiextensions.CustomResource`
  untyped args — deliberately: crd2pulumi flattens the CRD's name-keyed
  backup-storages map (nested s3/azure objects) into a flat string map
  that structurally cannot carry it.
- Root credential handle exported from the operator-managed
  `<name>-secrets` Secret (key `root`); no embedded ingress — exposure
  composes via the proxy Services.

### KubernetesPerconaMongoOperator (905, rebuilt)

- Rebuilt on the official `psmdb-operator` chart 1.22.0, release named
  `metadata.name`; typed surface mirrors the MySQL operator minus the
  chart values that don't exist there (correctly differentiated — e.g.
  `maxConcurrentReconciles` renders as the string this chart declares).

### KubernetesMongodb (906, rebuilt)

- Total redesign from the old 4-field container shape onto the
  `psmdb.percona.com/v1 PerconaServerMongoDB` CR (operator 1.22.0):
  replica sets (per-set mongod config, storage, arbiters, expose, PDB,
  scheduling), the sharding arm (config servers + mongos + balancer), PBM
  backups (S3/S3-compatible/GCS/Azure storages with per-arm credential
  Secrets, typed tasks with retention, PITR with oplog cadence), TLS
  modes with the cert-manager issuer seam, declarative users with roles,
  unsafe opt-ins, pause. Same untyped-CR Pulumi posture as MySQL, same
  reason.
- Outputs carry the driver-ready handles: the discovery Service (mongos
  when sharded, the replica set's headless Service otherwise), the
  `replicaSet` parameter, and the admin credential handle in the
  operator's `<name>-secrets` Secret.

### Shared-ownership disciplines (both database kinds)

- Declared user-password Secrets are CO-OWNED with the operator: it
  stamps a `percona.com/<cluster>-<user>-hash` rotation marker onto them.
  Both engines ignore annotation drift on exactly those Secrets (and
  nothing else), keeping blind import round-trips exact.
- Every CEL guardrail the specs promise is enforced and
  rejection-locked: Galera quorum / replica-set size vs `unsafe`,
  disabled TLS vs `unsafe.tls`, proxy sizing vs `unsafe.proxy_size`,
  exactly-one-main among multiple backup storages, required proxy oneof.
  The one upstream flag that is declared but unenforced in the operator
  source (`mongosSize`) is documented as such instead of faked.

## Proven live (persistent kind cluster, both engines)

- Operators 2×2 each, incl. the webhook lifecycle proof (present while
  deployed, gone after destroy).
- Valkey 3×2 incl. the persistence-durability proof (marker write with
  AOF → pod deleted → marker intact) and the replication proof (write via
  the primary → read via the read Service).
- MySQL 3×2 incl. the Galera durability proof (marker written through
  HAProxy → database node deleted → marker served during the outage →
  node rejoins) and a REAL XtraBackup to in-cluster MinIO.
- MongoDB 3×2 incl. the failover durability proof (majority write →
  primary deleted → survivors elect → marker read on the new primary) and
  a REAL PBM backup to in-cluster MinIO.
- Blind import round-trips proven for all five kinds (13 lanes), incl.
  the webhook's scoped recipe and the co-owned-Secret tolerance.

## Notes for module authors

- The kind-lane scenarios disable the operators' default hostname
  anti-affinity with the operators' own `"none"` switch — a 3-node
  database can never schedule on a single-node cluster otherwise; node
  spread is real-cluster material.
- `pkg/kubernetes/kubernetestypes` no longer generates the PXC/PSMDB type
  sets: the CRDs' map-of-nested-objects fields defeat crd2pulumi (the
  generated type panics at runtime), so both database modules render
  their CRs untyped — the reason is recorded in the kubernetestypes
  Makefile and both cluster builders.
