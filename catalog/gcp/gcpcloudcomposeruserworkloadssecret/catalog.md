# GCP Cloud Composer User Workloads Secret

Deploys a Kubernetes Secret inside a Cloud Composer environment's GKE cluster, managed as a first-class resource: DAGs read credentials from it at runtime, and rotation flows through the same declarative review process as everything else — no `kubectl` against the environment's cluster and no ad-hoc mutation of the environment. Values are base64-encoded at the source (Kubernetes Secret semantics), marked sensitive in IaC state, and deliberately never exported in outputs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **User Workloads Secret** — a Kubernetes Secret in the Composer environment's user-workloads namespace, created through the Cloud Composer API against the referenced environment. It holds the configured base64-encoded entries; KubernetesPodOperator tasks and Airflow connections consume it by name, and its lifecycle follows the environment's.

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** — an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** — required when using Runner-based credential delivery.

### GCP Project

- **A Cloud Composer environment** (Composer 3, or Composer 2 with user workloads support) in the target project and region — reference a GcpCloudComposerEnvironment Infra Component via ValueFromRef.
- **Cloud Composer API** — already enabled by the environment this Secret is delivered into; the module enables nothing itself.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Composer User Workloads Secret**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the environment reference, and the Secret's data entries. Start from the **Airflow Connection** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsSecret
metadata:
  name: orders-db-connection
  org: acme-corp
  env: prod
spec:
  region: us-central1
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
  secretName: orders-db-connection
  data:
    # base64 of "postgresql://appuser:example-password@10.20.0.12:5432/orders"
    connection: cG9zdGdyZXNxbDovL2FwcHVzZXI6ZXhhbXBsZS1wYXNzd29yZEAxMC4yMC4wLjEyOjU0MzIvb3JkZXJz
```

```shell
planton apply -f secret.yaml
```

This creates one Kubernetes Secret named `orders-db-connection` in the environment's user-workloads namespace, readable by DAG tasks from that moment on. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the environment by reference so the deploy graph orders itself:

```yaml
spec:
  region: us-central1
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
```

The InfraPipeline provisions the Composer environment first, then creates the Secret inside it.

## Key Configuration

These are the most important decisions when configuring a user workloads Secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The attachment pair is the identity** — `environment` + `secretName` locate the object; both are immutable after creation, as are `region` and `projectId`. Changing any of them replaces the Secret. The name follows Kubernetes object-name rules: lowercase letters, numbers, and hyphens, up to 63 characters.

**Values are base64 at the source** — every `data` value must be the BASE64-ENCODED material (`echo -n 'value' | base64`); the API rejects raw strings at deploy. The entries are held as secrets in IaC state and the decoded material is never placed in outputs — but the base64 text does sit in the manifest, so treat the manifest file with the same care as the credential it carries.

**Rotation is the day-2 lever** — `data` updates in place: change an entry and redeploy, and new task runs pick up the rotated value without recreating the Secret or touching the environment.

**`deletionPolicy` decides what a destroy does** — the default `DELETE` removes the Kubernetes Secret from the environment, and DAGs consuming it start failing immediately. Set `PREVENT` on credentials live pipelines depend on so the Secret cannot ride along with a stack teardown, or `ABANDON` to drop it from management while leaving it in the cluster.

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
| `name` | Fully qualified Secret resource name | Composer API handles, audit |
| `secret_name` | The short Kubernetes object name | The name DAG code mounts and reads |

The Secret's data is deliberately never exported.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Airflow connection** — a database or warehouse connection URI DAGs read as a mounted secret. Start from the **Airflow Connection** preset.

**API credentials** — third-party vendor keys tasks read as environment variables. Start from the **API Credentials** preset.

## Works With

- [**GCP Cloud Composer Environment**](/infra-catalog/gcp-cloud-composer-environment) — the environment this Secret lives in
- [**GCP Cloud Composer User Workloads ConfigMap**](/infra-catalog/gcp-cloud-composer-user-workloads-config-map) — the plain-text twin for non-sensitive settings
- [**GCP Project**](/infra-catalog/gcp-project) — provides the GCP project the environment runs in
