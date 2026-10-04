# AwsBatchJobDefinition

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsBatchJobDefinitionSpec defines an AWS Batch job definition: the
versioned container blueprint jobs are submitted from -- image, command,
sizing, IAM identities, retries, and timeout.

A job definition is the workload half of the Batch graph: the compute
environment provides capacity, the job queue routes, and the job
definition describes WHAT runs. It is referenced at submission time
(SubmitJob) and by EventBridge Batch targets, which is why it is a
first-class resource.

Job definitions are revisioned and revisions are immutable in AWS: every
change to this spec REGISTERS A NEW REVISION of the name rather than
mutating the old one. The name comes from metadata.name. Because the
job_definition_arn output carries the revision, an EventBridge rule
that references it by output picks up each new revision on its next
deployment -- "change the image tag, the schedule runs the new code"
falls out of the composition.

This kind models type "container" job definitions in both of their
workload arms: single-container ECS-based jobs (containerProperties --
the shape nearly every Batch workload uses, for EC2 and Fargate) and
Batch-on-EKS pod jobs (eksProperties -- the workload half of an
EKS-attached compute environment). Exactly one arm is set per
definition. Multi-node parallel jobs (nodeProperties, type "multinode")
and multi-container ECS jobs (ecsProperties) remain separate long-tail
workload shapes, deliberately not modeled.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBatchJobDefinition
metadata:
  name: test-batch-etl
  id: test-batch-etl
  org: test-org
  env: dev
  annotations:
    planton.dev/provisioner: pulumi
