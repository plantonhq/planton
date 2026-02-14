---
title: "CronJob"
description: "CronJob deployment documentation"
icon: "package"
order: 100
componentName: "kubernetescronjob"
---

# Kubernetes CronJob

Deploys a container as a Kubernetes CronJob with configurable scheduling, concurrency control, retry policies, environment variable and secret management, ConfigMap creation, and volume mounts. OpenMCF handles the creation of all supporting resources (namespace, secrets, image pull secrets, ConfigMaps) alongside the CronJob itself.

## What Gets Created

When you deploy a KubernetesCronJob resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **CronJob** — a Kubernetes CronJob with the specified container image, schedule, concurrency policy, resource limits, volume mounts, and restart policy
- **Secret** — an Opaque Secret containing environment secrets provided as direct string values, created only when `env.secrets` includes direct values
- **Image Pull Secret** — a `kubernetes.io/dockerconfigjson` Secret for pulling from private registries, created only when a Docker config JSON is provided
- **ConfigMaps** — one ConfigMap per entry in `configMaps`, available for mounting into the CronJob container
- **Auto-injected environment variables** — `HOSTNAME` (set to the pod IP) and `K8S_POD_ID` (set to the pod name) are added to every CronJob container automatically

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A container image** accessible from the cluster (public registry or with a configured image pull secret)

## Quick Start

Create a file `cronjob.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCronJob
metadata:
  name: my-cronjob
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesCronJob.my-cronjob
spec:
  namespace: my-namespace
  createNamespace: true
  schedule: "0 * * * *"
  image:
    repo: busybox
    tag: "1.36"
  command:
    - /bin/sh
    - -c
  args:
    - "echo Hello from CronJob"
```

Deploy:

```shell
openmcf apply -f cronjob.yaml
```

