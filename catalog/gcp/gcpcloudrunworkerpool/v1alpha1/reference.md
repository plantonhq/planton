# GcpCloudRunWorkerPool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudRunWorkerPoolSpec defines a Cloud Run worker pool
(`google_cloud_run_v2_worker_pool`): a pool of always-running container
instances with no HTTP front door -- queue consumers, schedulers, event
pullers, and background workers that do their own work instead of
answering requests.

A worker pool is a Cloud Run service with the request path removed. It
keeps the revision template (containers, volumes, VPC egress, GPU) and the
revision model (every template change stamps out a new immutable
revision, instance_splits decides how many instances each revision runs),
but it has no port, no ingress, no invoker IAM, no URL, and no
request-driven autoscaling: instances scale MANUALLY to a count you set,
or AUTOMATICALLY between bounds by a signal you drive (for example a
Pub/Sub backlog metric through Cloud Monitoring), never by traffic. CPU is
always allocated, so billing is per instance-hour, not per request.

Choose GcpCloudRun for anything that serves HTTP; GcpCloudRunJob for work
that runs to completion and exits; this kind for a process that should be
up all the time and pull its own work.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRunWorkerPool
metadata:
  name: orders-worker
spec:
  projectId:
    value: my-gcp-project
  workerPoolName: orders-worker
  description: Pulls order events from Pub/Sub and writes them to the orders database
  region: us-central1
  containers:
    - name: worker
      image: us-docker.pkg.dev/my-gcp-project/apps/orders-worker:1.4.0
      env:
        - name: SUBSCRIPTION
          value: projects/my-gcp-project/subscriptions/orders-events
        - name: DB_PASSWORD
          valueFromSecret:
            secret: orders-db-password
            version: latest
        # A Planton secret: the component keeps it in a Secret Manager
        # secret of its own, readable only by the pool's identity.
        - name: API_TOKEN
          secretValue: $secret/orders-api-token
      resources:
        cpu: "1"
        memory: 512Mi
      # A worker has no serving port; the probe names the health listener.
      livenessProbe:
        httpGet:
          path: /healthz
          port: 8081
        periodSeconds: 30
        failureThreshold: 3
  # A dedicated least-privilege identity, never the Compute default.
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: orders-worker-sa
      fieldPath: status.outputs.email
  # Direct VPC egress so the worker reaches Cloud SQL private IP and
  # Memorystore; public egress keeps Cloud Run's own path.
  vpcAccess:
    networkInterfaces:
      - subnetwork:
          valueFrom:
            kind: GcpSubnetwork
            name: workers-subnet
            fieldPath: status.outputs.subnetwork_name
    egress: PRIVATE_RANGES_ONLY
  # Exactly two instances, around the clock.
  scaling:
    scalingMode: MANUAL
    manualInstanceCount: 2
  labels:
    tier: production
  deletionProtection: true
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.workerPoolName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.containers` | `[]GcpCloudRunWorkerPoolContainer` | yes |  |  |
| `spec.containers[].name` | `string` |  |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].command` | `[]string` |  |  |  |
| `spec.containers[].args` | `[]string` |  |  |  |
| `spec.containers[].env` | `[]GcpCloudRunWorkerPoolEnvVar` |  |  |  |
| `spec.containers[].env[].name` | `string` | yes |  |  |
| `spec.containers[].env[].value` | `string` (no secrets: use `secretValue`) |  |  |  |
| `spec.containers[].env[].valueFromSecret` | `GcpCloudRunWorkerPoolSecretEnvSource` |  |  |  |
| `spec.containers[].env[].valueFromSecret.secret` | `string` | yes |  |  |
| `spec.containers[].env[].valueFromSecret.version` | `string` |  | `latest` |  |
| `spec.containers[].env[].secretValue` | `string` (sensitive) |  |  |  |
| `spec.containers[].resources` | `GcpCloudRunWorkerPoolContainerResources` |  |  |  |
| `spec.containers[].resources.cpu` | `string` |  |  |  |
| `spec.containers[].resources.memory` | `string` |  |  |  |
| `spec.containers[].volumeMounts` | `[]GcpCloudRunWorkerPoolVolumeMount` |  |  |  |
| `spec.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.containers[].workingDir` | `string` |  |  |  |
| `spec.containers[].startupProbe` | `GcpCloudRunWorkerPoolStartupProbe` |  |  |  |
| `spec.containers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].startupProbe.httpGet` | `GcpCloudRunWorkerPoolHttpGetAction` |  |  |  |
| `spec.containers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].startupProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders` | `[]GcpCloudRunWorkerPoolHttpHeader` |  |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.containers[].startupProbe.tcpSocket` | `GcpCloudRunWorkerPoolTcpSocketAction` |  |  |  |
| `spec.containers[].startupProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.grpc` | `GcpCloudRunWorkerPoolGrpcAction` |  |  |  |
| `spec.containers[].startupProbe.grpc.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.containers[].livenessProbe` | `GcpCloudRunWorkerPoolLivenessProbe` |  |  |  |
| `spec.containers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.httpGet` | `GcpCloudRunWorkerPoolHttpGetAction` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders` | `[]GcpCloudRunWorkerPoolHttpHeader` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.containers[].livenessProbe.grpc` | `GcpCloudRunWorkerPoolGrpcAction` |  |  |  |
| `spec.containers[].livenessProbe.grpc.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.containers[].dependsOn` | `[]string` |  |  |  |
| `spec.volumes` | `[]GcpCloudRunWorkerPoolVolume` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].cloudSqlInstance` | `GcpCloudRunWorkerPoolVolumeCloudSql` |  |  |  |
| `spec.volumes[].cloudSqlInstance.instances` | `[]string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.connection_name`) |
| `spec.volumes[].secret` | `GcpCloudRunWorkerPoolVolumeSecret` |  |  |  |
| `spec.volumes[].secret.secret` | `string` | yes |  |  |
| `spec.volumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.volumes[].secret.items` | `[]GcpCloudRunWorkerPoolVolumeSecretItem` |  |  |  |
| `spec.volumes[].secret.items[].path` | `string` | yes |  |  |
| `spec.volumes[].secret.items[].version` | `string` |  | `latest` |  |
| `spec.volumes[].secret.items[].mode` | `int32` |  |  |  |
| `spec.volumes[].emptyDir` | `GcpCloudRunWorkerPoolVolumeEmptyDir` |  |  |  |
| `spec.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.volumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.volumes[].gcs` | `GcpCloudRunWorkerPoolVolumeGcs` |  |  |  |
| `spec.volumes[].gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.volumes[].gcs.readOnly` | `bool` |  |  |  |
| `spec.volumes[].gcs.mountOptions` | `[]string` |  |  |  |
| `spec.volumes[].nfs` | `GcpCloudRunWorkerPoolVolumeNfs` |  |  |  |
| `spec.volumes[].nfs.server` | `string` | yes |  |  |
| `spec.volumes[].nfs.path` | `string` | yes |  |  |
| `spec.volumes[].nfs.readOnly` | `bool` |  |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.scaling` | `GcpCloudRunWorkerPoolScaling` |  |  |  |
| `spec.scaling.scalingMode` | `string` |  |  |  |
| `spec.scaling.manualInstanceCount` | `int32` |  |  |  |
| `spec.scaling.minInstanceCount` | `int32` |  |  |  |
| `spec.scaling.maxInstanceCount` | `int32` |  |  |  |
| `spec.instanceSplits` | `[]GcpCloudRunWorkerPoolInstanceSplit` |  |  |  |
| `spec.instanceSplits[].type` | `string` | yes |  |  |
| `spec.instanceSplits[].revision` | `string` |  |  |  |
| `spec.instanceSplits[].percent` | `int32` |  |  |  |
| `spec.encryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.encryptionKeyRevocationAction` | `string` |  |  |  |
| `spec.encryptionKeyShutdownDuration` | `string` |  |  |  |
| `spec.revision` | `string` |  |  |  |
| `spec.revisionLabels` | `map<string, string>` |  |  |  |
| `spec.revisionAnnotations` | `map<string, string>` |  |  |  |
| `spec.vpcAccess` | `GcpCloudRunWorkerPoolVpcAccess` |  |  |  |
| `spec.vpcAccess.connector` | `string \| valueFrom` |  |  | GcpServerlessVpcConnector (`status.outputs.self_link`) |
| `spec.vpcAccess.networkInterfaces` | `[]GcpCloudRunWorkerPoolNetworkInterface` |  |  |  |
| `spec.vpcAccess.networkInterfaces[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.vpcAccess.networkInterfaces[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.vpcAccess.networkInterfaces[].tags` | `[]string` |  |  |  |
| `spec.vpcAccess.egress` | `string` |  |  |  |
| `spec.nodeSelector` | `GcpCloudRunWorkerPoolNodeSelector` |  |  |  |
| `spec.nodeSelector.accelerator` | `string` | yes |  |  |
| `spec.gpuZonalRedundancyDisabled` | `bool` |  |  |  |
| `spec.launchStage` | `string` |  |  |  |
| `spec.binaryAuthorization` | `GcpCloudRunWorkerPoolBinaryAuthorization` |  |  |  |
| `spec.binaryAuthorization.useDefault` | `bool` |  |  |  |
| `spec.binaryAuthorization.policy` | `string` |  |  |  |
| `spec.binaryAuthorization.breakglassJustification` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the worker pool is created in. Accepts a literal
project ID or a reference to a GcpProject resource. If omitted, the
provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region the worker pool runs in, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.workerPoolName

`string`

Name of the worker pool in GCP. Immutable. Defaults to metadata.name.
1-63 characters: lowercase letters, digits, and hyphens, starting with
a letter and ending alphanumeric.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.description

`string`

Human-readable description, shown in the console.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.labels

`map<string, string>`

Labels on the worker pool object, shared with Google's billing system
for cost breakdowns. Keys in the `run.googleapis.com`,
`cloud.googleapis.com`, `serving.knative.dev`, and
`autoscaling.knative.dev` namespaces are rejected by the API.

### spec.annotations

`map<string, string>`

Annotations on the worker pool object: unstructured metadata for
external tools, never behavioral for Cloud Run. Same namespace
restrictions as labels. For metadata stamped on each REVISION use
revision_annotations.

### spec.containers

`[]GcpCloudRunWorkerPoolContainer` · required

The containers that make up one instance. The first container
conventionally does the work; additional containers are sidecars
(collectors, proxies) sharing the instance's localhost and volumes,
ordered by depends_on. No container exposes a port -- a worker pool
receives no requests. Deployment pipelines inject the built image into
containers whose image is left BLANK (the image-slot contract);
authored images are untouched.

- rule: {"repeated":{"minItems":"1"}}

### spec.containers[].name

`string`

Name of the container. Required when the pool runs more than one
container (depends_on refers to these names). If omitted for a single
container, Cloud Run assigns one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.containers[].image

`string` · required

Container image URL, e.g. "us-docker.pkg.dev/project/repo/worker:1.0.0".
Pin a digest or immutable tag for repeatable deploys -- Cloud Run
resolves the image to a digest at revision creation. Private images
are pulled only from Artifact Registry (or the legacy Container
Registry) the Cloud Run service agent can read; public Docker Hub and
GHCR images deploy directly.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].command

`[]string`

Entrypoint array -- overrides the image's ENTRYPOINT. Not executed in
a shell; variable references are not expanded.

### spec.containers[].args

`[]string`

Arguments to the entrypoint -- overrides the image's CMD.

### spec.containers[].env

`[]GcpCloudRunWorkerPoolEnvVar`

Environment variables. Each entry carries a literal value, a Secret
Manager secret you already own, or a secret value this component keeps
in Secret Manager for you.

- rule: an environment variable takes exactly one of a literal value, a Secret Manager reference, or a secret value

### spec.containers[].env[].name

`string` · required

Variable name, e.g. "QUEUE_SUBSCRIPTION". Must not start with a digit.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.containers[].env[].value

`string` · no secrets

Literal value, written into the worker pool's revision template where
anyone who can view the pool reads it, and every past revision keeps it.
Fine for configuration; never a credential -- a credential goes in
secret_value (or value_from_secret).

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretValue`, which keeps it in a secret store the workload reads by reference
### spec.containers[].env[].valueFromSecret

`GcpCloudRunWorkerPoolSecretEnvSource`

A Secret Manager secret you already own, resolved into the variable at
instance start. Rotation is Secret Manager's: with version "latest", new
instances pick up a new version without a deploy.

### spec.containers[].env[].valueFromSecret.secret

`string` · required

The secret: a short name for a secret in the same project
("db-password") or a full resource name (projects/*/secrets/*) for
cross-project reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env[].valueFromSecret.version

`string`

Secret version to resolve: a version number or "latest". GCP requires
an explicit version for env vars -- "latest" is the common choice, at
the cost of new instances silently picking up rotations.

- default: `latest`

### spec.containers[].env[].secretValue

`string` · sensitive

A secret value this component keeps in Secret Manager for you. It
creates one secret for this variable, replicated only in the pool's
region, stores the value as a version, grants the pool's runtime
identity secretAccessor on that secret alone, and points the variable at
that exact version -- the revision carries a reference, never the value.
A changed value adds a version and stamps a new revision, so rotation is
a deploy; destroying the worker pool removes the secret.

### spec.containers[].resources

`GcpCloudRunWorkerPoolContainerResources`

CPU and memory for this container. If omitted, Cloud Run defaults
apply (1 CPU, 512Mi). CPU is always allocated on a worker pool.

### spec.containers[].resources.cpu

`string`

CPU limit: "1", "2", "4", "6", "8" or a fraction like "0.5"/"500m".
4 CPU needs at least 2Gi of memory; 6 or more need 4Gi. GPU workers
need at least "4".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]+m|[0-9]+(\\.[0-9]+)?)$"}}

### spec.containers[].resources.memory

`string`

Memory limit with unit suffix, e.g. "512Mi", "2Gi". Minimums scale
with CPU; GPU workers need at least "16Gi".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.containers[].volumeMounts

`[]GcpCloudRunWorkerPoolVolumeMount`

Volumes (declared at the pool level) mounted into this container's
filesystem.

### spec.containers[].volumeMounts[].name

`string` · required

Name of a volume declared in spec.volumes.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].volumeMounts[].mountPath

`string` · required

Absolute path in the container to mount at. Cloud SQL volumes must
mount at "/cloudsql".

- rule: {"required":true,"string":{"pattern":"^/.*$"}}

### spec.containers[].volumeMounts[].subPath

`string`

Path WITHIN the volume to mount instead of its root. Relative; empty
mounts the volume root.

### spec.containers[].workingDir

`string`

Working directory for the entrypoint. If omitted, the image's WORKDIR
is used.

### spec.containers[].startupProbe

`GcpCloudRunWorkerPoolStartupProbe`

Probe that gates instance start: depends_on waiters stay blocked until
this succeeds, and an instance that never passes is killed. HTTP, TCP,
or gRPC. If omitted, Cloud Run considers the container started once
its process is running.

- rule: probe timeout_seconds cannot exceed period_seconds
- rule: the startup window (failure_threshold x period_seconds, defaults 3 x 10) cannot exceed 240 seconds

### spec.containers[].startupProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds to wait after container start before the first probe (0-240).

- rule: {"int32":{"lte":240,"gte":0}}

### spec.containers[].startupProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which a single probe attempt times out (1-240; GCP
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-240; GCP default 10).

- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures after which startup is considered failed and the
instance is killed (GCP default 3).

- rule: {"int32":{"gte":1}}

### spec.containers[].startupProbe.httpGet

`GcpCloudRunWorkerPoolHttpGetAction`

HTTP GET against a path on a port the container listens on; 2xx is
success. A worker has no serving port, so name the port explicitly.

### spec.containers[].startupProbe.httpGet.path

`string`

Path to probe, e.g. "/healthz". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.containers[].startupProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. A worker pool has no serving port to fall back to, so
set the port the worker's health listener binds.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.httpGet.httpHeaders

`[]GcpCloudRunWorkerPoolHttpHeader`

Custom header sent with the probe request (e.g. an auth header for a
protected health endpoint). At most one header: the pinned Pulumi SDK
models a worker-pool probe's headers as a single header, so both
engines accept the same manifests; the list widens when the SDK does.

- rule: {"repeated":{"maxItems":"1"}}

### spec.containers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].startupProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.containers[].startupProbe.tcpSocket

`GcpCloudRunWorkerPoolTcpSocketAction`

TCP connect to a port; a successful connection is success.

### spec.containers[].startupProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.grpc

`GcpCloudRunWorkerPoolGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check).

### spec.containers[].startupProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.grpc.service

`string`

Service name passed to the health check, letting one server report
per-service health. If empty, overall server health is checked.

### spec.containers[].livenessProbe

`GcpCloudRunWorkerPoolLivenessProbe`

Probe that monitors a running instance: on failure_threshold
consecutive failures the container is restarted. HTTP and gRPC only --
Cloud Run rejects TCP liveness probes. If omitted, instances are never
health-restarted.

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.containers[].livenessProbe.initialDelaySeconds

`int32` · optional (explicit presence)

Seconds to wait after container start before the first probe
(0-3600).

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.containers[].livenessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which a single probe attempt times out (1-3600; GCP
default 1). Must not exceed period_seconds.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.containers[].livenessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (1-3600; GCP default 10).

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.containers[].livenessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures after which the container is restarted (GCP
default 3).

- rule: {"int32":{"gte":1}}

### spec.containers[].livenessProbe.httpGet

`GcpCloudRunWorkerPoolHttpGetAction`

HTTP GET against a path on a port the container listens on; 2xx is
success.

### spec.containers[].livenessProbe.httpGet.path

`string`

Path to probe, e.g. "/healthz". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.containers[].livenessProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. A worker pool has no serving port to fall back to, so
set the port the worker's health listener binds.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].livenessProbe.httpGet.httpHeaders

`[]GcpCloudRunWorkerPoolHttpHeader`

Custom header sent with the probe request (e.g. an auth header for a
protected health endpoint). At most one header: the pinned Pulumi SDK
models a worker-pool probe's headers as a single header, so both
engines accept the same manifests; the list widens when the SDK does.

- rule: {"repeated":{"maxItems":"1"}}

### spec.containers[].livenessProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].livenessProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.containers[].livenessProbe.grpc

`GcpCloudRunWorkerPoolGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check).

### spec.containers[].livenessProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].livenessProbe.grpc.service

`string`

Service name passed to the health check, letting one server report
per-service health. If empty, overall server health is checked.

### spec.containers[].dependsOn

`[]string`

Names of containers this one waits for: this container starts only
after the listed containers pass their startup probes. Immutable --
changing the order replaces the worker pool.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.volumes

`[]GcpCloudRunWorkerPoolVolume`

Named volumes instances can mount: Cloud SQL sockets, Secret Manager
material, scratch space, GCS buckets (FUSE), and NFS shares. A volume
is inert until a container mounts it by name.

### spec.volumes[].name

`string` · required

Volume name referenced by volume_mounts entries.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.volumes[].cloudSqlInstance

`GcpCloudRunWorkerPoolVolumeCloudSql`

Cloud SQL Unix sockets, one per instance, under the mount path
(mount at "/cloudsql"; connect via
"/cloudsql/<project:region:instance>"). GCP manages the proxying.

### spec.volumes[].cloudSqlInstance.instances

`[]string | valueFrom` · required

Cloud SQL instance connection names (project:region:instance).
Accepts literal values or references to GcpCloudSql resources.

- references: GcpCloudSql (`status.outputs.connection_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.connection_name}} -- a bare string does not parse

### spec.volumes[].secret

`GcpCloudRunWorkerPoolVolumeSecret`

Secret Manager secret versions exposed as files.

### spec.volumes[].secret.secret

`string` · required

The secret: a short name for a secret in the same project or a full
resource name (projects/*/secrets/*).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.volumes[].secret.defaultMode

`int32` · optional (explicit presence)

Default Unix permission mode for projected files, in decimal (e.g. 292
= 0444). If unset, GCP defaults to 0444.

- rule: {"int32":{"lte":511,"gte":0}}

### spec.volumes[].secret.items

`[]GcpCloudRunWorkerPoolVolumeSecretItem`

Which versions land at which relative paths. If empty, the "latest"
version is projected at a file named after the secret.

### spec.volumes[].secret.items[].path

`string` · required

Relative path of the file under the volume's mount path.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.volumes[].secret.items[].version

`string`

Secret version to project: a version number or "latest".

- default: `latest`

### spec.volumes[].secret.items[].mode

`int32` · optional (explicit presence)

Unix permission mode for this file, in decimal. Overrides
default_mode.

- rule: {"int32":{"lte":511,"gte":0}}

### spec.volumes[].emptyDir

`GcpCloudRunWorkerPoolVolumeEmptyDir`

Ephemeral scratch space, in-memory (counts against the instance's
memory limit) or disk-backed.

### spec.volumes[].emptyDir.medium

`string`

Backing medium: MEMORY (default -- tmpfs; usage counts against the
containers' memory limits) or DISK.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","MEMORY","DISK"]}}

### spec.volumes[].emptyDir.sizeLimit

`string`

Capacity limit with unit suffix, e.g. "512Mi", "2Gi".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.volumes[].gcs

`GcpCloudRunWorkerPoolVolumeGcs`

A GCS bucket mounted via Cloud Storage FUSE; object storage
semantics apply (no POSIX locking; renames are copies).

### spec.volumes[].gcs.bucket

`string | valueFrom` · required

The bucket to mount. Accepts a literal bucket name or a reference to a
GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.volumes[].gcs.readOnly

`bool`

Mount read-only. Recommended unless the worker genuinely writes --
concurrent writers through FUSE are easy to get wrong.

### spec.volumes[].gcs.mountOptions

`[]string`

Flags passed to the gcsfuse command mounting this volume, without
leading dashes (e.g. "implicit-dirs", "only-dir=media").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.volumes[].nfs

`GcpCloudRunWorkerPoolVolumeNfs`

An NFS share (e.g. Filestore) mounted into the instance. Needs VPC
access to reach the server.

### spec.volumes[].nfs.server

`string` · required

Hostname or IP of the NFS server.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.volumes[].nfs.path

`string` · required

Exported path on the server, e.g. "/share1".

- rule: {"required":true,"string":{"pattern":"^/.*$"}}

### spec.volumes[].nfs.readOnly

`bool`

Mount read-only.

### spec.serviceAccount

`string | valueFrom`

Email of the IAM service account the instances run as -- the identity
whose permissions the code exercises. Accepts a literal email or a
GcpServiceAccount reference. If omitted, the project's Compute Engine
default service account is used -- fine for experiments, too broad for
production; give real workers a dedicated least-privilege identity.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.scaling

`GcpCloudRunWorkerPoolScaling`

How many instances run. MANUAL (Google's default) pins the total to
manual_instance_count; AUTOMATIC lets Cloud Run move between
min_instance_count and max_instance_count on a signal you drive.
If omitted, Google runs the pool in MANUAL mode at its default count.

- rule: manual_instance_count only applies when scaling_mode is MANUAL (or unset, which is MANUAL)
- rule: min_instance_count and max_instance_count only apply when scaling_mode is AUTOMATIC
- rule: min_instance_count cannot exceed max_instance_count

### spec.scaling.scalingMode

`string`

MANUAL (Google's default): the pool runs exactly manual_instance_count
instances. AUTOMATIC: Cloud Run moves the count between
min_instance_count and max_instance_count on a signal you drive.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","AUTOMATIC","MANUAL"]}}

### spec.scaling.manualInstanceCount

`int32` · optional (explicit presence)

Total instances across all revisions in MANUAL mode. 0 parks the pool
(no instances, no bill) without deleting it.

- rule: {"int32":{"gte":0}}

### spec.scaling.minInstanceCount

`int32` · optional (explicit presence)

Lower bound on instances in AUTOMATIC mode, distributed across
revisions by instance_splits.

- rule: {"int32":{"gte":0}}

### spec.scaling.maxInstanceCount

`int32` · optional (explicit presence)

Upper bound on instances in AUTOMATIC mode -- the cost circuit
breaker.

- rule: {"int32":{"gte":1}}

### spec.instanceSplits

`[]GcpCloudRunWorkerPoolInstanceSplit`

How instances are split across revisions. If empty, every instance
runs the latest ready revision -- the right default. Populate for a
gradual rollout: percentages must sum to 100 (enforced by the API at
deploy time).

- rule: REVISION splits must name a revision; LATEST splits must not

### spec.instanceSplits[].type

`string` · required

Assign to the latest ready revision (LATEST) or a named revision
(REVISION).

- rule: {"required":true,"string":{"in":["INSTANCE_SPLIT_ALLOCATION_TYPE_LATEST","INSTANCE_SPLIT_ALLOCATION_TYPE_REVISION"]}}

### spec.instanceSplits[].revision

`string`

Revision name for REVISION splits (see spec.revision for deterministic
naming).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.instanceSplits[].percent

`int32` · optional (explicit presence)

Percent of instances for this revision (0-100). All percents in the
list must sum to 100 -- enforced by the API at deploy time. Unset
means 0.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.encryptionKey

`string | valueFrom`

Customer-managed encryption key (CMEK) that encrypts the deployed
container images: a full crypto key ID
(projects/*/locations/*/keyRings/*/cryptoKeys/*) or a GcpKmsKey
reference. The key must be in the worker pool's region and the Cloud
Run service agent needs encrypter/decrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.encryptionKeyRevocationAction

`string`

What Cloud Run does with running instances if the CMEK is revoked:
  PREVENT_NEW -- no new instances start; running ones keep going
  SHUTDOWN    -- every instance is stopped after
                 encryption_key_shutdown_duration
Only meaningful with encryption_key.

- rule: encryption_key_revocation_action must be PREVENT_NEW or SHUTDOWN

### spec.encryptionKeyShutdownDuration

`string`

Grace period before instances are shut down after a key revocation
under SHUTDOWN, as a seconds duration in whole hours (e.g. "3600s").

- rule: encryption_key_shutdown_duration must be a seconds duration such as "3600s"

### spec.revision

`string`

Explicit name for the next revision, prefixed with the worker pool
name (e.g. "orders-worker-v42"). If omitted (recommended), Cloud Run
generates one. Pin names only when instance_splits routes by
revision; each template change then REQUIRES a new value here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.revisionLabels

`map<string, string>`

Labels stamped on every REVISION the template creates, as opposed to
`labels` on the worker pool object. Same namespace restrictions.

### spec.revisionAnnotations

`map<string, string>`

Annotations stamped on every REVISION the template creates, as
opposed to `annotations` on the worker pool object.

### spec.vpcAccess

`GcpCloudRunWorkerPoolVpcAccess`

Private networking for OUTBOUND traffic: Direct VPC egress
(network_interfaces -- recommended) or a Serverless VPC Access
connector. A worker pool that pulls from Memorystore, Cloud SQL private
IP, or an internal service needs this.

- rule: use direct VPC egress (network_interfaces) or a Serverless VPC Access connector, not both

### spec.vpcAccess.connector

`string | valueFrom`

Serverless VPC Access connector to route egress through. Full resource
name (projects/*/locations/*/connectors/*) or a reference to a
GcpServerlessVpcConnector resource.

- references: GcpServerlessVpcConnector (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServerlessVpcConnector, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.vpcAccess.networkInterfaces

`[]GcpCloudRunWorkerPoolNetworkInterface`

Direct VPC egress: instances get IPs in the subnetwork and reach VPC
resources with no connector infrastructure. The subnetwork needs free
address space for the instance fleet.

### spec.vpcAccess.networkInterfaces[].network

`string | valueFrom`

The VPC network. Accepts a literal network name or a reference to a
GcpVpcNetwork resource. May be omitted when subnetwork is set (the
network is inferred).

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.vpcAccess.networkInterfaces[].subnetwork

`string | valueFrom`

The subnetwork instances draw IPs from. Accepts a literal subnetwork
name or a reference to a GcpSubnetwork resource. Must be in the pool's
region.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.vpcAccess.networkInterfaces[].tags

`[]string`

Network tags applied to the instances -- how VPC firewall rules select
the pool's egress traffic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.vpcAccess.egress

`string`

Which egress traffic uses the VPC path: everything (ALL_TRAFFIC) or
only RFC1918/private destinations (PRIVATE_RANGES_ONLY -- the
default; public egress keeps Cloud Run's own path).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALL_TRAFFIC","PRIVATE_RANGES_ONLY"]}}

### spec.nodeSelector

`GcpCloudRunWorkerPoolNodeSelector`

Hardware requirement for GPU workers. Setting an accelerator (e.g.
"nvidia-l4") gives every instance one GPU; the containers' resource
limits must then meet Cloud Run's GPU minimums (4 CPU / 16Gi).

### spec.nodeSelector.accelerator

`string` · required

GPU accelerator type each instance gets, e.g. "nvidia-l4". GPU pools
need at least "4" CPU / "16Gi" memory and regional GPU quota.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.gpuZonalRedundancyDisabled

`bool`

Opts a GPU worker pool out of zonal redundancy: instances may be
served from a single zone, lowering GPU capacity cost for zonal-failure
risk. Only meaningful with node_selector.

### spec.launchStage

`string`

Launch-stage gate. Set BETA (or ALPHA) only when the spec uses preview
Cloud Run features the default GA stage rejects; the value is a
declaration, not a feature switch. The provider enum admits more
values, but Cloud Run supports only ALPHA, BETA, and GA -- this list
encodes the API truth.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALPHA","BETA","GA"]}}

### spec.binaryAuthorization

`GcpCloudRunWorkerPoolBinaryAuthorization`

Binary Authorization: only images that pass the policy's attestation
checks may deploy. The project default policy or a named one.

- rule: use the project default policy (use_default) or name a specific policy, not both

### spec.binaryAuthorization.useDefault

`bool`

Evaluate deploys against the project's default Binary Authorization
policy.

### spec.binaryAuthorization.policy

`string`

Evaluate deploys against a specific platform policy
(projects/*/platforms/cloudRun/policies/*).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.binaryAuthorization.breakglassJustification

`string`

Justification recorded when a break-glass deploy bypasses the policy.

### spec.deletionProtection

`bool` · optional (explicit presence)

Prevents the worker pool from being destroyed while true. Defaults to
true (Google's posture): a destroy fails until this is set to false.
Both engines send the value explicitly so a manifest that never
mentions it behaves the same everywhere.

- default: `true`

### spec.deletionPolicy

`string`

What happens to the worker pool when this resource is destroyed
(after deletion_protection allows the destroy):
  "" / "DELETE" -- the worker pool is deleted (default)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the worker pool leaves management but keeps running

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `gpu.redundancy_requires_accelerator`: gpu_zonal_redundancy_disabled only applies to GPU worker pools -- set node_selector.accelerator
- `encryption.shutdown_duration_requires_shutdown_action`: encryption_key_shutdown_duration only applies when encryption_key_revocation_action is SHUTDOWN
- `encryption.revocation_action_requires_key`: encryption_key_revocation_action only applies to a worker pool with an encryption_key

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudRunWorkerPool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name (projects/{project}/locations/{region}/workerPools/{name}). |
| `status.outputs.worker_pool_name` | `string` | Bare worker pool name in GCP -- the segment the API is keyed by. |
| `status.outputs.uid` | `string` | Server-generated unique identifier, stable across updates. |
| `status.outputs.location` | `string` | Region the worker pool runs in. |
| `status.outputs.project_id` | `string` | The GCP project the worker pool lives in. |
| `status.outputs.latest_created_revision` | `string` | Name of the most recently created revision (may still be rolling out). |
| `status.outputs.latest_ready_revision` | `string` | Name of the newest revision that is serving instances -- what a rollout has actually reached. |
| `status.outputs.observed_generation` | `string` | The generation Cloud Run's controller has reconciled (a decimal string, as Google reports it); equal to the pool's generation once a rollout is complete. |
| `status.outputs.etag` | `string` | Opaque version token for optimistic concurrency on the API. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.volumes[].cloudSqlInstance.instances` | GcpCloudSql | `status.outputs.connection_name` |
| `spec.volumes[].gcs.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.encryptionKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.vpcAccess.connector` | GcpServerlessVpcConnector | `status.outputs.self_link` |
| `spec.vpcAccess.networkInterfaces[].network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.vpcAccess.networkInterfaces[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_name` |

## See Also

- [Overview](../README.md)
