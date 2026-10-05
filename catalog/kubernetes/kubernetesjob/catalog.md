# Kubernetes Job

Runs work to completion on any Kubernetes cluster as a batch/v1 Job: pods are created, execute until they succeed or exhaust their retry budget, and are never restarted once the Job finishes. This is the kind for one-shot work — data migrations, backfills, report generation, parallel batch processing. Supports the complete batch/v1 surface: parallelism and completions, Indexed completion mode for partitioned workloads, per-index retry budgets, pod failure policies (fail fast on unrecoverable exit codes, forgive node disruptions), success policies for leader/worker topologies, active deadline enforcement, TTL-based cleanup, and the full shared workload surface (sidecars and init containers, all environment variable sources, probes, security contexts).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Kubernetes Job** -- the batch workload resource that creates pods and runs them to completion, with the configured fan-out, retry budgets, deadlines, failure/success policies, and container specification
- **Kubernetes Secret** -- created only when `container.app.env.secrets` carry literal values; stores them and mounts them as environment variables
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A container registry** accessible from the cluster for pulling the job image. If the registry is private, provide image pull credentials in the container image configuration.

## Deploy

### Console

Open the deployment store, find **Kubernetes Job**, and click **Deploy**. The creation wizard walks you through placement, the container specification, fan-out and retries, failure and success policies, deadlines, and cleanup. Start from the **Database Migration Job** preset for sequential one-shot work, or **Parallel Batch Job (Indexed)** for an Indexed fan-out, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesJob
metadata:
  name: etl-pipeline
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "jobs"
  createNamespace: true
  container:
    app:
      image:
        repo: ghcr.io/acme-corp/etl-runner
        tag: v2.0.0
      command: ["python", "run_etl.py"]
  backoffLimit: 3
  ttlSecondsAfterFinished: 3600
```

```shell
planton apply -f job.yaml
```

This creates a single-pod Job (parallelism and completions default to 1) with up to 3 retries on failure and automatic cleanup 1 hour after it finishes. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Job to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: jobs-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline resolves the dependency graph, deploys the namespace first, then provisions the Job with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Kubernetes Job. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Parallelism and completions** -- `completions` is the number of successful pods required; `parallelism` is how many run concurrently. `completions: 10` with `parallelism: 3` runs up to 3 pods at a time until 10 succeed. `completionMode: "Indexed"` assigns each pod a completion index (0 to completions-1, exposed via the batch.kubernetes.io/job-completion-index annotation and the pod hostname) so each pod claims its own partition of the work.

**Retries** -- `backoffLimit` (Kubernetes default 6) is the global retry budget. For partitioned work, `backoffLimitPerIndex` gives each index its OWN budget — one flaky partition exhausts only its own retries — and `maxFailedIndexes` bounds how many indexes may fail before the whole Job does. The per-index dials require Indexed mode and `restartPolicy: "Never"`.

**Pod failure policy** -- fine-grained per-failure handling, evaluated in order with first-match-wins: `FailJob` immediately on an unrecoverable exit code (a misconfiguration should not burn six retries), `Ignore` failures caused by node disruption (the DisruptionTarget pod condition) so infrastructure events never count against the budget, or `FailIndex` to stop retrying a single partition. Requires `restartPolicy: "Never"`.

**Success policy** -- Indexed-only early success: declare the Job succeeded once specific indexes (`succeededIndexes: "0"` for a leader) or enough indexes (`succeededCount`) finish; lingering pods are then terminated.

**Deadlines and cleanup** -- `activeDeadlineSeconds` is the hard wall-clock guard: past it, everything is killed regardless of retries (note that suspending the Job resets this timer). `ttlSecondsAfterFinished` deletes the finished Job and its pods automatically; unset preserves them for `kubectl logs` post-mortems.

**The work itself** -- `container.app.command`/`args` override the image entrypoint; the app container's exit code decides the pod's success. Sidecars need an exit story: a Job pod only completes when ALL its containers exit.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesServiceAccount** | `pod.serviceAccount` | `status.outputs.service_account_name` |
| **KubernetesSecret** | `pod.imagePullSecrets` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Kubernetes namespace the Job was created in | Other workloads deploying into the same namespace |
| `job_name` | Name of the Job object as created in the cluster | Monitoring dashboards and log queries |
| `selector_labels` | The controller-generated job-name selector as a `k=v,k=v` string | `kubectl get pods -l` / `kubectl logs -l` post-run inspection |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Database migration** -- Sequential one-shot work with a tight retry budget (retrying a half-applied migration is dangerous), an active deadline to catch runaway processes, and no TTL cleanup so the pods survive for verification. Start from the **Database Migration Job** preset.

**Parallel batch** -- An Indexed fan-out where each pod processes its own numbered partition, with a per-index retry budget so one flaky shard cannot fail the run. Start from the **Parallel Batch Job (Indexed)** preset.

**Resilient batch** -- A failure policy that fails fast on the "bad input" exit code and forgives node disruptions — the production posture for long-running batch work on spot capacity. Start from the **Resilient Batch Job (Failure Policy)** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the target namespace for the Job
- [**Kubernetes ServiceAccount**](/infra-catalog/kubernetes-service-account) -- keyless cloud access for the run's lifetime via workload identity
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) -- image pull secrets and referenced credential material
- [**Kubernetes CronJob**](/infra-catalog/kubernetes-cron-job) -- the scheduled twin: runs this same batch surface on a cron schedule
