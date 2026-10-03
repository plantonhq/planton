# Flink Deployment

Declares one Flink cluster -- the `FlinkDeployment` custom resource (`flinkdeployments.flink.apache.org/v1beta1`) that the Apache Flink Kubernetes Operator reconciles into a JobManager, its TaskManagers, and (in application mode) the job they run. Two shapes, one field apart: with `job` set this is an APPLICATION cluster -- the cluster exists to run that one job and follows its lifecycle, the recommended production shape, isolation per pipeline. Without `job` it is a SESSION cluster -- an empty Flink runtime that accepts jobs submitted separately (FlinkSessionJob CRs or the REST API), for many short-lived jobs sharing warm capacity. A KubernetesFlinkOperator whose watch scope covers this namespace is the PREREQUISITE: nothing reconciles this declaration without it, and a declaration outside the operator's fence is ignored without an error. Uses a Kubernetes Provider Connection for cluster access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **FlinkDeployment custom resource** -- the declaration itself, carrying the Flink version, image, job spec, sizing, state posture, and Flink configuration; the OPERATOR then reconciles it into:
  - JobManager pod(s) -- the cluster coordinator, with standbys behind HA metadata when `jobManager.replicas` exceeds 1
  - TaskManager pods -- the workers; requested on demand from the job's parallelism in `native` mode (the default), or the fixed `taskManager.replicas` count in `standalone` mode
  - The `<name>-rest` Service -- the Flink REST API and web UI on port 8081
  - In application mode: the job itself, submitted from `job.jarUri`
- **S3 credential wiring** -- when `state.s3` is set, the endpoint and credentials render as pod environment (`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` from Secret references) -- NEVER into `flinkConfiguration`, which renders into a ConfigMap in clear text
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

The apply is deliberately NON-blocking: the CR applies and the operator reconciles -- cluster readiness (image pulls, job submission, TaskManager registration) is the operator's business, and nothing blocks on a controller.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A running Flink Operator whose fence covers this namespace** -- deploy KubernetesFlinkOperator first. Under its fenced posture, a namespace missing from `watchNamespaces` looks like a deployment that never reconciles -- no error is raised.
- **The `flink` service account in this namespace** -- the operator's chart creates it with reconcile RBAC (its `job_service_account` output). Point `serviceAccount` elsewhere only if you provision equivalent RBAC yourself.
- **Object storage for stateful pipelines** -- checkpoint, savepoint, and HA directories must live on storage every pod can reach; S3-compatible object storage is the normal answer. Compose a KubernetesSeaweedFs via `state.s3`, or omit `state.s3` on clusters where pods reach object storage ambiently (IRSA / workload identity on a cloud bucket). Secret reads cannot cross namespaces: co-locate the deployment with its object-store credentials Secret or replicate it.
- **A name within budget** -- keep `metadata.name` at 45 characters or fewer: the operator derives `<name>-rest` and `<name>-taskmanager-N-M` child names against the Kubernetes 63-character cap. Both engines fail loudly over the budget.

## Deploy

### Console

Open the deployment store, find **Flink Deployment**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Application (stateless) preset** for the recommended one-cluster-per-pipeline grain, or the **Stateful HA-on-S3 preset** for the full durable-state posture, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkDeployment
metadata:
  name: orders-pipeline
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "stream-processing"
  createNamespace: true
  flinkVersion: v2_1
  image: registry.example.com/pipelines/orders:2.1
  job:
    jarUri: local:///opt/pipeline/orders.jar
    parallelism: 4
```

```shell
planton apply -f orders-pipeline.yaml
```

This declares an application cluster for one pipeline: a custom image with the job jar baked in (`local:///` paths point inside the image -- the production pattern), Flink 2.1, and four parallel subtasks. The operator picks up the declaration and reconciles the JobManager, TaskManagers, and job. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the deployment to a namespace and an object store managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: stream-processing-namespace
      fieldPath: spec.name
  createNamespace: false
  state:
    s3:
      endpoint:
        valueFrom:
          kind: KubernetesSeaweedFs
          name: objects
          fieldPath: status.outputs.s3_endpoint
```

The InfraPipeline deploys the namespace and the object store first, then declares the Flink cluster against them. The S3 credential selectors default to the same store's generated credentials Secret (keys `admin_access_key_id` / `admin_secret_access_key`).

## Key Configuration

These are the most important decisions when configuring the Flink Deployment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Two shapes, one field apart** -- `job` set = APPLICATION cluster: the cluster exists to run that one job, the recommended production grain. `job` absent = SESSION cluster: warm shared capacity for jobs submitted separately. `job.jarUri` is the required heart of the application arm: `local:///…` points inside the image (the production pattern); remote schemes download at submission.

**The state truth, read this first** -- These are the operator's own validator rules, mirrored into the spec so they fail at authoring time instead of at admission: any `job.upgradeMode` other than `stateless` REQUIRES `state.checkpointsDir`; `savepoint` mode additionally REQUIRES `state.savepointsDir`; JobManager replicas beyond 1 REQUIRE `state.highAvailability` (with its `storageDir`). And choose `upgradeMode` honestly: `stateless` (the default) restarts from clean state -- on a stateful job it silently discards state on every upgrade. `last-state` restores the newest checkpoint (fast); `savepoint` drains through a savepoint (the operationally safest).

