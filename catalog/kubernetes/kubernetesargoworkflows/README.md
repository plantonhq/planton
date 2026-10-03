# Kubernetes Argo Workflows

## When NOT to Use This

**One resource is ONE Argo Workflows engine** — the Kubernetes-native
executor for DAG/step pipelines (CI jobs, data and ML pipelines, batch
orchestration).

Not the right kind when:

- **You want to declare the pipelines it runs** — Workflows,
  WorkflowTemplates and CronWorkflows are plain custom resources once
  the engine runs; declare them via `KubernetesManifest`, a chart, or
  the Argo CLI/UI. This kind installs the engine, never your pipelines.
- **You need GitOps delivery** — that is `KubernetesArgocd`. Workflows
  runs jobs to completion; it does not converge cluster state.
- **A plain Job/CronJob is enough** — single-container batch work
  without steps, artifacts or fan-out needs no engine.

## The runner identity

Workflow pods run as the `workflow_service_account` (default
`argo-workflow`), never as the controller. Annotate THAT account for
IRSA/workload identity when pipelines touch cloud APIs, and grant extra
RBAC to it alone. `workflow_namespaces` places the runner identity into
additional namespaces (the install namespace is always included) — the
controller's watch stays cluster-wide either way.

## The two durability seams

`artifact_repository` gives steps a place to pass files (S3-compatible
— an in-cluster `KubernetesSeaweedFs` pairs by reference — GCS, or
Azure Blob; keyless ambient-identity arms on all three). `archive`
writes completed workflows to Postgres/MySQL (a `KubernetesPostgres`
pairs by reference) so history survives CR garbage collection. Without
them pipelines still run — files just cannot flow between steps and
history lives only as CRs.

## CRD install reaches the internet

The chart's default CRD arm downloads the full-schema CRDs from its
GitHub release at install time via a hook Job (they exceed inline
limits). Air-gapped clusters set `crds.full_schema: false` (templated
minified CRDs, no download) or point `crds.base_url` at a mirror.
`crds.keep` defaults true: destroy leaves the CRDs — and every Workflow
in the cluster — intact.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
