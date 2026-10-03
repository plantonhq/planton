# Kubernetes Tekton Operator

## When NOT to Use This

**One resource is the Tekton LIFECYCLE MANAGER** — the operator that
turns a `TektonConfig` declaration into running Tekton components and
keeps them converged. One install per cluster (an upstream contract).

Not the right kind when:

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

## Done means serving

The deploy completes only once the operator can serve a `TektonConfig`:
both of its Deployments have rolled out and the API server admits a
`TektonConfig` with the operator's defaults filled in (a server-side dry
run; nothing is written). The operator's webhook registers its own
admission configuration at runtime, so until it runs a `TektonConfig`
is refused — on a cluster whose node pool starts empty, for as long as
the first node takes to join. A `KubernetesTekton` applied right after
this resource therefore never meets a half-started operator. The wait
runs `kubectl` where the deploy runs (the Planton runner carries it); it
is bounded at 10 minutes for the pods and 5 for the webhook, and says
what it waited for when it gives up.

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
table built for another. The outputs name the registry and the four
images Tekton injects into build pods (entrypoint, nop, workingdirinit,
sidecarlogresults) exactly as the cluster pulls them, digests included --
the list to mirror and allow-list before the first build.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
