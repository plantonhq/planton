# Kubernetes GHA Runner Scale Set Controller

## When NOT to Use This

**One resource is the runner-fleet CONTROLLER** — GitHub's official
actions-runner-controller manager that reconciles runner scale sets
into listeners and ephemeral runner pods. One cluster-wide controller
is the sane default.

Not the right kind when:

- **You want actual runners** — that is `KubernetesGhaRunnerScaleSet`,
  one per repository/organization/enterprise registration. This kind
  installs the manager; it registers nothing with GitHub and holds no
  GitHub credential.
- **You run Tekton** — the `KubernetesTektonOperator` /
  `KubernetesTekton` pair is the Tekton path; this family is
  specifically GitHub Actions.
- **You need the legacy community ARC** — the old
  RunnerDeployment-style controller (cert-manager era) is a different,
  superseded product; this is GitHub's supported scale-set line.

## Watch scope and multi-tenancy

By default the controller watches ALL namespaces — every runner scale
set on the cluster is served by it, which is what almost everyone
wants. `flags.watch_single_namespace` fences a controller to one
namespace for hard multi-tenancy; every scale set outside the fence
then needs its own controller AND must name its controller explicitly
(the scale set kind's `controller_service_account`, wired from this
kind's `service_account_name` output).

## The destroy contract

The chart installs the `actions.github.com` CRDs and removes them with
the release — destroying the controller cascade-deletes every runner
scale set on the cluster. Destroy the scale sets first.

## Version lockstep

Chart and controller image move together, and GitHub supports
controller and scale-set charts only at MATCHING versions — keep this
kind's `chart_version` equal to every `KubernetesGhaRunnerScaleSet`'s.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