spec:
  region: us-west-2
  container:
    image: public.ecr.aws/amazonlinux/amazonlinux:2023
    command:
      - echo
      - hello
    vcpus: 1
    memoryMib: 2048
    environment:
      STAGE: dev
  retryStrategy:
    attempts: 3
    evaluateOnExit:
      - action: RETRY
        onStatusReason: Host EC2*
      - action: EXIT
        onExitCode: "1*"
  timeout:
    attemptDurationSeconds: 3600
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.container` | `AwsBatchJobDefinitionContainer` |  |  |  |
| `spec.container.image` | `string` | yes |  |  |
| `spec.container.command` | `[]string` |  |  |  |
| `spec.container.vcpus` | `double` |  |  |  |
| `spec.container.memoryMib` | `int32` |  |  |  |
| `spec.container.gpus` | `int32` |  |  |  |
| `spec.container.jobRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.container.executionRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.container.environment` | `map<string, string>` |  |  |  |
| `spec.container.secrets` | `map<string, string>` |  |  |  |
| `spec.container.logConfiguration` | `AwsBatchJobDefinitionLogConfiguration` |  |  |  |
| `spec.container.logConfiguration.logDriver` | `string` | yes |  |  |
| `spec.container.logConfiguration.options` | `map<string, string>` |  |  |  |
| `spec.container.logConfiguration.secretOptions` | `map<string, string>` |  |  |  |
| `spec.container.mountPoints` | `[]AwsBatchJobDefinitionMountPoint` |  |  |  |
| `spec.container.mountPoints[].sourceVolume` | `string` | yes |  |  |
| `spec.container.mountPoints[].containerPath` | `string` | yes |  |  |
| `spec.container.mountPoints[].readOnly` | `bool` |  |  |  |
| `spec.container.volumes` | `[]AwsBatchJobDefinitionVolume` |  |  |  |
| `spec.container.volumes[].name` | `string` | yes |  |  |
| `spec.container.volumes[].efs` | `AwsBatchJobDefinitionEfsVolume` |  |  |  |
| `spec.container.volumes[].efs.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.container.volumes[].efs.rootDirectory` | `string` |  |  |  |
| `spec.container.volumes[].efs.accessPointId` | `string \| valueFrom` |  |  | AwsEfsAccessPoint (`status.outputs.access_point_id`) |
| `spec.container.volumes[].efs.iamAuthorization` | `bool` |  |  |  |
| `spec.container.volumes[].hostPath` | `string` |  |  |  |
| `spec.container.ulimits` | `[]AwsBatchJobDefinitionUlimit` |  |  |  |
| `spec.container.ulimits[].name` | `string` | yes |  |  |
| `spec.container.ulimits[].softLimit` | `int32` |  |  |  |
| `spec.container.ulimits[].hardLimit` | `int32` |  |  |  |
| `spec.container.linuxParameters` | `AwsBatchJobDefinitionLinuxParameters` |  |  |  |
| `spec.container.linuxParameters.initProcessEnabled` | `bool` |  |  |  |
| `spec.container.linuxParameters.devices` | `[]AwsBatchJobDefinitionDevice` |  |  |  |
| `spec.container.linuxParameters.devices[].hostPath` | `string` | yes |  |  |
| `spec.container.linuxParameters.devices[].containerPath` | `string` |  |  |  |
| `spec.container.linuxParameters.devices[].permissions` | `[]string` |  |  |  |
| `spec.container.linuxParameters.sharedMemorySizeMib` | `int32` |  |  |  |
| `spec.container.linuxParameters.maxSwapMib` | `int32` |  |  |  |
| `spec.container.linuxParameters.swappiness` | `int32` |  |  |  |
| `spec.container.linuxParameters.tmpfs` | `[]AwsBatchJobDefinitionTmpfs` |  |  |  |
| `spec.container.linuxParameters.tmpfs[].containerPath` | `string` | yes |  |  |
| `spec.container.linuxParameters.tmpfs[].sizeMib` | `int32` |  |  |  |
| `spec.container.linuxParameters.tmpfs[].mountOptions` | `[]string` |  |  |  |
| `spec.container.privileged` | `bool` |  |  |  |
| `spec.container.user` | `string` |  |  |  |
| `spec.container.readonlyRootFilesystem` | `bool` |  |  |  |
| `spec.container.repositoryCredentialsSecretArn` | `string` |  |  |  |
| `spec.container.runtimePlatform` | `AwsBatchJobDefinitionRuntimePlatform` |  |  |  |
| `spec.container.runtimePlatform.cpuArchitecture` | `string` |  |  |  |
| `spec.container.runtimePlatform.operatingSystemFamily` | `string` |  |  |  |
| `spec.container.fargatePlatformVersion` | `string` |  |  |  |
| `spec.container.assignPublicIp` | `bool` |  |  |  |
| `spec.container.ephemeralStorageGib` | `int32` |  |  |  |
| `spec.platformCapabilities` | `[]string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.retryStrategy` | `AwsBatchJobDefinitionRetryStrategy` |  |  |  |
| `spec.retryStrategy.attempts` | `int32` |  |  |  |
| `spec.retryStrategy.evaluateOnExit` | `[]AwsBatchJobDefinitionEvaluateOnExit` |  |  |  |
| `spec.retryStrategy.evaluateOnExit[].action` | `string` | yes |  |  |
| `spec.retryStrategy.evaluateOnExit[].onExitCode` | `string` |  |  |  |
| `spec.retryStrategy.evaluateOnExit[].onReason` | `string` |  |  |  |
| `spec.retryStrategy.evaluateOnExit[].onStatusReason` | `string` |  |  |  |
| `spec.timeout` | `AwsBatchJobDefinitionTimeout` |  |  |  |
| `spec.timeout.attemptDurationSeconds` | `int32` |  |  |  |
| `spec.schedulingPriority` | `int32` |  |  |  |
| `spec.propagateTags` | `bool` |  |  |  |
| `spec.deregisterOnNewRevision` | `bool` |  | `true` |  |
| `spec.eks` | `AwsBatchJobDefinitionEks` |  |  |  |
| `spec.eks.containers` | `[]AwsBatchJobDefinitionEksContainer` | yes |  |  |
| `spec.eks.containers[].image` | `string` | yes |  |  |
| `spec.eks.containers[].name` | `string` |  |  |  |
| `spec.eks.containers[].command` | `[]string` |  |  |  |
| `spec.eks.containers[].args` | `[]string` |  |  |  |
| `spec.eks.containers[].env` | `map<string, string>` |  |  |  |
| `spec.eks.containers[].imagePullPolicy` | `string` |  |  |  |
| `spec.eks.containers[].resources` | `AwsBatchJobDefinitionEksResources` |  |  |  |
| `spec.eks.containers[].resources.limits` | `map<string, string>` |  |  |  |
| `spec.eks.containers[].resources.requests` | `map<string, string>` |  |  |  |
| `spec.eks.containers[].securityContext` | `AwsBatchJobDefinitionEksSecurityContext` |  |  |  |
| `spec.eks.containers[].securityContext.runAsUser` | `int64` |  |  |  |
| `spec.eks.containers[].securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.eks.containers[].securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.eks.containers[].securityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.eks.containers[].securityContext.privileged` | `bool` |  |  |  |
| `spec.eks.containers[].securityContext.readOnlyRootFileSystem` | `bool` |  |  |  |
| `spec.eks.containers[].volumeMounts` | `[]AwsBatchJobDefinitionEksVolumeMount` |  |  |  |
| `spec.eks.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.eks.containers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.eks.containers[].volumeMounts[].readOnly` | `bool` |  |  |  |
| `spec.eks.initContainers` | `[]AwsBatchJobDefinitionEksContainer` |  |  |  |
| `spec.eks.initContainers[].image` | `string` | yes |  |  |
| `spec.eks.initContainers[].name` | `string` |  |  |  |
| `spec.eks.initContainers[].command` | `[]string` |  |  |  |
| `spec.eks.initContainers[].args` | `[]string` |  |  |  |
| `spec.eks.initContainers[].env` | `map<string, string>` |  |  |  |
| `spec.eks.initContainers[].imagePullPolicy` | `string` |  |  |  |
| `spec.eks.initContainers[].resources` | `AwsBatchJobDefinitionEksResources` |  |  |  |
| `spec.eks.initContainers[].resources.limits` | `map<string, string>` |  |  |  |
| `spec.eks.initContainers[].resources.requests` | `map<string, string>` |  |  |  |
| `spec.eks.initContainers[].securityContext` | `AwsBatchJobDefinitionEksSecurityContext` |  |  |  |
| `spec.eks.initContainers[].securityContext.runAsUser` | `int64` |  |  |  |
| `spec.eks.initContainers[].securityContext.runAsGroup` | `int64` |  |  |  |
| `spec.eks.initContainers[].securityContext.runAsNonRoot` | `bool` |  |  |  |
| `spec.eks.initContainers[].securityContext.allowPrivilegeEscalation` | `bool` |  |  |  |
| `spec.eks.initContainers[].securityContext.privileged` | `bool` |  |  |  |
| `spec.eks.initContainers[].securityContext.readOnlyRootFileSystem` | `bool` |  |  |  |
| `spec.eks.initContainers[].volumeMounts` | `[]AwsBatchJobDefinitionEksVolumeMount` |  |  |  |
| `spec.eks.initContainers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.eks.initContainers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.eks.initContainers[].volumeMounts[].readOnly` | `bool` |  |  |  |
| `spec.eks.hostNetwork` | `bool` |  |  |  |
| `spec.eks.dnsPolicy` | `string` |  |  |  |
| `spec.eks.serviceAccountName` | `string` |  |  |  |
| `spec.eks.podLabels` | `map<string, string>` |  |  |  |
| `spec.eks.imagePullSecretNames` | `[]string` |  |  |  |
| `spec.eks.shareProcessNamespace` | `bool` |  |  |  |
| `spec.eks.volumes` | `[]AwsBatchJobDefinitionEksVolume` |  |  |  |
| `spec.eks.volumes[].name` | `string` | yes |  |  |
| `spec.eks.volumes[].emptyDir` | `AwsBatchJobDefinitionEksEmptyDir` |  |  |  |
| `spec.eks.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.eks.volumes[].emptyDir.sizeLimit` | `string` | yes |  |  |
| `spec.eks.volumes[].hostPath` | `string` |  |  |  |
| `spec.eks.volumes[].secret` | `AwsBatchJobDefinitionEksSecretVolume` |  |  |  |
| `spec.eks.volumes[].secret.secretName` | `string` | yes |  |  |
| `spec.eks.volumes[].secret.optional` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the job definition is registered in. A queue can only
run definitions registered in its own region.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.container

`AwsBatchJobDefinitionContainer`

The ECS-based container the job runs: image, command, sizing,
identities, logging, and storage. Exactly one of container or eks is
set -- this arm targets EC2/Fargate compute environments.

- rule: environment variable names must not start with 'AWS_BATCH' -- that prefix is reserved for variables AWS Batch sets on every job
- rule: ephemeral_storage_gib must be between 21 and 200 when set (Fargate includes 20 GiB by default)
- rule: every mount_points entry must reference the name of a volume declared in volumes

### spec.container.image

`string` · required

The container image, as a full reference: "<repository>:<tag>" or
"<repository>@<digest>". Up to 255 characters. Private ECR images
require execution_role (Fargate) or the instance role (EC2) to carry
pull permissions; other private registries use
repository_credentials_secret_arn. The image architecture must match
the compute it lands on (ARM images need ARM compute).
Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/etl:1.4.2".

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.container.command

`[]string`

Command override (Docker CMD) -- or the arguments to the image's
ENTRYPOINT. Supports "Ref::<key>" placeholders resolved from
spec.parameters and per-job SubmitJob overrides.
Example: ["python", "process.py", "Ref::input_path"].

### spec.container.vcpus

`double`

vCPUs reserved for the job (the VCPU resource requirement). EC2 jobs
take whole numbers; Fargate jobs take the Fargate sizes: 0.25, 0.5, 1,
2, 4, 8, or 16, each paired with a valid memory_mib range.

- rule: {"double":{"gt":0}}

### spec.container.memoryMib

`int32`

Memory hard limit in MiB (the MEMORY resource requirement) -- the job
is killed when it exceeds it. Fargate pairs memory with vcpus (e.g.
0.25 vCPU allows 512-2048 MiB in 1024-MiB steps).

- rule: {"int32":{"gte":4}}

### spec.container.gpus

`int32`

GPUs reserved for the job (the GPU resource requirement) -- pinned
exclusively to this job's container. EC2 GPU instance types only
(use an ECS_AL2_NVIDIA image type on the compute environment);
Fargate offers no GPUs.

- rule: {"int32":{"gte":0}}

### spec.container.jobRole

`string | valueFrom`

The IAM role the JOB'S CODE assumes at runtime -- its identity for
every AWS API call the workload makes (S3, DynamoDB, ...). Distinct
from execution_role by design: setup permissions and runtime
permissions should never be one role. Reference an AwsIamRole's
role_arn output or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.container.executionRole

`string | valueFrom`

The IAM role the ECS AGENT assumes to set the job up: pull private
images, resolve secrets, write logs. REQUIRED for Fargate job
definitions; on EC2 the instance profile usually covers it. Reference
an AwsIamRole's role_arn output or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.container.environment

`map<string, string>`

Plain-text environment variables (name -> value). For anything
sensitive use secrets instead -- environment values are visible to
anyone who can describe the job definition. Names must not start with
"AWS_BATCH" (reserved by the service).

### spec.container.secrets

`map<string, string>`

Secret environment variables (name -> the ARN of an AWS Secrets
Manager secret or SSM Parameter Store parameter). The agent resolves
each reference at job start -- via execution_role on Fargate, the
instance role on EC2 -- so the value never appears in the job
definition. Append ":<json-key>::" to a Secrets Manager ARN to inject
one key of a JSON secret.

### spec.container.logConfiguration

`AwsBatchJobDefinitionLogConfiguration`

Log driver configuration override. When unset, Batch sends container
logs to CloudWatch under the /aws/batch/job log group with zero
configuration -- set this only to change drivers or options.

### spec.container.logConfiguration.logDriver

`string` · required

The log driver: "awslogs" (CloudWatch -- the default even without this
block), "splunk", "fluentd", "gelf", "syslog", "journald", or
"json-file". The compute environment's instances must have the driver
available (Fargate supports awslogs and splunk).

- rule: {"required":true,"string":{"in":["awslogs","splunk","fluentd","gelf","syslog","journald","json-file"]}}

### spec.container.logConfiguration.options

`map<string, string>`

Driver-specific options -- e.g. awslogs-group / awslogs-stream-prefix
to redirect awslogs away from the /aws/batch/job default, or
splunk-url for Splunk.

### spec.container.logConfiguration.secretOptions

`map<string, string>`

Driver options whose values come from Secrets Manager / SSM (name ->
ARN) -- e.g. a splunk-token. Resolved by the agent at job start, never
stored in the definition.

### spec.container.mountPoints

`[]AwsBatchJobDefinitionMountPoint`

Mounts of the job's named volumes into the container's filesystem.

### spec.container.mountPoints[].sourceVolume

`string` · required

The name of a volume declared in container.volumes.

- rule: {"required":true}

### spec.container.mountPoints[].containerPath

`string` · required

The path inside the container where the volume mounts.
Example: "/mnt/data".

- rule: {"required":true}

### spec.container.mountPoints[].readOnly

`bool`

Mount read-only.

### spec.container.volumes

`[]AwsBatchJobDefinitionVolume`

Named volumes containers mount via mount_points: EFS file systems
(durable, shared, Fargate-supported) or container-instance host paths
(EC2 only).

- rule: a volume is backed by either efs or host_path, not both

### spec.container.volumes[].name

`string` · required

The volume's name, referenced by container.mount_points.

- rule: {"required":true}

### spec.container.volumes[].efs

`AwsBatchJobDefinitionEfsVolume`

Back the volume with an EFS file system.

### spec.container.volumes[].efs.fileSystemId

`string | valueFrom` · required

The EFS file system backing the volume. Reference an
AwsElasticFileSystem's file_system_id output or pass a literal file
system ID (e.g. "fs-0123456789abcdef0").

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.container.volumes[].efs.rootDirectory

`string`

The path within the file system to mount as the volume root. Ignored
when access_point_id is set (the access point defines the root).
Default: "/".

### spec.container.volumes[].efs.accessPointId

`string | valueFrom`

Mount through this EFS access point -- the recommended pattern: the
access point pins the POSIX identity and root path, so jobs cannot
wander the file system. Requires transit encryption (which the modules
enable whenever an access point or IAM authorization is used).
Reference an AwsEfsAccessPoint's access_point_id output or pass a
literal ID (e.g. "fsap-0123456789abcdef0").

- references: AwsEfsAccessPoint (`status.outputs.access_point_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEfsAccessPoint, name: <that resource's name>, fieldPath: status.outputs.access_point_id}} -- a bare string does not parse

