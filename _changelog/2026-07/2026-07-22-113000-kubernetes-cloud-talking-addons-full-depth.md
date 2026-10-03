# Kubernetes cloud-talking addons rebuilt: ExternalDNS and the External Secrets family at full depth on both engines

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetesexternaldns rebuilt; kubernetesexternalsecrets renamed to kubernetesexternalsecretsoperator and rebuilt; kubernetesclustersecretstore, kubernetessecretstore, kubernetesexternalsecret forged; new shared `external_secrets_store.proto`), `catalogkind` (foundation-addons band widened to 830–869, three family-adjacent inserts, observability/security kinds renumbered), `aa_e2e/verify` (external-dns + external-secrets verifiers), `e2e` + Makefile Tier-1, `pkg/outputs`, `pkg/iac/importmap` (two proven maps + ledger), `pkg/iac/pulumi/pulumimodule/provider/kubernetes/externalsecretsstore` (new shared builder), `pkg/kubernetes/kubernetestypes` (external-secrets pin v0.9.20 → v2.8.0, types regenerated at `external-secrets.io/v1`), site catalog, `_rules/catalog-kind` (forge + update lessons)

## What changed

The two addon families that reach OUT of the cluster to other clouds —
ExternalDNS (publishes DNS records for Services/Ingresses/routes into a DNS
provider) and the External Secrets family (syncs secrets from external
stores into Kubernetes Secrets) — rebuilt from the pinned upstreams to full
configuration depth, with dual-engine parity and live kind-cluster E2E
including a real end-to-end secret sync.

### KubernetesExternalDns (rebuilt — host and DNS provider fully decoupled)

- **The DNS provider is no longer welded to the host cloud.** The old spec's
  arms conflated "runs on EKS" with "publishes to Route 53". The rebuilt
  spec separates them: a typed `dns_provider` oneof (AWS Route 53, Google
  Cloud DNS, Azure DNS public/private, Cloudflare, the upstream webhook
  extension arm, and the in-memory sandbox arm) plus the shared
  workload-identity proto for keyless auth — so EKS + Cloudflare or GKE +
  Route 53 are first-class, not contortions.
- **Typed chart surface at full depth**: sources (validated against the
  upstream source list), sync policy, the TXT/DynamoDB ownership registry
  (owner id, prefix XOR suffix, table settings), domain/zone filtering with
  per-arm zone references (AwsRoute53Zone / GcpDnsZone / AzureDnsZone /
  CloudflareDnsZone foreign keys), annotation/label filters, managed record
  types, interval + event triggering, namespaced mode, logging, sizing,
  scheduling, ServiceMonitor, image override, and the `helm_values` escape
  hatch merged LAST with Helm `-f` semantics on both engines.
- **Declared credentials materialize as Secrets**: Cloudflare tokens, AWS
  static keys, GCP keys, and the rendered Azure `azure.json` land in
  deterministically-named Secrets wired into the controller via env/volume
  references — never in chart values or pod specs.
- **Multi-instance by design**: the release (and the chart fullname) is
  pinned to `metadata.name` — one instance per DNS provider/zone set,
  separated by TXT owner IDs, is the upstream pattern and now the module's.
- Chart pinned to what the repository actually serves (`external-dns`
  1.21.1, controller v0.21.0) — the vendored Chart.yaml at the app tag lags
  the served chart, a trap now taught by the update rule.

### KubernetesExternalSecretsOperator (renamed from KubernetesExternalSecrets, rebuilt)

- **Renamed**: the kind installs the External Secrets OPERATOR — upstream's
  own product name — and the old name sat one letter away from the new
  ExternalSecret kind, a typo-distance trap. Enum number (835) and
  id_prefix kept.
- **Pin moved v0.9.20 → v2.8.0** (the CRDs now serve `external-secrets.io/v1`);
  crd2pulumi types regenerated; the chart repo still serves from
  https://charts.external-secrets.io.
- **Typed spec at the operator's real surface**: CRD lifecycle — including
  keep-on-uninstall implemented via the `helm.sh/resource-policy: keep`
  annotation the chart forwards onto CRDs (the chart itself has no keep
  knob and would cascade-delete every ESO object on uninstall) — HA with
  enforced leader election, reconcile concurrency, controller-class
  sharding, namespace scoping with scoped RBAC, per-kind tuning
  (webhook, cert-controller), ambient workload identity for the controller
  ServiceAccount, scheduling, PDB, ServiceMonitor, image override,
  `helm_values` escape hatch.
- The stale copy-pasted outputs (Redis/GitLab-era `port_forward_command`,
  `ingress_endpoint`) replaced with the real contract: namespace, release
  name (fixed to `external-secrets` — one installation per cluster is an
  upstream constraint), controller ServiceAccount.

