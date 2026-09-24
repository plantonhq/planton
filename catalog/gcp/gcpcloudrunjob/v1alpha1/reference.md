# GcpCloudRunJob

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudRunJobSpec defines a Cloud Run job (`google_cloud_run_v2_job`) —
a run-to-completion container workload that executes a fixed number of
parallel tasks and exits. Unlike a Cloud Run service (request-serving,
always-on endpoint, traffic splitting), a job is batch-shaped: you define
a task template once, then each run (an "execution") stamps out
task_count tasks with up to parallelism running concurrently.

Jobs share the same container/volume/VPC vocabulary as GcpCloudRun but
omit serving concerns — no ingress, no traffic, no liveness/readiness
probes (a startup probe IS supported: it gates task start, and a task
that never passes it is shut down and retried). Trigger executions with
`gcloud run jobs execute`, Cloud Scheduler, Eventarc, or declaratively
at deploy time via start_execution_token / run_execution_token; this
resource owns the job definition, not individual runs.

## Example

```yaml
# Development manifest for GcpCloudRunJob — exercises batch semantics
# (multi-task parallelism, secret env, Cloud SQL and GCS volumes with a
# sub_path mount, a TCP startup probe against the declared port, execution
# metadata, a run-on-deploy token, direct VPC egress, and the teardown
# policy) for offline plan verification.
#
# Usage: planton tofu plan --manifest catalog/gcp/gcpcloudrunjob/e2e/manifest.yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunJob
metadata:
  name: hack-cloud-run-job
  id: cldrunjob-hack-001
  org: planton-dev
  env: dev
spec:
  region: us-central1
  jobName: hack-etl
  labels:
    team: platform
  executionLabels:
    cost-center: "1234"
  executionAnnotations:
    external-tool/trace: enabled
  template:
    containers:
      - name: worker
        image: us-docker.pkg.dev/cloudrun/container/job:latest
        env:
          - name: BATCH_MODE
            value: full
          - name: DB_PASSWORD
            valueFromSecret:
              secret: hack-db-password
              version: latest
          # A Planton secret: the component keeps it in a Secret Manager
          # secret of its own, readable only by the task identity.
          - name: API_TOKEN
            secretValue: $secret/hack-api-token
        resources:
          cpu: "2"
          memory: 2Gi
        volumeMounts:
          - name: cloudsql
            mountPath: /cloudsql
          - name: datalake
            mountPath: /data
            subPath: batch-input
        ports:
          containerPort: 8080
        startupProbe:
          periodSeconds: 2
          failureThreshold: 3
          tcpSocket: {}
    volumes:
      - name: cloudsql
        cloudSqlInstance:
          instances:
            - value: my-project:us-central1:hack-db
      - name: datalake
        gcs:
          bucket:
            value: hack-datalake-bucket
          readOnly: true
          mountOptions:
            - implicit-dirs
    timeoutSeconds: 1800
    maxRetries: 2
    vpcAccess:
      networkInterfaces:
        - network:
            value: hack-vpc
          subnetwork:
            value: hack-subnet
      egress: PRIVATE_RANGES_ONLY
    executionEnvironment: EXECUTION_ENVIRONMENT_GEN2
  taskCount: 5
  parallelism: 2
  startExecutionToken: hack-run-once
  deletionProtection: false
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.jobName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.template` | `GcpCloudRunJobTemplate` | yes |  |  |
| `spec.template.containers` | `[]GcpCloudRunJobContainer` | yes |  |  |
| `spec.template.containers[].name` | `string` |  |  |  |
| `spec.template.containers[].image` | `string` | yes |  |  |
| `spec.template.containers[].command` | `[]string` |  |  |  |
| `spec.template.containers[].args` | `[]string` |  |  |  |
| `spec.template.containers[].env` | `[]GcpCloudRunJobEnvVar` |  |  |  |
| `spec.template.containers[].env[].name` | `string` | yes |  |  |
| `spec.template.containers[].env[].value` | `string` (no secrets: use `secretValue`) |  |  |  |
| `spec.template.containers[].env[].valueFromSecret` | `GcpCloudRunJobSecretEnvSource` |  |  |  |
| `spec.template.containers[].env[].valueFromSecret.secret` | `string` | yes |  |  |
| `spec.template.containers[].env[].valueFromSecret.version` | `string` |  | `latest` |  |
| `spec.template.containers[].env[].secretValue` | `string` (sensitive) |  |  |  |
| `spec.template.containers[].resources` | `GcpCloudRunJobContainerResources` |  |  |  |
| `spec.template.containers[].resources.cpu` | `string` |  |  |  |
| `spec.template.containers[].resources.memory` | `string` |  |  |  |
| `spec.template.containers[].volumeMounts` | `[]GcpCloudRunJobVolumeMount` |  |  |  |
| `spec.template.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.template.containers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.template.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.template.containers[].workingDir` | `string` |  |  |  |
| `spec.template.containers[].dependsOn` | `[]string` |  |  |  |
| `spec.template.containers[].ports` | `GcpCloudRunJobContainerPort` |  |  |  |
| `spec.template.containers[].ports.containerPort` | `int32` |  |  |  |
| `spec.template.containers[].ports.name` | `string` |  |  |  |
| `spec.template.containers[].startupProbe` | `GcpCloudRunJobProbe` |  |  |  |
| `spec.template.containers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.httpGet` | `GcpCloudRunJobHttpGetAction` |  |  |  |
| `spec.template.containers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.template.containers[].startupProbe.httpGet.port` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.httpGet.httpHeaders` | `[]GcpCloudRunJobHttpHeader` |  |  |  |
| `spec.template.containers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.template.containers[].startupProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.template.containers[].startupProbe.tcpSocket` | `GcpCloudRunJobTcpSocketAction` |  |  |  |
| `spec.template.containers[].startupProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.grpc` | `GcpCloudRunJobGrpcAction` |  |  |  |
| `spec.template.containers[].startupProbe.grpc.port` | `int32` |  |  |  |
| `spec.template.containers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.template.volumes` | `[]GcpCloudRunJobVolume` |  |  |  |
| `spec.template.volumes[].name` | `string` | yes |  |  |
| `spec.template.volumes[].cloudSqlInstance` | `GcpCloudRunJobVolumeCloudSql` |  |  |  |
| `spec.template.volumes[].cloudSqlInstance.instances` | `[]string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.connection_name`) |
| `spec.template.volumes[].secret` | `GcpCloudRunJobVolumeSecret` |  |  |  |
| `spec.template.volumes[].secret.secret` | `string` | yes |  |  |
| `spec.template.volumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.template.volumes[].secret.items` | `[]GcpCloudRunJobVolumeSecretItem` |  |  |  |
| `spec.template.volumes[].secret.items[].path` | `string` | yes |  |  |
| `spec.template.volumes[].secret.items[].version` | `string` |  | `latest` |  |
| `spec.template.volumes[].secret.items[].mode` | `int32` |  |  |  |
| `spec.template.volumes[].emptyDir` | `GcpCloudRunJobVolumeEmptyDir` |  |  |  |
| `spec.template.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.template.volumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.template.volumes[].gcs` | `GcpCloudRunJobVolumeGcs` |  |  |  |
| `spec.template.volumes[].gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.template.volumes[].gcs.readOnly` | `bool` |  |  |  |
| `spec.template.volumes[].gcs.mountOptions` | `[]string` |  |  |  |
| `spec.template.volumes[].nfs` | `GcpCloudRunJobVolumeNfs` |  |  |  |
| `spec.template.volumes[].nfs.server` | `string` | yes |  |  |
| `spec.template.volumes[].nfs.path` | `string` | yes |  |  |
| `spec.template.volumes[].nfs.readOnly` | `bool` |  |  |  |
| `spec.template.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.template.executionEnvironment` | `enum` |  | `EXECUTION_ENVIRONMENT_GEN2` |  |
| `spec.template.encryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.template.timeoutSeconds` | `int32` |  |  |  |
| `spec.template.maxRetries` | `int32` |  |  |  |
| `spec.template.vpcAccess` | `GcpCloudRunJobVpcAccess` |  |  |  |
| `spec.template.vpcAccess.connector` | `string \| valueFrom` |  |  | GcpServerlessVpcConnector (`status.outputs.self_link`) |
| `spec.template.vpcAccess.networkInterfaces` | `[]GcpCloudRunJobNetworkInterface` |  |  |  |
| `spec.template.vpcAccess.networkInterfaces[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.template.vpcAccess.networkInterfaces[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.template.vpcAccess.networkInterfaces[].tags` | `[]string` |  |  |  |
| `spec.template.vpcAccess.egress` | `string` |  |  |  |
| `spec.template.nodeSelector` | `GcpCloudRunJobNodeSelector` |  |  |  |
| `spec.template.nodeSelector.accelerator` | `string` | yes |  |  |
| `spec.taskCount` | `int32` |  |  |  |
| `spec.parallelism` | `int32` |  |  |  |
| `spec.launchStage` | `string` |  |  |  |
| `spec.binaryAuthorization` | `GcpCloudRunJobBinaryAuthorization` |  |  |  |
| `spec.binaryAuthorization.useDefault` | `bool` |  |  |  |
| `spec.binaryAuthorization.policy` | `string` |  |  |  |
| `spec.binaryAuthorization.breakglassJustification` | `string` |  |  |  |
| `spec.gpuZonalRedundancyDisabled` | `bool` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.executionLabels` | `map<string, string>` |  |  |  |
| `spec.executionAnnotations` | `map<string, string>` |  |  |  |
| `spec.startExecutionToken` | `string` |  |  |  |
| `spec.runExecutionToken` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the job is created in. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region the job is deployed in, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.jobName

`string`

Name of the Cloud Run job in GCP. Immutable. If not specified,
defaults to metadata.name. Must be 1-63 characters: lowercase letters,
digits, and hyphens; starting with a letter.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.labels

`map<string, string>`

Labels applied to the job object. User labels are shared with
Google's billing system. Keys and values in the `run.googleapis.com`,
`cloud.googleapis.com`, `serving.knative.dev`, and
`autoscaling.knative.dev` namespaces are rejected by the API.

### spec.annotations

`map<string, string>`

Unstructured metadata preserved by external tools. Not queryable.
System namespaces (`run.googleapis.com`, etc.) are rejected on create.

### spec.template

`GcpCloudRunJobTemplate` · required

The task template every execution runs: containers, volumes,
networking, hardware, and per-task limits. Required.

- rule: {"required":true}

### spec.template.containers

`[]GcpCloudRunJobContainer` · required

The containers that make up one task. The first container is the main
worker; additional containers are sidecars sharing the task's network
namespace (localhost) and volumes. Use depends_on for startup ordering.

- rule: {"repeated":{"minItems":"1"}}
- rule: the startup window (failure_threshold × period_seconds, defaults 3 × 10) cannot exceed 240 seconds

### spec.template.containers[].name

`string`

Name of the container. Required when the task runs more than one
container (depends_on refers to these names).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.template.containers[].image

`string` · required

Container image URL, e.g. "us-docker.pkg.dev/project/repo/worker:1.0.0".

Cloud Run pulls PRIVATE images only from Artifact Registry (or the
legacy Container Registry) in a project its service agent can read;
public Docker Hub and GHCR images run directly. There is no field here
for a registry login and Google accepts none. For a private image in
another registry, push it to Artifact Registry, or declare a
GcpArtifactRegistryRepo in REMOTE_REPOSITORY mode that proxies that
registry and point this URL at it.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.containers[].command

`[]string`

Entrypoint array — overrides the image's ENTRYPOINT.

### spec.template.containers[].args

`[]string`

Arguments to the entrypoint — overrides the image's CMD.

### spec.template.containers[].env

`[]GcpCloudRunJobEnvVar`

Environment variables. Each entry carries either a literal value or a
Secret Manager reference resolved at task start.

- rule: an environment variable takes exactly one of a literal value, a Secret Manager reference, or a secret value

### spec.template.containers[].env[].name

`string` · required

Variable name, e.g. "BATCH_SIZE". Must not start with a digit.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.template.containers[].env[].value

`string` · no secrets

Literal value, written into the job's task template where anyone who
can view the job reads it. Fine for configuration; never a credential
-- a credential goes in secret_value (or value_from_secret).

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretValue`, which keeps it in a secret store the workload reads by reference
### spec.template.containers[].env[].valueFromSecret

`GcpCloudRunJobSecretEnvSource`

A Secret Manager secret you already own, resolved at task start.

### spec.template.containers[].env[].valueFromSecret.secret

`string` · required

The secret: a short name or full resource name (projects/*/secrets/*).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.containers[].env[].valueFromSecret.version

`string`

Secret version: a version number or "latest".

- default: `latest`

### spec.template.containers[].env[].secretValue

`string` · sensitive

A secret value this component keeps in Secret Manager for you. It
creates one secret for this variable, replicated only in the job's
region, stores the value as a version, grants the job's runtime
identity secretAccessor on that secret alone, and points the variable at
that exact version -- the task template carries a reference, never the
value. A changed value adds a version and updates the template, so
rotation is a deploy; destroying the job removes the secret.

### spec.template.containers[].resources

`GcpCloudRunJobContainerResources`

CPU and memory limits for this container.

### spec.template.containers[].resources.cpu

`string`

CPU limit, e.g. "1", "2", "4", "8" or a fraction like "0.5"/"500m".
GPU tasks need at least "4".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]+m|[0-9]+(\\.[0-9]+)?)$"}}

### spec.template.containers[].resources.memory

`string`

Memory limit with unit suffix, e.g. "512Mi", "2Gi", "16Gi".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.template.containers[].volumeMounts

`[]GcpCloudRunJobVolumeMount`

Volumes (declared at the template level) mounted into this container.

### spec.template.containers[].volumeMounts[].name

`string` · required

Name of a volume declared in template.volumes.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.containers[].volumeMounts[].mountPath

`string` · required

Absolute path in the container. Cloud SQL volumes must mount at "/cloudsql".

- rule: {"required":true,"string":{"pattern":"^/.*$"}}

### spec.template.containers[].volumeMounts[].subPath

`string`

Path WITHIN the volume to mount instead of its root — e.g. mount only
one secret item or one bucket directory. Relative path; empty mounts
the volume root.

### spec.template.containers[].workingDir

`string`

Working directory for the entrypoint.

### spec.template.containers[].dependsOn

`[]string`

Names of containers this one waits for before starting.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.template.containers[].ports

`GcpCloudRunJobContainerPort`

The port this container listens on — for jobs this exists to give
probes a target (there is no request traffic). If unset and a probe
needs a port, the probe names one explicitly.

### spec.template.containers[].ports.containerPort

`int32` · optional (explicit presence)

Port number the container listens on.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.template.containers[].ports.name

`string`

Protocol selector: "http1" (default) or "h2c".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","http1","h2c"]}}

### spec.template.containers[].startupProbe

`GcpCloudRunJobProbe`

Probe that gates task start: the task's work does not begin (and
depends_on waiters stay blocked) until this succeeds; a container
that never passes is shut down and the task retried per max_retries.
Supports HTTP, TCP, and gRPC checks — the ONLY probe type jobs have
(no liveness/readiness; those are serving concerns).

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.template.containers[].startupProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds to wait after container start before the first probe (0-240).

- rule: {"int32":{"lte":240,"gte":0}}

### spec.template.containers[].startupProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which a single probe attempt times out (1-240; GCP
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":240,"gte":1}}

### spec.template.containers[].startupProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-240; GCP default 10).

- rule: {"int32":{"lte":240,"gte":1}}

### spec.template.containers[].startupProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures after which the startup probe fails and the
container is shut down (GCP default 3).

- rule: {"int32":{"gte":1}}

### spec.template.containers[].startupProbe.httpGet

`GcpCloudRunJobHttpGetAction`

HTTP GET against a path on the container; 2xx is success. The job
code must expose the health endpoint itself.

### spec.template.containers[].startupProbe.httpGet.path

`string`

Path to probe, e.g. "/healthz". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.template.containers[].startupProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. If unset, the container's declared port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.template.containers[].startupProbe.httpGet.httpHeaders

`[]GcpCloudRunJobHttpHeader`

Custom headers sent with the probe request.

### spec.template.containers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.containers[].startupProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.template.containers[].startupProbe.tcpSocket

`GcpCloudRunJobTcpSocketAction`

TCP connect to a port; a successful connection is success. The
simplest probe — no code changes needed beyond listening.

### spec.template.containers[].startupProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to. If unset, the container's declared port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.template.containers[].startupProbe.grpc

`GcpCloudRunJobGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check),
which the job code must implement.

### spec.template.containers[].startupProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on. If unset, the container's
declared port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.template.containers[].startupProbe.grpc.service

`string`

Service name passed to the health check. If empty, overall server
health is checked.

### spec.template.volumes

`[]GcpCloudRunJobVolume`

Named volumes tasks can mount: Cloud SQL sockets, Secret Manager
material, scratch space, GCS buckets (FUSE), and NFS shares.

### spec.template.volumes[].name

`string` · required

Volume name referenced by volume_mounts entries.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.template.volumes[].cloudSqlInstance

`GcpCloudRunJobVolumeCloudSql`

Cloud SQL Unix sockets under the mount path.

### spec.template.volumes[].cloudSqlInstance.instances

`[]string | valueFrom` · required

Cloud SQL instance connection names (project:region:instance).

- references: GcpCloudSql (`status.outputs.connection_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.connection_name}} -- a bare string does not parse

### spec.template.volumes[].secret

`GcpCloudRunJobVolumeSecret`

Secret Manager secret versions exposed as files.

### spec.template.volumes[].secret.secret

`string` · required

The secret: a short name or full resource name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.volumes[].secret.defaultMode

`int32` · optional (explicit presence)

Default Unix permission mode for projected files, in decimal (e.g. 292 = 0444).

- rule: {"int32":{"lte":511,"gte":0}}

### spec.template.volumes[].secret.items

`[]GcpCloudRunJobVolumeSecretItem`

Which versions land at which relative paths.

### spec.template.volumes[].secret.items[].path

`string` · required

Relative path under the volume's mount path.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.volumes[].secret.items[].version

`string`

Secret version: a version number or "latest".

- default: `latest`

### spec.template.volumes[].secret.items[].mode

`int32` · optional (explicit presence)

Unix permission mode for this file, in decimal.

- rule: {"int32":{"lte":511,"gte":0}}

### spec.template.volumes[].emptyDir

`GcpCloudRunJobVolumeEmptyDir`

Ephemeral scratch space, in-memory or disk-backed.

### spec.template.volumes[].emptyDir.medium

`string`

Backing medium: MEMORY (default) or DISK.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","MEMORY","DISK"]}}

### spec.template.volumes[].emptyDir.sizeLimit

`string`

Capacity limit with unit suffix, e.g. "512Mi", "2Gi".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.template.volumes[].gcs

`GcpCloudRunJobVolumeGcs`

A GCS bucket mounted via Cloud Storage FUSE (requires GEN2).

### spec.template.volumes[].gcs.bucket

`string | valueFrom` · required

The bucket to mount. Accepts a literal name or a GcpGcsBucket reference.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.template.volumes[].gcs.readOnly

`bool`

Mount read-only.

### spec.template.volumes[].gcs.mountOptions

`[]string`

Flags passed to the gcsfuse command mounting this volume, without
leading dashes (e.g. "implicit-dirs", "only-dir=media",
"file-cache-max-size-mb=512").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.template.volumes[].nfs

`GcpCloudRunJobVolumeNfs`

An NFS share mounted into the task (requires GEN2 and VPC reachability).

### spec.template.volumes[].nfs.server

`string` · required

Hostname or IP of the NFS server.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.template.volumes[].nfs.path

`string` · required

Exported path on the server, e.g. "/share1".

- rule: {"required":true,"string":{"pattern":"^/.*$"}}

### spec.template.volumes[].nfs.readOnly

`bool`

Mount read-only.

### spec.template.serviceAccount

`string | valueFrom`

Email of the IAM service account each task runs as. Accepts a literal
email or a reference to a GcpServiceAccount resource. If omitted, the
project's Compute Engine default service account is used.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.template.executionEnvironment

`enum`

Sandbox generation tasks execute in. GEN2 (recommended) offers full
Linux compatibility (required for GCS/NFS volumes); GEN1 starts faster
with a gVisor-restricted syscall surface.

- default: `EXECUTION_ENVIRONMENT_GEN2`

Allowed values (use exactly as shown):

- `EXECUTION_ENVIRONMENT_UNSPECIFIED` -- Unspecified — GCP selects per workload.
- `EXECUTION_ENVIRONMENT_GEN1` -- First generation: gVisor-restricted, no network file systems.
- `EXECUTION_ENVIRONMENT_GEN2` -- Second generation: full Linux; required for GCS and NFS volumes.

### spec.template.encryptionKey

`string | valueFrom`

Customer-managed encryption key (CMEK) encrypting deployed container
images. Accepts a full crypto key ID or a reference to a GcpKmsKey
resource. The key must be in the same region as the job.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.template.timeoutSeconds

`int32` · optional (explicit presence)

Maximum time one task attempt may run before GCP marks it failed, in
seconds (1-86400). If unset, GCP defaults to 600 (10 minutes). Each
retry gets a fresh timeout budget.

- rule: {"int32":{"lte":86400,"gte":1}}

### spec.template.maxRetries

`int32` · optional (explicit presence)

Retries per task before marking it failed (>= 0). GCP defaults to 3.
Set 0 for fail-fast batch work with no retries.

- rule: {"int32":{"gte":0}}

### spec.template.vpcAccess

`GcpCloudRunJobVpcAccess`

Private networking for OUTBOUND traffic from tasks.

- rule: use direct VPC egress (network_interfaces) or a Serverless VPC Access connector, not both

### spec.template.vpcAccess.connector

`string | valueFrom`

Serverless VPC Access connector (legacy mechanism). Full resource name
(projects/*/locations/*/connectors/*) or a reference to a
GcpServerlessVpcConnector resource.

- references: GcpServerlessVpcConnector (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServerlessVpcConnector, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.template.vpcAccess.networkInterfaces

`[]GcpCloudRunJobNetworkInterface`

Direct VPC egress: tasks get IPs in the subnetwork.

### spec.template.vpcAccess.networkInterfaces[].network

`string | valueFrom`

The VPC network. Accepts a literal name or a GcpVpcNetwork reference.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.template.vpcAccess.networkInterfaces[].subnetwork

`string | valueFrom`

The subnetwork tasks draw IPs from. Must be in the job's region.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.template.vpcAccess.networkInterfaces[].tags

`[]string`

Network tags applied to tasks — how VPC firewall rules select egress.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.template.vpcAccess.egress

`string`

Which egress uses the VPC path: ALL_TRAFFIC or PRIVATE_RANGES_ONLY.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALL_TRAFFIC","PRIVATE_RANGES_ONLY"]}}

### spec.template.nodeSelector

`GcpCloudRunJobNodeSelector`

Hardware requirements for GPU batch workloads. Setting an accelerator
(e.g. "nvidia-l4") gives each task one GPU; container resource limits
must meet Cloud Run's GPU minimums (4 CPU / 16Gi recommended).

### spec.template.nodeSelector.accelerator

`string` · required

GPU accelerator type each task gets, e.g. "nvidia-l4".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.taskCount

`int32` · optional (explicit presence)

Desired number of tasks each execution runs (>= 1). Setting 1 means
success of that single task signals execution success. If unset, GCP
defaults to 1.

- rule: {"int32":{"gte":1}}

### spec.parallelism

`int32` · optional (explicit presence)

Maximum tasks running concurrently during an execution (>= 0). Must be
<= task_count when both are set. If 0 or unset, GCP uses the maximum
possible for that run.

- rule: {"int32":{"gte":0}}

### spec.launchStage

`string`

Launch-stage gate for preview Cloud Run features. Set BETA (or ALPHA)
only when the spec uses features GCP rejects at the default GA stage.
The provider enum admits more values (UNIMPLEMENTED, PRELAUNCH,
EARLY_ACCESS, DEPRECATED), but Cloud Run itself supports only
ALPHA/BETA/GA — this list encodes the API truth, deliberately.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALPHA","BETA","GA"]}}

### spec.binaryAuthorization

`GcpCloudRunJobBinaryAuthorization`

Binary Authorization: only container images that pass the policy's
attestation checks may deploy.

- rule: use the project default policy (use_default) or name a specific policy, not both

### spec.binaryAuthorization.useDefault

`bool`

Evaluate deploys against the project's default Binary Authorization policy.

### spec.binaryAuthorization.policy

`string`

Evaluate deploys against a specific platform policy.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.binaryAuthorization.breakglassJustification

`string`

Justification recorded when a break-glass deploy bypasses the policy.

### spec.gpuZonalRedundancyDisabled

`bool`

Opts a GPU job out of zonal redundancy: tasks may run from a single
zone for cheaper GPU capacity. Only meaningful when
template.node_selector is set.

### spec.deletionProtection

`bool` · optional (explicit presence)

Prevents the job from being destroyed while true. Defaults to true
(matching GCP's posture): a delete fails until this is set to false.

- default: `true`

### spec.executionLabels

`map<string, string>`

Labels stamped on every EXECUTION the job creates (visible on
execution objects and in billing breakdowns), as opposed to `labels`,
which lands on the job object itself. Same namespace restrictions as
job labels.

### spec.executionAnnotations

`map<string, string>`

Annotations stamped on every EXECUTION the job creates, as opposed to
`annotations`, which lands on the job object. Same namespace
restrictions.

### spec.startExecutionToken

`string`

Declarative run-on-deploy: a unique suffix string that triggers a new
execution when the job is created or updated, with the job counted
READY as soon as the execution successfully STARTS. Changing the
token on a later update triggers another run. The combined length of
job name and token must stay under 63 characters. Set at most one of
the two token fields.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63"}}

### spec.runExecutionToken

`string`

Declarative run-on-deploy like start_execution_token, but the job is
counted READY only when the triggered execution successfully
COMPLETES — deploy-and-verify semantics for migrations and one-shot
setup work. The combined length of job name and token must stay under
63 characters. Set at most one of the two token fields.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63"}}

### spec.deletionPolicy

`string`

What happens to the Cloud Run job when this resource is destroyed:
  "" / "DELETE" -- the job is deleted (default)
  "PREVENT"     -- destroy operations fail while this is set
  "ABANDON"     -- the job is removed from management but left
                   running in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the job at creation, as a map of
tagKeys/{tag_key_id} to tagValues/{tag_value_id} — the tag bindings
that organization policies, IAM conditions, and cost reports key on.
Immutable: changing the map replaces the job (Cloud Run applies tags
only at create), so plan tag changes as a recreate.

- rule: {"map":{"keys":{"string":{"pattern":"^tagKeys/[0-9]+$"}},"values":{"string":{"pattern":"^tagValues/[0-9]+$"}}}}

## Validation Rules

- `gpu.redundancy_requires_accelerator`: gpu_zonal_redundancy_disabled only applies to GPU jobs — set template.node_selector.accelerator
- `parallelism_lte_task_count`: parallelism cannot exceed task_count when both are set
- `execution_token.start_xor_run`: start_execution_token and run_execution_token conflict — a deploy triggers at most one kind of execution

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudRunJob, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.job_name` | `string` | Name of the Cloud Run job as created in GCP — the handle gcloud and Scheduler reference when triggering executions. |
| `status.outputs.location` | `string` | Region the job is deployed in (plain region name). |
| `status.outputs.uid` | `string` | Server-assigned unique identifier of the job, stable across its lifetime and never reused after deletion. |
| `status.outputs.latest_created_execution` | `string` | Name of the most recently created execution, if any. Empty until the job has been run at least once. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.template.volumes[].cloudSqlInstance.instances` | GcpCloudSql | `status.outputs.connection_name` |
| `spec.template.volumes[].gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.template.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.template.encryptionKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.template.vpcAccess.connector` | GcpServerlessVpcConnector | `status.outputs.self_link` |
| `spec.template.vpcAccess.networkInterfaces[].network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.template.vpcAccess.networkInterfaces[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_name` |

## See Also

- [Overview](../README.md)