This creates a CronJob in the `my-namespace` namespace that runs every hour using a busybox container, with default resource limits (1000m CPU, 1Gi memory) and a `Forbid` concurrency policy.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the CronJob. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `schedule` | `string` | Cron schedule expression in standard cron format (e.g., `"0 0 * * *"`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `image.repo` | `string` | — | Container image repository (e.g., `busybox`, `gcr.io/project/image`). |
| `image.tag` | `string` | — | Container image tag (e.g., `latest`, `1.36`). |
| `image.pullSecretName` | `string` | — | Name of an existing image pull secret in the namespace. |
| `resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation. |
| `resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation. |
| `resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU. |
| `resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory. |
| `env.variables` | `map<string, StringValueOrRef>` | `{}` | Environment variables. Each value can be a direct string via `value` or a reference to another resource via `valueFrom`. |
| `env.secrets` | `map<string, KubernetesSensitiveValue>` | `{}` | Secret environment variables. Each value can be a direct string via `value` (auto-stored in a Kubernetes Secret) or a reference to an existing Kubernetes Secret via `secretRef`. |
| `concurrencyPolicy` | `string` | `Forbid` | How concurrent job runs are handled. Valid values: `Allow`, `Forbid`, `Replace`. |
| `startingDeadlineSeconds` | `uint64` | `0` | Deadline in seconds for starting the job if it misses its scheduled time. `0` means no deadline. |
| `suspend` | `bool` | `false` | When `true`, no subsequent runs are scheduled. |
| `successfulJobsHistoryLimit` | `uint32` | `3` | Number of successful finished jobs to retain. |
| `failedJobsHistoryLimit` | `uint32` | `1` | Number of failed finished jobs to retain. |
| `backoffLimit` | `uint32` | `6` | Number of retries before marking the job as failed. |
| `restartPolicy` | `string` | `Never` | Pod restart policy. Valid values: `Always`, `OnFailure`, `Never`. |
| `command` | `string[]` | `[]` | Overrides the container image's ENTRYPOINT. |
| `args` | `string[]` | `[]` | Overrides the container image's CMD. |
| `configMaps` | `map<string, string>` | `{}` | ConfigMaps to create alongside the CronJob. Key is the ConfigMap name, value is the content. These can be referenced in `volumeMounts`. |
| `volumeMounts` | `VolumeMount[]` | `[]` | Volume mounts supporting ConfigMap, Secret, HostPath, EmptyDir, and PVC sources. |

## Examples

### Periodic Log Cleanup

A CronJob that runs daily at midnight to clean up old log files:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCronJob
metadata:
  name: log-cleanup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesCronJob.log-cleanup
spec:
  namespace: maintenance
  createNamespace: true
  schedule: "0 0 * * *"
  image:
    repo: busybox
    tag: "1.36"
  command:
    - /bin/sh
    - -c
  args:
    - "find /var/log -name '*.log' -mtime +7 -delete && echo 'Cleanup complete'"
  resources:
    limits:
      cpu: "200m"
      memory: "128Mi"
    requests:
      cpu: "50m"
      memory: "64Mi"
```

### Database Backup with Environment Variables and Secrets

A CronJob that performs a nightly database backup, using `valueFrom` to resolve the database host from another OpenMCF resource and `secretRef` to retrieve the password from an existing Kubernetes Secret:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCronJob
metadata:
  name: db-backup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesCronJob.db-backup
spec:
  namespace: backups
  schedule: "0 2 * * *"
  image:
    repo: postgres
    tag: "16-alpine"
  resources:
    limits:
      cpu: "500m"
      memory: "512Mi"
    requests:
      cpu: "100m"
      memory: "256Mi"
  command:
    - /bin/sh
    - -c
  args:
    - "pg_dump -h $DATABASE_HOST -U $DATABASE_USER -d $DATABASE_NAME > /tmp/backup.sql && echo 'Backup complete'"
  env:
    variables:
      DATABASE_HOST:
        valueFrom:
          kind: KubernetesPostgres
          name: my-postgres
          field: status.outputs.service
      DATABASE_NAME:
        value: "app_production"
      DATABASE_USER:
        value: "backup_user"
      BACKUP_RETENTION_DAYS:
        value: "30"
    secrets:
      DATABASE_PASSWORD:
        secretRef:
          name: postgres-credentials
          key: password
  concurrencyPolicy: Forbid
  backoffLimit: 3
  restartPolicy: OnFailure
```

### Scheduled Script with ConfigMap and Volume Mounts

A CronJob that runs a backup script stored in a ConfigMap, with a PVC for persistent backup storage and tuned scheduling parameters:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCronJob
metadata:
  name: scheduled-backup
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCronJob.scheduled-backup
spec:
  namespace: production
  schedule: "0 3 * * 0"
  image:
    repo: gcr.io/my-project/backup-runner
    tag: "v2.0.0"
  resources:
    limits:
      cpu: "2000m"
      memory: "2Gi"
    requests:
      cpu: "500m"
      memory: "512Mi"
  command:
    - /bin/bash
    - /scripts/backup.sh
  env:
    variables:
      BACKUP_BUCKET:
        value: "s3://my-backups"
      NOTIFICATION_URL:
        value: "https://hooks.slack.com/services/XXX"
    secrets:
      AWS_ACCESS_KEY_ID:
        secretRef:
          name: aws-credentials
          key: access-key-id
      AWS_SECRET_ACCESS_KEY:
        secretRef:
          name: aws-credentials
          key: secret-access-key
  configMaps:
    backup-script: |
      #!/bin/bash
      set -euo pipefail
      echo "Starting weekly backup..."
      pg_dumpall > /backup/full-dump.sql
      aws s3 cp /backup/full-dump.sql $BACKUP_BUCKET/$(date +%Y-%m-%d).sql
      echo "Backup uploaded successfully"
  volumeMounts:
    - name: backup-script
      mountPath: /scripts/backup.sh
      subPath: backup.sh
      configMap:
        name: backup-script
        key: backup-script
        path: backup.sh
        defaultMode: 493
    - name: backup-storage
      mountPath: /backup
      pvc:
        claimName: backup-pvc
  concurrencyPolicy: Forbid
  suspend: false
  startingDeadlineSeconds: 600
  successfulJobsHistoryLimit: 5
  failedJobsHistoryLimit: 3
  backoffLimit: 2
  restartPolicy: Never
```

### Suspended CronJob for Manual Triggering

A CronJob defined in a suspended state, ready to be un-suspended or manually triggered as needed:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCronJob
metadata:
  name: data-migration
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesCronJob.data-migration
spec:
  namespace: migrations
  schedule: "0 0 * * *"
  image:
    repo: gcr.io/my-project/migrator
    tag: "v1.4.0"
  args:
    - "--source=old-db"
    - "--target=new-db"
    - "--batch-size=1000"
  env:
    variables:
      SOURCE_DB_HOST:
        valueFrom:
          kind: KubernetesPostgres
          name: old-postgres
          field: status.outputs.service
      TARGET_DB_HOST:
        valueFrom:
          kind: KubernetesPostgres
          name: new-postgres
          field: status.outputs.service
    secrets:
      SOURCE_DB_PASSWORD:
        value: "temp-migration-password"
      TARGET_DB_PASSWORD:
        secretRef:
          name: new-db-credentials
          key: password
  suspend: true
  concurrencyPolicy: Forbid
  backoffLimit: 0
  restartPolicy: Never
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the CronJob is created |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — for long-running workloads instead of scheduled jobs
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — commonly referenced for database connection environment variables
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — commonly referenced for cache connection environment variables
- [KubernetesSecret](/docs/catalog/kubernetes/secret) — for managing secrets that CronJobs reference via `secretRef`
