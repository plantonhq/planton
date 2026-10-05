# Kubernetes cert-manager family rebuilt: controller, issuers, and certificates at full depth on both engines

**Date**: 2026-07-22
**Scope**: `apis/dev/planton/provider/kubernetes` (kubernetescertmanager, kubernetesclusterissuer, kubernetesissuer, kubernetescertificate — all rebuilt; new shared `cert_manager_issuer.proto`), `catalogkind` (registry prerequisites), `aa_e2e/verify` (cert-manager verifiers), `aa_import` (certmanager map), `e2e` + Makefile Tier-1, `pkg/outputs`, `pkg/iac/importmap` (ledger), `pkg/iac/pulumi/pulumimodule/provider/kubernetes/certmanagerissuer` (new shared builder), `pkg/kubernetes/kubernetestypes` (cert-manager pin + regenerated types), site catalog, `_rules/catalog-kind/update`

## What changed

The four kinds that give a cluster automatic TLS — the cert-manager
controller install, the two issuer kinds (who signs), and the certificate
kind (what is issued) — rebuilt from the pinned upstream (chart values +
CRD schemas) to full configuration depth, with dual-engine parity and live
kind-cluster E2E including real certificate issuance.

### KubernetesCertManager (rebuilt — the first Helm-backed typed kind at the current bar)

- **Typed spec over the chart's meaningful values surface**: CRD lifecycle
  (install-with-release default true, keep-on-uninstall true — deleting the
  CRDs cascades to every certificate object cluster-wide), controller
  tuning, cluster-resource namespace, DNS-01 self-check resolvers (the
  split-horizon fix), feature gates, webhook host-network/secure-port (the
  EKS-custom-CNI fix), cainjector, startupapicheck, prometheus/
  ServiceMonitor, scheduling, image registry. `helm_values` merges LAST
  with Helm `-f` semantics on both engines — the escape hatch, never the
  interface.
- **Keyless DNS-01 via the shared workload-identity proto**: the chart owns
  the ServiceAccount; the identity annotation rides
  `serviceAccount.annotations` (AKS also gets the required pod label).
  The SA name and the resolved cluster-resource namespace are exported as
  composition seams.
- **The known Terraform crash fixed in its owning session**: the old
  locals read optional nested blocks through `x != null && x.y != null`
  ternaries — HCL's `&&` does not short-circuit, so an absent
  workload-identity block crashed the plan. All reads are `try()`-guarded.
- Both engines install a real Helm release and wait for the
  startupapicheck hook Job (the webhook-actually-serves proof); atomic +
  cleanup-on-fail.

### KubernetesClusterIssuer + KubernetesIssuer (rebuilt on one shared config)

- **Upstream gives the two kinds an IDENTICAL spec; the rebuild makes that
  structural**: one shared `CertManagerIssuerConfig` proto, one shared
  Pulumi spec builder (`pkg/.../certmanagerissuer`), twin Terraform locals.
  The kinds cannot drift.
- **Grain change (ClusterIssuer)**: the issuer is named after the resource
  (`metadata.name`) like every other kind — the old one-issuer-per-DNS-domain
  shape is gone. Solver SELECTORS (dns_zones/dns_names/match_labels) scope
  challenge strategies within one issuer, which is upstream's own model.
- **Full backend surface**: ACME (multiple solvers, HTTP-01 via Ingress or
  Gateway HTTPRoute, DNS-01 across Cloudflare, Route53, Azure DNS, Cloud
  DNS, DigitalOcean, RFC 2136, acme-dns, Akamai, plus the `webhook`
  extension point; External Account Binding; profiles; preferred chains),
  CA, SelfSigned, and Vault (token/AppRole/Kubernetes auth). Venafi is
  deliberately unmodeled (proprietary platform; recorded in the research
  doc). ACME requires ≥1 solver — a solver-less issuer is an
  always-misconfiguration, rejected at validation instead of hanging at
  issuance.
- **Declared-credential model**: wherever upstream expects a secretRef, the
  spec takes the credential VALUE (sensitive); modules materialize
  deterministic Secrets (`<name>-solver<N>-<provider>`) in the namespace
  cert-manager reads from, identically on both engines. Keyless paths
  (IRSA / Workload Identity / Managed Identity) leave credentials empty.
