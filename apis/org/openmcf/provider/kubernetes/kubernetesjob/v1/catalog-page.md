# Kubernetes Job

Deploys a one-shot batch workload to Kubernetes as a Job with configurable parallelism, completion tracking, retry policies, environment variable and secret management, ConfigMap creation, and volume mounts. The Job runs pods to completion and then stops, making it suitable for data migrations, ETL pipelines, backup operations, and any task that must finish before the process exits.

## What Gets Created

When you deploy a KubernetesJob resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Job** — a Kubernetes batch/v1 Job with the specified container image, resource limits, parallelism, completion requirements, retry policy, and optional deadline
- **ConfigMaps** — one ConfigMap per entry in `configMaps`, available for mounting into the Job container
- **Secret** — an Opaque Secret containing environment secrets provided as direct string values, created only when `env.secrets` includes direct values
- **Image Pull Secret** — a `kubernetes.io/dockerconfigjson` Secret for pulling from private registries, created only when Docker config JSON is provided via the provider

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A container image** accessible from the cluster (public registry or with a configured image pull secret)

## Quick Start

Create a file `job.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJob
metadata:
  name: db-migrate
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesJob.db-migrate
spec:
  namespace: my-namespace
  createNamespace: true
  image:
    repo: my-registry/db-migrate
    tag: "v1.0.0"
```

Deploy:

```shell
openmcf apply -f job.yaml
```

This creates a Job that runs one pod to completion in the `my-namespace` namespace, using default resource limits (1000m CPU, 1Gi memory), a backoff limit of 6, and a restart policy of `Never`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the job. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `image.repo` | `string` | Container image repository (e.g., `my-registry/worker`, `alpine`). | Required, non-empty |
| `image.tag` | `string` | Container image tag (e.g., `latest`, `v2.0.0`). | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation. |
| `resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation. |
| `resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU. |
| `resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory. |
| `env.variables` | `map<string, StringValueOrRef>` | `{}` | Environment variables. Each value can be a direct string via `value` or a reference to another resource via `valueFrom`. |
| `env.secrets` | `map<string, KubernetesSensitiveValue>` | `{}` | Secret environment variables. Each value can be a direct string via `value` (auto-stored in a Kubernetes Secret) or a reference to an existing Kubernetes Secret via `secretRef`. |
| `parallelism` | `uint32` | `1` | Number of pods to run in parallel. Set higher for parallel batch processing. |
| `completions` | `uint32` | `1` | Number of successful pod completions required before the job is considered complete. |
| `backoffLimit` | `uint32` | `6` | Number of retries before the job is marked as failed. |
| `activeDeadlineSeconds` | `uint64` | `0` | Maximum duration in seconds for the job. The job is terminated if it exceeds this limit. `0` means no deadline. |
| `ttlSecondsAfterFinished` | `uint32` | `0` | Seconds to retain the job after completion before automatic cleanup. `0` means no automatic cleanup. |
| `completionMode` | `string` | `NonIndexed` | Completion mode. `NonIndexed`: all pods are equivalent. `Indexed`: each pod gets an index from 0 to completions-1. |
| `restartPolicy` | `string` | `Never` | Pod restart policy. Allowed values: `OnFailure`, `Never`. |
| `command` | `string[]` | `[]` | Overrides the container image ENTRYPOINT. |
| `args` | `string[]` | `[]` | Overrides the container image CMD. |
| `configMaps` | `map<string, string>` | `{}` | ConfigMaps to create alongside the job. Key is the ConfigMap name, value is the content. These can be referenced in `volumeMounts`. |
| `volumeMounts` | `VolumeMount[]` | `[]` | Volume mounts supporting ConfigMap, Secret, HostPath, EmptyDir, and PVC sources. |
| `suspend` | `bool` | `false` | When `true`, prevents pod creation. Existing pods are not affected. |

## Examples

### Database Migration

A one-shot migration job that connects to a database using a referenced host and a secret password:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJob
metadata:
  name: db-migrate
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesJob.db-migrate
spec:
  namespace: backend
  image:
    repo: my-registry/db-migrate
    tag: "v2.1.0"
  command:
    - /bin/sh
    - -c
  args:
    - "flyway -url=jdbc:postgresql://$DATABASE_HOST:5432/mydb migrate"
  env:
    variables:
      DATABASE_HOST:
        valueFrom:
          kind: KubernetesPostgres
          name: my-postgres
          field: status.outputs.service
    secrets:
      DATABASE_PASSWORD:
        secretRef:
          name: postgres-credentials
          key: password
  backoffLimit: 3
  activeDeadlineSeconds: 600
  ttlSecondsAfterFinished: 3600
```

### Parallel Batch Processing

An indexed parallel job that processes partitioned data across multiple pods, each receiving its own index via the `JOB_COMPLETION_INDEX` environment variable:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJob
metadata:
  name: batch-processor
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesJob.batch-processor
spec:
  namespace: data-pipeline
  createNamespace: true
  image:
    repo: my-registry/batch-worker
    tag: "v1.3.0"
  resources:
    limits:
      cpu: "2000m"
      memory: "4Gi"
    requests:
      cpu: "500m"
      memory: "1Gi"
  env:
    variables:
      INPUT_BUCKET:
        value: "s3://my-bucket/input"
      OUTPUT_BUCKET:
        value: "s3://my-bucket/output"
    secrets:
      AWS_SECRET_ACCESS_KEY:
        secretRef:
          name: aws-credentials
          key: secret-access-key
  parallelism: 5
  completions: 20
  completionMode: Indexed
  backoffLimit: 3
  restartPolicy: OnFailure
  activeDeadlineSeconds: 7200
```

### Script Job with ConfigMap and Volume Mount

A job that mounts a user-defined script from a ConfigMap and executes it, with an emptyDir volume for scratch space:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesJob
metadata:
  name: etl-pipeline
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesJob.etl-pipeline
spec:
  namespace: data-ops
  image:
    repo: python
    tag: "3.12-slim"
  command:
    - /bin/sh
    - -c
  args:
    - "pip install pandas && python /scripts/etl.py"
  resources:
    limits:
      cpu: "4000m"
      memory: "8Gi"
    requests:
      cpu: "1000m"
      memory: "2Gi"
  env:
    variables:
      SOURCE_DB:
        valueFrom:
          kind: KubernetesPostgres
          name: source-db
          field: status.outputs.service
      TARGET_DB:
        valueFrom:
          kind: KubernetesPostgres
          name: target-db
          field: status.outputs.service
  configMaps:
    etl-script: |
      import pandas as pd
      import os
      source = os.environ["SOURCE_DB"]
      target = os.environ["TARGET_DB"]
      print(f"Extracting from {source}, loading to {target}")
      # ... ETL logic ...
  volumeMounts:
    - name: etl-script
      mountPath: /scripts/etl.py
      subPath: etl.py
      configMap:
        name: etl-script
        key: etl-script
        path: etl.py
    - name: scratch
      mountPath: /tmp/scratch
      emptyDir:
        sizeLimit: "10Gi"
  backoffLimit: 2
  activeDeadlineSeconds: 3600
  ttlSecondsAfterFinished: 86400
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the job is created |
| `jobName` | `string` | Name of the Kubernetes Job resource |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesCronJob](/docs/catalog/kubernetes/kubernetescronjob) — runs jobs on a recurring schedule rather than as a one-shot execution
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — runs containers continuously as long-lived services
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — commonly referenced for database connection environment variables
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — commonly referenced for cache connection environment variables
