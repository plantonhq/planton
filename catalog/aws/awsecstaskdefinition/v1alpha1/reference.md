# AwsEcsTaskDefinition

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsEcsTaskDefinitionSpec defines an ECS task definition: the immutable,
versioned blueprint that describes the containers a task runs -- images,
ports, environment, secrets, health checks, sizing, volumes, and the IAM
identities the task assumes.

A task definition is the composition anchor of ECS compute, the same way a
launch template anchors EC2 fleets. It has its own lifecycle and is
referenced from the places that RUN it: an ECS service (steady-state
workloads), EventBridge scheduled tasks, and one-off RunTask calls. That is
why it is a first-class resource rather than a detail of the service that
happens to run it.

Task definitions are revisioned and revisions are immutable in AWS: every
change to this spec registers a NEW revision of the family rather than
mutating the old one. The family name comes from metadata.name. Because the
task_definition_arn stack output carries the revision, a service that
references it by output picks up each new revision on its next deployment
-- "change the image tag, the service rolls" falls out of the composition.

Task and container sizing interact: on Fargate, task-level cpu and memory
are REQUIRED and constrain the valid combinations (e.g. 256 CPU pairs with
512-2048 MiB); on EC2, task-level sizing is optional and per-container
cpu/memory govern bin-packing onto instances.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsTaskDefinition
metadata:
  name: awsecstaskdefinition-demo
