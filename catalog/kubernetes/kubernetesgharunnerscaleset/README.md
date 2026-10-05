# Kubernetes GHA Runner Scale Set

## When NOT to Use This

**One resource is ONE runner fleet** — an autoscaling pool of ephemeral
self-hosted GitHub Actions runners registered against one repository,
organization or enterprise. Each runner pod executes exactly one job
and is replaced.

Not the right kind when:

- **The controller is missing** —
  `KubernetesGhaRunnerScaleSetController` is the registry prerequisite;
  this kind renders the fleet declaration it reconciles.
- **You want one fleet for several unrelated registrations** — a scale
  set binds to exactly one GitHub URL. Declare one resource per
  registration; an ORGANIZATION registration already serves all its
  repositories (fence access with `runner_group`).
- **Your CI is not GitHub Actions** — Tekton runs through the
  `KubernetesTekton` pair.

## Workflows target the NAME

The fleet registers in GitHub under `runner_scale_set_name` (default:
`metadata.name`, at most 45 characters — a GitHub limit) and workflows
select it with `runs-on: <that name>`. Labels are not how scale sets
route; the name is the whole contract, and it is exported as a
output.

## The credential never rides a manifest

`auth` is secret-native on every arm: reference a pre-created Secret
(recommended), or declare a PAT / GitHub App inline and the module
materializes the `<name>-github-auth` Secret itself — rendered chart
values only ever carry the Secret's NAME. A GitHub App is the
production posture: fine-grained permissions and expiring tokens.

## Docker builds need a container mode

The plain runner executes shell/tool steps only. `container_mode.mode:
dind` enables `docker build`/`docker run` via a privileged sidecar (the
cluster must allow privileged pods); `kubernetes` runs container jobs
as separate UNPRIVILEGED pods through the container hook (requires the
per-runner work volume); `kubernetes-novolume` is the same hook for
jobs that never share files between containers.

## Sizing is about the jobs

`runner.resources` bounds the BUILDS, not the agent — a starved runner
is a slow pipeline. `min_runners` trades idle cost against per-job pod
cold start; `max_runners` caps the blast radius (excess jobs queue in
GitHub).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
