# KubernetesJob: Technical Research Documentation

## Introduction

Kubernetes Jobs are a fundamental workload controller that runs pods to completion, making them the building blocks for batch processing in Kubernetes. Unlike Deployments that maintain a desired number of running pods indefinitely, or CronJobs that trigger on a schedule, Jobs create pods that execute a task and then stop.

This document provides comprehensive research into Kubernetes Jobs, their deployment landscape, implementation approaches, and the design decisions behind Planton's KubernetesJob component.

## What is a Kubernetes Job?

A Kubernetes Job creates one or more pods and ensures that a specified number of them successfully terminate. When a successful number of completions is reached, the job is complete. Jobs can run:

1. **Single Pods**: One pod runs to completion (default behavior)
2. **Parallel Pods (Work Queue)**: Multiple pods process a shared queue until empty
3. **Parallel Pods (Fixed Count)**: A specific number of pods each process a portion of work
4. **Indexed Parallel Jobs**: Each pod gets a unique index for processing partitioned data

### Core Concepts

**Completion**: A job is complete when the required number of pods have successfully terminated (exit code 0). The `completions` field specifies how many successful completions are needed.

**Parallelism**: The `parallelism` field specifies the maximum number of pods that can run simultaneously. For work queue patterns, this is typically less than completions.

**Backoff Limit**: The `backoffLimit` field specifies how many times Kubernetes retries creating a pod before marking the job as failed. Each retry uses exponential backoff.

**Active Deadline**: The `activeDeadlineSeconds` field sets an absolute deadline for the job. If the job runs longer, it's terminated and marked as failed.

**TTL After Finished**: The `ttlSecondsAfterFinished` field enables automatic cleanup of completed jobs after a specified duration.

### Job Completion Modes

Kubernetes 1.21+ introduced completion modes:

1. **NonIndexed (default)**: All pods are interchangeable. The job completes when `.spec.completions` pods succeed.

2. **Indexed**: Each pod gets a unique index (0 to completions-1) via the `JOB_COMPLETION_INDEX` environment variable. The job completes when each index has exactly one successful pod.

Indexed mode is particularly useful for:
- Processing partitioned datasets
- Sharded database operations
- Parallel processing with explicit coordination

## Deployment Landscape

### Manual Deployment Methods

**kubectl CLI**:
```bash
kubectl create job my-job --image=busybox -- echo "Hello"
kubectl get jobs
kubectl describe job my-job
kubectl logs job/my-job
kubectl delete job my-job
```

**YAML Manifests**:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: my-job
spec:
  template:
    spec:
      containers:
      - name: main
        image: busybox
        command: ["echo", "Hello"]
      restartPolicy: Never
  backoffLimit: 4
```

### Infrastructure-as-Code Tools

**Terraform (kubernetes provider)**:
```hcl
resource "kubernetes_job" "example" {
  metadata {
    name = "my-job"
  }
  spec {
    template {
      spec {
        container {
          name    = "main"
          image   = "busybox"
          command = ["echo", "Hello"]
        }
        restart_policy = "Never"
      }
    }
    backoff_limit = 4
  }
}
```

**Pulumi (Go)**:
```go
job, err := batchv1.NewJob(ctx, "my-job", &batchv1.JobArgs{
    Spec: &batchv1.JobSpecArgs{
        Template: &corev1.PodTemplateSpecArgs{
            Spec: &corev1.PodSpecArgs{
                Containers: corev1.ContainerArray{
                    &corev1.ContainerArgs{
                        Name:    pulumi.String("main"),
                        Image:   pulumi.String("busybox"),
                        Command: pulumi.StringArray{pulumi.String("echo"), pulumi.String("Hello")},
                    },
                },
                RestartPolicy: pulumi.String("Never"),
            },
        },
        BackoffLimit: pulumi.Int(4),
    },
})
```

### Specialized Tools

**Helm Charts**: Many applications include job templates for migrations, setup tasks, or backups.

**Argo Workflows**: Extends Jobs into complex DAG-based workflows with dependencies.

**Tekton**: CI/CD focused task runner built on Kubernetes primitives.

**Kueue**: Kubernetes-native job queueing system for batch workloads.

## Comparative Analysis

| Aspect | kubectl/YAML | Terraform | Pulumi | Planton |
|--------|--------------|-----------|--------|-----------------|
| Learning Curve | Low | Medium | Medium-High | Low |
| Type Safety | None | Limited (HCL) | Full (Go) | Full (Protobuf) |
| Multi-Cluster | Manual | Via providers | Via providers | Built-in |
| Secrets Management | External | Via providers | Via providers | Integrated |
| State Management | None | Remote state | Remote state | Integrated |
| Validation | kubectl dry-run | terraform validate | Compile-time | Protobuf + CEL |
| Documentation | Manual | terraform-docs | Code comments | Auto-generated |

## Planton's Approach

### Design Philosophy

Planton's KubernetesJob follows the 80/20 principle, exposing the configuration options that address the most common use cases while providing sensible defaults for advanced settings.

### Key Design Decisions

1. **Unified Container Model**: Uses the same container image, resources, and environment variable patterns as KubernetesDeployment and KubernetesCronJob for consistency.

2. **Foreign Key References**: Environment variables can reference outputs from other Planton resources, enabling dynamic configuration.

3. **Secret References**: Supports both direct secret values (for development) and Kubernetes Secret references (for production).

4. **Volume Mounts**: Supports ConfigMaps, Secrets, PVCs, HostPaths, and EmptyDirs with a unified interface.

5. **Sensible Defaults**:
   - `parallelism: 1` - Sequential execution by default
   - `completions: 1` - Single completion by default
   - `backoffLimit: 6` - Standard Kubernetes default
   - `restartPolicy: Never` - Job-level retries preferred over pod-level

### Fields Included (80% Use Cases)

| Field | Purpose |
|-------|---------|
| `namespace` | Target namespace with reference support |
| `createNamespace` | Optionally create namespace |
| `image` | Container image configuration |
| `resources` | CPU and memory limits/requests |
| `env` | Environment variables and secrets |
| `parallelism` | Concurrent pod count |
| `completions` | Required successful completions |
| `backoffLimit` | Retry count before failure |
| `activeDeadlineSeconds` | Maximum job duration |
| `ttlSecondsAfterFinished` | Automatic cleanup timer |
| `completionMode` | NonIndexed or Indexed |
| `restartPolicy` | Never or OnFailure |
| `command` / `args` | Container entry point override |
| `configMaps` | Create ConfigMaps for the job |
| `volumeMounts` | Mount various volume types |
| `suspend` | Pause job creation |

### Fields Excluded (Advanced/Rare)

| Field | Reason for Exclusion |
|-------|----------------------|
| `selector` | Auto-generated, rarely customized |
| `manualSelector` | Advanced use case |
| `podFailurePolicy` | Complex, Kubernetes 1.26+ |
| `successPolicy` | Complex, Kubernetes 1.30+ |
| `backoffLimitPerIndex` | Indexed mode advanced config |
| `maxFailedIndexes` | Indexed mode advanced config |
| `podReplacementPolicy` | Advanced pod scheduling |
| `managedBy` | External controller integration |

## Implementation Architecture

### Resource Creation Flow

```
User Manifest → Orchestrator → Stack Input → IaC Module → Kubernetes API
                    ↓
            Resolve References
            Apply Defaults
            Validate Schema