spec:
  region: us-west-2
  cpu: 256
  memory: 512
  executionRole:
    value: arn:aws:iam::123456789012:role/ecsTaskExecutionRole
  containers:
    - name: app
      image: public.ecr.aws/nginx/nginx:stable
      portMappings:
        - containerPort: 80
          name: http
          appProtocol: http
      environment:
        APP_ENV: demo
      # A Planton secret: the component keeps it in a Secrets Manager secret
      # only the execution role reads, and injects it by ARN.
      secretEnvironment:
        API_TOKEN: $secret/demo-api-token
      mountPoints:
        - sourceVolume: data
          containerPath: /var/data
  # A launch-time volume: only the NAME is declared here -- the ECS
  # service running this task supplies the managed-EBS backing through
  # its volume_configuration under the same name.
  volumes:
    - name: data
      configureAtLaunch: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.containers` | `[]AwsEcsTaskDefinitionContainer` | yes |  |  |
| `spec.containers[].name` | `string` | yes |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].essential` | `bool` |  |  |  |
| `spec.containers[].cpu` | `int32` |  |  |  |
| `spec.containers[].memory` | `int32` |  |  |  |
| `spec.containers[].memoryReservation` | `int32` |  |  |  |
| `spec.containers[].portMappings` | `[]AwsEcsTaskDefinitionPortMapping` |  |  |  |
| `spec.containers[].portMappings[].containerPort` | `int32` |  |  |  |
| `spec.containers[].portMappings[].protocol` | `string` |  |  |  |
| `spec.containers[].portMappings[].name` | `string` |  |  |  |
| `spec.containers[].portMappings[].appProtocol` | `string` |  |  |  |
| `spec.containers[].entryPoint` | `[]string` |  |  |  |
| `spec.containers[].command` | `[]string` |  |  |  |
| `spec.containers[].workingDirectory` | `string` |  |  |  |
| `spec.containers[].environment` | `map<string, string>` (no secrets: use `secretEnvironment`) |  |  |  |
| `spec.containers[].secrets` | `map<string, string>` |  |  |  |
| `spec.containers[].secretEnvironment` | `map<string, string>` (sensitive) |  |  |  |
| `spec.containers[].environmentFiles` | `[]string` |  |  |  |
| `spec.containers[].healthCheck` | `AwsEcsTaskDefinitionHealthCheck` |  |  |  |
| `spec.containers[].healthCheck.command` | `[]string` | yes |  |  |
| `spec.containers[].healthCheck.intervalSeconds` | `int32` |  |  |  |
| `spec.containers[].healthCheck.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].healthCheck.retries` | `int32` |  |  |  |
| `spec.containers[].healthCheck.startPeriodSeconds` | `int32` |  |  |  |
| `spec.containers[].dependsOn` | `[]AwsEcsTaskDefinitionContainerDependency` |  |  |  |
| `spec.containers[].dependsOn[].containerName` | `string` | yes |  |  |
| `spec.containers[].dependsOn[].condition` | `string` |  |  |  |
| `spec.containers[].mountPoints` | `[]AwsEcsTaskDefinitionMountPoint` |  |  |  |
| `spec.containers[].mountPoints[].sourceVolume` | `string` | yes |  |  |
| `spec.containers[].mountPoints[].containerPath` | `string` | yes |  |  |
| `spec.containers[].mountPoints[].readOnly` | `bool` |  |  |  |
| `spec.containers[].logConfiguration` | `AwsEcsTaskDefinitionLogConfiguration` |  |  |  |
| `spec.containers[].logConfiguration.logDriver` | `string` | yes |  |  |
| `spec.containers[].logConfiguration.options` | `map<string, string>` |  |  |  |
| `spec.containers[].logConfiguration.secretOptions` | `map<string, string>` |  |  |  |
| `spec.containers[].firelensConfiguration` | `AwsEcsTaskDefinitionFirelens` |  |  |  |
| `spec.containers[].firelensConfiguration.type` | `string` |  |  |  |
| `spec.containers[].firelensConfiguration.options` | `map<string, string>` |  |  |  |
| `spec.containers[].repositoryCredentialsSecretArn` | `string` |  |  |  |
| `spec.containers[].user` | `string` |  |  |  |
| `spec.containers[].readonlyRootFilesystem` | `bool` |  |  |  |
| `spec.containers[].privileged` | `bool` |  |  |  |
| `spec.containers[].initProcessEnabled` | `bool` |  |  |  |
| `spec.containers[].gpuCount` | `int32` |  |  |  |
| `spec.containers[].ulimits` | `[]AwsEcsTaskDefinitionUlimit` |  |  |  |
| `spec.containers[].ulimits[].name` | `string` | yes |  |  |
| `spec.containers[].ulimits[].softLimit` | `int32` |  |  |  |
| `spec.containers[].ulimits[].hardLimit` | `int32` |  |  |  |
| `spec.containers[].dockerLabels` | `map<string, string>` |  |  |  |
| `spec.containers[].startTimeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].stopTimeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].restartPolicy` | `AwsEcsTaskDefinitionRestartPolicy` |  |  |  |
| `spec.containers[].restartPolicy.enabled` | `bool` |  |  |  |
| `spec.containers[].restartPolicy.ignoredExitCodes` | `[]int32` |  |  |  |
| `spec.containers[].restartPolicy.restartAttemptPeriodSeconds` | `int32` |  |  |  |
| `spec.requiresCompatibilities` | `[]string` |  |  |  |
| `spec.cpu` | `int32` |  |  |  |
| `spec.memory` | `int32` |  |  |  |
| `spec.networkMode` | `string` |  | `awsvpc` |  |
| `spec.executionRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.taskRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.runtimePlatform` | `AwsEcsTaskDefinitionRuntimePlatform` |  |  |  |
| `spec.runtimePlatform.cpuArchitecture` | `string` |  |  |  |
| `spec.runtimePlatform.operatingSystemFamily` | `string` |  |  |  |
| `spec.ephemeralStorageGib` | `int32` |  |  |  |
| `spec.volumes` | `[]AwsEcsTaskDefinitionVolume` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].efs` | `AwsEcsTaskDefinitionEfsVolume` |  |  |  |
| `spec.volumes[].efs.fileSystemId` | `string \| valueFrom` | yes |  | AwsElasticFileSystem (`status.outputs.file_system_id`) |
| `spec.volumes[].efs.rootDirectory` | `string` |  |  |  |
| `spec.volumes[].efs.accessPointId` | `string \| valueFrom` |  |  | AwsEfsAccessPoint (`status.outputs.access_point_id`) |
| `spec.volumes[].efs.iamAuthorization` | `bool` |  |  |  |
| `spec.volumes[].efs.transitEncryptionPort` | `int32` |  |  |  |
| `spec.volumes[].hostPath` | `string` |  |  |  |
| `spec.volumes[].configureAtLaunch` | `bool` |  |  |  |
| `spec.volumes[].docker` | `AwsEcsTaskDefinitionDockerVolume` |  |  |  |
| `spec.volumes[].docker.autoprovision` | `bool` |  |  |  |
| `spec.volumes[].docker.driver` | `string` |  |  |  |
| `spec.volumes[].docker.driverOpts` | `map<string, string>` |  |  |  |
| `spec.volumes[].docker.labels` | `map<string, string>` |  |  |  |
| `spec.volumes[].docker.scope` | `string` |  |  |  |
| `spec.volumes[].s3files` | `AwsEcsTaskDefinitionS3FilesVolume` |  |  |  |
| `spec.volumes[].s3files.fileSystemArn` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_arn`) |
| `spec.volumes[].s3files.accessPointArn` | `string` |  |  |  |
| `spec.volumes[].s3files.rootDirectory` | `string` |  |  |  |
| `spec.volumes[].s3files.transitEncryptionPort` | `int32` |  |  |  |
| `spec.logging` | `AwsEcsTaskDefinitionLogging` |  |  |  |
| `spec.logging.disabled` | `bool` |  |  |  |
| `spec.logging.logGroup` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.logging.retentionDays` | `int32` |  | `30` |  |
| `spec.skipDestroy` | `bool` |  |  |  |
| `spec.enableFaultInjection` | `bool` |  |  |  |
| `spec.ipcMode` | `string` |  |  |  |
| `spec.pidMode` | `string` |  |  |  |
| `spec.placementConstraints` | `[]AwsEcsTaskDefinitionPlacementConstraint` |  |  |  |
| `spec.placementConstraints[].type` | `string` |  |  |  |
| `spec.placementConstraints[].expression` | `string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the task definition is registered in. A task definition
is a regional object: a service can only run revisions registered in
its own region.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.containers