### spec.container.volumes[].efs.iamAuthorization

`bool`

Authorize the mount with the job's IAM role (job_role must carry
elasticfilesystem:ClientMount/ClientWrite). Requires transit
encryption, which the modules enable automatically.

### spec.container.volumes[].hostPath

`string`

Back the volume with a path on the container instance (EC2 only).
Example: "/mnt/scratch".

### spec.container.ulimits

`[]AwsBatchJobDefinitionUlimit`

Resource limits (ulimits) for the container, e.g. raise "nofile" for
connection-heavy jobs. EC2 only -- Fargate rejects ulimit overrides.

### spec.container.ulimits[].name

`string` · required

The limit name: "nofile" (open files -- the common override), "core",
"cpu", "data", "fsize", "locks", "memlock", "msgqueue", "nice",
"nproc", "rss", "rtprio", "rttime", "sigpending", or "stack".

- rule: {"required":true}

### spec.container.ulimits[].softLimit

`int32`

The soft limit, enforced but raisable by the process up to hard_limit.

### spec.container.ulimits[].hardLimit

`int32`

The hard ceiling.

### spec.container.linuxParameters

`AwsBatchJobDefinitionLinuxParameters`

Linux host-level settings: device mappings, tmpfs mounts, shared
memory, and swap. EC2 only.

