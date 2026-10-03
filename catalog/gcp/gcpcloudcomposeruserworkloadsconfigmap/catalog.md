# GCP Cloud Composer User Workloads ConfigMap

Deploys a Kubernetes ConfigMap into a Cloud Composer environment's GKE cluster, managed as a first-class resource instead of a `kubectl` side channel. DAGs and KubernetesPodOperator tasks read its entries at runtime, and configuration changes flow through the same declarative review process as everything else — no drift against the environment's cluster. The name, environment, region, and project are immutable after creation; `data` is the one field that updates in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **User Workloads ConfigMap** -- a Kubernetes ConfigMap in the Composer environment's user-workloads namespace, holding the configured key-value data
- **Environment Attachment** -- the ConfigMap is created through the Cloud Composer API against the referenced environment, so its lifecycle follows the environment's

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Cloud Composer environment** (Composer 3, or Composer 2 with user workloads support) in the target project and region — reference a GcpCloudComposerEnvironment Infra Component via ValueFromRef. The Composer API is necessarily already enabled: a ConfigMap cannot exist without an environment, so the module deliberately performs no API enablement of its own.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Composer User Workloads ConfigMap**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **DAG Configuration** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsConfigMap
metadata:
  name: dag-config
  org: acme-corp
  env: prod
spec:
  region: us-central1
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
  configMapName: dag-config
  data:
    batch_size: "100"
    retry_count: "3"
```

```shell
planton apply -f config-map.yaml
```

This creates a ConfigMap named `dag-config` in the `prod-airflow` environment's cluster, holding two tuning values DAGs read at runtime (`projectId` omitted falls back to the provider's default project). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying alongside the Composer environment itself, wire the reference with ValueFromRef:

```yaml
spec:
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
  configMapName: etl-tuning
  data:
    batch_size: "500"
```

The InfraPipeline resolves the dependency graph, provisions the Composer environment first, then creates the ConfigMap inside it.

## Key Configuration

These are the most important decisions when configuring a user workloads ConfigMap. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The attachment pair is the identity** -- `environment` + `configMapName` locate the object; both are immutable (pointing the same data at a different environment is a new ConfigMap there). The name follows Kubernetes object-name rules: lowercase letters, numbers, and hyphens, up to 63 characters. Name for the configuration's role (`etl-tuning`, `feature-flags`) — a rename is a new ConfigMap plus a DAG reference update.

**Data is the day-2 lever** -- `data` requires at least one entry and updates the Kubernetes object in place; running tasks keep the values they already read, new tasks see the new values. Nothing restarts by itself — roll a change by applying it and letting the next scheduled runs pick it up. All values are strings: quote numbers and booleans, or YAML parses them as scalars the ConfigMap rejects.

**Plain text only** -- ConfigMap values are visible to anyone with cluster read access, and that visibility is the point: behavior changes stay reviewable as text. Anything you would not paste into a code review belongs in a GcpCloudComposerUserWorkloadsSecret instead.

**Deletion policy** -- a destroyed ConfigMap fails or silently changes every DAG that reads it, often less visibly than a missing Secret (defaults kick in). Set `deletionPolicy: PREVENT` for configuration live pipelines depend on so a stack teardown cannot take it along; `ABANDON` drops it from management but leaves the object in the cluster for a handover.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCloudComposerEnvironment** | `environment` | `status.outputs.environment_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | Fully qualified resource name (`projects/.../userWorkloadsConfigMaps/...`) | Addressing the ConfigMap in direct Composer API calls |
| `config_map_name` | The short Kubernetes object name | The name DAG code and KubernetesPodOperator tasks mount and read — consume this output rather than re-typing the literal |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**DAG configuration** -- per-environment tuning values (batch sizes, retry counts, endpoints, timezones) DAGs read at runtime instead of hard-coding, so a tuning change is an apply rather than a DAG redeploy. Start from the **DAG Configuration** preset.

**Feature flags** -- boolean-style flags that gate pipeline behavior (incremental load, data-quality checks, `dry_run`) between dev and prod without code changes. Keep the values quoted — unquoted `true`/`false` parse as YAML booleans and ConfigMap values must be strings. Start from the **Feature Flags** preset.

## Works With

- [**GCP Cloud Composer Environment**](/infra-catalog/gcp-cloud-composer-environment) -- the environment this ConfigMap lives in
- [**GCP Cloud Composer User Workloads Secret**](/infra-catalog/gcp-cloud-composer-user-workloads-secret) -- the sensitive twin for credentials
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project the environment runs in
