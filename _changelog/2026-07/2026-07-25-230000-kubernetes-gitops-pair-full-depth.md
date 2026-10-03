# Kubernetes GitOps: Argo CD rebuilt and Argo Workflows forged at full depth

## What changed

- **KubernetesArgocd (950, rebuilt)** — deploys Argo CD, the declarative
  GitOps delivery engine, from the official `argo-cd` Helm chart pinned
  10.2.1 (Argo CD v3.4.5, argoproj index). The previous surface — a
  single container-resources block with an embedded ingress — is
  replaced end to end: seven typed kinds (application controller,
  API/UI server, repo server, ApplicationSet controller, notifications,
  dex, commit server), a three-arm Redis cache oneof (bundled single pod
  XOR the redis-ha Sentinel subchart XOR an external endpoint that
  default-references a KubernetesValkey resource), SSO via direct OIDC
  (PKCE or secret-backed) or dex connectors, Argo CD's CSV RBAC surface,
  public-repository registrations, and templated-CRD keep semantics.

- **Nothing credential-bearing renders into values on either kind.** The
  Argo CD admin password is APPLICATION-owned — Argo CD generates it at
  first start into the fixed-name `argocd-initial-admin-secret`,
  exported as an output handle; SSO client secrets ride Argo CD's own
  `$<secret-name>:<key>` runtime indirection against labeled Secrets;
  private-repo credentials are labeled Secrets composed outside the
  spec; Argo Workflows artifact-store and archive-database credentials
  ride the chart's secret selectors, with keyless ambient-identity arms
  (IRSA / workload identity on the runner ServiceAccount) on every
  cloud backend.

- **The templated-CRD ownership truth is taught where users read.** The
  argo-cd chart templates its CRDs, so kept CRDs carry the keeping
  release's Helm ownership metadata — a later release cannot adopt them
  and must install with `crds.install: false`. The spec field teaches
  the class; destroy semantics (keep by default — removing the CRDs
  cascades to every Application in the cluster) are explicit on the
  field and in the docs.

- **KubernetesArgoWorkflows (951, new)** — deploys Argo Workflows, the
  Kubernetes-native DAG/step pipeline engine, from the official
  `argo-workflows` chart pinned 1.0.23 (v4.0.8). Controller depth
  (hot-standby replica truth, an instance-ID claim for multi-engine
  clusters — the spec's plain string maps onto the chart's structured
  `{enabled, explicitID}` block, parallelism caps, retention), the Argo
  server with honest auth-mode teaching (`client` default — callers act
  with their own Kubernetes permissions), the runner-identity contract
  (the chart defaults `workflow.serviceAccount.create` to false — the
  module always creates the runner ServiceAccount; the
  `workflow_namespaces` list PLACES runner RBAC and is always rendered
  so the chart's `["default"]` default never silently plants
  permissions in the cluster's `default` namespace), an
  artifact-repository oneof (S3-compatible — a KubernetesSeaweedFs
  pairs by reference — GCS, Azure) and a Postgres/MySQL workflow
  archive that default-references a KubernetesPostgres resource.

- **The CRD install's internet dependency is explicit.** The chart's
  default arm downloads full-schema CRDs from its GitHub release via a
  hook Job at install time; the spec models the air-gap fallback
  (templated minified CRDs) and a mirror base URL, with the trade-offs
  taught on the fields.

- **Enum housekeeping:** KubernetesArgoWorkflows takes 951 next to
  KubernetesArgocd; the unbuilt GitOps-band placeholders
  (TektonOperator, Tekton, GhaRunnerScaleSetController,
  GhaRunnerScaleSet, Harbor, Jenkins) renumber to 952–957.

- **E2E surfaces authored** (compiled, not yet live-run — the profiles
  ship `pending_proof`): product-grade verifiers driving Argo CD's own
  session and applications APIs (login as the generated admin, a GitOps
  sync of a public repository asserted on the synced workload, recovery
  across a server pod replacement) and Argo Workflows' engine
  (an authenticated version-API round-trip as the runner identity via a
  TokenRequest token, a Workflow run to Succeeded under the runner
  ServiceAccount, recovery across a controller pod replacement), plus
  import maps, presets, docs and catalog pages for both kinds.

## Why

GitOps delivery and pipeline execution are the heart of how teams
operate Kubernetes; both kinds now hold the catalog bar: typed depth
over the official charts' meaningful surface, secret-by-default
credential paths, cross-engine parity by construction, and composition
handles that let infra charts wire Argo Workflows' artifacts and
history to in-catalog object storage and Postgres by reference.

## Impact

- Manifests: `KubernetesArgocd` specs from the previous surface are not
  compatible (pre-launch, no adoption); `KubernetesArgoWorkflows` is
  new.
- Both kinds await live proving; their CI lanes enter the matrix when
  the live runs go green (the previous KubernetesArgocd CI entry — keyed
  to the pre-rebuild surface — is dropped until then).