### spec.container.linuxParameters.initProcessEnabled

`bool`

Run an init process (PID 1) inside the container to reap zombie
processes -- maps to Docker's --init. Recommended for images whose
entrypoint spawns child processes.

### spec.container.linuxParameters.devices

`[]AwsBatchJobDefinitionDevice`

Host devices mapped into the container.

### spec.container.linuxParameters.devices[].hostPath

`string` · required

The device path on the container instance. Example: "/dev/xvdf".

- rule: {"required":true}

### spec.container.linuxParameters.devices[].containerPath

`string`

The path the device is exposed at inside the container. Defaults to
host_path when omitted.

### spec.container.linuxParameters.devices[].permissions

`[]string`

The cgroup permissions granted: any of "READ", "WRITE", "MKNOD".
Defaults to all three when empty.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["READ","WRITE","MKNOD"]}}}}

### spec.container.linuxParameters.sharedMemorySizeMib

`int32`

The /dev/shm size in MiB. Raise it for scientific/ML workloads that
use shared memory heavily.

### spec.container.linuxParameters.maxSwapMib

`int32`

The container's total swap budget in MiB. 0 disables swap; omit to
inherit the instance's configuration. Swap must be enabled on the
instance (via the compute environment's launch template) to take
effect.

### spec.container.linuxParameters.swappiness

`int32`

Swap aggressiveness, 0-100 (0 = swap only under pressure, 100 = swap
aggressively). Only meaningful when max_swap_mib is positive. AWS
default: 60.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.container.linuxParameters.tmpfs

`[]AwsBatchJobDefinitionTmpfs`

tmpfs (in-memory) mounts inside the container.

### spec.container.linuxParameters.tmpfs[].containerPath

`string` · required

The mount path inside the container. Example: "/tmp/scratch".

- rule: {"required":true}

### spec.container.linuxParameters.tmpfs[].sizeMib

`int32`

The tmpfs size in MiB.

- rule: {"int32":{"gt":0}}

### spec.container.linuxParameters.tmpfs[].mountOptions

`[]string`

Mount options (e.g. "noexec", "nosuid", "uid=1000").

### spec.container.privileged

`bool`

Run the container with elevated host permissions (root-equivalent).
EC2 only; reserved for host-integration workloads.

### spec.container.user

`string`

Run the container process as this user ("uid", "uid:gid", or a
username present in the image).

### spec.container.readonlyRootFilesystem

`bool`

Mount the container's root filesystem read-only -- writable paths must
come from volumes. A strong hardening default for jobs that only read
inputs and write to mounted storage.

### spec.container.repositoryCredentialsSecretArn

`string`

Credentials for pulling the image from a private NON-ECR registry: the
ARN of an AWS Secrets Manager secret holding {"username","password"}
-- a reference resolved at job start, never the credential itself.
ECR images need no credentials; grant the execution/instance role pull
access instead.

### spec.container.runtimePlatform

`AwsBatchJobDefinitionRuntimePlatform`

CPU architecture and OS for Fargate jobs. Set cpu_architecture to
"ARM64" to run on Graviton (cheaper per vCPU; images must be built for
arm64).

### spec.container.runtimePlatform.cpuArchitecture

`string`

"X86_64" (default) or "ARM64" (Graviton -- cheaper per vCPU; the image
must be built for arm64).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["X86_64","ARM64"]}}

