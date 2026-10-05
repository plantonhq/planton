# Kubernetes Postgres rebuilt on CloudNativePG at full depth: the operator kind forged, plugin-based backups, a live failover data-durability proof, and the legacy Postgres operators retired

**Date**: 2026-07-23
**Scope**: `apis/dev/planton/shared/catalogkind` (KubernetesCloudNativePgOperator = 900 forged; KubernetesPostgres renumbered 902 → 901 with its prerequisite flipped; KubernetesZalandoPostgresOperator + KubernetesPerconaPostgresOperator retired), `apis/dev/planton/provider/kubernetes` (kubernetescloudnativepgoperator forged; kubernetespostgres rebuilt end to end; two kind folders removed), `pkg/kubernetes/kubernetestypes` (cloudnativepg generation set: Cluster + ScheduledBackup + ObjectStore typed Pulumi SDK; zalandooperator + perconapostgres sets removed), `aa_e2e/verify` (CloudNativePG operator + cluster verifiers incl. the behavioral failover and backup proofs), `e2e` (four new entrypoints; six retired), `pkg/outputs` (Postgres conformance case rebuilt + operator case added), `pkg/iac/importmap` (+2 proven round-trips ledgered; count-indexed-singleton guidance), site catalog, `_rules/catalog-kind` (update rule: deprecation-fork modeling; from_address_key vs count-indexed singletons)

## What changed

### KubernetesCloudNativePgOperator (900, new)

- Installs CloudNativePG — the CNCF PostgreSQL operator — from the official
  Helm chart (`cloudnative-pg` 0.29.0 = operator 1.30.0 at
  https://cloudnative-pg.github.io/charts, chart-index-verified). Fixed
  release name `cnpg`: the operator's CRDs are cluster-scoped and its
  webhook service name is baked into the webhook certificate — one
  installation per cluster.
- Typed surface: CRD lifecycle (the chart stamps
  `helm.sh/resource-policy: keep` on every CRD unconditionally — uninstall
  never cascade-deletes databases), replicas/resources, watch scoping
  (cluster-wide or namespace-fenced, with the typed field owning the
  WATCH_NAMESPACE config key), operator-config map, reconcile concurrency,
  monitoring (PodMonitor + Grafana dashboard), scheduling, image override,
  and a `helm_values` escape hatch merged last with Helm `-f` semantics.
- **The Barman Cloud plugin as a typed arm**: CloudNativePG's built-in
  object-store backup support is deprecated upstream (removal announced),
  so backups run through the Barman Cloud CNPG-I plugin — installed here as
  a SECOND fixed-name Helm release (`plugin-barman-cloud` 0.7.0 = v0.13.0)
  beside the operator (upstream forbids folding the two into one release).
  The plugin's internal TLS is issued by cert-manager; the dependency is
  documented, never hidden.

### KubernetesPostgres (901, rebuilt)

- The spec now declares a CloudNativePG-managed PostgreSQL cluster: the
  module renders a `postgresql.cnpg.io/v1` Cluster resource plus — when
  backups are declared — the Barman Cloud `ObjectStore` and per-schedule
  `ScheduledBackup` companions, all from one naming root (metadata.name;
  services `<name>-rw/-ro/-r`, credential Secrets `<name>-app` /
  `<name>-superuser`).
- Full surface: instances with streaming replication and automated
  failover; storage + optional dedicated WAL volume (StorageClass FKs);
  postgresql.conf parameters, pg_hba/pg_ident, preload libraries,
  SYNCHRONOUS replication (quorum/priority + data-durability posture);
  bootstrap arms — fresh initdb (with owner credentials, locales,
  checksums, post-init SQL, and LOGICAL IMPORT from an existing server —
  the RDS/Cloud SQL migration path), recovery from an object-store backup
  (with PITR targets), and pg_basebackup streaming; external-cluster
  connection descriptors with declared passwords as Secrets; declarative
  managed roles; superuser posture; plugin-based backups with
  S3/S3-compatible (MinIO, R2)/GCS/Azure-Blob arms in keyless
  (workload-identity) or declared-key postures, WAL/base-backup tuning,
  retention, and typed schedules (six-field crons); TLS with the
  cert-manager seam (server certificate FK); metrics exporter tuning;
  scheduling incl. anti-affinity strength; rolling-update strategy.
- The embedded ingress block and its LoadBalancer Service are GONE:
  exposure composes from first-class kinds against the exported service
  handles — consistent with every workload kind in the catalog.
- Engines: Pulumi renders through the typed crd2pulumi CloudNativePG SDK
  (compile-time schema fidelity); Terraform is a hand-authored
  15-resource `kubectl_manifest` twin on the null-prune idiom with
  deterministic satellite names. Declared credentials only ever travel as
  Kubernetes Secrets referenced by `secretKeyRef` — never plaintext in a
  rendered custom resource.

### Retirements

- `KubernetesZalandoPostgresOperator` and `KubernetesPerconaPostgresOperator`
  are removed: CloudNativePG is the catalog's one first-class Postgres
  path; three overlapping operator kinds served nobody. Their enum slots
  were reassigned within the data-platforms band (operator-then-workload
  teaching order), their crd2pulumi type sets and E2E/verifier/matrix/site
  registrations removed.

## Proven live (kind cluster, both engines, six-phase runner)

- Operator: minimal + plugin scenarios (the plugin scenario chains the
  cert-manager fixture) — 2×2 green; blind import round-trip green
  (two tofu_resource_name-scoped Helm releases + namespace).
- Postgres: minimal, behavioral-failover, and with-backup scenarios — 3×2
  green. The **failover data-durability proof** writes a marker row through
  the primary under synchronous replication, DELETES the primary pod, waits
  for the operator to promote the replica, and reads the marker back
  through the new primary. The **backup proof** drives a real base backup
  to Completed against an in-cluster MinIO store through the Barman Cloud
  plugin. Blind import round-trips green for all three scenarios — the
  with-backup lane's 10-resource re-import (Cluster + ObjectStore +
  ScheduledBackup + credential/region Secrets + namespace) is the largest
  kubectl_manifest family proven to date.
- Zero orphaned resources after every lane (CRDs survive by the designed
  keep posture).

## Workflow guidance updated

- Update rule: read upstream DEPRECATION notices at every pin reconcile and
  model the REPLACEMENT architecture, never a surface with an announced
  removal (CloudNativePG's in-tree backup stanza vs the plugin is the
  worked example); `from_address_key` only serves `for_each` resources —
  count-indexed singleton satellites need scoped `from_metadata_name_suffix`
  declarations (the failure is live-only; the round-trip lane caught it).