`[]AwsEcsTaskDefinitionContainer` · required

The containers the task runs. Most tasks run one application container;
add sidecars (a log router, an OpenTelemetry collector, a proxy) as
additional entries and order their startup with depends_on. At least
one container must be essential -- when an essential container exits,
the whole task stops. Deployment pipelines inject the built image into
containers whose image is left BLANK (the image-slot contract);
authored images are untouched.

- rule: {"repeated":{"minItems":"1"}}
- rule: memory_reservation (the soft reservation) must not exceed memory (the hard limit)
- rule: a variable in secret_environment is also named in environment or secrets -- give each variable its value from exactly one of them

### spec.containers[].name

`string` · required

The container's name, unique within the task definition. Referenced by
an ECS service's load_balancers.container_name, by sibling containers'
depends_on, and used as the CloudWatch log stream prefix under the
shared task log group.

- rule: {"required":true}

### spec.containers[].image

`string` · required

The container image, as a full reference: "<repository>:<tag>" or
"<repository>@<digest>". Private ECR images require execution_role to
carry pull permissions; other private registries use
repository_credentials.
Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/api:1.4.2".

- rule: {"required":true}

### spec.containers[].essential

`bool` · optional (explicit presence)

Whether this container is essential. When an essential container
exits, ECS stops the whole task; non-essential sidecars can exit or
fail without killing the application container. AWS default: true.
Optional so an explicit "this sidecar is not essential" (false) is
distinguishable from unset.

### spec.containers[].cpu

`int32`

CPU units reserved for this container (1024 = 1 vCPU). On Fargate this
subdivides the task-level cpu between containers (optional -- unset
containers share what is left); on EC2 it drives bin-packing.

### spec.containers[].memory

`int32`

Hard memory limit for this container, in MiB -- the container is
killed when it exceeds it. On Fargate the task-level memory already
caps the task; set per-container limits to fence sidecars off from
the application's share.

### spec.containers[].memoryReservation

`int32`

Soft memory reservation, in MiB: the scheduler reserves this much but
lets the container burst up to memory (or the task limit). Set
reservation at the expected footprint and the limit at the tolerable
ceiling.

### spec.containers[].portMappings

`[]AwsEcsTaskDefinitionPortMapping`

Ports the container exposes. On awsvpc networking the container port
IS the host port (each task has its own ENI). Name a port to make it
referenceable from Service Connect.

- rule: protocol must be 'tcp' or 'udp' when set
- rule: app_protocol must be 'http', 'http2', or 'grpc' when set
- rule: app_protocol only applies to named ports -- set name as well

### spec.containers[].portMappings[].containerPort

`int32`

The port the application listens on inside the container.
Example: 8080.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].portMappings[].protocol

`string`

Layer-4 protocol: "tcp" (default) or "udp".

### spec.containers[].portMappings[].name

`string`

A name for this port, referenced by Service Connect
(service_connect.services[].port_name) and by sibling tooling. Name
the port whenever the service participates in Service Connect.

### spec.containers[].portMappings[].appProtocol

`string`

The application protocol ECS uses for Service Connect telemetry and
routing: "http", "http2", or "grpc". Only meaningful on named ports.

### spec.containers[].entryPoint

`[]string`

Entry point override (Docker ENTRYPOINT). Leave empty to use the
image's own.

### spec.containers[].command

`[]string`

Command override (Docker CMD) -- or the arguments to entry_point.
Leave empty to use the image's own.

### spec.containers[].workingDirectory

`string`

Working directory override for the command.

### spec.containers[].environment

`map<string, string>` · no secrets