### spec.container.runtimePlatform.operatingSystemFamily

`string`

OS family; "LINUX" (the default) for almost everything. Windows
containers on Fargate use the WINDOWS_SERVER_* families.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LINUX","WINDOWS_SERVER_2019_CORE","WINDOWS_SERVER_2019_FULL","WINDOWS_SERVER_2022_CORE","WINDOWS_SERVER_2022_FULL"]}}

### spec.container.fargatePlatformVersion

`string`

The Fargate platform version (e.g. "1.4.0" or "LATEST"). Fargate only;
omit to let AWS pick LATEST.

### spec.container.assignPublicIp

`bool`

Give the Fargate job's ENI a public IP -- required for internet access
from PUBLIC subnets without a NAT gateway. Fargate only; jobs in
private subnets should route through NAT instead.

### spec.container.ephemeralStorageGib

`int32`

Ephemeral scratch storage for the Fargate job, in GiB (21-200).
Fargate includes 20 GiB at no charge; set this only when the job needs
more (large intermediate files). Fargate only.

### spec.platformCapabilities

`[]string`

Where the job may run: "EC2" (default when empty) and/or "FARGATE".
A Fargate job definition additionally requires container.execution_role
and uses the Fargate-only knobs (platform version, public IP,
ephemeral storage, runtime platform); the EC2-only knobs (GPUs,
privileged, ulimits, Linux parameters) are rejected for it.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["EC2","FARGATE"]}}}}

### spec.parameters

`map<string, string>`

Default placeholder values for the job definition's parameter
substitution: a command like ["python", "run.py", "Ref::dataset"]
resolves "Ref::dataset" from this map, and SubmitJob can override each
key per job -- one definition, many parameterized runs.

### spec.retryStrategy

`AwsBatchJobDefinitionRetryStrategy`

How failed attempts are retried. When unset, jobs get a single attempt.

### spec.retryStrategy.attempts

`int32`

Total attempts, 1-10 (1 = no retries). Attempts re-run the whole job;
the workload must be idempotent or checkpoint-aware.

- rule: {"int32":{"lte":10,"gte":1}}

### spec.retryStrategy.evaluateOnExit

`[]AwsBatchJobDefinitionEvaluateOnExit`

Ordered conditions evaluated against a FAILED attempt's exit code and
status reasons; the FIRST match decides RETRY or EXIT, and a failure
matching nothing behaves like EXIT. Up to 5 conditions. The classic
use: RETRY on "Host EC2*" status reasons (Spot reclaims) while EXITing
on real application failures. CREATE-TIME per revision: changing these
registers a new revision.

- rule: {"repeated":{"maxItems":"5"}}

### spec.retryStrategy.evaluateOnExit[].action

`string` · required

The decision when this condition matches: "RETRY" (consume another
attempt) or "EXIT" (fail the job immediately).

- rule: {"required":true,"string":{"in":["RETRY","EXIT"]}}

### spec.retryStrategy.evaluateOnExit[].onExitCode

`string`

Glob match on the container's decimal exit code. Only a trailing "*"
wildcard is allowed. Example: "137" (SIGKILL / OOM), "1*".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[0-9]*\\*?$"}}

### spec.retryStrategy.evaluateOnExit[].onReason

`string`

Glob match on the attempt's reason (the container runtime's message,
e.g. "DockerTimeoutError*"). Only a trailing "*" wildcard.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[0-9A-Za-z.:\\s]*\\*?$"}}

### spec.retryStrategy.evaluateOnExit[].onStatusReason

`string`

Glob match on the attempt's status reason (Batch's own message, e.g.
"Host EC2*" for Spot interruptions). Only a trailing "*" wildcard.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^[0-9A-Za-z.:\\s]*\\*?$"}}

### spec.timeout

`AwsBatchJobDefinitionTimeout`

The hard wall-clock limit per job attempt. Attempts running longer are
terminated by Batch (and retried per retry_strategy). SubmitJob can
override it per job.

### spec.timeout.attemptDurationSeconds

`int32`

Seconds an attempt may run before Batch terminates it. Minimum 60.

- rule: {"int32":{"gte":60}}

### spec.schedulingPriority

