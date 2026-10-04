# GcpCloudComposerUserWorkloadsSecret

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudComposerUserWorkloadsSecretSpec defines a Kubernetes Secret
delivered into a Cloud Composer environment's workloads namespace.

User workloads Secrets are how Airflow DAGs receive credentials —
database passwords, API tokens, Airflow connection URIs — without
baking them into DAG code or environment variables. Composer manages
the underlying Kubernetes Secret in the environment's GKE cluster;
KubernetesPodOperator tasks and Airflow connections consume it by name.

The Secret's data updates in place; the name, environment, region,
and project are immutable after creation. Deleting this resource
deletes the Kubernetes Secret from the environment.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsSecret
metadata:
  name: test-airflow-secret
spec:
  # GCP project of the Composer environment. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # Region of the Composer environment.
  region: us-central1

  # The Composer environment the Secret is delivered into — its
  # environment_name. Replace with your environment.
  environment:
    value: test-composer

  # Kubernetes Secret name — what DAGs reference.
  secretName: test-db-credentials

  # Values MUST be base64-encoded (echo -n 'value' | base64).
  data:
    # base64 of "postgresql://airflow:pass@10.0.0.5/mydb"
    connection: cG9zdGdyZXNxbDovL2FpcmZsb3c6cGFzc0AxMC4wLjAuNS9teWRi
    # base64 of "sk-live-4f9a2b7c8d1e"
    api-token: c2stbGl2ZS00ZjlhMmI3YzhkMWU=

  # Explicit DELETE (the provider default) proves the round-trip.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.environment` | `string \| valueFrom` | yes |  | GcpCloudComposerEnvironment (`status.outputs.environment_name`) |
| `spec.secretName` | `string` | yes |  |  |
| `spec.data` | `map<string, string>` (sensitive) | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project of the Composer environment.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region of the Composer environment (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.environment

`string | valueFrom` · required

The Composer environment the Secret is delivered into. Resolves to
the environment's name. Immutable after creation.

- references: GcpCloudComposerEnvironment (`status.outputs.environment_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudComposerEnvironment, name: <that resource's name>, fieldPath: status.outputs.environment_name}} -- a bare string does not parse

### spec.secretName

`string` · required

Name of the Kubernetes Secret. Must be lowercase letters, numbers,
and hyphens; start with a letter; end with a letter or number.
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.data

`map<string, string>` · required · sensitive

The Secret's key-value entries. Values MUST be base64-encoded
(Kubernetes Secret semantics — e.g. `echo -n 'postgresql://...' |
base64`); the API rejects raw values. The decoded material (Airflow
connection URIs, passwords, tokens) is never placed in stack
outputs, and the entries are held as secrets in IaC state.

- rule: {"map":{"minPairs":"1","values":{"cel":[{"id":"data_value_base64","message":"each data value must be base64-encoded (e.g. echo -n 'value' | base64)","expression":"this.matches('^[A-Za-z0-9+/]+={0,2}$')"}]}}}

### spec.deletionPolicy

`string`

Deletion policy for the Secret — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the Kubernetes Secret is removed from the
               environment; DAGs consuming it start failing
  "PREVENT" -- destroy FAILS; protects credentials live pipelines
               depend on from riding along with a stack teardown
  "ABANDON" -- the Secret is removed from management but stays in
               the environment's cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudComposerUserWorkloadsSecret, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Fully qualified resource name. Format: projects/{project}/locations/{region}/environments/{environment}/userWorkloadsSecrets/{name} |
| `status.outputs.secret_name` | `string` | The Kubernetes Secret name (same as the spec's secret_name input) — what KubernetesPodOperator tasks and Airflow connections reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.environment` | GcpCloudComposerEnvironment | `status.outputs.environment_name` |

## See Also

- [Overview](../README.md)
