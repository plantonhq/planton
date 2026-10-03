# GCP Cloud Composer User Workloads ConfigMap

Deploys a Kubernetes ConfigMap (`google_composer_user_workloads_config_map`) into a Cloud Composer environment's workloads namespace — non-secret configuration for Airflow DAGs (feature flags, endpoints, tuning parameters) without baking values into DAG code.

## What Gets Created

When you deploy a GcpCloudComposerUserWorkloadsConfigMap resource, Planton provisions:

- **User Workloads ConfigMap** — a Kubernetes ConfigMap that Composer manages in the environment's GKE cluster; DAGs and `KubernetesPodOperator` tasks consume it by name

No API enablement is needed: the Composer API is enabled by the environment the ConfigMap is delivered into.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing Composer environment** (deploy via GcpCloudComposerEnvironment first) — referenced via `environment`
- **IAM permissions** — [`iac/permissions.yaml`](iac/permissions.yaml) lists the exact least-privilege permissions

## Quick Start

Create a file `configmap.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsConfigMap
metadata:
  name: orders-dag-config
spec:
  region: us-central1
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
  configMapName: orders-dag-config
  data:
    api_endpoint: https://api.example.com/v2
    batch_size: "500"
```

Deploy:

```shell
planton apply -f configmap.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Region of the Composer environment (e.g., `us-central1`). Immutable. | Required, region pattern |
| `environment` | `StringValueOrRef` | The Composer environment the ConfigMap is delivered into. References a GcpCloudComposerEnvironment's `environment_name` output. Immutable. | Required |
| `configMapName` | `string` | Kubernetes ConfigMap name — what DAGs reference. Immutable. | Required, lowercase letters/numbers/hyphens, starts with a letter, ends with a letter or number |
| `data` | `map<string,string>` | Plain key-value configuration entries. Mutable. | Min 1 entry |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project of the Composer environment. |
| `deletionPolicy` | `string` | `DELETE` | What a destroy does: `DELETE` the ConfigMap, `PREVENT` (fail — protects configuration live pipelines depend on), or `ABANDON` (keep it in the cluster, drop from management). |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Fully qualified resource name (`projects/{p}/locations/{r}/environments/{e}/userWorkloadsConfigMaps/{n}`) |
| `config_map_name` | `string` | The Kubernetes ConfigMap name — what DAGs reference |

## Important Notes

- **Plain data only**: entries are readable by anything with cluster access — for credentials, connection URIs, and tokens, use [GcpCloudComposerUserWorkloadsSecret](/docs/catalog/gcp/gcpcloudcomposeruserworkloadssecret) instead.
- **All values are strings**: quote numbers and booleans in YAML (`"500"`, `"true"`; YAML 1.1 also treats unquoted `on`/`off`/`yes`/`no` as booleans).
- **Data updates in place**; `configMapName`, `environment`, `region`, and `projectId` are immutable.
- **Deleting this resource deletes the Kubernetes ConfigMap** from the environment.

## Related Kinds

- [GcpCloudComposerEnvironment](/docs/catalog/gcp/gcpcloudcomposerenvironment) — the environment the ConfigMap is delivered into
- [GcpCloudComposerUserWorkloadsSecret](/docs/catalog/gcp/gcpcloudcomposeruserworkloadssecret) — the secret-bearing sibling for credentials
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project the environment lives in

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

For copy-paste ready manifests, see the [presets](presets/).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
