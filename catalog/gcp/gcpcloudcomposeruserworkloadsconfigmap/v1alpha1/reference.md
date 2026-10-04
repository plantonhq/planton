# GcpCloudComposerUserWorkloadsConfigMap

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudComposerUserWorkloadsConfigMapSpec defines a Kubernetes
ConfigMap delivered into a Cloud Composer environment's workloads
namespace.

User workloads ConfigMaps carry non-secret configuration for Airflow
DAGs and KubernetesPodOperator tasks — feature flags, endpoints,
tuning parameters — without baking values into DAG code. Composer
manages the underlying Kubernetes ConfigMap in the environment's GKE
cluster; workloads consume it by name. For credentials and other
secret material use GcpCloudComposerUserWorkloadsSecret instead.

The ConfigMap's data updates in place; the name, environment, region,
and project are immutable after creation. Deleting this resource
deletes the Kubernetes ConfigMap from the environment.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudComposerUserWorkloadsConfigMap
metadata:
  name: test-dag-config
spec:
  # GCP project of the Composer environment. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  # Region of the Composer environment.
  region: us-central1

  # The Composer environment the ConfigMap is delivered into — its
  # environment_name. Replace with your environment.
  environment:
    value: test-composer

  # Kubernetes ConfigMap name — what DAGs reference.
  configMapName: test-dag-config

  # Plain configuration data. Values are strings: quote numerals and
  # true/false so they stay strings (on/off/yes/no need no quoting -
  # manifests parse under YAML 1.2).
  data:
    api_endpoint: https://api.example.com/v2
    batch_size: "500"
    enable_new_flow: "true"

  # Explicit DELETE (the provider default) proves the round-trip.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.environment` | `string \| valueFrom` | yes |  | GcpCloudComposerEnvironment (`status.outputs.environment_name`) |
| `spec.configMapName` | `string` | yes |  |  |
| `spec.data` | `map<string, string>` | yes |  |  |
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

The Composer environment the ConfigMap is delivered into. Resolves
to the environment's name. Immutable after creation.

- references: GcpCloudComposerEnvironment (`status.outputs.environment_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudComposerEnvironment, name: <that resource's name>, fieldPath: status.outputs.environment_name}} -- a bare string does not parse

### spec.configMapName

`string` · required

Name of the Kubernetes ConfigMap. Must be lowercase letters,
numbers, and hyphens; start with a letter; end with a letter or
number. Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.data

`map<string, string>` · required

The ConfigMap's key-value entries (plain configuration data).

- rule: {"map":{"minPairs":"1"}}

### spec.deletionPolicy

`string`

Deletion policy for the ConfigMap — what happens when this resource
is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the Kubernetes ConfigMap is removed from the
               environment; DAGs consuming it start failing
  "PREVENT" -- destroy FAILS; protects configuration live pipelines
               depend on from riding along with a stack teardown
  "ABANDON" -- the ConfigMap is removed from management but stays
               in the environment's cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudComposerUserWorkloadsConfigMap, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Fully qualified resource name. Format: projects/{project}/locations/{region}/environments/{environment}/userWorkloadsConfigMaps/{name} |
| `status.outputs.config_map_name` | `string` | The Kubernetes ConfigMap name (same as the spec's config_map_name input) — what KubernetesPodOperator tasks reference. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.environment` | GcpCloudComposerEnvironment | `status.outputs.environment_name` |

## See Also

- [Overview](../README.md)
