# KubernetesFlinkDeployment

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesFlinkDeploymentSpec** declares a Flink cluster — the
Kubernetes CR (`flinkdeployments.flink.apache.org/v1beta1`) that
the Flink Kubernetes Operator reconciles into a JobManager, its
TaskManagers, and (in application mode) the job they run.
PREREQUISITE: a KubernetesFlinkOperator whose watch scope covers
this namespace.

TWO SHAPES, ONE FIELD APART: with `job` set this is an APPLICATION
cluster — the cluster exists to run that one job and follows its
lifecycle (the recommended production shape: isolation per
pipeline). Without `job` it is a SESSION cluster — an empty Flink
runtime that accepts jobs submitted separately (FlinkSessionJob
CRs or the REST API); use it for many short-lived jobs sharing
warm capacity.

STATE TRUTH, READ THIS FIRST (mirrored from the operator's own
validator at 1.15.0): any `job.upgrade_mode` other than
"stateless" REQUIRES `state.checkpoints_dir`, and "savepoint"
additionally REQUIRES `state.savepoints_dir` — both on storage
every pod can reach (S3-compatible object storage is the normal
answer; compose a KubernetesSeaweedFs via `state.s3`). Without
them the operator rejects the deployment. "last-state" upgrades
restore from the newest checkpoint and pair naturally with
`state.high_availability` (also required for the operator's
rollback feature).

CONFIG OWNERSHIP: `flink_configuration` is Flink's own key space
(taskmanager.numberOfTaskSlots, state backends, restart
strategies…). The operator FORBIDS keys it owns
(kubernetes.cluster-id, kubernetes.namespace, the HA cluster id) —
they are derived from this resource; setting them is rejected at
admission. NEVER put credentials in it — it renders into a
ConfigMap in clear text; S3 credentials ride `state.s3`'s Secret
selectors as pod environment instead.

