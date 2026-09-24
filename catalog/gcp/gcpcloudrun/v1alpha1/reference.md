# GcpCloudRun

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudRunSpec defines a Cloud Run service (`google_cloud_run_v2_service`) —
a fully managed, request-serving container deployment that scales from zero
to thousands of instances.

A Cloud Run service is two things at once: a stable serving endpoint
(ingress, invoker policy, traffic splitting across revisions) and a revision
template (containers, volumes, scaling, networking) that describes the next
revision to stamp out. Every spec change that touches the template creates a
new immutable revision; the traffic block decides how requests are split
between revisions, which is what makes blue/green and canary rollouts a
first-class, declarative operation.

Sidecar containers are first-class: a service can run an application
container alongside collectors, proxies, or auth helpers, with explicit
startup ordering (depends_on) and shared volumes. Private egress uses
Direct VPC (network_interfaces) — no connector infrastructure needed.

Custom domains are deliberately not part of this resource. The
production-grade path is composition: a serverless network endpoint group
(GcpRegionNetworkEndpointGroup) bridges the service into a backend service,
URL map, target proxy, and forwarding rule, with DNS managed by
GcpDnsRecord.

## Example

```yaml
# Development manifest for GcpCloudRun — exercises a broad slice of the spec
# (sidecar with startup ordering, secret env, Cloud SQL volume, all three
# probe types, scaling with the service-wide max, traffic split, direct VPC
# egress, annotations at both object levels, sub_path mounts, and the
# teardown policy) for offline plan verification.
#
# Usage: planton tofu plan --manifest catalog/gcp/gcpcloudrun/e2e/manifest.yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudRun
metadata:
  name: hack-cloud-run
  id: cldrun-hack-001
  org: planton-dev
  env: dev
spec:
  # project_id omitted: the service lands in the provider's default project.
  region: us-central1
  serviceName: hack-api
  description: Development service exercising the full template surface
  labels:
    team: platform
  annotations:
    external-tool/owner: platform-team
  revisionLabels:
    cost-center: "1234"
  revisionAnnotations:
    external-tool/trace: enabled
  containers:
    - name: app
      image: us-docker.pkg.dev/cloudrun/container/hello:latest
      env:
        - name: LOG_LEVEL
          value: debug
        - name: DB_PASSWORD
          valueFromSecret:
            secret: hack-db-password
            version: latest
        # A Planton secret: the component keeps it in a Secret Manager
        # secret of its own, readable only by the runtime identity.
        - name: API_TOKEN
          secretValue: $secret/hack-api-token
      ports:
        containerPort: 8080
      resources:
        cpu: "1"
        memory: 512Mi
        startupCpuBoost: true
      volumeMounts:
        - name: cloudsql
          mountPath: /cloudsql
        - name: assets
          mountPath: /media
          subPath: static
      startupProbe:
        periodSeconds: 5
        httpGet:
          path: /
      livenessProbe:
        httpGet:
          path: /
      readinessProbe:
        periodSeconds: 10
        httpGet:
          path: /
      dependsOn:
        - collector
    - name: collector
      image: us-docker.pkg.dev/cloudrun/container/hello:latest
  volumes:
    - name: cloudsql
      cloudSqlInstance:
        instances:
          - value: my-project:us-central1:hack-db
    - name: assets
      gcs:
        bucket:
          value: hack-assets-bucket
        readOnly: true
        mountOptions:
          - implicit-dirs
          - only-dir=media
  scaling:
    minInstanceCount: 0
    maxInstanceCount: 10
  serviceScaling:
    maxInstanceCount: 20
  maxInstanceRequestConcurrency: 80
  timeoutSeconds: 300
  executionEnvironment: EXECUTION_ENVIRONMENT_GEN2
  sessionAffinity: true
  vpcAccess:
    networkInterfaces:
      - network:
          value: hack-vpc
        subnetwork:
          value: hack-subnet
        tags:
          - cloud-run-egress
    egress: PRIVATE_RANGES_ONLY
  ingress: INGRESS_TRAFFIC_ALL
  allowUnauthenticated: true
  traffic:
    - type: TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST
      percent: 100
  deletionProtection: false
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.serviceName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.containers` | `[]GcpCloudRunContainer` | yes |  |  |
| `spec.containers[].name` | `string` |  |  |  |
| `spec.containers[].image` | `string` | yes |  |  |
| `spec.containers[].command` | `[]string` |  |  |  |
| `spec.containers[].args` | `[]string` |  |  |  |
| `spec.containers[].env` | `[]GcpCloudRunEnvVar` |  |  |  |
| `spec.containers[].env[].name` | `string` | yes |  |  |
| `spec.containers[].env[].value` | `string` (no secrets: use `secretValue`) |  |  |  |
| `spec.containers[].env[].valueFromSecret` | `GcpCloudRunSecretEnvSource` |  |  |  |
| `spec.containers[].env[].valueFromSecret.secret` | `string` | yes |  |  |
| `spec.containers[].env[].valueFromSecret.version` | `string` |  | `latest` |  |
| `spec.containers[].env[].secretValue` | `string` (sensitive) |  |  |  |
| `spec.containers[].ports` | `GcpCloudRunContainerPort` |  |  |  |
| `spec.containers[].ports.containerPort` | `int32` |  | `8080` |  |
| `spec.containers[].ports.name` | `string` |  |  |  |
| `spec.containers[].resources` | `GcpCloudRunContainerResources` |  |  |  |
| `spec.containers[].resources.cpu` | `string` |  |  |  |
| `spec.containers[].resources.memory` | `string` |  |  |  |
| `spec.containers[].resources.cpuIdle` | `bool` |  | `true` |  |
| `spec.containers[].resources.startupCpuBoost` | `bool` |  |  |  |
| `spec.containers[].volumeMounts` | `[]GcpCloudRunVolumeMount` |  |  |  |
| `spec.containers[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.containers[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.containers[].workingDir` | `string` |  |  |  |
| `spec.containers[].startupProbe` | `GcpCloudRunStartupProbe` |  |  |  |
| `spec.containers[].startupProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].startupProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].startupProbe.httpGet` | `GcpCloudRunHttpGetAction` |  |  |  |
| `spec.containers[].startupProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].startupProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders` | `[]GcpCloudRunHttpHeader` |  |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.containers[].startupProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.containers[].startupProbe.tcpSocket` | `GcpCloudRunTcpSocketAction` |  |  |  |
| `spec.containers[].startupProbe.tcpSocket.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.grpc` | `GcpCloudRunGrpcAction` |  |  |  |
| `spec.containers[].startupProbe.grpc.port` | `int32` |  |  |  |
| `spec.containers[].startupProbe.grpc.service` | `string` |  |  |  |
| `spec.containers[].livenessProbe` | `GcpCloudRunLivenessProbe` |  |  |  |
| `spec.containers[].livenessProbe.initialDelaySeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.httpGet` | `GcpCloudRunHttpGetAction` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders` | `[]GcpCloudRunHttpHeader` |  |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders[].name` | `string` | yes |  |  |
| `spec.containers[].livenessProbe.httpGet.httpHeaders[].value` | `string` |  |  |  |
| `spec.containers[].livenessProbe.grpc` | `GcpCloudRunGrpcAction` |  |  |  |
| `spec.containers[].livenessProbe.grpc.port` | `int32` |  |  |  |
| `spec.containers[].livenessProbe.grpc.service` | `string` |  |  |  |
| `spec.containers[].dependsOn` | `[]string` |  |  |  |
| `spec.containers[].baseImageUri` | `string` |  |  |  |
| `spec.containers[].readinessProbe` | `GcpCloudRunReadinessProbe` |  |  |  |
| `spec.containers[].readinessProbe.timeoutSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.periodSeconds` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.failureThreshold` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.httpGet` | `GcpCloudRunReadinessHttpGetAction` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.path` | `string` |  |  |  |
| `spec.containers[].readinessProbe.httpGet.port` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.grpc` | `GcpCloudRunGrpcAction` |  |  |  |
| `spec.containers[].readinessProbe.grpc.port` | `int32` |  |  |  |
| `spec.containers[].readinessProbe.grpc.service` | `string` |  |  |  |
| `spec.containers[].sandboxLauncher` | `bool` |  |  |  |
| `spec.volumes` | `[]GcpCloudRunVolume` |  |  |  |
| `spec.volumes[].name` | `string` | yes |  |  |
| `spec.volumes[].cloudSqlInstance` | `GcpCloudRunVolumeCloudSql` |  |  |  |
| `spec.volumes[].cloudSqlInstance.instances` | `[]string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.connection_name`) |
| `spec.volumes[].secret` | `GcpCloudRunVolumeSecret` |  |  |  |
| `spec.volumes[].secret.secret` | `string` | yes |  |  |
| `spec.volumes[].secret.defaultMode` | `int32` |  |  |  |
| `spec.volumes[].secret.items` | `[]GcpCloudRunVolumeSecretItem` |  |  |  |
| `spec.volumes[].secret.items[].path` | `string` | yes |  |  |
| `spec.volumes[].secret.items[].version` | `string` |  | `latest` |  |
| `spec.volumes[].secret.items[].mode` | `int32` |  |  |  |
| `spec.volumes[].emptyDir` | `GcpCloudRunVolumeEmptyDir` |  |  |  |
| `spec.volumes[].emptyDir.medium` | `string` |  |  |  |
| `spec.volumes[].emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.volumes[].gcs` | `GcpCloudRunVolumeGcs` |  |  |  |
| `spec.volumes[].gcs.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.volumes[].gcs.readOnly` | `bool` |  |  |  |
| `spec.volumes[].gcs.mountOptions` | `[]string` |  |  |  |
| `spec.volumes[].nfs` | `GcpCloudRunVolumeNfs` |  |  |  |
| `spec.volumes[].nfs.server` | `string` | yes |  |  |
| `spec.volumes[].nfs.path` | `string` | yes |  |  |
| `spec.volumes[].nfs.readOnly` | `bool` |  |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.scaling` | `GcpCloudRunRevisionScaling` |  |  |  |
| `spec.scaling.minInstanceCount` | `int32` |  |  |  |
| `spec.scaling.maxInstanceCount` | `int32` |  |  |  |
| `spec.serviceScaling` | `GcpCloudRunServiceScaling` |  |  |  |
| `spec.serviceScaling.scalingMode` | `string` |  |  |  |
| `spec.serviceScaling.manualInstanceCount` | `int32` |  |  |  |
| `spec.serviceScaling.minInstanceCount` | `int32` |  |  |  |
| `spec.serviceScaling.maxInstanceCount` | `int32` |  |  |  |
| `spec.maxInstanceRequestConcurrency` | `int32` |  |  |  |
| `spec.timeoutSeconds` | `int32` |  |  |  |
| `spec.executionEnvironment` | `enum` |  | `EXECUTION_ENVIRONMENT_GEN2` |  |
| `spec.sessionAffinity` | `bool` |  |  |  |
| `spec.encryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.revision` | `string` |  |  |  |
| `spec.vpcAccess` | `GcpCloudRunVpcAccess` |  |  |  |
| `spec.vpcAccess.connector` | `string \| valueFrom` |  |  | GcpServerlessVpcConnector (`status.outputs.self_link`) |
| `spec.vpcAccess.networkInterfaces` | `[]GcpCloudRunNetworkInterface` |  |  |  |
| `spec.vpcAccess.networkInterfaces[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.vpcAccess.networkInterfaces[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.vpcAccess.networkInterfaces[].tags` | `[]string` |  |  |  |
| `spec.vpcAccess.egress` | `string` |  |  |  |
| `spec.nodeSelector` | `GcpCloudRunNodeSelector` |  |  |  |
| `spec.nodeSelector.accelerator` | `string` | yes |  |  |
| `spec.gpuZonalRedundancyDisabled` | `bool` |  |  |  |
| `spec.ingress` | `enum` |  | `INGRESS_TRAFFIC_ALL` |  |
| `spec.allowUnauthenticated` | `bool` |  | `true` |  |
| `spec.invokerIamDisabled` | `bool` |  |  |  |
| `spec.customAudiences` | `[]string` |  |  |  |
| `spec.traffic` | `[]GcpCloudRunTrafficTarget` |  |  |  |
| `spec.traffic[].type` | `string` | yes |  |  |
| `spec.traffic[].revision` | `string` |  |  |  |
| `spec.traffic[].percent` | `int32` |  |  |  |
| `spec.traffic[].tag` | `string` |  |  |  |
| `spec.launchStage` | `string` |  |  |  |
| `spec.binaryAuthorization` | `GcpCloudRunBinaryAuthorization` |  |  |  |
| `spec.binaryAuthorization.useDefault` | `bool` |  |  |  |
| `spec.binaryAuthorization.policy` | `string` |  |  |  |
| `spec.binaryAuthorization.breakglassJustification` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.revisionLabels` | `map<string, string>` |  |  |  |
| `spec.revisionAnnotations` | `map<string, string>` |  |  |  |
| `spec.buildConfig` | `GcpCloudRunBuildConfig` |  |  |  |
| `spec.buildConfig.sourceLocation` | `string` |  |  |  |
| `spec.buildConfig.functionTarget` | `string` |  |  |  |
| `spec.buildConfig.imageUri` | `string` |  |  |  |
| `spec.buildConfig.baseImage` | `string` |  |  |  |
| `spec.buildConfig.enableAutomaticUpdates` | `bool` |  |  |  |
| `spec.buildConfig.environmentVariables` | `map<string, string>` |  |  |  |
| `spec.buildConfig.workerPool` | `string` |  |  |  |
| `spec.buildConfig.serviceAccount` | `string` |  |  |  |
| `spec.iapEnabled` | `bool` |  |  |  |
| `spec.defaultUriDisabled` | `bool` |  |  |  |
| `spec.healthCheckDisabled` | `bool` |  |  |  |
| `spec.multiRegionSettings` | `GcpCloudRunMultiRegionSettings` |  |  |  |
| `spec.multiRegionSettings.regions` | `[]string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.sandboxTemplates` | `[]GcpCloudRunSandboxTemplate` |  |  |  |
| `spec.sandboxTemplates[].name` | `string` | yes |  |  |
| `spec.sandboxTemplates[].image` | `string` | yes |  |  |
| `spec.sandboxTemplates[].command` | `[]string` |  |  |  |
| `spec.sandboxTemplates[].args` | `[]string` |  |  |  |
| `spec.sandboxTemplates[].env` | `[]GcpCloudRunSandboxEnvVar` |  |  |  |
| `spec.sandboxTemplates[].env[].name` | `string` | yes |  |  |
| `spec.sandboxTemplates[].env[].value` | `string` |  |  |  |
| `spec.sandboxTemplates[].volumeMounts` | `[]GcpCloudRunVolumeMount` |  |  |  |
| `spec.sandboxTemplates[].volumeMounts[].name` | `string` | yes |  |  |
| `spec.sandboxTemplates[].volumeMounts[].mountPath` | `string` | yes |  |  |
| `spec.sandboxTemplates[].volumeMounts[].subPath` | `string` |  |  |  |
| `spec.sandboxTemplates[].workingDir` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the service is created in. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region the service is deployed in, e.g. "us-central1". Immutable.
Cloud Run services are regional; the one exception is multi-region
services (multi_region_settings), which are created through the
special "global" region with the actual serving regions listed in
multi_region_settings.regions.

- rule: {"required":true,"string":{"pattern":"^([a-z]+-[a-z]+[0-9]+|global)$"}}

### spec.serviceName

`string`

Name of the Cloud Run service in GCP. Immutable. If not specified,
defaults to metadata.name. Must be 1-63 characters: lowercase letters,
digits, and hyphens; starting with a letter.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.description

`string`

Human-readable description of the service, shown in the console.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.labels

`map<string, string>`

Labels applied to the service object. User labels are shared with
Google's billing system, so they can filter or break down billing
charges by team, component, or environment. Keys and values in the
`run.googleapis.com`, `cloud.googleapis.com`, `serving.knative.dev`,
and `autoscaling.knative.dev` namespaces are rejected by the API.

### spec.containers

`[]GcpCloudRunContainer` · required

The containers that make up one instance of the service. The first
container conventionally serves requests; additional containers are
sidecars (log collectors, auth proxies, service meshes) that share the
instance's network namespace (localhost) and volumes. Exactly one
container may expose a port. Use depends_on for startup ordering.
Deployment pipelines inject the built image into containers whose image
is left BLANK (the image-slot contract); authored images are untouched.

- rule: {"repeated":{"minItems":"1"}}

### spec.containers[].name

`string`

Name of the container. Required when the service runs more than one
container (depends_on refers to these names). If omitted for a
single-container service, Cloud Run assigns one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.containers[].image

`string` · required

Container image URL, e.g. "us-docker.pkg.dev/project/repo/app:1.0.0".
Pin a digest or immutable tag for repeatable deploys — Cloud Run
resolves the image to a digest at revision creation, so a moving tag
only takes effect on the NEXT deploy anyway.

Cloud Run pulls PRIVATE images only from Artifact Registry (or the
legacy Container Registry) in a project its service agent can read;
public Docker Hub and GHCR images deploy directly. There is no field
here for a registry login and Google accepts none. For a private image
in another registry, push it to Artifact Registry, or declare a
GcpArtifactRegistryRepo in REMOTE_REPOSITORY mode that proxies that
registry and point this URL at it.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].command

`[]string`

Entrypoint array — overrides the image's ENTRYPOINT. Not executed in a
shell; variable references are not expanded.

### spec.containers[].args

`[]string`

Arguments to the entrypoint — overrides the image's CMD.

### spec.containers[].env

`[]GcpCloudRunEnvVar`

Environment variables. Each entry carries either a literal value or a
Secret Manager reference resolved at instance start.

- rule: an environment variable takes exactly one of a literal value, a Secret Manager reference, or a secret value

### spec.containers[].env[].name

`string` · required

Variable name, e.g. "DATABASE_URL". Must not start with a digit.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.containers[].env[].value

`string` · no secrets

Literal value, written into the revision where anyone who can view the
service reads it. Fine for configuration; never a credential -- a
credential goes in secret_value (or value_from_secret).

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretValue`, which keeps it in a secret store the workload reads by reference
### spec.containers[].env[].valueFromSecret

