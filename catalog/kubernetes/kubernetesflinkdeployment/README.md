# Kubernetes Flink Deployment

## When NOT to Use This

**One resource is ONE Flink cluster** — the declaration of the
FlinkDeployment CR (`flinkdeployments.flink.apache.org/v1beta1`) that
the Flink Kubernetes Operator reconciles into a JobManager, its
TaskManagers, and (in application mode) the job they run.

Not the right kind when:

- **The operator is missing** — a KubernetesFlinkOperator whose watch
  scope covers this namespace is the PREREQUISITE. Nothing reconciles
  this declaration without it.
- **You want the operator** — that is KubernetesFlinkOperator; this
  resource declares one Flink cluster against it.

## Two shapes, one field apart

With `job` set this is an APPLICATION cluster — the cluster exists to
run that one job and follows its lifecycle. That is the recommended
production shape: isolation per pipeline. Without `job` it is a
SESSION cluster — an empty Flink runtime that accepts jobs submitted
separately (FlinkSessionJob CRs or the REST API); use it for many
short-lived jobs sharing warm capacity.

## The state truth, read this first

These are the operator's own validator rules, mirrored into this
spec so they fail at authoring time instead of at admission:

- **Any `job.upgrade_mode` other than `stateless` REQUIRES
  `state.checkpoints_dir`** — checkpoints are where the state to
  carry across upgrades lives.
- **`savepoint` mode additionally REQUIRES `state.savepoints_dir`** —
  the operator triggers a savepoint there on every upgrade.
- **JobManager `replicas` beyond 1 REQUIRE
  `state.high_availability`** — standbys coordinate through HA
  metadata, and HA is also required for the operator's rollback
  feature; it pairs naturally with `last-state` upgrades.

All directories must live on storage every pod can reach —
S3-compatible object storage is the normal answer; compose a
KubernetesSeaweedFs via `state.s3`. And choose `upgrade_mode`
honestly: `stateless` (the default) restarts from clean state,
correct ONLY for pipelines with no state worth carrying — on a
stateful job it silently discards state on every upgrade.
`last-state` restores the newest checkpoint (fast); `savepoint`
drains through a savepoint (the operationally safest).

## The S3 seam is Secret-native

`state.s3` carries the endpoint and credentials as SECRET REFERENCES,
rendered as pod environment (`AWS_ACCESS_KEY_ID` /
`AWS_SECRET_ACCESS_KEY` from secretKeyRefs) — NEVER into
`flink_configuration`, which renders into a ConfigMap in clear text.
The foreign-key defaults compose a KubernetesSeaweedFs (its S3
endpoint and its generated credentials Secret). Secret reads cannot
cross namespaces: co-locate the deployment with its object-store
credentials or replicate the Secret. Omit `state.s3` entirely on
clusters where pods reach object storage ambiently (IRSA/workload
identity on a cloud bucket).

**The plugin truth:** the official Flink images ship the S3
filesystem plugin DISABLED. `state.s3.builtin_plugin_jar` names the
exact jar under `/opt/flink/opt` in YOUR image (e.g.
`flink-s3-fs-hadoop-2.1.3.jar` for the current `flink:2.1` image —
the version in the name must match the image's Flink patch version).
A patch mismatch CrashLoopBackOffs the JobManager with no jar name
in the operator error — `ls /opt/flink/opt` is the diagnosis.
Without the field, every `s3://` path fails at runtime with
"unsupported filesystem scheme". Leave it empty only for custom
images that bake the plugin into `/opt/flink/plugins`.

## The slot-sizing arithmetic

Total task slots = TaskManager count ×
`taskmanager.numberOfTaskSlots` (a `flink_configuration` key, default
1), and a job needs `parallelism` slots to run. An under-slotted
cluster holds the job in a scheduling wait, not an error. In `native`
mode (the default) Flink requests TaskManagers on demand from the
job's parallelism — leave `task_manager.replicas` empty there
(setting it is ignored); in `standalone` mode it is the fixed worker
count.

## Version and image move in lockstep

`flink_version` names the Flink line in the CR's own enum form
(`v2_1` = Flink 2.1) and the default image derives from it
(`flink:2.1`). A custom image — your job jar and connectors baked
in — must carry exactly that Flink version: the operator shapes its
submission protocol from the declared version, and a mismatch fails
at runtime, not at apply. The operator also REFUSES changing
`flink_version` after a last-state suspend — restore first, then
upgrade.

## Config ownership

`flink_configuration` is Flink's own key space
(`taskmanager.numberOfTaskSlots`, restart strategies, state-backend
tuning). The operator FORBIDS keys it owns (`kubernetes.cluster-id`,
`kubernetes.namespace`, the HA cluster id) — they derive from this
resource and setting them is rejected at admission. The state
directories have first-class fields under `state` — prefer them; the
module-rendered state keys merge last, so colliding entries lose
deliberately. Never put credentials in it.

## The 45-character name budget

The operator derives `<name>-rest` and `<name>-taskmanager-N-M` child
names, and Kubernetes names cap at 63 characters. Both modules fail
loudly past the budget.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to deploy into — must be inside the
  operator's watch scope (literal or a KubernetesNamespace reference)
- **`spec.flink_version`**: the Flink line (`v1_13`…`v1_20`, `v2_0`,
  `v2_1`, `v2_2`; `v1_20` is the last 1.x LTS line)

### Common

- **`spec.job`**: the application-cluster arm — `jar_uri` (required;
  `local:///…` points inside the image, the production pattern),
  `entry_class`, `args`, `parallelism`, `state`
  (`running`/`suspended` — the declarative pause), `upgrade_mode`,
  `initial_savepoint_path` (bootstrap the FIRST start when migrating
  a job in), `allow_non_restored_state`, `savepoint_trigger_nonce`
  (the declarative manual savepoint, into `state.savepoints_dir`)
- **`spec.job_manager`**: resources (default 1 CPU / 2Gi; Flink
  derives its JVM sizing from container memory — resize instead of
  tuning JVM flags) and `replicas` (>1 = warm standbys, HA required)
- **`spec.task_manager`**: resources (default 1 CPU / 2Gi) and
  `replicas` (standalone mode only)
- **`spec.state`**: the durable-state posture (see the state truth
  and the S3 seam above)
- **`spec.flink_configuration`**: Flink's own key space (see config
  ownership above)