### KubernetesClusterSecretStore + KubernetesSecretStore (forged on one shared config)

- **Upstream gives the two store kinds an IDENTICAL spec; the forge makes
  that structural**: one shared `ExternalSecretsStoreConfig` proto, one
  shared Pulumi builder, twin Terraform locals — the cluster and namespaced
  grains cannot drift.
- **Typed backend oneof**: AWS Secrets Manager / Parameter Store /
  Certificate Manager, GCP Secret Manager, Azure Key Vault, Vault/OpenBao
  (token, AppRole, Kubernetes auth), remote-Kubernetes, and the fake
  sandbox backend. Uniform authentication model per backend: keyless via a
  referenced ServiceAccount (foreign key to KubernetesServiceAccount, whose
  workload-identity arms carry the cloud binding), the operator's ambient
  identity when auth is empty, or declared static credentials materialized
  as one deterministic `<name>-credentials` Secret the CR references.
- **The cluster kind adds the multi-tenancy fence**: namespace conditions
  (names, label selector, regexes) controlling which namespaces may sync
  from the store.
- Terraform applies the CRs through `kubectl_manifest` (plannable before
  the operator exists — single-run infra charts and offline proofs); the
  null-prune rendering idiom keeps numbers and booleans typed.

### KubernetesExternalSecret (forged — the sync declaration)

- **The complete `external-secrets.io/v1` request surface**: store
  reference (foreign key to either store kind), refresh interval/policy,
  the target Secret with creation/deletion policies, immutability, and a
  template (type, merge policy, metadata, Go-template data — e.g. assemble
  a connection string from synced parts), explicit `data` entries
  (property/version/decoding per entry) and bulk `dataFrom` pulls (extract
  a structured document or find by name pattern/tags, with regexp key
  rewriting).
- Outputs export the materialized Secret's name — the handle workloads
  wire env/volume references to.

### Registry: foundation-addons band widened, ESO family adjacent

The 830–859 foundation-addons sub-band was exactly full once the remaining
addon forges were counted, so it widened to 830–869: observability moved to
870–889 (Prometheus, Grafana, Signoz renumbered) and security to 890–899
(Keycloak, OpenBao, OpenFga renumbered), keeping every band decade-aligned
with everything from 900 up untouched. The three new ESO-family kinds sit
family-adjacent at 836–838; ingress-nginx through EnvoyFilter shifted by
three. Names and id_prefixes everywhere unchanged; the kind map and E2E
matrix regenerated.

## Validation

- Five spec-test suites green (every CEL contract locked by a rejection
  case); per-kind + release-entrypoint builds; `make build-go`;
  secret-coverage; validate-refs; outputs conformance (+5 cases, one stale
  case retired); import-map conformance (the pre-existing aws/awsecrrepo
  failure remains the AWS program's, untouched); kind map; e2e matrix; site
  catalog regenerated (five pages). 18 presets machine-validated; every
  scenario/prerequisite/hack manifest CLI-validated.
- Offline engine proofs: `tofu validate` ×5; full-surface AND
  optionals-absent `tofu plan` proofs ×5 kinds (numeric/bool type fidelity
  spot-checked in the rendered plans).
- Live on the kind cluster, BOTH engines, full six-phase runner: 18
  scenario-engine lanes green — ExternalDNS (in-memory minimal + tuned,
  AWS keyless soft-fail posture) ×2 engines; operator (minimal + tuned) ×2;
  cluster store + namespaced store over the fake backend (Ready condition
  verified) ×2; ExternalSecret minimal + full-surface ×2 — the full-surface
  lane proves the REAL sync loop end-to-end: cluster-store fixture chain,
  JSON extraction, key rewriting, and templating, all asserted against the
  materialized Secret's contents. Zero orphans.
- Blind import round-trips green: kubernetesexternaldns (3 scenarios),
  kubernetesexternalsecretsoperator (2 scenarios); the three CR kinds'
  maps deliberately deferred pending the kubectl_manifest composed-ID
  catalog uplift (ledgered).
- NOT verified (recorded): real DNS record writes and real cloud secret
  reads (need cloud accounts — batched real-cluster lanes); the webhook
  provider arm live (no public webhook image fits a kind lane; offline
  proofs cover the rendering).

## Workflow lessons folded into the rules

- Update rule: pin Helm charts against the chart repository INDEX — the
  vendored Chart.yaml at an app tag routinely lags the served chart.
- Forge rule: model the upstream's built-in sandbox arm (in-memory DNS,
  fake secret store) when one exists — a safe user evaluation arm AND a
  fully verifiable kind-cluster E2E lane; and read the controller's error
  handling before designing credential-less lanes (crash-loop vs
  log-and-retry providers).
