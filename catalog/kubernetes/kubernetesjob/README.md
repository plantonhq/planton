# Kubernetes Job

## Overview

**KubernetesJob** is a catalog kind that runs work to completion on a Kubernetes cluster as a batch/v1 Job. Pods are created, execute until they succeed (or exhaust their retry budget), and are never restarted once the Job finishes. This is the kind for one-shot work: database migrations, backfills, report generation, parallel batch processing.

The kind covers the complete batch/v1 JobSpec surface that matters for declarative batch work — parallelism, completions, Indexed completion mode, global and per-index retry budgets, deadlines, TTL cleanup, suspension, pod failure policies, and success policies — on top of the same fully-modeled container and pod core shared by every Planton workload kind.

For work that runs on a schedule, use **KubernetesCronJob**. For always-on services, use **KubernetesDeployment**.

## Run-to-Completion Semantics

A Job is not a service. Its pods exist to finish:

- The Job controller creates pods until the required number of successful completions is reached (`completions`, default 1), running at most `parallelism` pods at a time (default 1)
- A pod succeeds when **all** of its containers exit with code 0. This includes sidecars — a sidecar that never exits keeps the Job running forever. Use termination-aware sidecars, or have the app signal them to shut down when its work is done
- Failed pods are retried within a budget: `backoff_limit` globally (Kubernetes defaults to 6, with exponential back-off between retries) or `backoff_limit_per_index` per partition in Indexed mode
- `active_deadline_seconds` is the hard wall-clock guard: when exceeded, all running pods are killed and the Job fails regardless of remaining retries
- `ttl_seconds_after_finished` deletes the finished Job and its pods automatically; keep some retention if you rely on `kubectl logs` for post-mortems

## Batch Controls

- **`parallelism` / `completions`** — how many pods run at once and how many successes finish the Job
- **`completion_mode`** — `NonIndexed` (default): pods are interchangeable. `Indexed`: each pod receives a completion index (0 to completions−1) via the `batch.kubernetes.io/job-completion-index` annotation and the `JOB_COMPLETION_INDEX` environment variable, so each pod can claim its own partition of the work
- **`backoff_limit_per_index` + `max_failed_indexes`** — Indexed-mode retry budgeting: one flaky partition exhausts only its own retries, and the Job tolerates a bounded number of permanently failed indexes
- **`restart_policy`** — `Never` (default): each failed attempt leaves its pod behind for debugging and the controller creates a replacement. `OnFailure`: the container restarts in place, reusing the pod. `Always` is invalid for Jobs — it would restart even successfully completed containers
- **`pod_failure_policy`** — classify failures before they burn retries: fail the Job immediately on an unrecoverable exit code, ignore failures caused by node disruption. Requires `restart_policy: Never`
- **`success_policy`** — declare Indexed Jobs succeeded early, e.g. when the leader index finishes
- **`suspend`** — create the Job without running it; note that suspending a running Job deletes its active pods and resets the `active_deadline_seconds` timer

## Composition Story

This kind deliberately creates **no Service and no ingress**: Job pods are batch workers, not endpoints. The exported `selector_labels` output locates the pods for log inspection, and that is the only network-adjacent handle a Job needs.

Identity is composed, never bundled:

- **KubernetesServiceAccount** — reference it from `spec.pod.service_account` (literal name or resource reference); workload-identity annotations and pull-secret attachment live on the identity
- **KubernetesRbac** — grants permissions to that identity; the Job itself never creates RBAC objects
- **KubernetesSecret / KubernetesConfigMap** — consumed via `secret_ref` env entries, `env_from` imports, or volume mounts
- **KubernetesNamespace** — `spec.namespace` accepts a reference, so an infra chart creates the namespace and the Job in one run

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`namespace`** — the namespace the Job was created in
- **`job_name`** — the name of the Job object in the cluster
- **`selector_labels`** — the job-name label selector as a `k=v` string, ready for `kubectl get pods -l` / `kubectl logs -l` post-run inspection

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference), creating it when `create_namespace` is true
2. Materialize literal secret env values into a workload-scoped Kubernetes Secret and the registry logins declared on `spec.pod.image_registries` into an image-pull Secret
3. Build the pod template from the shared workload container/pod core: app container, sidecars, init containers, scheduling, and security hardening
4. Create the batch/v1 Job with the full set of batch controls
5. Export the namespace, Job name, and selector labels for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesJob** when you need:

- One-shot work that must run to completion: migrations, backfills, imports/exports
- Parallel processing of partitioned data with Indexed completions
- Failure-aware batch runs on clusters with spot nodes or frequent drains

**Do NOT use** when:

- The work recurs on a schedule — use **KubernetesCronJob**
- The process should run indefinitely — use **KubernetesDeployment** (or **KubernetesStatefulSet** for stateful services)

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist, be created via `create_namespace`, or be referenced as a `KubernetesNamespace` resource in the same chart

## Best Practices

1. **Set resource requests and limits**: batch work without limits can starve neighbors on shared nodes
2. **Prefer `restart_policy: Never`**: one pod per attempt keeps every failure inspectable
3. **Always set a deadline for release-blocking Jobs**: `active_deadline_seconds` turns a hang into a bounded failure
4. **Set `ttl_seconds_after_finished`**: finished Jobs otherwise accumulate until deleted manually
5. **Make sidecars exit**: a Job pod only completes when all containers exit — design sidecars to terminate with the app
6. **Classify failures with `pod_failure_policy`**: fail fast on unrecoverable exit codes, ignore node disruption
7. **Reference secrets instead of embedding them**: use `secret_ref` env entries against a `KubernetesSecret`

## References

- [Kubernetes Jobs Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Indexed Jobs for Parallel Processing](https://kubernetes.io/docs/tasks/job/indexed-parallel-processing-static/)
- [Pod Failure Policy](https://kubernetes.io/docs/concepts/workloads/controllers/job/#pod-failure-policy)
- [Job API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/job-v1/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
