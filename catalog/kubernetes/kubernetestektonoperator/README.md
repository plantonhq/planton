# Kubernetes Tekton Operator

## When NOT to Use This

**One resource is the Tekton LIFECYCLE MANAGER** — the operator that
turns a `TektonConfig` declaration into running Tekton components and
keeps them converged. One install per cluster (an upstream contract).

Not the right component when:

- **You want Tekton itself** — that is `KubernetesTekton`: the
  declaration of which components run (Pipelines, Triggers, Dashboard,
  Chains), their feature flags and the pruner policy. Installing this
  operator alone deploys NO Tekton components — by design.
- **You want to run pipelines** — PipelineRuns/TaskRuns are plain
  custom resources once `KubernetesTekton` converges; declare them via
  `KubernetesManifest`, your platform, or the Tekton CLI.
- **You want GitHub Actions runners** — that is the
  `KubernetesGhaRunnerScaleSetController` / `KubernetesGhaRunnerScaleSet`
  pair; Tekton is its own pipeline ecosystem.

## Why auto-install is disabled

The upstream release ships with the operator auto-creating a default
`TektonConfig` (profile `all`) at startup. This module always disables
that: two managers writing one object fight through server-side apply,
and the cluster's Tekton shape would depend on install order. Here the
`KubernetesTekton` resource is the single owner of the configuration —
this operator only reconciles it.

## The destroy contract

The operator's CRDs delete with it, which cascade-deletes any
`TektonConfig`. Always destroy the `KubernetesTekton` resource FIRST:
its teardown blocks until the operator finishes removing the components
(the `TektonInstallerSet` finalizers are processed by the RUNNING
operator — removing the operator first strands them, which is exactly
the hang this ordering exists to prevent).

## Distribution

Installed from the official single-file release manifest at the pinned
tag (the in-repo Helm chart is unpublished). The namespace is the
manifest's fixed `tekton-operator`; the spec deliberately has no
version field — the `TektonConfig` surface `KubernetesTekton` models is
designed against the pinned release. `image_registry` points every image
Tekton publishes at a mirror of ghcr.io; the modules read the images the
pinned release installs from a table built for that release, and refuse a
table built for another.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