`GcpCloudRunSecretEnvSource`

A Secret Manager secret you already own, resolved into the variable at
instance start. Rotation is Secret Manager's: with version "latest",
new instances pick up a new version without a deploy.

### spec.containers[].env[].valueFromSecret.secret

`string` · required

The secret: a short name for a secret in the same project ("my-secret")
or a full resource name (projects/*/secrets/*) for cross-project reads.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].env[].valueFromSecret.version

`string`

Secret version to resolve: a version number or "latest". If unset,
GCP requires an explicit version for env vars — "latest" is the common
choice, at the cost of new instances silently picking up rotations.

- default: `latest`

### spec.containers[].env[].secretValue

`string` · sensitive

A secret value this component keeps in Secret Manager for you. It
creates one secret for this variable, replicated only in the service's
region(s), stores the value as a version, grants the service's runtime
identity secretAccessor on that secret alone, and points the variable at
that exact version -- the revision carries a reference, never the value.
A changed value adds a version and stamps a new revision, so rotation is
a deploy; destroying the service removes the secret.

### spec.containers[].ports

`GcpCloudRunContainerPort`

The port this container listens on for requests. At most ONE container
in the service may declare a port (that container receives traffic).
If no container declares one, Cloud Run injects PORT=8080 into the
first container.

### spec.containers[].ports.containerPort

`int32` · optional (explicit presence)

Port number the container listens on. Cloud Run injects it as $PORT.

- default: `8080`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].ports.name

`string`

Protocol selector: "http1" (default) or "h2c" for end-to-end HTTP/2
(required for serving gRPC streams).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","http1","h2c"]}}

### spec.containers[].resources

`GcpCloudRunContainerResources`

CPU and memory for this container, plus the CPU-allocation levers.
If omitted, Cloud Run defaults apply (1 CPU, 512Mi).

### spec.containers[].resources.cpu

`string`

CPU limit, e.g. "1", "2", "4", "8" or a fraction like "0.5"/"500m"
(fractions below 1 require max_instance_request_concurrency <= 1 for
some features). GPU services need at least "4".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-9]+m|[0-9]+(\\.[0-9]+)?)$"}}

### spec.containers[].resources.memory

`string`

Memory limit with unit suffix, e.g. "512Mi", "2Gi". Minimums scale
with CPU (1 CPU >= 128Mi, 4 CPU >= 2Gi, GPU >= 16Gi).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.containers[].resources.cpuIdle

`bool` · optional (explicit presence)

Keep CPU allocated only while requests are in flight (true — the
default, request-based billing) or allocate it for the instance's
whole lifetime (false — instance-based billing, required for real
background work between requests).

- default: `true`

### spec.containers[].resources.startupCpuBoost

`bool`

Temporarily boost CPU during instance startup, cutting cold-start
latency for JIT-heavy runtimes (JVM, .NET) at no idle cost.

### spec.containers[].volumeMounts

`[]GcpCloudRunVolumeMount`

Volumes (declared at the service level) mounted into this container's
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

Path WITHIN the volume to mount instead of its root — e.g. mount only
one secret item or one bucket directory. Relative path; empty mounts
the volume root.

### spec.containers[].workingDir

`string`

Working directory for the entrypoint. If omitted, the image's WORKDIR
is used.

### spec.containers[].startupProbe

`GcpCloudRunStartupProbe`

Probe that gates instance start: the container receives no traffic
(and depends_on waiters stay blocked) until this succeeds. Supports
HTTP, TCP, and gRPC checks. If omitted, Cloud Run performs a default
TCP check on the container port.

- rule: probe timeout_seconds cannot exceed period_seconds
- rule: the startup window (failure_threshold × period_seconds, defaults 3 × 10) cannot exceed 240 seconds

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

Seconds between probe attempts (1-240; GCP default 10). Bounded
further by the 240-second startup window.

- rule: {"int32":{"lte":240,"gte":1}}

### spec.containers[].startupProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures after which startup is considered failed and the
instance is killed (GCP default 3).

- rule: {"int32":{"gte":1}}

### spec.containers[].startupProbe.httpGet

`GcpCloudRunHttpGetAction`

HTTP GET against a path on the container; 2xx is success.

### spec.containers[].startupProbe.httpGet.path

`string`

Path to probe, e.g. "/healthz". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.containers[].startupProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. If unset, the container's serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.httpGet.httpHeaders

`[]GcpCloudRunHttpHeader`

Custom headers sent with the probe request (e.g. an auth header for a
protected health endpoint).

### spec.containers[].startupProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].startupProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.containers[].startupProbe.tcpSocket

`GcpCloudRunTcpSocketAction`

TCP connect to a port; a successful connection is success. Startup
is the only probe type Cloud Run accepts TCP on.

### spec.containers[].startupProbe.tcpSocket.port

`int32` · optional (explicit presence)

Port to connect to. If unset, the container's serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.grpc

`GcpCloudRunGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check).

### spec.containers[].startupProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on. If unset, the container's
serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].startupProbe.grpc.service

`string`

Service name passed to the health check, letting one server report
per-service health. If empty, overall server health is checked.

### spec.containers[].livenessProbe

`GcpCloudRunLivenessProbe`

Probe that monitors a running instance: on failure_threshold
consecutive failures the container is restarted. Supports HTTP and
gRPC checks only — the message has no TCP arm because Cloud Run
rejects TCP liveness probes. If omitted, instances are never
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

`GcpCloudRunHttpGetAction`

HTTP GET against a path on the container; 2xx is success.

### spec.containers[].livenessProbe.httpGet.path

`string`

Path to probe, e.g. "/healthz". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.containers[].livenessProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. If unset, the container's serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].livenessProbe.httpGet.httpHeaders

`[]GcpCloudRunHttpHeader`

Custom headers sent with the probe request (e.g. an auth header for a
protected health endpoint).

### spec.containers[].livenessProbe.httpGet.httpHeaders[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.containers[].livenessProbe.httpGet.httpHeaders[].value

`string`

Header value.

### spec.containers[].livenessProbe.grpc

`GcpCloudRunGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check).

### spec.containers[].livenessProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on. If unset, the container's
serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].livenessProbe.grpc.service

`string`

Service name passed to the health check, letting one server report
per-service health. If empty, overall server health is checked.

### spec.containers[].dependsOn

`[]string`

Names of containers this one waits for: this container starts only
after the listed containers pass their startup probes. The mechanism
that makes proxy/collector sidecars start before (and stop after) the
serving container.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.containers[].baseImageUri

`string`

Base image for automatic base-image updates on source-deployed
services (see spec.build_config): Google patches the OS/runtime
layers of the running image without a redeploy. A serverless
runtimes image URI, e.g.
"us-central1-docker.pkg.dev/serverless-runtimes/google-24-full/runtimes/nodejs24".

### spec.containers[].readinessProbe

`GcpCloudRunReadinessProbe`

Probe that gates whether this container receives traffic: instances
whose readiness check fails are pulled from serving without being
restarted (unlike liveness, which restarts). Supports HTTP and gRPC
checks only, and unlike the other probes has no initial delay — it
starts with the container and runs for the instance's lifetime.

- rule: probe timeout_seconds cannot exceed period_seconds

### spec.containers[].readinessProbe.timeoutSeconds

`int32` · optional (explicit presence)

Seconds after which a single probe attempt times out (GCP default 1).
Must not exceed period_seconds.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.containers[].readinessProbe.periodSeconds

`int32` · optional (explicit presence)

Seconds between probe attempts (GCP default 10).

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.containers[].readinessProbe.failureThreshold

`int32` · optional (explicit presence)

Consecutive failures after which the instance is pulled from serving
(GCP default 3). The instance is NOT restarted — that is the liveness
probe's job.

- rule: {"int32":{"gte":1}}

### spec.containers[].readinessProbe.httpGet

`GcpCloudRunReadinessHttpGetAction`

HTTP GET against a path on the container; 2xx is success.

### spec.containers[].readinessProbe.httpGet.path

`string`

Path to probe, e.g. "/ready". Defaults to "/".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^/.*$"}}

### spec.containers[].readinessProbe.httpGet.port

`int32` · optional (explicit presence)

Port to probe. If unset, the container's serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].readinessProbe.grpc

`GcpCloudRunGrpcAction`

Standard gRPC health-check protocol (grpc.health.v1.Health/Check).

### spec.containers[].readinessProbe.grpc.port

`int32` · optional (explicit presence)

Port the gRPC health service listens on. If unset, the container's
serving port is used.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.containers[].readinessProbe.grpc.service

`string`

Service name passed to the health check, letting one server report
per-service health. If empty, overall server health is checked.

### spec.containers[].sandboxLauncher

`bool`

Marks this container as the sandbox supervisor: the one process in
the instance allowed to launch the isolated sandboxes declared in
spec.sandbox_templates (through the Cloud Run sandbox CLI/API). The
pattern for agent workloads that run untrusted, model-generated code:
the supervisor orchestrates, each sandbox executes in isolation.

### spec.volumes

`[]GcpCloudRunVolume`

Named volumes instances can mount: Cloud SQL sockets, Secret Manager
material, in-memory or disk scratch space, GCS buckets (FUSE), and NFS
shares. A volume is inert until a container mounts it by name via
volume_mounts.

### spec.volumes[].name

`string` · required

Volume name referenced by volume_mounts entries.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.volumes[].cloudSqlInstance

`GcpCloudRunVolumeCloudSql`

Cloud SQL Unix sockets, one per instance, under the mount path
(mount at "/cloudsql"; connect via
"/cloudsql/<project:region:instance>"). GCP manages the proxying —
no sidecar, no VPC needed.

### spec.volumes[].cloudSqlInstance.instances

`[]string | valueFrom` · required

Cloud SQL instance connection names (project:region:instance).
Accepts literal values or references to GcpCloudSql resources.

- references: GcpCloudSql (`status.outputs.connection_name`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.connection_name}} -- a bare string does not parse

### spec.volumes[].secret

`GcpCloudRunVolumeSecret`

Secret Manager secret versions exposed as files.

### spec.volumes[].secret.secret

`string` · required

The secret: a short name for a secret in the same project or a full
resource name (projects/*/secrets/*).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.volumes[].secret.defaultMode

`int32` · optional (explicit presence)

Default Unix permission mode for projected files, in decimal (e.g. 292
= 0444). If unset, GCP defaults to 0444 (read-only for all).

- rule: {"int32":{"lte":511,"gte":0}}

### spec.volumes[].secret.items

`[]GcpCloudRunVolumeSecretItem`

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

Unix permission mode for this file, in decimal. Overrides default_mode.

- rule: {"int32":{"lte":511,"gte":0}}

### spec.volumes[].emptyDir

`GcpCloudRunVolumeEmptyDir`

Ephemeral scratch space, in-memory (counts against the instance's
memory limit) or disk-backed.

### spec.volumes[].emptyDir.medium

`string`

Backing medium: MEMORY (default — tmpfs; usage counts against the
containers' memory limits) or DISK.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","MEMORY","DISK"]}}

### spec.volumes[].emptyDir.sizeLimit

`string`

Capacity limit with unit suffix, e.g. "512Mi", "2Gi". For MEMORY
volumes, leave unset to let GCP cap it sensibly relative to instance
memory.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(Ki|Mi|Gi|Ti|K|M|G|T)$"}}

### spec.volumes[].gcs

`GcpCloudRunVolumeGcs`

A GCS bucket mounted via Cloud Storage FUSE. Requires the GEN2
execution environment; object storage semantics apply (no POSIX
locking; renames are copies).

### spec.volumes[].gcs.bucket

`string | valueFrom` · required

The bucket to mount. Accepts a literal bucket name or a reference to a
GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.volumes[].gcs.readOnly

`bool`

Mount read-only. Recommended unless the service genuinely writes —
concurrent writers through FUSE are easy to get wrong.

### spec.volumes[].gcs.mountOptions

`[]string`

Flags passed to the gcsfuse command mounting this volume, without
leading dashes (e.g. "implicit-dirs", "only-dir=media",
"file-cache-max-size-mb=512"). The tuning surface for FUSE caching
and directory semantics.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.volumes[].nfs

`GcpCloudRunVolumeNfs`

An NFS share (e.g. Filestore) mounted into the instance. Requires
the GEN2 execution environment and VPC access to reach the server.

### spec.volumes[].nfs.server

`string` · required

Hostname or IP of the NFS server, e.g. a Filestore instance's IP.

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

Email of the IAM service account the revisions run as — the identity
whose permissions the code exercises when calling other GCP APIs.
Accepts a literal email or a reference to a GcpServiceAccount resource.
If omitted, the project's Compute Engine default service account is
used — fine for experiments, too broad for production; give real
services a dedicated least-privilege identity.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.scaling

`GcpCloudRunRevisionScaling`

Per-revision instance bounds: how far a single revision scales in and
out. If omitted, the service scales zero-to-100. min_instance_count > 0
keeps instances warm to eliminate cold starts — at idle cost.

- rule: min_instance_count cannot exceed max_instance_count

### spec.scaling.minInstanceCount

`int32` · optional (explicit presence)

Instances kept warm even with zero traffic. 0 (default) scales to
zero — cheapest, with cold starts; 1+ eliminates cold starts at idle
cost.

- rule: {"int32":{"gte":0}}

### spec.scaling.maxInstanceCount

`int32` · optional (explicit presence)

Upper bound on instances for this revision — the cost/overload
circuit breaker. If unset, GCP's default cap (100) applies.

- rule: {"int32":{"gte":1}}

### spec.serviceScaling

`GcpCloudRunServiceScaling`

Service-level scaling posture across ALL revisions. The MANUAL mode
pins total instance count regardless of traffic — an emergency brake
or a load-test lever. Distinct from the per-revision bounds in
`scaling`.

- rule: manual_instance_count only applies when scaling_mode is MANUAL

### spec.serviceScaling.scalingMode

`string`

AUTOMATIC (default): Cloud Run scales with traffic. MANUAL: the
service runs exactly manual_instance_count instances regardless of
traffic — a load-test lever or emergency brake.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","AUTOMATIC","MANUAL"]}}

### spec.serviceScaling.manualInstanceCount

`int32` · optional (explicit presence)

Exact total instance count in MANUAL mode.

- rule: {"int32":{"gte":0}}

### spec.serviceScaling.minInstanceCount

`int32` · optional (explicit presence)

Service-level minimum instances, distributed across serving revisions
(unlike scaling.min_instance_count, which is per revision). Useful
during gradual rollouts so warm capacity follows the traffic split.

- rule: {"int32":{"gte":0}}

### spec.serviceScaling.maxInstanceCount

`int32` · optional (explicit presence)

Service-level maximum across ALL revisions combined — the total-cost
circuit breaker during rollouts, when two revisions serve at once and
per-revision caps (scaling.max_instance_count) would otherwise double
the worst case.

- rule: {"int32":{"gte":1}}

### spec.maxInstanceRequestConcurrency

`int32` · optional (explicit presence)

Maximum concurrent requests one instance serves before Cloud Run scales
out (1-1000). If unset, GCP defaults to 80 for instances with >= 1 CPU
and 1 below that. Lower values isolate requests (memory-heavy or
CPU-bound work); higher values improve utilization for I/O-bound
services.

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.timeoutSeconds

`int32` · optional (explicit presence)

Maximum time a request may run before Cloud Run cancels it, in seconds
(1-3600). If unset, GCP defaults to 300 (5 minutes). Also bounds
startup: an instance that cannot serve within this window is killed.

- rule: {"int32":{"lte":3600,"gte":1}}

### spec.executionEnvironment

`enum`

Sandbox generation the revisions execute in. GEN2 (recommended) offers
full Linux compatibility (required for GCS/NFS volumes and network file
systems) at slightly slower cold starts; GEN1 starts faster with a
gVisor-restricted syscall surface. If unset, GCP picks per workload.

- default: `EXECUTION_ENVIRONMENT_GEN2`

Allowed values (use exactly as shown):

- `EXECUTION_ENVIRONMENT_UNSPECIFIED` -- Unspecified — GCP selects per workload.
- `EXECUTION_ENVIRONMENT_GEN1` -- First generation: fastest cold starts, gVisor-restricted syscall surface, no network file systems.
- `EXECUTION_ENVIRONMENT_GEN2` -- Second generation: full Linux compatibility; required for GCS and NFS volumes.

### spec.sessionAffinity

`bool`

Routes requests from the same client to the same instance on a
best-effort basis (session affinity cookie). Useful for local caches;
never a correctness guarantee — instances still scale in.

### spec.encryptionKey

`string | valueFrom`

Customer-managed encryption key (CMEK) used to encrypt the deployed
container images. Accepts a full crypto key ID
(projects/*/locations/*/keyRings/*/cryptoKeys/*) or a reference to a
GcpKmsKey resource. The key must be in the same region as the service,
and the Cloud Run service agent needs encrypter/decrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.revision

`string`

Explicit name for the next revision. Must be prefixed with the service
name (e.g. "my-api-v42"). If omitted (recommended), Cloud Run
auto-generates revision names. Pin revision names only when the traffic
block routes by revision name — deterministic names make declarative
blue/green possible, but each template change then REQUIRES a new
unique value here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.vpcAccess

`GcpCloudRunVpcAccess`

Private networking for OUTBOUND traffic: route egress into a VPC either
through Direct VPC egress (network_interfaces — no extra infrastructure,
recommended) or a Serverless VPC Access connector. Inbound restriction
is `ingress`, not this.

- rule: use direct VPC egress (network_interfaces) or a Serverless VPC Access connector, not both

### spec.vpcAccess.connector

`string | valueFrom`

Serverless VPC Access connector to route egress through — the legacy
mechanism, still required for some org constraints. Full resource name
(projects/*/locations/*/connectors/*) or a reference to a
GcpServerlessVpcConnector resource.

- references: GcpServerlessVpcConnector (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServerlessVpcConnector, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.vpcAccess.networkInterfaces

`[]GcpCloudRunNetworkInterface`

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
name or a reference to a GcpSubnetwork resource. Must be in the
service's region.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.vpcAccess.networkInterfaces[].tags

`[]string`

Network tags applied to the instances — how VPC firewall rules select
Cloud Run egress traffic.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.vpcAccess.egress

`string`

Which egress traffic uses the VPC path: everything (ALL_TRAFFIC) or
only RFC1918/private destinations (PRIVATE_RANGES_ONLY — the default;
public egress keeps Cloud Run's own path).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALL_TRAFFIC","PRIVATE_RANGES_ONLY"]}}

### spec.nodeSelector

`GcpCloudRunNodeSelector`

Hardware requirements for GPU inference workloads. Setting an
accelerator (e.g. "nvidia-l4") gives every instance one GPU; the
containers' resource limits must then meet Cloud Run's GPU minimums
(4 CPU / 16Gi recommended).

### spec.nodeSelector.accelerator

`string` · required

GPU accelerator type each instance gets, e.g. "nvidia-l4". GPU
services need GEN2-class resources (at least "4" CPU / "16Gi" memory)
and regional GPU quota.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.gpuZonalRedundancyDisabled

`bool`

Opts a GPU service out of zonal redundancy: instances may be served
from a single zone, which lowers the price of GPU capacity in exchange
for zonal-failure risk. Only meaningful when node_selector is set.

### spec.ingress

`enum`

Which network paths may reach the service: everything, only internal
traffic, or internal plus Cloud Load Balancing. Services behind the
composed HTTPS load balancer should use INTERNAL_LOAD_BALANCER so the
default run.app URL stops accepting public traffic.

- default: `INGRESS_TRAFFIC_ALL`

Allowed values (use exactly as shown):

- `INGRESS_TRAFFIC_UNSPECIFIED` -- Unspecified — GCP defaults to INGRESS_TRAFFIC_ALL.
- `INGRESS_TRAFFIC_ALL` -- Accept traffic from all sources, including the public internet.
- `INGRESS_TRAFFIC_INTERNAL_ONLY` -- Accept only internal traffic (VPC, internal load balancers in the project/VPC-SC perimeter, Pub/Sub push, Eventarc).
- `INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER` -- Accept internal traffic plus traffic arriving through Cloud Load Balancing — the right setting behind the composed HTTPS load balancer.

### spec.allowUnauthenticated

`bool`

If true, grants roles/run.invoker to allUsers so unauthenticated
callers can reach the service — the standard shape for public HTTP
APIs and websites. If false, callers must present an IAM-authorized
identity token.

- default: `true`

### spec.invokerIamDisabled

`bool`

Disables the IAM run.routes.invoke permission check entirely. Unlike
allow_unauthenticated (which grants access THROUGH IAM), this switches
the check off — required by some org policies that forbid allUsers
grants while still serving public traffic. Set at most one of the two.

### spec.customAudiences

`[]string`

Additional audience values accepted in the OAuth/OIDC tokens of
authenticated callers (beyond the default service URL). Lets callers
mint one token for a stable custom audience instead of the run.app URL.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.traffic

`[]GcpCloudRunTrafficTarget`

How traffic is split across revisions. If empty, 100% of traffic goes
to the latest ready revision — the right default for most services.
Populate for gradual rollouts: percentages must sum to 100 (enforced
by the API at deploy time), and tagged entries get stable
<tag>---<host> preview URLs that receive no percentage-based traffic
unless assigned some.

- rule: REVISION targets must name a revision; LATEST targets must not

### spec.traffic[].type

`string` · required

Route to the latest ready revision (LATEST) or a specific named
revision (REVISION).

- rule: {"required":true,"string":{"in":["TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST","TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION"]}}

### spec.traffic[].revision

`string`

Revision name for REVISION targets (see spec.revision for
deterministic naming).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.traffic[].percent

`int32` · optional (explicit presence)

Percent of traffic for this target (0-100). All percents in the
traffic block must sum to 100 — GCP enforces this at deploy time.

- rule: {"int32":{"lte":100,"gte":0}}

### spec.traffic[].tag

`string`

Tag that gives this target a stable preview URL
(https://<tag>---<service-host>) receiving only requests addressed to
it — how a canary is smoke-tested before percent-based traffic moves.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.launchStage

`string`

Launch-stage gate for the service. Set BETA (or ALPHA) only when the
spec uses preview Cloud Run features that GCP rejects at the default
GA stage; the value is a declaration, not a feature switch. The
provider enum admits more values (UNIMPLEMENTED, PRELAUNCH,
EARLY_ACCESS, DEPRECATED), but Cloud Run itself supports only
ALPHA/BETA/GA — this list encodes the API truth, deliberately.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["","ALPHA","BETA","GA"]}}

### spec.binaryAuthorization

`GcpCloudRunBinaryAuthorization`

Binary Authorization: only container images that pass the policy's
attestation checks may deploy. Use the project default policy or name
a specific one.

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

Prevents the service from being destroyed while true. Defaults to true
(matching GCP's posture): a delete of this resource fails until this is
set to false. Deleting a service tears down its endpoint and every
revision; keep this on for anything real.

- default: `true`

### spec.annotations

`map<string, string>`

Annotations on the SERVICE object: unstructured metadata read by
external tools (never queryable, never behavioral for Cloud Run
itself). Keys in the `run.googleapis.com`, `cloud.googleapis.com`,
`serving.knative.dev`, and `autoscaling.knative.dev` namespaces are
rejected by the API. For metadata stamped on each REVISION instead,
use revision_annotations.

### spec.revisionLabels

`map<string, string>`

Labels stamped on every REVISION the template creates (visible on
revision objects and in billing breakdowns), as opposed to `labels`,
which lands on the service object itself. Same namespace
restrictions as service labels.

### spec.revisionAnnotations

`map<string, string>`

Annotations stamped on every REVISION the template creates, as
opposed to `annotations`, which lands on the service object. Same
namespace restrictions.

### spec.buildConfig

`GcpCloudRunBuildConfig`

Deploy-from-source: have Cloud Build produce the serving image from
source code (the Cloud Run functions build path) instead of deploying
a prebuilt image. The built artifact deploys as the service's
container image; containers[].base_image_uri pairs with
enable_automatic_updates for managed base-image patching.

### spec.buildConfig.sourceLocation

`string`

Cloud Storage URI of the source code to build, e.g.
"gs://my-bucket/source.zip". Uploading source is the caller's job;
this field tells the build where it lives.

### spec.buildConfig.functionTarget

`string`

Name of the function (as defined in the source code) to execute.
Defaults to the resource name suffix; the API falls back to a
function named "function" when the named one is not found.

### spec.buildConfig.imageUri

`string`

Artifact Registry URI where the built image is stored, e.g.
"us-docker.pkg.dev/project/repo/image".

### spec.buildConfig.baseImage

`string`

Base image used to build the function — a serverless runtimes image
(see containers[].base_image_uri for the serving-side pairing).

### spec.buildConfig.enableAutomaticUpdates

`bool`

Whether the function receives automatic base-image updates: Google
patches OS/runtime layers of the deployed image without a redeploy.

### spec.buildConfig.environmentVariables

`map<string, string>`

Build-time environment variables visible to the build process (NOT
to the running service — runtime env lives on the containers).

### spec.buildConfig.workerPool

`string`

Cloud Build Custom Worker Pool to run the build in, as
"projects/{project}/locations/{region}/workerPools/{workerPool}".
For builds that must run inside a private network perimeter.

### spec.buildConfig.serviceAccount

`string`

Service account the BUILD runs as, in the full resource form
"projects/{projectId}/serviceAccounts/{email}" (note: not a bare
email — this is the build-time identity, distinct from the runtime
service_account).

### spec.iapEnabled

`bool`

Puts the service behind Identity-Aware Proxy: callers authenticate
with Google identities and IAP enforces access policy before requests
reach the service — Google's managed login wall, with no code changes.

### spec.defaultUriDisabled

`bool`

Disables the default *.run.app URL, leaving the service reachable only
through custom domains or the load-balancer path. Removes the
often-forgotten second front door on locked-down services.

### spec.healthCheckDisabled

`bool`

Disables ALL health checking (startup and liveness probes) for the
revisions. An escape hatch for workloads whose serving model breaks
the probe contract — not a performance optimization.

### spec.multiRegionSettings

`GcpCloudRunMultiRegionSettings`

Multi-region service: one service identity serving from several
regions at once. Requires region = "global"; the actual serving
regions are listed here.

### spec.multiRegionSettings.regions

`[]string` · required

The regions the multi-region service deploys to, e.g.
["us-central1", "europe-west1"].

- rule: {"repeated":{"minItems":"1","unique":true,"items":{"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}}}

### spec.deletionPolicy

`string`

What happens to the Cloud Run service when this resource is
destroyed:
  "" / "DELETE" -- the service is deleted (default)
  "PREVENT"     -- destroy operations fail while this is set
  "ABANDON"     -- the service is removed from management but left
                   running in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.sandboxTemplates

`[]GcpCloudRunSandboxTemplate`

Sandbox templates the instance's supervisor container (the one with
sandbox_launcher) may launch on demand: isolated, short-lived
containers for executing untrusted or model-generated code beside
the serving container without exposing it. Each template names the
image and startup shape a sandbox runs with; the supervisor picks a
template by name at launch time. Requires exactly one container with
sandbox_launcher set.

### spec.sandboxTemplates[].name

`string` · required

Template name the supervisor launches by, a DNS label (RFC 1123):
lowercase letters, digits, hyphens; starts and ends alphanumeric.

- rule: {"required":true,"string":{"maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.sandboxTemplates[].image

`string` · required

Container image the sandbox runs, e.g.
"us-docker.pkg.dev/project/repo/sandbox:1.0.0". A bare name without a
registry host is pulled from Docker Hub.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sandboxTemplates[].command

`[]string`

Entrypoint array, not run through a shell. Empty uses the image's
ENTRYPOINT.

### spec.sandboxTemplates[].args

`[]string`

Arguments to the entrypoint. Empty uses the image's CMD.

### spec.sandboxTemplates[].env

`[]GcpCloudRunSandboxEnvVar`

Environment variables set in the sandbox: literal values only.

### spec.sandboxTemplates[].env[].name

`string` · required

Variable name, e.g. "PYTHONUNBUFFERED". Must not start with a digit.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_.-]*$"}}

### spec.sandboxTemplates[].env[].value

`string`

Literal value (up to 32768 characters). Never place credentials here.

- rule: {"string":{"maxLen":"32768"}}

### spec.sandboxTemplates[].volumeMounts

`[]GcpCloudRunVolumeMount`

Volumes (declared in spec.volumes) mounted into the sandbox's
filesystem.

### spec.sandboxTemplates[].volumeMounts[].name

`string` · required

Name of a volume declared in spec.volumes.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.sandboxTemplates[].volumeMounts[].mountPath

`string` · required

Absolute path in the container to mount at. Cloud SQL volumes must
mount at "/cloudsql".

- rule: {"required":true,"string":{"pattern":"^/.*$"}}

### spec.sandboxTemplates[].volumeMounts[].subPath

`string`

Path WITHIN the volume to mount instead of its root — e.g. mount only
one secret item or one bucket directory. Relative path; empty mounts
the volume root.

### spec.sandboxTemplates[].workingDir

`string`

Working directory for the entrypoint. Empty uses the image's WORKDIR.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the service at creation, as a map of
tagKeys/{tag_key_id} to tagValues/{tag_value_id} — the tag bindings
that organization policies, IAM conditions, and cost reports key on.
Immutable: changing the map replaces the service (Cloud Run applies
tags only at create), so plan tag changes as a redeploy.

- rule: {"map":{"keys":{"string":{"pattern":"^tagKeys/[0-9]+$"}},"values":{"string":{"pattern":"^tagValues/[0-9]+$"}}}}

## Validation Rules

- `auth.allow_unauthenticated_xor_invoker_disabled`: allow_unauthenticated grants public access through IAM; invoker_iam_disabled turns the IAM check off entirely — set at most one
- `gpu.redundancy_requires_accelerator`: gpu_zonal_redundancy_disabled only applies to GPU services — set node_selector.accelerator
- `multi_region.requires_global_region`: multi-region services deploy through the global endpoint — set region to "global" when using multi_region_settings
- `sandbox_templates_require_launcher`: sandbox_templates need a supervisor -- set sandbox_launcher on exactly one container

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudRun, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.url` | `string` | Canonical serving URL of the service (https://<service>-<hash>-<region>.run.app). |
| `status.outputs.service_name` | `string` | Name of the Cloud Run service as created in GCP — the handle serverless network endpoint groups and gcloud commands reference. |
| `status.outputs.revision` | `string` | Name of the latest ready revision — the revision serving LATEST traffic. |
| `status.outputs.location` | `string` | Region the service is deployed in (plain region name). |
| `status.outputs.uid` | `string` | Server-assigned unique identifier of the service, stable across its lifetime and never reused after deletion. |
| `status.outputs.urls` | `[]string` | Every URL serving this service: the canonical run.app URL plus any deterministic URLs. |
| `status.outputs.project_id` | `string` | GCP project the service is created in — with location and service_name, the service's full provider-side coordinate (projects/{project}/locations/{location}/services/{name}). |

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

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudRunDomainMapping | `spec.route` | `status.outputs.service_name` |
| GcpEventarcTrigger | `spec.destination.cloudRunService.service` | `status.outputs.service_name` |
| GcpIamOauthClient | `spec.allowedRedirectUris` | `status.outputs.url` |
| GcpPubSubSubscription | `spec.pushConfig.pushEndpoint` | `status.outputs.url` |
| GcpRegionNetworkEndpointGroup | `spec.cloudRun.service` | `status.outputs.service_name` |

## See Also

- [Overview](../README.md)