- **`spec.mode`**: `native` (default — the operator-recommended mode)
  or `standalone` (fixed resources, no Flink↔Kubernetes API traffic)
- **`spec.service_account`**: default `flink` — the account the
  KubernetesFlinkOperator's chart creates with reconcile RBAC (its
  `job_service_account` output); point elsewhere only if you
  provision equivalent RBAC yourself
- **`spec.log_configuration`**: Log4j overrides (file name → content)
- **`spec.scheduling`** / **`spec.image_pull_secrets`**: rendered
  through the CR's pod template, applied to JobManager and
  TaskManager pods alike
- **`spec.restart_nonce`**: change it to force a restart without any
  spec change (the declarative "kick")

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the Flink cluster runs in |
| `rest_service` | The JobManager REST Service (`<name>-rest`) — the Flink REST API and web UI (port 8081); where session-mode jobs submit and where job status reads from |
| `rest_endpoint` | In-cluster REST endpoint (`<rest_service>.<namespace>.svc.cluster.local:8081`) |
| `port_forward_command` | kubectl port-forward one-liner for reaching the Flink UI from a workstation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **KubernetesFlinkOperator is the prerequisite** — its watch scope
  must cover this namespace, and the pods run as the `flink` service
  account its chart creates.
- **`state.s3.endpoint`** is a foreign key (default kind
  KubernetesSeaweedFs, field path `status.outputs.s3_endpoint`), and
  the credential selectors default to the same store's generated
  credentials Secret (keys `admin_access_key_id` /
  `admin_secret_access_key` for the admin identity). Deploy the store
  first, in this same namespace, and declare the bucket on the
  store — Flink does not create buckets.
- **The apply is deliberately non-blocking**: cluster readiness
  depends on the operator (image pulls, job submission, TaskManager
  registration) — the CR applies and the operator reconciles; nothing
  blocks on a controller.

## Examples

### Application cluster, stateless pipeline

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkDeployment
metadata:
  name: orders-pipeline
spec:
  namespace:
    value: stream-processing
  createNamespace: true
  flinkVersion: v2_1
  image: registry.example.com/pipelines/orders:2.1
  job:
    jarUri: local:///opt/pipeline/orders.jar
    parallelism: 4
  taskManager:
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "2"
        memory: 4Gi
```

### Stateful, HA, savepoint upgrades on composed S3

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkDeployment
metadata:
  name: orders-pipeline
spec:
  namespace:
    value: stream-processing
  createNamespace: true
  flinkVersion: v2_1
  image: registry.example.com/pipelines/orders:2.1
  job:
    jarUri: local:///opt/pipeline/orders.jar
    parallelism: 4
    upgradeMode: savepoint
  jobManager:
    replicas: 2
  flinkConfiguration:
    taskmanager.numberOfTaskSlots: "2"
    execution.checkpointing.interval: 30s
  state:
    checkpointsDir: s3://flink-state/checkpoints/orders
    savepointsDir: s3://flink-state/savepoints/orders
    highAvailability:
      enabled: true
      storageDir: s3://flink-state/ha/orders
    s3:
      endpoint:
        valueFrom:
          name: objects
      accessKeySecret:
        name:
          valueFrom:
            name: objects
        key: admin_access_key_id
      secretKeySecret:
        name:
          valueFrom:
            name: objects
        key: admin_secret_access_key
      builtinPluginJar: flink-s3-fs-hadoop-2.1.3.jar
  taskManager:
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "2"
        memory: 4Gi
```

### Session cluster

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkDeployment
metadata:
  name: flink-session
spec:
  namespace:
    value: stream-processing
  createNamespace: true
  flinkVersion: v2_1
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