Plain-text environment variables (name -> value), written into the task
definition where anyone who can describe it reads them -- and kept in
every revision for good, since revisions are immutable. Configuration
only; a credential goes in secret_environment (or secrets).

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretEnvironment`, which keeps it in a secret store the workload reads by reference
### spec.containers[].secrets

`map<string, string>`

Secret environment variables backed by secrets YOU already own (name ->
the ARN of an AWS Secrets Manager secret or SSM Parameter Store
parameter). The ECS agent resolves each reference at task start using
execution_role, so the value never appears in the task definition.
Append ":<json-key>::" to a Secrets Manager ARN to inject one key of a
JSON secret.

### spec.containers[].secretEnvironment

`map<string, string>` · sensitive

Secret environment variables whose VALUES this component keeps in AWS
Secrets Manager for you (name -> value). Per entry it creates one secret
named "<family>/<container>/<name>", stores the value, and attaches a
resource policy that lets only execution_role read it; the container's
secrets list then carries that secret's ARN pinned to the stored
version, so the task definition holds a reference, never the value. A
changed value registers a new revision (a deploy is the rotation), and
destroying the task definition deletes the secrets with no recovery
window -- the value's source of truth is whoever supplied it here.
Requires execution_role.

### spec.containers[].environmentFiles

`[]string`

Environment files loaded from S3 (each entry an S3 object ARN of a
.env file). Applied before environment/secrets; later sources win.
execution_role must be able to read the objects.

- rule: {"repeated":{"unique":true}}

### spec.containers[].healthCheck

`AwsEcsTaskDefinitionHealthCheck`

Container-level health check ECS runs INSIDE the container (distinct
from any load-balancer target health check). Sibling containers can
gate their startup on it via depends_on condition "HEALTHY", and the
task reports unhealthy when it fails.

### spec.containers[].healthCheck.command

`[]string` · required

The probe command, in Docker exec form. The first element is "CMD"
(exec directly) or "CMD-SHELL" (run through the shell).
Example: ["CMD-SHELL", "curl -f http://localhost:8080/healthz || exit 1"].

- rule: {"repeated":{"minItems":"1"}}

### spec.containers[].healthCheck.intervalSeconds

`int32`

Seconds between probes. AWS default: 30.

### spec.containers[].healthCheck.timeoutSeconds

`int32`

Seconds before an unanswered probe counts as a failure. AWS default: 5.

### spec.containers[].healthCheck.retries

`int32`

Consecutive failures before the container is marked unhealthy. AWS
default: 3.

### spec.containers[].healthCheck.startPeriodSeconds

`int32`

Grace period after container start during which failures do not count
-- give slow-booting apps room before probes matter. AWS default: 0.

### spec.containers[].dependsOn

`[]AwsEcsTaskDefinitionContainerDependency`

Startup ordering against sibling containers: wait for a dependency to
START, become HEALTHY (requires its health_check), COMPLETE, or exit
with SUCCESS before this container starts. Shutdown runs in reverse.

### spec.containers[].dependsOn[].containerName

`string` · required

The name of the sibling container this one waits for.

- rule: {"required":true}

### spec.containers[].dependsOn[].condition

`string`

The state to wait for: "START" (the dependency has started),
"HEALTHY" (its health_check passes -- the dependency must define one),
"COMPLETE" (it exited, any code), or "SUCCESS" (it exited 0).

- rule: {"string":{"in":["START","HEALTHY","COMPLETE","SUCCESS"]}}

### spec.containers[].mountPoints

`[]AwsEcsTaskDefinitionMountPoint`

Mounts of the task's named volumes into this container's filesystem.

### spec.containers[].mountPoints[].sourceVolume

`string` · required

The name of a volume declared in spec.volumes.

- rule: {"required":true}

### spec.containers[].mountPoints[].containerPath

`string` · required

The path inside the container where the volume mounts.
Example: "/var/data".

- rule: {"required":true}

### spec.containers[].mountPoints[].readOnly

`bool`

Mount read-only.

### spec.containers[].logConfiguration

`AwsEcsTaskDefinitionLogConfiguration`

Per-container log configuration override. When set, this container
logs with exactly this driver/options instead of the task-level
logging default -- the escape hatch for shipping to Splunk, Fluentd,
or a FireLens router.

- rule: log_driver must be one of: awslogs, awsfirelens, splunk, fluentd, gelf, syslog, journald, json-file

### spec.containers[].logConfiguration.logDriver

`string` · required

The log driver: "awslogs" (CloudWatch), "awsfirelens" (route through a
FireLens sibling container), "splunk", "fluentd", "gelf", "syslog",
"journald", or "json-file" (EC2 only).

- rule: {"required":true}

### spec.containers[].logConfiguration.options

`map<string, string>`

Driver-specific options (e.g. awslogs-group / awslogs-region /
awslogs-stream-prefix for awslogs; Name/host/port for awsfirelens
outputs).

### spec.containers[].logConfiguration.secretOptions

`map<string, string>`

Driver options whose values come from Secrets Manager / SSM (name ->
ARN) -- e.g. a Splunk HEC token. Resolved by the agent at task start
via execution_role.

### spec.containers[].firelensConfiguration

`AwsEcsTaskDefinitionFirelens`

Marks this container as a FireLens log router (fluentbit or fluentd).
Sibling containers then use log_configuration with driver
"awsfirelens" to route their logs through it.

### spec.containers[].firelensConfiguration.type

`string`

The router type: "fluentbit" (the AWS-recommended lightweight router)
or "fluentd".

- rule: {"string":{"in":["fluentbit","fluentd"]}}

### spec.containers[].firelensConfiguration.options

`map<string, string>`

Router options -- e.g. enable-ecs-log-metadata: "true", or
config-file-type/config-file-value for a custom parsing config.

### spec.containers[].repositoryCredentialsSecretArn

`string`

Credentials for pulling the image from a private NON-ECR registry:
the ARN of an AWS Secrets Manager secret holding {"username","password"}
-- a reference the ECS agent resolves at task start, never the
credential itself. ECR images need no credentials; grant
execution_role pull access instead.

### spec.containers[].user

`string`

Run the container process as this user ("uid", "uid:gid", or a
username present in the image). Fargate: uid-based values only.

### spec.containers[].readonlyRootFilesystem

`bool`

Mount the container's root filesystem read-only -- writable paths must
come from volumes. A strong hardening default for stateless services.

### spec.containers[].privileged

`bool`

Give the container elevated privileges on the host (EC2 launch type
only; not supported on Fargate). Reserved for host-integration agents.

### spec.containers[].initProcessEnabled

`bool`

Run an init process (PID 1) inside the container to reap zombie
processes -- maps to Docker's --init. Recommended for images whose
entrypoint spawns child processes.

### spec.containers[].gpuCount

`int32`

Number of GPUs reserved for this container (EC2 GPU instance types
only; Fargate does not offer GPUs).

### spec.containers[].ulimits

`[]AwsEcsTaskDefinitionUlimit`

Resource limits (ulimits) for the container, e.g. raise "nofile" for
high-connection proxies. Fargate supports only "nofile" overrides
(default soft 1024 / hard 65535 there).

### spec.containers[].ulimits[].name

`string` · required

The limit name: "nofile" (open files -- the common override), "core",
"cpu", "data", "fsize", "locks", "memlock", "msgqueue", "nice",
"nproc", "rss", "rtprio", "rttime", "sigpending", or "stack".

- rule: {"required":true}

### spec.containers[].ulimits[].softLimit

`int32`

The soft limit, enforced but raisable by the process up to hard_limit.

### spec.containers[].ulimits[].hardLimit

`int32`

The hard ceiling.

### spec.containers[].dockerLabels

`map<string, string>`

Docker labels applied to the container (key -> value) -- consumed by
on-host tooling and some log routers.

### spec.containers[].startTimeoutSeconds

`int32`

Seconds ECS waits for this container's depends_on conditions before
giving up on starting it.

### spec.containers[].stopTimeoutSeconds

`int32`

Seconds ECS waits after SIGTERM before SIGKILL at shutdown. Default
30; raise it for workloads that need a longer graceful drain (capped
at 120 on Fargate).

- rule: {"int32":{"lte":120,"gte":0}}

### spec.containers[].restartPolicy

`AwsEcsTaskDefinitionRestartPolicy`

Restart the container in place when it exits, without replacing the
whole task -- faster recovery for a crashing sidecar than a full task
cycle.

- rule: restart_attempt_period_seconds must be between 60 and 1800 when set

### spec.containers[].restartPolicy.enabled

`bool`

Enable in-place restarts for this container.

### spec.containers[].restartPolicy.ignoredExitCodes

`[]int32`

Exit codes that should NOT trigger a restart (e.g. 0 for a batch
sidecar that finished cleanly).

### spec.containers[].restartPolicy.restartAttemptPeriodSeconds

`int32`

Minimum seconds the container must run before a restart attempt is
made (60-1800). AWS default: 300.

### spec.requiresCompatibilities

`[]string`

Launch types the task definition validates against: "FARGATE",
"EC2", "EXTERNAL" (ECS Anywhere), and/or "MANAGED_INSTANCES" (ECS
Managed Instances capacity). AWS registers the definition for those
environments and rejects incompatible settings at registration time
instead of at run time. When empty, both modules register for
["FARGATE"] -- the serverless launch type that needs no instance
management.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["FARGATE","EC2","EXTERNAL","MANAGED_INSTANCES"]}}}}

### spec.cpu

`int32`

Total CPU for the task, in CPU units (1024 = 1 vCPU). REQUIRED for
Fargate, where it selects the task size: 256, 512, 1024, 2048, 4096,
8192, or 16384. Optional on EC2 (containers bin-pack by their own cpu).
Example: 512.

### spec.memory

`int32`

Total memory for the task, in MiB. REQUIRED for Fargate and constrained
by cpu (e.g. 256 CPU pairs with 512-2048 MiB, 1024 CPU with 2048-8192
MiB). Optional on EC2, where it caps the sum of container memory.
Example: 1024.

### spec.networkMode

`string`

Docker networking mode: "awsvpc" (each task gets its own ENI and
private IP -- required for Fargate and the modern default everywhere),
"bridge" (EC2 docker0 bridge), "host" (EC2 host network stack), or
"none". Default: "awsvpc".

- default: `awsvpc`

### spec.executionRole

`string | valueFrom`

The IAM role the ECS AGENT assumes to set the task up: pull private
images from ECR, fetch secrets for the secrets/repository_credentials
fields, and create/write CloudWatch log streams. Reference an
AwsIamRole's role_arn output or pass a literal role ARN. AWS REQUIRES
it at registration time for Fargate tasks that use the awslogs driver
-- which the default logging wiring does -- and in practice whenever
the task uses ECR images or secrets.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.taskRole

`string | valueFrom`

The IAM role the APPLICATION code assumes at runtime -- the task's
identity for every AWS API call the workload itself makes (S3, SQS,
DynamoDB, ...). Distinct from execution_role by design: the agent's
setup permissions and the app's runtime permissions should never be
one role. Omit when the app calls no AWS APIs.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.runtimePlatform

`AwsEcsTaskDefinitionRuntimePlatform`

CPU architecture and OS the task runs on. Set cpu_architecture to
"ARM64" to run on Graviton (Fargate ARM pricing is ~20% below x86 for
the same vCPU/memory) -- the images must be built for arm64.

- rule: cpu_architecture must be 'X86_64' or 'ARM64' when set
- rule: operating_system_family must be LINUX or a supported WINDOWS_SERVER_* family when set

### spec.runtimePlatform.cpuArchitecture

`string`

"X86_64" (default) or "ARM64" (Graviton -- cheaper per vCPU, images
must be multi-arch or arm64-built).

### spec.runtimePlatform.operatingSystemFamily

`string`

OS family; "LINUX" (default) for almost everything. Windows families
("WINDOWS_SERVER_2019_CORE", "WINDOWS_SERVER_2019_FULL",
"WINDOWS_SERVER_2022_CORE", "WINDOWS_SERVER_2022_FULL") are supported
on Fargate for Windows containers.

### spec.ephemeralStorageGib

`int32`

Ephemeral scratch storage shared by the task's containers, in GiB
(21-200). Fargate default: 20 GiB at no charge; set this only when the
workload needs more (image processing, builds, large temp files).

### spec.volumes

`[]AwsEcsTaskDefinitionVolume`

Named volumes containers mount via mount_points. Fargate supports EFS
and S3-backed volumes (durable, shared across tasks) and the implicit
ephemeral storage; host-path and Docker volumes are EC2-only. A
configure_at_launch volume declares only the NAME here -- the
AwsEcsService running the task supplies the backing (managed EBS) at
deployment time through its volume_configuration.

- rule: a volume is backed by at most one of: efs, host_path, docker, s3files
- rule: a configure_at_launch volume must not declare a backing here -- the AwsEcsService supplies it via volume_configuration

### spec.volumes[].name

`string` · required

The volume's name, referenced by containers' mount_points.

- rule: {"required":true}

### spec.volumes[].efs

`AwsEcsTaskDefinitionEfsVolume`

Back the volume with an EFS file system.

### spec.volumes[].efs.fileSystemId

`string | valueFrom` · required

The EFS file system backing the volume. Reference an AwsElasticFileSystem
resource or pass a literal file system ID (e.g. "fs-0123456789abcdef0").

- references: AwsElasticFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.volumes[].efs.rootDirectory

`string`

The path within the file system to mount as the volume root. Ignored
when access_point_id is set (the access point defines the root).
Default: "/".

### spec.volumes[].efs.accessPointId

`string | valueFrom`

Mount through this EFS access point -- the recommended pattern:
the access point pins the POSIX identity and root path, so tasks
cannot wander the file system. Reference an AwsEfsAccessPoint resource
or pass a literal access point ID (e.g. "fsap-0123456789abcdef0").

- references: AwsEfsAccessPoint (`status.outputs.access_point_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEfsAccessPoint, name: <that resource's name>, fieldPath: status.outputs.access_point_id}} -- a bare string does not parse