`int32`

The job's scheduling priority WITHIN a fair-share queue (0-9999, higher
is sooner within the job's share). Only consulted when the queue has a
scheduling policy; FIFO queues ignore it.

- rule: {"int32":{"lte":9999,"gte":0}}

### spec.propagateTags

`bool`

Propagate the job definition's tags to the ECS task (and, from there,
to cost reports and IAM tag conditions on the running task).

### spec.deregisterOnNewRevision

`bool` · optional (explicit presence)

Whether registering a new revision deregisters the previous one
(marks it INACTIVE). The default (true) keeps exactly one ACTIVE
revision -- the one this resource manages. Set false when out-of-band
consumers (a manual SubmitJob against a pinned revision) must keep
running old revisions.

- default: `true`

### spec.eks

`AwsBatchJobDefinitionEks`

The Batch-on-EKS pod the job runs: containers, pod networking, and
Kubernetes-native volumes. Exactly one of container or eks is set --
this arm targets compute environments attached to an EKS cluster
(eks_configuration on AwsBatchComputeEnvironment). Jobs are submitted
the same way; Batch translates the definition into a pod on the
attached cluster.

- rule: every volume_mounts entry (containers and init_containers) must reference the name of a volume declared in volumes

### spec.eks.containers

`[]AwsBatchJobDefinitionEksContainer` · required

The pod's main containers (1-10). Batch watches these to decide job
success: the job completes when every main container exits.

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}
- rule: environment variable names must not start with 'AWS_BATCH' -- that prefix is reserved for variables AWS Batch sets on every job

### spec.eks.containers[].image

`string` · required

The container image, as a full reference: "<repository>:<tag>" or
"<repository>@<digest>". The image architecture must match the
cluster's node architecture.
Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/genomics:2.1".

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.eks.containers[].name

`string`

The container's name -- a Kubernetes DNS-1123 label (lowercase
alphanumerics and hyphens, max 63 chars). Required by Kubernetes when
the pod has more than one container; Batch names a lone unnamed
container "default".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.eks.containers[].command

`[]string`

Entrypoint override (Kubernetes command / Docker ENTRYPOINT).
Supports "Ref::<key>" placeholders resolved from spec.parameters.

### spec.eks.containers[].args

`[]string`

Arguments to the entrypoint (Kubernetes args / Docker CMD). Supports
"Ref::<key>" placeholders resolved from spec.parameters.

### spec.eks.containers[].env

`map<string, string>`

Plain-text environment variables (name -> value). Names must not
start with "AWS_BATCH" (reserved by the service). For secrets, mount
a Kubernetes secret volume instead -- EKS jobs have no ECS-style
secrets injection.

### spec.eks.containers[].imagePullPolicy

`string`

When Kubernetes pulls the image: "Always", "IfNotPresent", or
"Never". AWS defaults to Always, matching Kubernetes for :latest
tags -- pinned tags commonly use IfNotPresent to spare registry
traffic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Always","IfNotPresent","Never"]}}

### spec.eks.containers[].resources

`AwsBatchJobDefinitionEksResources`

The container's compute sizing -- Kubernetes resource requests and
limits. Batch schedules the job by these (its EKS counterpart of the
container arm's vcpus/memory_mib/gpus).

- rule: resources must set at least one of limits or requests (Batch requires cpu and memory sizing for EKS jobs)

### spec.eks.containers[].resources.limits

`map<string, string>`

Hard caps (Kubernetes limits): keys "cpu" (e.g. "1", "500m"),
"memory" (e.g. "2Gi", "512Mi"), and "nvidia.com/gpu" for GPU nodes.
Batch treats limits as the job's sizing; GPU quantities must be
whole numbers.

### spec.eks.containers[].resources.requests

`map<string, string>`

Scheduling reservations (Kubernetes requests), same keys as limits.
When both are set for a key, request must not exceed limit; Batch
fills a missing request from the limit.

### spec.eks.containers[].securityContext

`AwsBatchJobDefinitionEksSecurityContext`

The container's Kubernetes securityContext -- run-as identity and
privilege hardening.

### spec.eks.containers[].securityContext.runAsUser

`int64` · optional (explicit presence)

Run the container process as this numeric UID. 0 (root) is a legal
explicit value -- UNSET leaves the image's own USER in effect, which
is why presence matters here.

- rule: {"int64":{"gte":"0"}}

### spec.eks.containers[].securityContext.runAsGroup

`int64` · optional (explicit presence)

Run the container process as this numeric GID. As with run_as_user,
0 is legal and distinct from unset.

- rule: {"int64":{"gte":"0"}}

### spec.eks.containers[].securityContext.runAsNonRoot

`bool`

Have Kubernetes REJECT the pod at start if the effective user
resolves to root -- an assertion, not an identity setting.

### spec.eks.containers[].securityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

Whether the process may gain more privileges than its parent
(setuid binaries, file capabilities). UNSET means Kubernetes'
default (allowed, unless the container is otherwise restricted);
an explicit false is the hardening posture.

### spec.eks.containers[].securityContext.privileged

`bool`

Run the container privileged (root-equivalent on the node).
Default false, like Kubernetes.

### spec.eks.containers[].securityContext.readOnlyRootFileSystem

`bool`

Mount the container's root filesystem read-only -- writable paths
must come from volumes.

### spec.eks.containers[].volumeMounts

`[]AwsBatchJobDefinitionEksVolumeMount`

Mounts of the pod's declared volumes into this container's
filesystem.

### spec.eks.containers[].volumeMounts[].name

`string` · required

The name of a volume declared in eks.volumes.

- rule: {"required":true}

### spec.eks.containers[].volumeMounts[].mountPath

`string` · required

The path inside the container where the volume mounts.
Example: "/mnt/data".

- rule: {"required":true}

### spec.eks.containers[].volumeMounts[].readOnly

`bool`

Mount read-only.

### spec.eks.initContainers

`[]AwsBatchJobDefinitionEksContainer`

Init containers (0-10), run sequentially to completion before the
main containers start -- setup steps like fetching data or waiting
for a dependency.

- rule: {"repeated":{"maxItems":"10"}}
- rule: environment variable names must not start with 'AWS_BATCH' -- that prefix is reserved for variables AWS Batch sets on every job

### spec.eks.initContainers[].image

`string` · required

The container image, as a full reference: "<repository>:<tag>" or
"<repository>@<digest>". The image architecture must match the
cluster's node architecture.
Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/genomics:2.1".

- rule: {"required":true,"string":{"maxLen":"255"}}

### spec.eks.initContainers[].name

`string`

The container's name -- a Kubernetes DNS-1123 label (lowercase
alphanumerics and hyphens, max 63 chars). Required by Kubernetes when
the pod has more than one container; Batch names a lone unnamed
container "default".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.eks.initContainers[].command