- **Issuer readiness is never awaited** (ACME/Vault/DNS reachability is not
  part of applying the resource) — the same posture as Ingress.

### KubernetesCertificate (rebuilt to the complete cert-manager.io/v1 surface)

- All SAN types (DNS/IP/URI/email/otherName), subject attributes,
  `literal_subject` (LDAP DN, order-preserving), issuer selection incl. an
  `external` arm for third-party issuer controllers, duration +
  renew-before OR renew-before-percentage, private key
  (algorithm/size/encoding/rotation with per-algorithm size contracts),
  usages (exact x509 vocabulary), `is_ca` + X.509 name constraints,
  JKS/PKCS#12 keystores with inline sensitive passwords, DER/combined-PEM
  output formats, secret template, revision history.
- Pulumi renders through the typed crd2pulumi SDK (types REGENERATED from
  the newly pinned cert-manager release — the old types were generated from
  a release line four minors behind the module default; pin and generator
  URLs reconciled in `pkg/kubernetes/kubernetestypes/Makefile`).

### Terraform engine: custom resources migrated to kubectl_manifest

The three CR kinds' modules previously used `kubernetes_manifest`, which
requires a reachable cluster AND the CRD's OpenAPI schema at PLAN time — an
issuer could never be planned before cert-manager existed, breaking
single-run Infra Charts and offline plan proofs. All three now apply
through `kubectl_manifest` (server-side apply, no plan-time cluster
dependency), consistent with the raw-manifest kind.

### HCL type-fidelity discipline (live E2E caught a real bug)

The Certificate full-surface Terraform lane failed server-side with
`.spec.privateKey.size: expected numeric, got string`: conditional-merge
assembly (`merge(concat(cond ? [{…}] : [], …)...)`) silently UNIFIES
primitive-only sibling objects into `map(string)`, stringifying numbers and
booleans; plain `cond ? {…} : {}` ternaries fail plan type-checking
outright for heterogeneous branches. All CR-spec/values rendering now uses
the null-prune idiom (`key = cond ? value : null` in one object literal,
pruned by a for-expression), which preserves every value's type. The
update rule now teaches all three idioms and their failure modes.

### Registry prerequisites + E2E machinery

- The three CR kinds declare `prerequisites: [KubernetesCertManager]` — the
  harness installs the controller fixture before their scenarios, the same
  mechanism the Gateway API kinds use. Fixture manifests chain further
  fixtures via the `planton.dev/e2e-prerequisites` annotation (the
  Certificate root-CA fixture pulls in the self-signed Issuer fixture it
  references).
- New verifiers: cert-manager install (three Deployments Available + CRDs
  Established), issuer Ready, and certificate issuance (Ready PLUS
  tls.crt/tls.key present in the target Secret — the customer-grade proof).

## Validation

- Spec tests all four kinds green (valid arms + every CEL contract's
  rejection); per-kind + release-entrypoint builds; `make build-go`;
  secret-coverage; validate-refs; outputs conformance (+4 cases); kind map;
  e2e matrix; site catalog regenerated; 15 presets validated against the
  new specs.
- Offline: tofu validate + full-surface AND optionals-absent plan proofs
  for all four kinds (numeric/bool fidelity verified in the rendered CR).
- Live on the kind cluster, BOTH engines, full six-phase runner: CertManager
  minimal + tuned; ClusterIssuer self-signed; Issuer self-signed + the
  composed CA-chain (self-signed Issuer fixture → root CA Certificate
  fixture → CA Issuer Ready — the family's composition story, FK-wired end
  to end); Certificate minimal + full-surface with REAL issuance verified.
  Blind import round-trip green for the certmanager map (both scenarios).
  Zero orphan namespaces/issuers/certificates (kept CRDs are the designed
  keep-on-uninstall behavior, proven idempotent across twelve reinstalls).
- Not verified (recorded honestly): ACME live issuance against a public CA
  (needs a real domain + DNS credentials — deferred to the
  batched real-cluster/Cloudflare lanes); import round-trips for the three
  kubectl_manifest-backed CR kinds (the catalog needs a composed-ID
  derivation first — deferred with the reason in the importmap README, not
  silently skipped); no parity exceptions shipped (both engines express the
  full surface — the CR specs are byte-identical by construction).