### spec.volumes[].efs.iamAuthorization

`bool`

Authorize the mount with the task's IAM role (execution/task role
must carry elasticfilesystem:ClientMount/ClientWrite). Requires
transit encryption, which the modules enable automatically.

### spec.volumes[].efs.transitEncryptionPort

`int32`

The port for the encrypted transit tunnel between host and EFS.
0 (unset) lets AWS choose an ephemeral port -- the right answer
unless a host firewall requires a fixed one.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.volumes[].hostPath

`string`

Back the volume with a path on the container instance (EC2 launch
type only). Example: "/mnt/data".

### spec.volumes[].configureAtLaunch

`bool`

Defer the volume's configuration to launch time: the AwsEcsService
running this task supplies the backing through its
volume_configuration (managed EBS -- a fresh, service-owned EBS
volume per task) under this same volume name. Set it true and leave
every backing here unset; the service side names the volume in
spec.volume_configuration.name.

### spec.volumes[].docker

`AwsEcsTaskDefinitionDockerVolume`

Back the volume with a Docker volume on the container instance (EC2
launch type only) -- named Docker-managed storage, optionally via a
volume driver plugin.

- rule: scope must be 'task' or 'shared' when set
- rule: autoprovision only applies to 'shared' scope volumes -- a task-scoped volume always exists exactly as long as its task