VERSION/IMAGE LOCKSTEP: `flink_version` names the Flink line
("v2_1" style, the CR's own enum) and the default image derives
from it (`flink:<major>.<minor>` — e.g. v2_1 → flink:2.1). A
custom image must carry exactly that Flink version — the operator
shapes its submission protocol from the declared version and a
mismatch fails at runtime, not at apply.

NAME BUDGET: keep this resource's name at 45 characters or fewer —
the operator derives `<name>-rest`/`<name>-taskmanager-…` child
names and Kubernetes names cap at 63. Both modules fail loudly
past the budget.

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: every
# module-rendered arm of the deployment expressed at once — an
# application-mode job on the savepoint upgrade path (exercising both
# state-directory requirements), JobManager standbys behind Kubernetes
# HA, explicit Flink configuration, the S3 credential seam as Secret
# references with the plugin activation, log configuration, pull
# secrets, scheduling, and both declarative nonces.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkDeployment
metadata:
  name: orders-pipeline-full
spec:
  namespace:
    value: stream-processing
  createNamespace: true
  flinkVersion: v2_1
  image: flink:2.1
  job:
    jarUri: local:///opt/flink/examples/streaming/StateMachineExample.jar
    entryClass: org.apache.flink.streaming.examples.statemachine.StateMachineExample
    args:
      - --error-rate
      - "0.0"
    parallelism: 4
    state: running
    upgradeMode: savepoint
    allowNonRestoredState: false
    savepointTriggerNonce: 1
  jobManager:
    replicas: 2
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "1"
        memory: 2Gi
  taskManager:
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "2"
        memory: 4Gi
  flinkConfiguration:
    taskmanager.numberOfTaskSlots: "2"
    restart-strategy.type: exponential-delay
  state:
    checkpointsDir: s3://flink-state/checkpoints/orders
    savepointsDir: s3://flink-state/savepoints/orders
    highAvailability:
      enabled: true
      storageDir: s3://flink-state/ha/orders
    s3:
      endpoint:
        value: http://objects-s3.stream-processing.svc.cluster.local:8333
      pathStyleAccess: true
      accessKeySecret:
        name:
          value: objects-s3-secret
        key: admin_access_key_id
      secretKeySecret:
        name:
          value: objects-s3-secret
        key: admin_secret_access_key
      builtinPluginJar: flink-s3-fs-hadoop-2.1.3.jar
  mode: native
  serviceAccount: flink
  logConfiguration:
    log4j-console.properties: |
      rootLogger.level = INFO
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      workload: stream
  restartNonce: 1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.flinkVersion` | `string` | yes |  |  |
| `spec.image` | `string` |  |  |  |
| `spec.job` | `KubernetesFlinkDeploymentJob` |  |  |  |
| `spec.job.jarUri` | `string` | yes |  |  |
| `spec.job.entryClass` | `string` |  |  |  |
| `spec.job.args` | `[]string` |  |  |  |
| `spec.job.parallelism` | `int32` |  | `1` |  |
| `spec.job.state` | `string` |  | `running` |  |
| `spec.job.upgradeMode` | `string` |  | `stateless` |  |
| `spec.job.initialSavepointPath` | `string` |  |  |  |
| `spec.job.allowNonRestoredState` | `bool` |  |  |  |
| `spec.job.savepointTriggerNonce` | `int64` |  |  |  |
| `spec.jobManager` | `KubernetesFlinkDeploymentJobManager` |  |  |  |
| `spec.jobManager.resources` | `ContainerResources` |  |  |  |
| `spec.jobManager.resources.limits` | `CpuMemory` |  |  |  |
| `spec.jobManager.resources.limits.cpu` | `string` |  |  |  |
| `spec.jobManager.resources.limits.memory` | `string` |  |  |  |
| `spec.jobManager.resources.requests` | `CpuMemory` |  |  |  |
| `spec.jobManager.resources.requests.cpu` | `string` |  |  |  |
| `spec.jobManager.resources.requests.memory` | `string` |  |  |  |
| `spec.jobManager.replicas` | `int32` |  | `1` |  |
| `spec.taskManager` | `KubernetesFlinkDeploymentTaskManager` |  |  |  |
| `spec.taskManager.resources` | `ContainerResources` |  |  |  |
| `spec.taskManager.resources.limits` | `CpuMemory` |  |  |  |
| `spec.taskManager.resources.limits.cpu` | `string` |  |  |  |
| `spec.taskManager.resources.limits.memory` | `string` |  |  |  |
| `spec.taskManager.resources.requests` | `CpuMemory` |  |  |  |
| `spec.taskManager.resources.requests.cpu` | `string` |  |  |  |
| `spec.taskManager.resources.requests.memory` | `string` |  |  |  |
| `spec.taskManager.replicas` | `int32` |  |  |  |
| `spec.flinkConfiguration` | `map<string, string>` |  |  |  |
| `spec.state` | `KubernetesFlinkDeploymentState` |  |  |  |
| `spec.state.checkpointsDir` | `string` |  |  |  |
| `spec.state.savepointsDir` | `string` |  |  |  |
| `spec.state.highAvailability` | `KubernetesFlinkDeploymentHighAvailability` |  |  |  |
| `spec.state.highAvailability.enabled` | `bool` |  |  |  |
| `spec.state.highAvailability.storageDir` | `string` |  |  |  |
| `spec.state.s3` | `KubernetesFlinkDeploymentS3` |  |  |  |
| `spec.state.s3.endpoint` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_endpoint`) |
| `spec.state.s3.pathStyleAccess` | `bool` |  | `true` |  |
| `spec.state.s3.accessKeySecret` | `KubernetesFlinkDeploymentSecretSelector` | yes |  |  |
| `spec.state.s3.accessKeySecret.name` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`) |
| `spec.state.s3.accessKeySecret.key` | `string` | yes |  |  |
| `spec.state.s3.secretKeySecret` | `KubernetesFlinkDeploymentSecretSelector` | yes |  |  |
| `spec.state.s3.secretKeySecret.name` | `string \| valueFrom` | yes |  | KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`) |
| `spec.state.s3.secretKeySecret.key` | `string` | yes |  |  |
| `spec.state.s3.builtinPluginJar` | `string` |  |  |  |
| `spec.mode` | `string` |  | `native` |  |
| `spec.serviceAccount` | `string` |  | `flink` |  |
| `spec.logConfiguration` | `map<string, string>` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesFlinkDeploymentScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.restartNonce` | `int64` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource. Must be inside the
KubernetesFlinkOperator's watch scope. NOTE the `state.s3`
credential Secrets ride secretKeyRefs, readable only from this
same namespace — co-locate the deployment with its object store
credentials or replicate the Secret.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the CR and deleted with the resource.
When false, the namespace must already exist.

### spec.flinkVersion

`string` · required

Flink version line, in the CR's own enum form (e.g. "v2_1" =
Flink 2.1, "v1_20" = the last 1.x LTS line). Drives the default
image and the operator's submission protocol. The operator
REFUSES changing this after a last-state suspend — restore
first, then upgrade.

- rule: flink_version must be one of the CR enum values: v1_13…v1_20, v2_0, v2_1, v2_2.
- rule: {"required":true}

### spec.image

`string`

Container image for every Flink pod. Empty = the official image
derived from flink_version (`flink:<major>.<minor>` — e.g.
flink:2.1). Set for custom images with your job jar or
connectors baked in — and keep the Flink inside identical to
flink_version.

### spec.job

`KubernetesFlinkDeploymentJob`

The job — set it for an APPLICATION cluster (the cluster runs
exactly this job and follows its lifecycle); omit it for a
SESSION cluster (an empty runtime accepting external
submissions).

### spec.job.jarUri

`string` · required

URI of the job jar. `local:///…` paths point INSIDE the image
(the production pattern: bake the jar); the Flink images also
ship examples under /opt/flink/examples. Remote schemes
(https://, s3://) download at submission.

- rule: {"required":true}

### spec.job.entryClass

`string`

Fully qualified main class. Empty = the jar manifest's
Main-Class.

### spec.job.args

`[]string`

Program arguments passed to main().

### spec.job.parallelism

`int32` · optional (explicit presence)

Job parallelism — the slots the job needs (see the TaskManager
sizing truth). Empty = 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.job.state

`string` · optional (explicit presence)

Desired state: "running" (empty default) or "suspended" — the
declarative pause. Suspending with a stateful upgrade_mode
retains state for the resume; how much depends on the mode
(savepoint drains to a savepoint; last-state relies on the
newest checkpoint).

- default: `running`
- rule: Job state must be "running" or "suspended".

### spec.job.upgradeMode

`string` · optional (explicit presence)

How spec changes carry job state across the restart:
"stateless" (empty default — restart from clean state; correct
for stateless pipelines only), "last-state" (restore the newest
checkpoint — fast, needs state.checkpoints_dir), or "savepoint"
(drain through a savepoint — the operationally safest; needs
both directories). The spec-level rules enforce the directory
requirements.

- default: `stateless`
- rule: upgrade_mode must be "stateless", "last-state", or "savepoint".

### spec.job.initialSavepointPath

`string`

Savepoint path to bootstrap the FIRST start from (migrating a
job into this declaration). Later restores are governed by
upgrade_mode.

### spec.job.allowNonRestoredState

`bool`

Allow restoring from a savepoint that carries state for
operators no longer in the job graph (after removing an
operator from the pipeline). Off = the restore fails instead of
silently dropping state.

### spec.job.savepointTriggerNonce

`int64`

Change this value to trigger a savepoint NOW (the CR's
savepointTriggerNonce) — the declarative manual savepoint, into
state.savepoints_dir.

### spec.jobManager

`KubernetesFlinkDeploymentJobManager`

The JobManager — the cluster's coordinator. Always exactly one
ACTIVE JobManager; replicas beyond 1 are warm standbys and
require state.high_availability (mirrored from the operator's
validator).

### spec.jobManager.resources

`ContainerResources`

CPU and memory for the JobManager. Empty = 1 CPU / 2Gi (the
upstream example sizing). Flink derives its JVM sizing from the
container memory — resize instead of tuning JVM flags.

### spec.jobManager.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.jobManager.resources.limits.cpu

`string`

### spec.jobManager.resources.limits.memory

`string`

### spec.jobManager.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.jobManager.resources.requests.cpu

`string`

### spec.jobManager.resources.requests.memory

`string`

### spec.jobManager.replicas

`int32` · optional (explicit presence)

JobManager count. Empty = 1. More than 1 = warm standbys —
requires state.high_availability (validated at the spec level).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.taskManager

`KubernetesFlinkDeploymentTaskManager`

The TaskManagers — the worker processes. Sizing truth: total
task slots = replicas × "taskmanager.numberOfTaskSlots" (a
flink_configuration key, default 1), and a job needs
`parallelism` slots to run — an under-slotted cluster holds the
job in a scheduling wait, not an error.

### spec.taskManager.resources

`ContainerResources`

CPU and memory for each TaskManager. Empty = 1 CPU / 2Gi. Flink
derives managed/network memory from the container memory —
resize instead of tuning JVM flags.

### spec.taskManager.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.taskManager.resources.limits.cpu

`string`

### spec.taskManager.resources.limits.memory

`string`

### spec.taskManager.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.taskManager.resources.requests.cpu

`string`

### spec.taskManager.resources.requests.memory

`string`

### spec.taskManager.replicas

`int32` · optional (explicit presence)

TaskManager count. In native mode (the default) leave this
empty — Flink requests TaskManagers on demand from the job's
parallelism; setting it there is ignored. In standalone mode it
is the fixed worker count (empty = 1).

- rule: {"int32":{"gte":1}}

### spec.flinkConfiguration

`map<string, string>`

Flink configuration, Flink's own key space (e.g.
"taskmanager.numberOfTaskSlots": "2", restart strategies, state
backend tuning). Keys the operator owns (kubernetes.cluster-id,
kubernetes.namespace, the HA cluster id) are REJECTED at
admission — they derive from this resource. State/HA
directories have first-class fields under `state` — prefer
them. NEVER put credentials here (renders into a ConfigMap).