**The S3 seam is Secret-native** -- `state.s3` carries the endpoint and credentials as SECRET REFERENCES, rendered as pod environment -- never into `flinkConfiguration`, which renders into a ConfigMap in clear text. The foreign-key defaults compose a KubernetesSeaweedFs. Omit `state.s3` entirely where pods reach object storage ambiently.

**The plugin truth** -- The official Flink images ship the S3 filesystem plugin DISABLED. `state.s3.builtinPluginJar` names the exact jar under `/opt/flink/opt` in YOUR image (e.g. `flink-s3-fs-hadoop-2.1.3.jar` for the `flink:2.1` image -- the version in the name must match the image's Flink patch version). A patch mismatch CrashLoopBackOffs the JobManager with no jar name in the operator error -- `ls /opt/flink/opt` is the diagnosis. Without the field, every `s3://` path fails at runtime with "unsupported filesystem scheme". Leave it empty only for custom images that bake the plugin into `/opt/flink/plugins`.

**Version and image move in lockstep** -- `flinkVersion` names the Flink line in the CR's own enum form (`v2_1` = Flink 2.1) and the default image derives from it (`flink:2.1`). A custom image must carry exactly that Flink version: the operator shapes its submission protocol from the declared version, and a mismatch fails at runtime, not at apply. The operator also REFUSES changing `flinkVersion` after a last-state suspend -- restore first, then upgrade.

**The slot-sizing arithmetic** -- Total task slots = TaskManager count × `taskmanager.numberOfTaskSlots` (a `flinkConfiguration` key, default 1), and a job needs `parallelism` slots to run. An under-slotted cluster holds the job in a scheduling wait, not an error. In `native` mode (the default) Flink requests TaskManagers on demand from the job's parallelism -- leave `taskManager.replicas` empty there (setting it is ignored); in `standalone` mode it is the fixed worker count.

**Sizing is JVM sizing** -- JobManager and TaskManager default to 1 CPU / 2Gi each. Flink derives its JVM sizing from container memory -- resize the containers instead of tuning JVM flags.

**Config ownership** -- `flinkConfiguration` is Flink's own key space (`taskmanager.numberOfTaskSlots`, restart strategies, state-backend tuning). The operator FORBIDS keys it owns (`kubernetes.cluster-id`, `kubernetes.namespace`, the HA cluster id) -- setting them is rejected at admission. The state directories have first-class fields under `state` -- prefer them; the module-rendered state keys merge last, so colliding entries lose deliberately. Never put credentials in it: it renders into a clear-text ConfigMap.

**Declarative operations** -- `job.state: suspended` is the declarative pause; `job.savepointTriggerNonce` triggers a manual savepoint into `state.savepointsDir` when changed; `restartNonce` forces a restart without any spec change (the declarative "kick"); `job.initialSavepointPath` bootstraps the FIRST start when migrating a job in, with `job.allowNonRestoredState` for dropped-operator migrations.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSeaweedFs** | `state.s3.endpoint` | `status.outputs.s3_endpoint` |
| **KubernetesFlinkOperator** (runtime prerequisite) | -- | its watch scope must cover this namespace; pods run as its `job_service_account` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the Flink cluster runs in | Co-locating FlinkSessionJob submissions and diagnostics |
| `rest_service` | The JobManager REST Service (`<name>-rest`) -- the Flink REST API and web UI (port 8081) | Where session-mode jobs submit and where job status reads from |
| `rest_endpoint` | In-cluster REST endpoint (`<rest_service>.<namespace>.svc.cluster.local:8081`) | Wiring dashboards, job submitters, and health checks |
| `port_forward_command` | kubectl port-forward one-liner for the Flink UI | Reaching the web UI from a workstation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Application cluster, stateless pipeline** -- One cluster per pipeline, a custom image with the job jar baked in, no durable state: correct ONLY for pipelines with no state worth carrying across upgrades. Start from the **Application (stateless) preset**.

**Stateful, HA, savepoint upgrades on composed S3** -- The full durable-state posture: checkpoints, savepoints, and HA metadata on S3-compatible storage, a standby JobManager, `savepoint` upgrade mode, and the S3 seam composed from a KubernetesSeaweedFs. Start from the **Stateful HA-on-S3 preset**.

**Session cluster** -- `flinkVersion` and sizing only, no `job`: an empty Flink runtime for many short-lived jobs sharing warm capacity, submitted via FlinkSessionJob CRs or the REST API at `rest_endpoint`.

## Works With

- [**Flink Operator**](/infra-catalog/kubernetes-flink-operator) -- the PREREQUISITE: reconciles this declaration; its watch scope must cover this namespace
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the deployment
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) -- the composed S3-compatible object store for checkpoints, savepoints, and HA metadata; the `state.s3` foreign-key defaults point at it
