# Kubernetes CronJob

Runs work on a recurring schedule on any Kubernetes cluster as a batch/v1 CronJob: at each scheduled time the controller creates a Job from the job template, and that Job runs pods to completion. This is the kind for scheduled work — nightly backups, report generation, periodic cleanup. The spec splits cleanly in two: scheduling controls (a 5-field cron expression, an explicit IANA time zone, overlap handling that deliberately defaults SAFER than upstream, history retention) live at the top level, and everything about the work itself lives in the job template, which mirrors the complete KubernetesJob batch surface (parallelism, Indexed mode, per-index retries, pod failure and success policies, sidecars, probes, security contexts). Credentials are delivered through a Kubernetes Provider Connection or Runner-based delivery.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Kubernetes CronJob** -- the scheduled workload resource; at each matching time it stamps a Job from the template, and the history limits bound how many finished Jobs survive
- **Kubernetes Secret** -- created only when `jobTemplate.container.app.env.secrets` carry literal values; stores them and mounts them as environment variables
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A container registry** accessible from the cluster for pulling the run's image. If the registry is private, provide image pull credentials in the container image configuration.

## Deploy

### Console

Open the deployment store, find **Kubernetes CronJob**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Nightly Backup CronJob** preset for the classic 03:00 backup window, or **Frequent Sync CronJob** for a tight-interval schedule, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCronJob
metadata:
  name: nightly-backup
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "databases"
  createNamespace: false
  schedule: "0 3 * * *"
  timeZone: "America/New_York"
  jobTemplate:
    container:
      app:
        image:
          repo: ghcr.io/acme-corp/pg-backup
          tag: v2.1.0
        command: ["/backup.sh"]
    activeDeadlineSeconds: 5400
    backoffLimit: 2
```

```shell
planton apply -f cronjob.yaml
```

This runs the backup daily at 03:00 New York time, skipping a run if the previous one is still going (the Forbid default), with a 90-minute deadline so a hung run can never block the schedule. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the CronJob to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: databases-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the CronJob with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Kubernetes CronJob. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Schedule and time zone** -- `schedule` is a standard 5-field cron expression: minute, hour, day of month, month, day of week (e.g. `"0 3 * * *"` daily at 03:00, `"*/15 * * * *"` every 15 minutes). Use standard cron only — @-style macros are not portable across controller versions. Set `timeZone` (an IANA name) whenever the wall-clock time matters: unset means the schedule runs in the cluster controller's LOCAL zone, usually UTC but not guaranteed.

**Concurrency** -- `concurrencyPolicy` decides what happens when the next run comes due while the previous one is still going. The default here is `"Forbid"` (skip the new run) — deliberately SAFER than upstream Kubernetes, which defaults to Allow: overlapping cron runs are the classic scheduled-workload incident, two backups writing the same target. `"Replace"` cancels the running job and starts the new one; `"Allow"` runs them concurrently.

**Missed runs and pausing** -- `startingDeadlineSeconds` bounds how late a missed run may start before it counts as failed and is skipped; the controller stops scheduling after 100 consecutive missed runs, so frequent schedules benefit from an explicit deadline. `suspend` pauses the SCHEDULE — running Jobs are unaffected.

**The job template** -- everything about the work mirrors KubernetesJob: `jobTemplate.container.app` carries the image and command (the exit code decides each run's success), the batch dials (parallelism, completions, Indexed mode, per-index retries), a pod failure policy (fail fast on unrecoverable exit codes, forgive node disruptions), and a success policy for Indexed runs. Set `jobTemplate.activeDeadlineSeconds` especially under Forbid — a hung run silently blocks every subsequent run without it.

**History** -- `successfulJobsHistoryLimit` (default 3) and `failedJobsHistoryLimit` (default 1) bound how many finished Jobs survive for logs and post-mortems. Keep at least one failed Job.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesServiceAccount** | `jobTemplate.pod.serviceAccount` | `status.outputs.service_account_name` |
| **KubernetesSecret** | `jobTemplate.pod.imagePullSecrets` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace the CronJob was created in | Other workloads deploying into the same namespace |
| `cron_job_name` | Name of the CronJob object as created in the cluster | Monitoring dashboards and run inspection |
| `schedule` | The effective cron expression the CronJob runs on | Audits and dependents read the DEPLOYED truth, not the spec |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Nightly backup** -- Daily at 03:00 in an explicit time zone, Forbid concurrency untouched, a run deadline so a hung backup can never block the schedule, and failed-run history kept for post-mortems. Start from the **Nightly Backup CronJob** preset.

**Frequent sync** -- A tight-interval schedule with a starting deadline (bounding the missed-run counter) and Replace concurrency — only the newest run matters. Start from the **Frequent Sync CronJob** preset.

**Monthly report** -- A low-frequency schedule with generous history retention and an Indexed fan-out template for partitioned report generation. Start from the **Monthly Report CronJob (Indexed)** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the target namespace for the CronJob
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- keyless cloud access for each run via workload identity
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- image pull secrets and referenced credential material
- [**Kubernetes Job**](/infra-catalog/kubernetes-job) -- the one-shot twin: the same batch surface without the schedule