### spec.state

`KubernetesFlinkDeploymentState`

Durable state posture: checkpoint/savepoint directories, HA
metadata, and the S3 credential seam. Required (in parts) by
every non-stateless upgrade mode — see the spec-level STATE
TRUTH.

- rule: High availability needs high_availability.storage_dir — the recovery metadata must live on storage every JobManager can reach (s3://… on a composed object store).

### spec.state.checkpointsDir

`string`

Checkpoint directory (e.g. "s3://flink-state/checkpoints/
my-pipeline"). Required for last-state and savepoint upgrade
modes. Every pod must reach it — object storage, not a local
path.

### spec.state.savepointsDir

`string`

Savepoint directory (e.g. "s3://flink-state/savepoints/
my-pipeline"). Required for the savepoint upgrade mode and for
savepoint_trigger_nonce.

### spec.state.highAvailability

`KubernetesFlinkDeploymentHighAvailability`

Kubernetes-native high availability: JobManager metadata in a
ConfigMap, recovery state in storage_dir. Required for
JobManager standbys and the operator's rollback feature; pairs
naturally with last-state upgrades.

### spec.state.highAvailability.enabled

`bool`

Enable Kubernetes HA services.

### spec.state.highAvailability.storageDir

`string`

Recovery-state directory (e.g. "s3://flink-state/ha/
my-pipeline"). Required when enabled.

### spec.state.s3

`KubernetesFlinkDeploymentS3`

S3-compatible access for the s3:// directories above —
endpoint + credentials as SECRET REFERENCES (rendered as pod
environment, never into Flink config). Compose a
KubernetesSeaweedFs: the FK default wires its S3 endpoint.
Omit on clusters where pods reach object storage ambiently
(IRSA/workload identity on a cloud bucket).

### spec.state.s3.endpoint

`string | valueFrom` · required

S3 endpoint URL. Accepts a literal (e.g. an AWS regional
endpoint) or a reference to a KubernetesSeaweedFs resource (its
in-cluster S3 endpoint).

- references: KubernetesSeaweedFs (`status.outputs.s3_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_endpoint}} -- a bare string does not parse

### spec.state.s3.pathStyleAccess

`bool` · optional (explicit presence)

Path-style addressing (bucket in the path, not the hostname).
Empty = true — correct for in-cluster object stores
(SeaweedFS/MinIO-compatible); set false for AWS S3 itself.

- default: `true`

### spec.state.s3.accessKeySecret

`KubernetesFlinkDeploymentSecretSelector` · required

Access-key id, read from a Secret (on a KubernetesSeaweedFs:
its S3 credentials Secret).

- rule: {"required":true}

### spec.state.s3.accessKeySecret.name

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesSeaweedFs resource (its S3 credentials Secret).

- references: KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_credentials_secret_name}} -- a bare string does not parse

### spec.state.s3.accessKeySecret.key

`string` · required

Key within the Secret (on a SeaweedFS credentials Secret:
"admin_access_key_id" / "admin_secret_access_key" for the admin
identity, or your declared identity's keys).

- rule: {"required":true}

### spec.state.s3.secretKeySecret

`KubernetesFlinkDeploymentSecretSelector` · required

Secret access key, read from a Secret.

- rule: {"required":true}

### spec.state.s3.secretKeySecret.name

`string | valueFrom` · required

Secret name. Accepts a literal or a reference to a
KubernetesSeaweedFs resource (its S3 credentials Secret).

- references: KubernetesSeaweedFs (`status.outputs.s3_credentials_secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSeaweedFs, name: <that resource's name>, fieldPath: status.outputs.s3_credentials_secret_name}} -- a bare string does not parse

### spec.state.s3.secretKeySecret.key

`string` · required

Key within the Secret (on a SeaweedFS credentials Secret:
"admin_access_key_id" / "admin_secret_access_key" for the admin
identity, or your declared identity's keys).

- rule: {"required":true}

### spec.state.s3.builtinPluginJar

`string`

Name of the S3 filesystem plugin jar to activate from the image's
bundled plugin set (e.g. "flink-s3-fs-hadoop-2.1.3.jar" — the
exact file under /opt/flink/opt in YOUR image; the version in the
name must match the image's Flink patch version). A patch
mismatch CrashLoopBackOffs the JobManager; the operator error
never names the jar — `ls /opt/flink/opt` in the image is the
diagnosis. The official images ship the plugin DISABLED: without
this (or a custom image that bakes the plugin into
/opt/flink/plugins) every s3:// path fails at runtime with
"unsupported filesystem scheme". Leave empty only for images
that pre-bake the plugin.

### spec.mode

`string` · optional (explicit presence)

Deployment mode: "native" (empty default — Flink asks Kubernetes
for TaskManagers on demand; the operator-recommended mode) or
"standalone" (fixed resources, no Flink↔Kubernetes API traffic).

- default: `native`
- rule: mode must be "native" or "standalone".

### spec.serviceAccount

`string` · optional (explicit presence)

Service account the Flink pods run as. Empty = "flink" — the
account the KubernetesFlinkOperator's chart creates with
reconcile RBAC (its job_service_account output). Point elsewhere
only if you provision equivalent RBAC yourself.

- default: `flink`

### spec.logConfiguration

`map<string, string>`

Log4j configuration overrides (file name → content), e.g.
"log4j-console.properties" → rootLogger tuning. Empty = the
image defaults.

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to every Flink pod.

### spec.scheduling

`KubernetesFlinkDeploymentScheduling`

Scheduling for every Flink pod (rendered through the CR's pod
template).

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for tainted nodes.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

Priority class name for the pods.

### spec.restartNonce

`int64`

Change this value to force the operator to restart the
deployment without any spec change (the CR's restartNonce —
a declarative "kick").

## Validation Rules

- `spec.state.checkpoints_for_stateful_upgrades`: last-state and savepoint upgrade modes need state.checkpoints_dir (the operator rejects the deployment without it — checkpoints are where the state to carry across upgrades lives). Use s3://… on a composed object store.
- `spec.state.savepoints_dir_for_savepoint_mode`: The savepoint upgrade mode needs state.savepoints_dir — the operator triggers a savepoint there on every upgrade and rejects the deployment without it.
- `spec.job_manager.standby_needs_ha`: JobManager replicas beyond 1 are standbys coordinated through HA metadata — enable state.high_availability (the operator rejects standby replicas without it).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesFlinkDeployment, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the Flink cluster runs in. |
| `status.outputs.rest_service` | `string` | the JobManager REST Service name (`<name>-rest`) — the Flink REST API and web UI (port 8081); where session-mode jobs submit and where job status reads from. |
| `status.outputs.rest_endpoint` | `string` | in-cluster REST endpoint (`<rest_service>.<namespace>.svc.cluster.local:8081`). |
| `status.outputs.port_forward_command` | `string` | kubectl port-forward one-liner for reaching the Flink UI from a workstation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.state.s3.endpoint` | KubernetesSeaweedFs | `status.outputs.s3_endpoint` |
| `spec.state.s3.accessKeySecret.name` | KubernetesSeaweedFs | `status.outputs.s3_credentials_secret_name` |
| `spec.state.s3.secretKeySecret.name` | KubernetesSeaweedFs | `status.outputs.s3_credentials_secret_name` |

## See Also

- [Overview](../README.md)
