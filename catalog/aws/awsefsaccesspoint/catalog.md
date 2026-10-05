# AWS EFS Access Point

Deploys an EFS access point — an application-specific entry point into an existing Elastic File System that enforces a POSIX user/group identity and pins the visible root directory. Access points are the recommended way to give Lambda functions, ECS tasks, and Batch jobs file system access: the NFS client's identity is overridden with the enforced POSIX user and the visible tree is restricted to the root directory, so applications cannot wander the file system regardless of what they run as. The entire access point is create-time immutable — only tags change in place — so the identity and root path are decisions to make up front.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EFS Access Point** -- an entry point into the referenced file system (up to 1,000 access points per file system)
- **Enforced POSIX Identity** -- configured only when `posixUser` is provided; every file operation through the access point uses this UID/GID (plus up to 16 secondary GIDs), regardless of the client's own identity
- **Root Directory Pinning** -- configured only when `rootDirectory` is provided; the given path is exposed as `/` when mounting through the access point
- **Directory Creation Info** -- configured only when `rootDirectory.creationInfo` is provided; EFS creates the root directory with the given ownership and permissions on first mount if it does not exist yet
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An EFS file system** -- the access point enters exactly one file system. Reference an [AWS Elastic File System](/infra-catalog/aws-elastic-file-system) Infra Component or provide a literal file system ID (`fs-...`).
- **A plan for the root path** -- if `rootDirectory.path` does not exist on the file system yet, provide `creationInfo` (owner UID/GID + octal permissions). Without it, mounting an access point whose path does not exist fails — AWS validates existence at mount time, not create time.

## Deploy

### Console

Open the deployment store, find **AWS EFS Access Point**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application Data Access Point** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEfsAccessPoint
metadata:
  name: app-data
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  fileSystemId:
    valueFrom:
      kind: AwsElasticFileSystem
      name: app-shared-storage
      fieldPath: status.outputs.file_system_id
  posixUser:
    uid: 1000
    gid: 1000
  rootDirectory:
    path: /app-data
    creationInfo:
      ownerUid: 1000
      ownerGid: 1000
      permissions: "0755"
```

```shell
planton apply -f efs-access-point.yaml
```

This creates an access point that enforces UID/GID 1000 for all file operations and exposes `/app-data` as the root, creating the directory with `0755` permissions on first mount. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the access point sits between the file system and its consumers — the InfraPipeline deploys the file system first, then the access point, then the Lambda function or ECS task definition that mounts it:

```yaml
spec:
  fileSystemId:
    valueFrom:
      kind: AwsElasticFileSystem
      name: app-shared-storage
      fieldPath: status.outputs.file_system_id
```

## Key Configuration

These are the most important decisions when configuring an EFS access point. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Everything is create-time immutable** -- the file system, POSIX identity, and root directory all replace the access point when changed (only tags are mutable). Consumers referencing the access point's outputs re-resolve on redeploy, but plan the identity and root path upfront.

**POSIX identity** -- when `posixUser` is set, EFS overrides the NFS client's identity: every operation uses the enforced UID/GID no matter what the client claims. This is most of the point of an access point — omit it only when clients should keep their own identity (rare). UID/GID range is 0–4294967295, with at most 16 secondary GIDs (the NFS AUTH_SYS limit).

**Root directory** -- `path` must be absolute (starts with `/`), at most 100 characters and 4 subdirectories deep. The path is exposed as `/` to everything mounting through this access point. Omit to expose the entire file system.

**Creation info** -- required in practice whenever the path does not already exist: EFS creates the directory with the given `ownerUid`/`ownerGid` and octal `permissions` (3–4 digits, e.g. `0755`) on first mount. Without it, mounts against a nonexistent path fail.

**One access point per application** -- one file system serves many applications, each behind its own access point with its own identity and directory. Pair with the file system's resource policy (condition on `elasticfilesystem:AccessPointArn`) to make the isolation IAM-enforced.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsElasticFileSystem** | `fileSystemId` | `status.outputs.file_system_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `access_point_id` | Access point identifier (`fsap-...`) | ECS task definition EFS volumes, Batch job definition volumes |
| `access_point_arn` | Amazon Resource Name of the access point | Lambda file system configuration, IAM policy conditions |
| `file_system_id` | The file system this access point enters | ECS EFS volumes, which need both the file system and access point IDs — wire both from this one node |
| `file_system_arn` | ARN of the file system | IAM policies for resource-level permissions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**App data directory** -- an access point enforcing a service account identity (UID/GID 1000) over a dedicated `/app-data` directory with `0755` permissions. The standard shape for one application's slice of a shared file system. Start from the **Application Data Access Point** preset.

**Lambda shared models** -- an access point over `/models` mounted by one or more Lambda functions at `/mnt/models` — large ML models loaded once and shared across invocations. The function references `access_point_arn`.

**ECS task volume** -- an ECS task definition EFS volume referencing `access_point_id`, with transit encryption on. Each service gets its own access point rather than sharing the file system root.

## Works With

- [**AWS Elastic File System**](/infra-catalog/aws-elastic-file-system) -- the file system this access point enters
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- mounts the access point via its file system config (`access_point_arn`)
- [**AWS ECS Task Definition**](/infra-catalog/aws-ecs-task-definition) -- mounts the access point in an EFS volume (`access_point_id`)
- [**AWS Batch Job Definition**](/infra-catalog/aws-batch-job-definition) -- mounts the access point in container EFS volumes (`access_point_id`)
