# Kubernetes GitOps, CI-runner and app-platform tiers live-proven: Argo CD, Argo Workflows, Temporal, NATS, Tekton Operator, Tekton, GHA runner controller and scale set

## What changed

- **Eight kinds proven against a live cluster, both engines** —
  KubernetesArgocd, KubernetesArgoWorkflows, KubernetesTemporal,
  KubernetesNats, KubernetesTektonOperator, KubernetesTekton,
  KubernetesGhaRunnerScaleSetController, KubernetesGhaRunnerScaleSet.
  Every behavioral promise ran with verifier-output evidence: an
  API-declared Argo CD Application auto-synced a public repository into
  the namespace and stayed Synced/Healthy through a server pod
  replacement; a fresh Argo Workflow ran to Succeeded after a controller
  pod replacement; a Temporal workflow executed to completion through a
  real SDK worker against a composed catalog PostgreSQL, and the
  completed workflow still described after history + frontend pod
  replacements; a NATS JetStream marker survived a server pod
  replacement through its PersistentVolume, with the auth gate rejecting
  unauthenticated connections; a Tekton TaskRun ran to Succeeded on
  every lane, with the two-kind destroy ordering proven live
  (TektonConfig teardown completes while the operator runs, zero
  stranded InstallerSets); the GHA controller reconciled a declared
  runner scale set end to end. Blind import round-trips proved every
  kind's recipes, including the first live keyed-collection secret
  import (per-user generated passwords re-imported under their own
  username keys). All eight entered the green E2E CI matrix.

- **Operator+CRD manifest bundles now apply as three ordered groups on
  both engines** (namespace → workloads → CRDs) so destroy runs the
  reverse: the CRDs drain FIRST while the operator still lives to
  process its runtime CRs' finalizers, then the workloads, then the
  fixed namespace awaited to fully gone. Verified live in both
  directions: flat parallel application raced namespaced documents
  against their Namespace, and one-pass deletion wedged the
  tektoninstallersets CRD on an operator-processed finalizer for the
  full timeout on both engines. Terraform chains the groups with
  depends_on plus delete-waits; Pulumi chains three classic-yaml
  ConfigGroups with a group-scoped skip-await on the workloads (the
  webhook binary fatals until its CRD is served, so both the Deployment
  await and the webhook Service's endpoints await deadlock a waiting
  chain — read in the operator source and verified live).

- **The import round-trip now finishes with a reconcile-apply** —
  import → apply → operate, the real adopter workflow. An import cannot
  read CONFIG-ONLY attributes back from the cluster, and the round-trip
  hands its freshly-imported state to the destroy phase: delete-waits
  and cascade modes silently degraded to defaults on every Terraform
  round-trip lane (verified live: a namespace delete-wait that blocked
  correctly from applied state returned instantly from imported state).
  Applying the already-proven zero-change plan writes those attributes
  into state; the update rule teaches the class.

- **NATS generated user passwords are letters-only (length 40)** —
  nats-server re-parses env-referenced config values through its own
  config parser, so randomly-drawn passwords containing digit-led or
  structural tokens crash-looped the server INTERMITTENTLY (flaky by
  draw; read in the server's conf parser and verified live). Both
  engines generate parser-safe values; the modules and docs teach the
  server contract. A publish-allowlisted NATS user is additionally
  fenced from the JetStream API unless the allowlist grants
  `$JS.API.>` and `$JS.ACK.>` — the server silently drops denied
  publishes, so a fenced user's stream calls time out rather than fail
  loudly; taught on the spec's publish_allow field, the docs, and the
  scenario that now proves the recipe live.

- **The GHA controller's CRD posture corrected to the live truth** —
  the chart ships its four actions.github.com CRDs in the `crds/`
  directory: Helm installs them once and NEVER removes them. The spec,
  docs and destroy verification previously claimed release-owned
  deletion; all three now teach (and assert) the keep posture.

- **The Temporal verifier probes the UI server's own settings API**
  instead of matching brand text in the SPA document — the ui-server
  registers `GET /api/v1/settings` unconditionally and answers its
  SettingsResponse contract (read at the pinned ui-server source; the
  2.52.0 index.html carries no product-identifying text at all).

- **Scenarios needing owner-arranged external credentials now declare
  them via the `planton.dev/e2e-required-env` annotation** and lanes
  skip them honestly (with the missing variable names) instead of
  failing token expansion — the GHA scale set's live GitHub
  registration proof is the first user.

- **A new tier-wiring guard** (`hack/guards/ensure_e2e_tier_wiring.sh`
  plus its lint workflow) asserts every Kubernetes kind with a
  runnable E2E profile has BOTH engine test entrypoints and appears in
  its Makefile tier regexes. Its first run caught five silent drifts —
  a missing Terraform entrypoint that made one kind's lane unrunnable,
  and four kinds absent from the local tier sweeps (all fixed in this
  change).

- **Verifier hygiene hardened from live evidence**: proof objects are
  deleted CONFIRMED within a bounded budget while the stack serving
  their finalizers is still alive (a fire-and-forget TaskRun delete
  left Tekton Results' archival finalizer holding the target namespace
  in Terminating); teardown-cascade assertions are bounded convergence
  waits, never one-shot snapshots; and the Tekton assertions exclude
  the operator's own webhook InstallerSet, which belongs to the
  operator's lifetime rather than any TektonConfig.

## Why

"E2E passed" is a customer-grade promise: every one of these kinds
can be deployed by a customer today, on either engine, with its
behavioral guarantees — GitOps convergence, workflow durability,
messaging durability, pipeline execution, runner reconciliation —
demonstrated against a live cluster rather than claimed. The live lanes
exist to catch exactly what they caught here: ordering hazards no
offline plan can see, upstream postures that differ from their
documentation, and harness blind spots that would have silently
weakened every future proof.
