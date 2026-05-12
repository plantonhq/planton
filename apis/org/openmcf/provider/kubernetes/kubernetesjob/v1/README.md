# KubernetesJob

## Overview

KubernetesJob is a OpenMCF component that deploys one-shot batch workloads to Kubernetes clusters. Unlike Deployments that run continuously or CronJobs that run on a schedule, a Job executes immediately when created and runs pods to completion before stopping.

Jobs are the building blocks for batch processing in Kubernetes, providing a reliable way to run tasks that need to complete successfully a specific number of times.

## Purpose

KubernetesJob simplifies the deployment of batch workloads by:

- **Declarative Configuration**: Define your job's container, resources, and execution parameters in a simple YAML manifest
- **Automatic Retries**: Configure backoff limits for automatic retry on failure
- **Parallel Execution**: Run multiple pods in parallel for faster batch processing
- **Indexed Jobs**: Assign unique indexes to pods for partitioned data processing
- **Automatic Cleanup**: Optionally delete jobs after completion with TTL settings
- **Consistent Experience**: Use the same OpenMCF workflow for jobs as other workloads

## Key Features

- **One-Shot Execution**: Runs pods to completion, ideal for tasks that shouldn't run indefinitely
- **Parallelism Control**: Configure how many pods run simultaneously
- **Completion Tracking**: Specify how many successful completions are required
- **Failure Handling**: Set backoff limits and restart policies for resilience
- **Deadline Enforcement**: Set maximum runtime to prevent runaway jobs
- **TTL-Based Cleanup**: Automatically delete completed jobs after a specified time
- **Environment Variables**: Support for both direct values and secret references
- **Volume Mounts**: Mount ConfigMaps, Secrets, PVCs, and more

## Use Cases

- **Database Migrations**: Run schema migrations during deployments
- **Data Processing**: Process batches of data, files, or messages
- **ETL Jobs**: Extract, transform, and load data between systems
- **Backup Operations**: Create backups of databases or file systems
- **One-Time Setup**: Initialize resources or configurations
- **Report Generation**: Generate periodic reports or analytics
- **Cleanup Tasks**: Remove old data or resources

## Example Usage

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJob
metadata:
  name: data-migration
spec:
  namespace:
    value: batch-jobs
  createNamespace: true
  image:
    repo: myregistry/migration-runner
    tag: v1.2.0
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 500m
      memory: 1Gi
  env:
    variables:
      - name: DATABASE_URL
        value: postgres://localhost:5432/mydb
      - name: BATCH_SIZE
        value: "1000"
    secrets:
      - name: DATABASE_PASSWORD
        secretRef:
          name: db-credentials
          key: password
  backoffLimit: 3
  activeDeadlineSeconds: 3600
  ttlSecondsAfterFinished: 86400
  command:
    - python
    - /app/migrate.py
```

Deploy with:

```bash
openmcf pulumi up --manifest job.yaml
```

## Best Practices

1. **Set Resource Limits**: Always specify CPU and memory limits to prevent resource exhaustion
2. **Use Backoff Limits**: Configure appropriate retry counts for transient failures
3. **Set Deadlines**: Use `activeDeadlineSeconds` to prevent stuck jobs
4. **Enable TTL Cleanup**: Set `ttlSecondsAfterFinished` to automatically clean up completed jobs
5. **Use Indexed Mode**: For parallel processing of partitioned data, use `completionMode: Indexed`
6. **Reference Secrets**: Use `secretRef` for sensitive values instead of direct strings
7. **Choose Restart Policy**: Use `OnFailure` for in-pod retries, `Never` for job-level retries

## Related Components

- [KubernetesCronJob](../kubernetescronjob/v1/README.md) - For scheduled recurring jobs
- [KubernetesDeployment](../kubernetesdeployment/v1/README.md) - For long-running services
- [KubernetesStatefulSet](../kubernetesstatefulset/v1/README.md) - For stateful applications