`[]string`

Entrypoint override (Kubernetes command / Docker ENTRYPOINT).
Supports "Ref::<key>" placeholders resolved from spec.parameters.

### spec.eks.initContainers[].args

`[]string`

Arguments to the entrypoint (Kubernetes args / Docker CMD). Supports
"Ref::<key>" placeholders resolved from spec.parameters.

### spec.eks.initContainers[].env

`map<string, string>`

Plain-text environment variables (name -> value). Names must not
start with "AWS_BATCH" (reserved by the service). For secrets, mount
a Kubernetes secret volume instead -- EKS jobs have no ECS-style
secrets injection.

### spec.eks.initContainers[].imagePullPolicy

`string`

When Kubernetes pulls the image: "Always", "IfNotPresent", or
"Never". AWS defaults to Always, matching Kubernetes for :latest
tags -- pinned tags commonly use IfNotPresent to spare registry
traffic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Always","IfNotPresent","Never"]}}

### spec.eks.initContainers[].resources

`AwsBatchJobDefinitionEksResources`

The container's compute sizing -- Kubernetes resource requests and
limits. Batch schedules the job by these (its EKS counterpart of the
container arm's vcpus/memory_mib/gpus).

- rule: resources must set at least one of limits or requests (Batch requires cpu and memory sizing for EKS jobs)

### spec.eks.initContainers[].resources.limits

`map<string, string>`

Hard caps (Kubernetes limits): keys "cpu" (e.g. "1", "500m"),
"memory" (e.g. "2Gi", "512Mi"), and "nvidia.com/gpu" for GPU nodes.
Batch treats limits as the job's sizing; GPU quantities must be
whole numbers.

### spec.eks.initContainers[].resources.requests

`map<string, string>`

Scheduling reservations (Kubernetes requests), same keys as limits.
When both are set for a key, request must not exceed limit; Batch
fills a missing request from the limit.

### spec.eks.initContainers[].securityContext

`AwsBatchJobDefinitionEksSecurityContext`

The container's Kubernetes securityContext -- run-as identity and
privilege hardening.

### spec.eks.initContainers[].securityContext.runAsUser

`int64` · optional (explicit presence)

Run the container process as this numeric UID. 0 (root) is a legal
explicit value -- UNSET leaves the image's own USER in effect, which
is why presence matters here.

- rule: {"int64":{"gte":"0"}}

### spec.eks.initContainers[].securityContext.runAsGroup

`int64` · optional (explicit presence)

Run the container process as this numeric GID. As with run_as_user,
0 is legal and distinct from unset.

- rule: {"int64":{"gte":"0"}}

### spec.eks.initContainers[].securityContext.runAsNonRoot

`bool`

Have Kubernetes REJECT the pod at start if the effective user
resolves to root -- an assertion, not an identity setting.

### spec.eks.initContainers[].securityContext.allowPrivilegeEscalation

`bool` · optional (explicit presence)

Whether the process may gain more privileges than its parent
(setuid binaries, file capabilities). UNSET means Kubernetes'
default (allowed, unless the container is otherwise restricted);
an explicit false is the hardening posture.

### spec.eks.initContainers[].securityContext.privileged

`bool`

Run the container privileged (root-equivalent on the node).
Default false, like Kubernetes.

### spec.eks.initContainers[].securityContext.readOnlyRootFileSystem

`bool`

Mount the container's root filesystem read-only -- writable paths
must come from volumes.

### spec.eks.initContainers[].volumeMounts

`[]AwsBatchJobDefinitionEksVolumeMount`

Mounts of the pod's declared volumes into this container's
filesystem.

### spec.eks.initContainers[].volumeMounts[].name

`string` · required

The name of a volume declared in eks.volumes.

- rule: {"required":true}

### spec.eks.initContainers[].volumeMounts[].mountPath

`string` · required

The path inside the container where the volume mounts.
Example: "/mnt/data".

- rule: {"required":true}

### spec.eks.initContainers[].volumeMounts[].readOnly

`bool`

Mount read-only.

### spec.eks.hostNetwork

`bool` · optional (explicit presence)

Whether the pod uses the NODE's network namespace (Kubernetes
hostNetwork). UNSET means AWS's default, which is TRUE for Batch
pods -- the opposite of the plain-Kubernetes default -- so an
explicit false is a real choice: it gives the pod its own namespace
(required for VPC-CNI pod networking with security groups per pod).

