# Kubernetes CronJob

## Overview

**KubernetesCronJob** is a catalog kind that runs work on a recurring schedule as a batch/v1 CronJob. At each scheduled time, the controller creates a Job from the template, and that Job runs pods to completion. This is the kind for scheduled work: nightly backups, report generation, periodic cleanup, data synchronization.

For one-shot work, use **KubernetesJob**. For always-on services, use **KubernetesDeployment**.

## The Scheduling Split

The spec divides cleanly in two, and everything follows from the split:

- **Top-level scheduling controls** — `schedule` (a standard 5-field cron expression), `time_zone`, `starting_deadline_seconds`, `concurrency_policy`, `suspend`, and the two history limits. These describe *when* work runs and how runs relate to each other
- **`job_template`** — everything about the work itself: containers, pod configuration, and the full set of batch controls (parallelism, completions, Indexed mode, retry budgets, deadlines, failure and success policies). Anything a standalone `KubernetesJob` can express is equally expressible here, with one exception: `suspend` lives at the CronJob level, where it pauses the schedule itself

Set `time_zone` (an IANA name like `America/New_York`) whenever the wall-clock time matters — without it, the schedule follows the controller's local clock, which is usually UTC but not guaranteed.

## Concurrency: Why the Default Is Forbid

`concurrency_policy` decides what happens when the next run comes due while the previous run is still going:

- **`Forbid` (this kind's default)** — skip the new run; the previous one keeps going
- **`Allow`** — run them concurrently
- **`Replace`** — cancel the running Job and start the new one

Upstream Kubernetes defaults to `Allow`. This component deliberately defaults to `Forbid`: overlapping cron runs are the classic scheduled-workload incident — two backups writing the same target, two migrations racing — so overlap is opt-in here rather than a surprise. Pair `Forbid` with `active_deadline_seconds` in the template, so a hung run cannot silently block every subsequent run.

## Run-to-Completion Semantics

Each scheduled run inherits Job semantics wholesale: a pod succeeds when **all** of its containers exit 0, failures retry within the template's budget, and `restart_policy` is `Never` or `OnFailure` (never `Always`). The sidecar warning is sharper here than on a plain Job: a sidecar that never exits keeps every run from completing — and with `Forbid`, blocks all future runs too. Use termination-aware sidecars, or have the app signal them to shut down.

Retention is two-layered: the CronJob keeps the last `successful_jobs_history_limit` (default 3) and `failed_jobs_history_limit` (default 1) Jobs for log inspection, while the template's `ttl_seconds_after_finished` can delete individual Jobs sooner. The usual choice is to leave the TTL unset and let the history limits bound retention.

## Composition Story

This kind deliberately creates **no Service and no ingress** — scheduled pods are batch workers, not endpoints. Identity is composed, never bundled:

- **KubernetesServiceAccount** — reference it from `spec.job_template.pod.service_account`; workload-identity annotations and pull-secret attachment live on the identity
- **KubernetesRbac** — grants permissions to that identity
- **KubernetesSecret / KubernetesConfigMap** — consumed via `secret_ref` env entries, `env_from` imports, or volume mounts
- **KubernetesNamespace** — `spec.namespace` accepts a reference, so one chart creates the namespace and the CronJob together

## Service Hub Deploy Target

KubernetesCronJob is a Service Hub deployment target for scheduled workloads built from user code. Deployment pipelines inject the freshly built artifact at **`spec.job_template.container.app.image`** (repository + tag). That path is part of the kind's public contract.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`namespace`** — the namespace the CronJob was created in
- **`cron_job_name`** — the name of the CronJob object in the cluster
- **`schedule`** — the effective cron expression, exported so dependents and audits read the deployed truth rather than the spec

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference), creating it when `create_namespace` is true
2. Materialize literal secret env values into a workload-scoped Kubernetes Secret and the registry logins declared on `spec.job_template.pod.image_registries` into an image-pull Secret
3. Build the Job template from the shared workload container/pod core plus the full batch-control surface
4. Create the batch/v1 CronJob with the scheduling controls
5. Export the namespace, CronJob name, and effective schedule

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesCronJob** when you need:

- Recurring scheduled work: backups, reports, cleanup, synchronization
- Scheduled parallel batch runs (Indexed templates)
- A Service Hub deploy target for scheduled workloads built from your own code

**Do NOT use** when:

- The work runs once — use **KubernetesJob**
- The process should react to events rather than a clock — use a queue-driven worker on **KubernetesDeployment**

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist, be created via `create_namespace`, or be referenced as a `KubernetesNamespace` resource in the same chart

## Best Practices

1. **Set `time_zone` whenever the run time matters**: the controller's local clock is an implementation detail, not a contract
2. **Keep `Forbid` unless you have a reason not to**: choose `Replace` only for idempotent latest-state-wins work, `Allow` only when overlap is genuinely safe
3. **Always set `active_deadline_seconds` in the template under `Forbid`**: a hung run must not block the schedule forever
4. **Set `starting_deadline_seconds` on frequent schedules**: it skips stale runs and keeps the controller's 100-consecutive-missed-runs cutoff at bay
5. **Make sidecars exit**: one never-exiting sidecar stalls every run — and, under `Forbid`, the entire schedule
6. **Size history limits to your debugging habits**: keep at least one failed Job so failure logs survive for post-mortems
7. **Reference secrets instead of embedding them**: use `secret_ref` env entries against a `KubernetesSecret`

## References

- [Kubernetes CronJobs Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Running Automated Tasks with a CronJob](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/)
- [Kubernetes Jobs Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [CronJob API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/cron-job-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