```

### Created Kubernetes Resources

1. **Namespace** (optional): If `createNamespace: true`
2. **ConfigMaps**: From `spec.configMaps`
3. **Secret** (internal): For `env.secrets` with direct values
4. **Image Pull Secret** (optional): If Docker credentials provided
5. **ServiceAccount**: For pod identity
6. **Job**: The main batch workload

### Output Values

| Output | Description |
|--------|-------------|
| `namespace` | Kubernetes namespace name |
| `job_name` | Created job name |

## Best Practices

### Resource Management

1. **Always Set Limits**: Jobs can consume significant resources; set CPU and memory limits
2. **Set Active Deadline**: Prevent runaway jobs with `activeDeadlineSeconds`
3. **Enable TTL Cleanup**: Use `ttlSecondsAfterFinished` to prevent job accumulation

### Reliability

1. **Configure Backoff**: Set appropriate `backoffLimit` for transient failures
2. **Choose Restart Policy Carefully**:
   - `Never`: Job controller handles retries (new pod each attempt)
   - `OnFailure`: kubelet restarts container in same pod (preserves local state)
3. **Use Indexed Mode**: For partitioned processing with failure isolation

### Security

1. **Use Secret References**: Avoid direct secret values in manifests
2. **Limit Namespace Access**: Run jobs in dedicated namespaces
3. **Use Service Accounts**: Configure appropriate RBAC permissions

### Observability

1. **Log Aggregation**: Ensure job logs are captured before pod cleanup
2. **Set Meaningful Names**: Use descriptive job names for debugging
3. **Monitor Job Metrics**: Track job success/failure rates

## Comparison: Job vs CronJob

| Aspect | Job | CronJob |
|--------|-----|---------|
| Trigger | Immediate on creation | Schedule-based |
| Use Case | One-time tasks | Recurring tasks |
| Cleanup | Manual or TTL | History limits |
| Concurrency | N/A | Allow/Forbid/Replace |
| Schedule | N/A | Cron expression |

Choose **Job** when:
- Task is triggered by an event (deployment, user action)
- Exact timing isn't important
- Task should run once

Choose **CronJob** when:
- Task needs to run on a schedule
- Recurring execution is required
- Time-based triggering is needed

## Conclusion

Planton's KubernetesJob component provides a streamlined, type-safe interface for deploying batch workloads to Kubernetes. By focusing on the 80/20 of configuration options while maintaining consistency with other Kubernetes workload components, it enables platform teams to standardize job deployments across their infrastructure.

The integration with Planton's resource reference system allows jobs to dynamically reference configuration from other resources, while the dual IaC support (Pulumi and Terraform) ensures flexibility in deployment tooling preferences.

## References

- [Kubernetes Jobs Documentation](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [Kubernetes Indexed Jobs](https://kubernetes.io/docs/tasks/job/indexed-parallel-processing-static/)
- [Kubernetes TTL Controller](https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/)
- [Batch Processing Patterns](https://kubernetes.io/docs/concepts/workloads/controllers/job/#job-patterns)