### spec.eks.dnsPolicy

`string`

The pod's DNS resolution policy. AWS defaults to "ClusterFirst"
(resolve through the cluster's DNS first); "Default" inherits the
NODE's resolution; "ClusterFirstWithHostNet" is the cluster-first
behavior for pods running with host_network.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Default","ClusterFirst","ClusterFirstWithHostNet"]}}

### spec.eks.serviceAccountName

`string`

The Kubernetes service account the pod runs as -- the EKS-native way
to grant the JOB's code AWS permissions (IRSA / Pod Identity), the
counterpart of the container arm's job_role.

### spec.eks.podLabels

`map<string, string>`

Labels applied to the pod's metadata -- Kubernetes selectors,
cost-allocation, and policy engines key off these.
Example: {"team": "genomics", "workload": "batch"}.

### spec.eks.imagePullSecretNames

`[]string`

Names of Kubernetes imagePullSecrets in the job's namespace, for
pulling from private non-ECR registries (the EKS counterpart of the
container arm's repository_credentials_secret_arn). ECR images need
no secret -- the node role's pull access covers them.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.eks.shareProcessNamespace

`bool`

Share one process namespace across the pod's containers (Kubernetes
shareProcessNamespace) -- lets a sidecar signal or observe the main
container's processes. Default false, like plain Kubernetes.

### spec.eks.volumes

`[]AwsBatchJobDefinitionEksVolume`

Kubernetes-native volumes the pod's containers mount by name:
emptyDir scratch space, node hostPath directories, or Kubernetes
secrets. (EFS rides the cluster's CSI driver and static
PersistentVolumes -- outside the job definition's surface.)

- rule: a volume is backed by exactly one of empty_dir, host_path, or secret

### spec.eks.volumes[].name

`string` · required

The volume's name (a DNS-1123 label), referenced by containers'
volume_mounts.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.eks.volumes[].emptyDir

`AwsBatchJobDefinitionEksEmptyDir`

Scratch space that lives and dies with the job's pod.

### spec.eks.volumes[].emptyDir.medium

`string`

Where the scratch lives: unset backs it with node storage; "Memory"
backs it with tmpfs (fast, counts against the container's memory
sizing).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"const":"Memory"}}

### spec.eks.volumes[].emptyDir.sizeLimit

`string` · required

The scratch size cap, as a Kubernetes quantity.
Example: "1Gi", "500Mi".

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ki|Mi|Gi|Ti|K|M|G|T)?$"}}

### spec.eks.volumes[].hostPath

`string`

A directory on the NODE's filesystem (Kubernetes hostPath.path) --
data outlives the pod but is pinned to whichever node ran it.
Example: "/mnt/scratch".

### spec.eks.volumes[].secret

`AwsBatchJobDefinitionEksSecretVolume`

Project a Kubernetes secret (from the job's namespace) into the
volume.

### spec.eks.volumes[].secret.secretName

`string` · required

The name of the Kubernetes secret in the job's namespace (the
namespace comes from the compute environment's eks_configuration).

- rule: {"required":true}

### spec.eks.volumes[].secret.optional

`bool`

Mount successfully even when the secret does not exist yet (an empty
volume) instead of failing the pod.

## Validation Rules

- `exactly_one_of_container_or_eks`: set exactly one workload arm: container (ECS-based jobs on EC2/Fargate) or eks (pod jobs on an EKS-attached compute environment)
- `eks_forbids_ecs_only_toggles`: platform_capabilities and propagate_tags apply only to container (ECS-based) job definitions -- AWS rejects them for EKS jobs
- `fargate_requires_execution_role`: Fargate job definitions need container.execution_role -- reference an AwsIamRole carrying AmazonECSTaskExecutionRolePolicy (image pull + log write permissions)
- `fargate_forbids_ec2_only_container_fields`: Fargate job definitions cannot use the EC2-only container fields -- remove gpus, privileged, ulimits, and linux_parameters, or drop FARGATE from platform_capabilities
- `ec2_forbids_fargate_only_container_fields`: fargate_platform_version, assign_public_ip, ephemeral_storage_gib, and runtime_platform only apply to Fargate job definitions -- add FARGATE to platform_capabilities or remove them

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBatchJobDefinition, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.job_definition_arn` | `string` | The full ARN of the registered revision, including the revision number (e.g. "arn:aws:batch:us-west-2:123456789012:job-definition/etl:7"). The primary handle consumers reference -- because it carries the revision, each newly registered revision changes this output and rolls referencing EventBridge rules on their next deployment. |
| `status.outputs.arn_without_revision` | `string` | The ARN without the revision suffix (e.g. "arn:aws:batch:us-west-2: 123456789012:job-definition/etl"), for consumers that should always track the name's latest ACTIVE revision instead of a pinned one. |
| `status.outputs.job_definition_name` | `string` | The job definition name (metadata.name) the revisions are registered under. |
| `status.outputs.revision` | `int64` | The revision number this deployment registered (e.g. 7). Revisions are immutable; every spec change registers the next number. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.container.jobRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.container.executionRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.container.volumes[].efs.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |
| `spec.container.volumes[].efs.accessPointId` | AwsEfsAccessPoint | `status.outputs.access_point_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEventBridgeRule | `spec.targets[].batchTarget.jobDefinition` | `status.outputs.job_definition_arn` |

## See Also

- [Overview](../README.md)
