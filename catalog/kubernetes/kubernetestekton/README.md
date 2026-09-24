# Kubernetes Tekton

## When NOT to Use This

**One resource is THE cluster's Tekton installation** — the declaration
of which components run (Pipelines, Triggers, Dashboard, Chains), their
feature flags, execution defaults and cleanup policy. Exactly one per
cluster (the operator's own admission rule).

Not the right component when:

- **The operator is missing** — `KubernetesTektonOperator` is the
  registry prerequisite; this resource is the declaration it
  reconciles.
- **You want to declare pipelines** — Tasks, Pipelines and their runs
  are plain custom resources once this converges; declare them via
  `KubernetesManifest`, your platform, or the Tekton CLI. This kind
  shapes the engine, never your pipelines.
- **You want GitHub Actions runners** — that is the
  `KubernetesGhaRunnerScaleSet` family; Tekton is its own ecosystem.

## The profile honesty

`lite` = Pipelines. `basic` = + Triggers. `all` (the default) = +
Dashboard. Chains installs on `basic`/`all` unless `chain.disabled`.
The dashboard has NO authentication of its own — its Service stays
ClusterIP and exposure composes from first-class kinds over the
exported handle; set `dashboard.readonly` before putting it anywhere
people can reach.

## One CloudEvents sink per cluster

`pipeline.cloud_events_sink_url` is Tekton's single, cluster-global
event destination — every run in every namespace reports there, and
per-namespace sinks do not exist. Planton's own runners need none: a
build-capable runner watches its build namespace's PipelineRuns and
TaskRuns directly and reports to its own control plane, so a cluster
serving several Planton control planes leaves this field unset. Set it
only for an event consumer of your own, and put a fan-out service at
the URL if more than one consumer needs the stream (each event carries
its source namespace).

## Two fields you cannot change in place

`target_namespace` is immutable (the operator rejects the update —
destroy and recreate to move), and the rendered TektonConfig is always
named `config` (the operator's singleton rule; your `metadata.name`
keys the Planton resource only).

## Destroy while the operator lives

Deleting this resource makes the operator tear down every component it
installed, and the deletion BLOCKS until that finishes. Never destroy
the `KubernetesTektonOperator` first — without a running operator the
teardown finalizers strand.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
