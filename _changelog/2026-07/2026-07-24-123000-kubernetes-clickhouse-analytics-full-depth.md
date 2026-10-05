# Kubernetes analytics database: Altinity operator + ClickHouse — two kinds rebuilt at full depth

## What changed

- **KubernetesAltinityOperator (918, rebuilt)** — installs the Altinity
  ClickHouse operator from its official Helm chart, pinned to 0.27.2
  (chart version = app version; the upstream repository publishes
  `release-X.Y.Z` tags only). The chart owns the CRD lifecycle — a
  `crds/` directory installs the four ClickHouse CRDs and a
  pre-install/pre-upgrade hook job keeps them current across chart
  upgrades; the hook's default image (`bitnami/kubectl:latest`, frozen
  upstream since the Bitnami catalog retirement) is exposed as a typed
  override rather than forked. Typed surface: watch-namespace regexps
  (empty = own-namespace-only — ClickHouseInstallations elsewhere would
  silently never reconcile, taught on the field), namespace-scoped RBAC,
  operator API credentials (unset = the chart's publicly documented
  defaults, flagged unsafe on the field), metrics exporter with
  ServiceMonitor, per-half image overrides, and scheduling. The chart
  fullname is pinned to the release name and re-pinned after the
  `helm_values` escape-hatch merge — every output handle derives from it.

- **KubernetesClickHouse (919, rebuilt)** — declares one ClickHouse
  installation as a `ClickHouseInstallation` custom resource at the full
  meaningful surface: shards-by-replicas topology with the operator's
  cluster-name budget mirrored in validation, a coordination contract
  (unset = a managed ClickHouse Keeper is deployed exactly when the
  topology needs one; explicit managed-Keeper, external-Keeper,
  external-ZooKeeper, and none arms) whose managed arm renders a
  `ClickHouseKeeperInstallation` and wires the CHI `keeper:` reference to
  the Keeper's client Service, users with Secret-only passwords
  (path-keyed settings, profiles, quotas, grants, network restrictions,
  access management), profiles/quotas/settings/configuration-files
  pass-throughs on the operator's own path vocabulary, version pinning,
  PodDisruptionBudget, the operator's `stop` pause switch, PVC-backed
  data and log volumes with Retain reclaim, resources, scheduling with
  anti-affinity arms, and service exposure. Pulumi renders through a
  typed SDK generated from the pinned CRDs, with a fail-loud adapter for
  the CRD's polymorphic string-boolean fields; Terraform through a
  hand-authored `kubectl_manifest` twin covering the CHI, the
  conditional Keeper, the auth Secret, and the anchor namespace — the
  import map is the catalog's first two-GVK composed-identity map with
  per-resource state-address scoping.

- **E2E surface (authored and compiled; live proving runs separately)**:
  dedicated verifiers — the operator install (Deployment Available, all
  four CRDs Established) and the installation (host readiness, a live
  SQL round-trip over the HTTP interface as a declared user, and a
  replica-loss durability arm that deletes a replica pod mid-proof and
  reads replicated rows back) — plus five scenarios across the two
  kinds, a consumer-scoped operator prerequisite watching all
  namespaces, and import maps for both. Profiles ship as
  `pending_proof`.

## Validation

Spec tests for both kinds (every CEL rule accept+reject locked); offline
`tofu` plan and `pulumi preview` proofs across full-surface and minimal
shapes for all four modules, including a hostile `helm_values`
fullname-hijack repel; secret-coverage, reference, containment,
import-map and outputs conformance gates; repo-wide Bazel build;
e2e-build/e2e-vet; license footers; all presets and scenario manifests
CLI-validated.
