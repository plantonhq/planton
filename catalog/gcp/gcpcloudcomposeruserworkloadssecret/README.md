# GCP Cloud Composer User Workloads Secret

Deploys a Kubernetes Secret (`google_composer_user_workloads_secret`) into a Cloud Composer environment's workloads namespace — how Airflow DAGs receive credentials without baking them into DAG code or environment variables.

## What Gets Created

When you deploy a GcpCloudComposerUserWorkloadsSecret resource, Planton provisions:

- **User Workloads Secret** — a Kubernetes Secret that Composer manages in the environment's GKE cluster; `KubernetesPodOperator` tasks and Airflow connections consume it by name

No API enablement is needed: the Composer API is enabled by the environment the Secret is delivered into.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing Composer environment** (deploy via GcpCloudComposerEnvironment first) — referenced via `environment`
- **IAM permissions** — [`iac/permissions.yaml`](iac/permissions.yaml) lists the exact least-privilege permissions

## Quick Start

Create a file `secret.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsSecret
metadata:
  name: orders-db-connection
spec:
  region: us-central1
  environment:
    valueFrom:
      kind: GcpCloudComposerEnvironment
      name: prod-airflow
      fieldPath: status.outputs.environment_name
  secretName: orders-db-connection
  data:
    # base64 of "postgresql://appuser:s3cr3t@10.20.0.12:5432/orders"
    connection: cG9zdGdyZXNxbDovL2FwcHVzZXI6czNjcjN0QDEwLjIwLjAuMTI6NTQzMi9vcmRlcnM=
```

Deploy:

```shell
planton apply -f secret.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Region of the Composer environment (e.g., `us-central1`). Immutable. | Required, region pattern |
| `environment` | `StringValueOrRef` | The Composer environment the Secret is delivered into. References a GcpCloudComposerEnvironment's `environment_name` output. Immutable. | Required |
| `secretName` | `string` | Kubernetes Secret name — what DAGs reference. Immutable. | Required, lowercase letters/numbers/hyphens, starts with a letter, ends with a letter or number |
| `data` | `map<string,string>` | Key-value entries. Values MUST be base64-encoded (`echo -n 'value' | base64`); raw values are rejected. Mutable. | Min 1 entry, each value base64 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default | GCP project of the Composer environment. |
| `deletionPolicy` | `string` | `DELETE` | What a destroy does: `DELETE` the Secret, `PREVENT` (fail — protects credentials live pipelines depend on), or `ABANDON` (keep it in the cluster, drop from management). |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | Fully qualified resource name (`projects/{p}/locations/{r}/environments/{e}/userWorkloadsSecrets/{n}`) |
| `secret_name` | `string` | The Kubernetes Secret name — what DAGs reference |

The Secret's data is deliberately never exported.

## Important Notes

- **Values are base64-encoded** — Kubernetes Secret semantics. Encode with `echo -n 'value' | base64`; the API rejects raw strings.
- **The material stays out of outputs**: decoded values never appear in outputs; the entries are held as secrets in IaC state (Terraform marks the attribute sensitive; Pulumi wraps the map with `ToSecret`).
- **Data updates in place**; `secretName`, `environment`, `region`, and `projectId` are immutable.
- **How DAGs consume it**: mount it into `KubernetesPodOperator` tasks (as env vars or files) or point an Airflow connection/secret backend at it by `secret_name`.
- **Deleting this resource deletes the Kubernetes Secret** from the environment.

## Related Kinds

- [GcpCloudComposerEnvironment](/docs/catalog/gcp/gcpcloudcomposerenvironment) — the environment the Secret is delivered into
- [GcpCloudComposerUserWorkloadsConfigMap](/docs/catalog/gcp/gcpcloudcomposeruserworkloadsconfigmap) — the non-secret sibling for plain configuration
- [GcpProject](/docs/catalog/gcp/gcpproject) — the project the environment lives in

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

For copy-paste ready manifests, see the [presets](presets/).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