### spec.volumes[].docker.autoprovision

`bool`

Provision the volume automatically if it does not already exist.
Only meaningful with scope "shared".

### spec.volumes[].docker.driver

`string`

The Docker volume driver. Default: "local". Must match a driver
installed on the container instance.

### spec.volumes[].docker.driverOpts

`map<string, string>`

Driver-specific options passed at volume creation.

### spec.volumes[].docker.labels

`map<string, string>`

Custom metadata labels applied to the volume.

### spec.volumes[].docker.scope

`string`

The volume's lifetime: "task" (AWS default -- destroyed when the task
stops) or "shared" (persists on the instance across tasks).

### spec.volumes[].s3files

`AwsEcsTaskDefinitionS3FilesVolume`

Back the volume with an S3 bucket mounted as a file system
(Mountpoint for Amazon S3) -- read-heavy shared data (models,
reference datasets) without provisioning EFS.

### spec.volumes[].s3files.fileSystemArn

`string | valueFrom` · required

The S3 bucket (by ARN) mounted as the volume. Reference an
AwsS3Bucket's bucket_arn output or pass a literal bucket ARN.

- references: AwsS3Bucket (`status.outputs.bucket_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_arn}} -- a bare string does not parse

### spec.volumes[].s3files.accessPointArn

`string`

An S3 access point ARN to mount through instead of the bucket
directly -- scopes the mount to the access point's policy.

### spec.volumes[].s3files.rootDirectory

`string`

The path within the bucket mounted as the volume root. Default: "/".

### spec.volumes[].s3files.transitEncryptionPort

`int32`

The port for encrypted transit between host and mount target.
0 (unset) lets AWS choose.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.logging

`AwsEcsTaskDefinitionLogging`

Default CloudWatch logging for every container that does not declare
its own log_configuration. Enabled by default: the modules create ONE
log group named "/ecs/<family>" (30-day retention unless overridden)
and each container logs under its own name as the stream prefix --
so a task's containers land in one predictable place with zero
configuration.

### spec.logging.disabled

`bool`

Disable the default log wiring entirely. Containers without their own
log_configuration then produce no logs -- almost never what you want.

### spec.logging.logGroup

`string | valueFrom`

Use an existing CloudWatch log group instead of creating one.
Reference an AwsCloudwatchLogGroup's log_group_name output or pass a
literal group name. When unset, the modules create a group named
"/ecs/<family>" and manage its lifecycle with the task definition.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.logging.retentionDays

`int32` · optional (explicit presence)

Retention, in days, for the auto-created log group. Ignored when
log_group references an existing group (that group owns its own
retention).

- default: `30`

### spec.skipDestroy

`bool`

Keep old revisions registered when this resource is destroyed, instead
of deregistering every revision of the family. Useful when other
consumers (a scheduled task, a manual RunTask) may still reference
older revisions.

### spec.enableFaultInjection

`bool`

Enable AWS Fault Injection Service (FIS) actions against this task's
containers -- the opt-in that lets chaos experiments (CPU stress,
network latency, process kill) target the task. Off by default; only
enable on definitions you deliberately run experiments against.

### spec.ipcMode

`string`

IPC namespace sharing for the task's containers: "host" (share the
instance's IPC namespace -- weakest isolation), "task" (containers
share one namespace within the task), or "none" (each container
isolated). EC2 tasks only -- Fargate rejects it at registration.
Unset keeps Docker's default (private namespace per container).

### spec.pidMode

`string`

Process namespace sharing: "host" (containers see the instance's
processes -- weakest isolation) or "task" (containers within the task
share one PID namespace -- what lets a sidecar observe the app's
processes). On Fargate only "task" is supported (platform 1.4+).
Unset keeps Docker's default (private namespace per container).

### spec.placementConstraints

`[]AwsEcsTaskDefinitionPlacementConstraint`

Task-level placement constraints, evaluated at RunTask/service
scheduling time on EC2 container instances (at most 10). EC2 tasks
only -- Fargate rejects them at registration. Service-level
constraints live on the AwsEcsService; declare here only what is
intrinsic to the task itself (e.g. an instance-attribute expression
every consumer must honor).

- rule: {"repeated":{"maxItems":"10"}}

### spec.placementConstraints[].type

`string`

The constraint type. "memberOf" (restrict placement to instances
matching the expression) is the only type AWS supports at the task
definition level.

- rule: {"string":{"in":["memberOf"]}}

### spec.placementConstraints[].expression

`string` · required

A cluster query language expression, e.g.
"attribute:ecs.instance-type =~ t2.*" or
"attribute:ecs.availability-zone in [us-west-2a, us-west-2b]".

- rule: {"required":true}

## Validation Rules

- `fargate_requires_awsvpc`: Fargate task definitions must use the 'awsvpc' network mode -- leave network_mode unset (it defaults to awsvpc) or set it to 'awsvpc'
- `fargate_requires_task_sizing`: Fargate task definitions must set task-level cpu and memory (e.g. cpu: 256, memory: 512)
- `network_mode_valid`: network_mode must be one of: awsvpc, bridge, host, none
- `ephemeral_storage_range`: ephemeral_storage_gib must be between 21 and 200 when set (Fargate includes 20 GiB by default)
- `at_least_one_essential_container`: at least one container must be essential (essential defaults to true when unset) -- when every essential container exits, the task stops
- `fargate_awslogs_requires_execution_role`: Fargate tasks using the awslogs driver (which the default logging wiring does) need an execution_role that can write logs -- reference an AwsIamRole carrying AmazonECSTaskExecutionRolePolicy, or disable logging
- `ipc_mode_valid`: ipc_mode must be 'host', 'task', or 'none' when set
- `pid_mode_valid`: pid_mode must be 'host' or 'task' when set
- `fargate_forbids_ipc_mode`: ipc_mode is EC2-only -- Fargate task definitions may not set it
- `fargate_pid_mode_task_only`: on Fargate pid_mode may only be 'task' -- 'host' requires an EC2-only task definition
- `secret_environment_requires_execution_role`: a container sets secret_environment, which the ECS agent reads as the task's execution_role -- set execution_role (an AwsIamRole's role_arn)
- `fargate_forbids_placement_constraints`: placement_constraints are EC2-only -- Fargate task definitions may not declare them

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEcsTaskDefinition, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.task_definition_arn` | `string` | The full ARN of the registered revision, including the revision number (e.g. "arn:aws:ecs:us-west-2:123456789012:task-definition/api:7"). The primary handle an ECS service references via status.outputs.task_definition_arn -- because it carries the revision, each newly registered revision changes this output and rolls the referencing service on its next deployment. |
| `status.outputs.arn_without_revision` | `string` | The ARN without the revision suffix (e.g. "arn:aws:ecs:us-west-2: 123456789012:task-definition/api"), for consumers that should always track the family's latest ACTIVE revision instead of a pinned one. |
| `status.outputs.family` | `string` | The family name (metadata.name) the revisions are registered under. |
| `status.outputs.revision` | `int64` | The revision number this deployment registered (e.g. 7). Revisions are immutable; every spec change registers the next number. |
| `status.outputs.log_group_name` | `string` | The name of the CloudWatch log group the task's containers log to -- the auto-created "/ecs/<family>" group or the referenced existing one. Empty when task-level logging is disabled and no container configures its own awslogs driver. |
| `status.outputs.log_group_arn` | `string` | The ARN of that CloudWatch log group. Empty under the same conditions as log_group_name, and when the group is referenced by name rather than created by the modules. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.executionRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.taskRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.volumes[].efs.fileSystemId` | AwsElasticFileSystem | `status.outputs.file_system_id` |
| `spec.volumes[].efs.accessPointId` | AwsEfsAccessPoint | `status.outputs.access_point_id` |
| `spec.volumes[].s3files.fileSystemArn` | AwsS3Bucket | `status.outputs.bucket_arn` |
| `spec.logging.logGroup` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEcsService | `spec.taskDefinition` | `status.outputs.task_definition_arn` |
| AwsEventBridgePipe | `spec.targetParameters.ecsTask.taskDefinitionArn` | `status.outputs.task_definition_arn` |
| AwsEventBridgeRule | `spec.targets[].ecsTarget.taskDefinitionArn` | `status.outputs.task_definition_arn` |
| AwsEventBridgeScheduler | `spec.target.ecsParameters.taskDefinitionArn` | `status.outputs.task_definition_arn` |

## See Also

- [Overview](../README.md)
