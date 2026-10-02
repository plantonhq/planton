# GcpColabRuntime

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpColabRuntimeSpec defines a Colab Enterprise runtime
(`google_colab_runtime`) -- a notebook VM assigned to one user, built
from a GcpColabRuntimeTemplate. Declaring runtimes ahead of time gives a
team ready-to-use, correctly configured machines (a GPU box for a new
hire, a runtime per course participant) and lets the block start or stop
them on purpose: desired_state STOPPED keeps the disk and stops the
compute bill.

Immutable: location, runtime_id, runtime_user, runtime_template,
display_name, and description. Mutable in place: desired_state and
auto_upgrade.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntime
metadata:
  name: alice-notebooks
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  runtimeTemplate:
    valueFrom:
      kind: GcpColabRuntimeTemplate
      name: standard-runtime
      fieldPath: status.outputs.name
  runtimeUser: alice@example.com
  displayName: Alice's notebooks
  # Keep the disk, stop the compute bill until Alice needs it.
  desiredState: STOPPED
  autoUpgrade: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.runtimeId` | `string` |  |  |  |
| `spec.runtimeTemplate` | `string \| valueFrom` | yes |  | GcpColabRuntimeTemplate (`status.outputs.name`) |
| `spec.runtimeUser` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.desiredState` | `string` |  |  |  |
| `spec.autoUpgrade` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the runtime lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Colab Enterprise region, e.g. "us-central1" -- the template's
region. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.runtimeId

`string`

The runtime's id -- the last segment of its resource name. Lowercase
letters, digits, and hyphens. Defaults to metadata.name. Immutable.

- rule: runtime_id must be lowercase letters, digits, and hyphens, starting with a letter or digit, up to 128 characters

### spec.runtimeTemplate

`string | valueFrom` · required

The template the runtime is built from: a GcpColabRuntimeTemplate
reference or a literal
projects/{project}/locations/{location}/notebookRuntimeTemplates/{id}.
Required: Colab Enterprise assigns every runtime from a template.
Immutable.

- references: GcpColabRuntimeTemplate (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpColabRuntimeTemplate, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.runtimeUser

`string` · required

The email of the user the runtime belongs to -- only this user can
connect notebooks to it. Immutable.

- rule: {"required":true,"string":{"email":true}}

### spec.displayName

`string`

The name shown in Colab Enterprise -- up to 128 characters. Defaults to
metadata.name. Immutable.

- rule: {"string":{"maxLen":"128"}}

### spec.description

`string`

What the runtime is for. Immutable.

### spec.desiredState

`string`

Whether the runtime should be running:
  "RUNNING" -- started (the default); compute bills
  "STOPPED" -- stopped; the disk stays and bills, compute does not
The block starts or stops the runtime to match on every apply. Unset
leaves the runtime running.

- rule: desired_state must be RUNNING or STOPPED

### spec.autoUpgrade

`bool`

Upgrade the runtime to the latest Colab image whenever it is started
and Google reports it upgradable.

### spec.deletionPolicy

`string`

What happens to the runtime when this resource is destroyed:
  "" / "DELETE" -- the runtime and its disk are deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the runtime leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpColabRuntime, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/notebookRuntimes/{runtime_id}. |
| `status.outputs.runtime_id` | `string` | The runtime's id (the last segment of name). |
| `status.outputs.location` | `string` | The region the runtime lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.runtimeTemplate` | GcpColabRuntimeTemplate | `status.outputs.name` |

## See Also

- [Overview](../README.md)
