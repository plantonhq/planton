# GcpKmsAutokeyConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpKmsAutokeyConfigSpec switches Cloud KMS Autokey on for a folder or a
single project (`google_kms_autokey_config` / `google_kms_project_autokey_config`).

With Autokey on, a team that creates a bucket, disk, dataset, topic, or
another compatible resource asks for a customer-managed key through a
GcpKmsKeyHandle instead of designing key rings and grants itself: Autokey
creates the key ring (named "autokey"), an HSM key (AES-256, rotated
yearly) for that resource type and location, and grants the resource's
service agent encrypt and decrypt on it.

Two storage models (Google's terms):
  - Same-project storage (key_project_resolution_mode RESOURCE_PROJECT):
    keys live in the project of the resource they protect. Works on a
    folder or a project. The Cloud KMS API must be on in each project;
    the module enables it on a project-scoped config's project. Google
    creates the KMS service agent when it is first needed.
  - Dedicated-project storage (DEDICATED_KEY_PROJECT, folders only): keys
    for every project in the folder live in one key project, the
    separation-of-duties model. One-time setup outside this block: create
    the key project's Cloud KMS service agent and grant it
    roles/cloudkms.admin on the key project (see the GUIDE).

Google keeps exactly one Autokey configuration per folder and per
project, and a project's configuration overrides its folder's. Applying
this block takes over whatever configuration the scope already had.
Destroying it (deletion_policy DELETE, the default) clears the
configuration, which turns Autokey off for the scope; keys Autokey
already created stay and keep protecting their resources.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsAutokeyConfig
metadata:
  name: workloads-folder-autokey
spec:
  scope:
    folderId:
      value: "123456789012"
  keyProjectResolutionMode: DEDICATED_KEY_PROJECT
  keyProject:
    value: security-keys
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpKmsAutokeyConfigScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.keyProjectResolutionMode` | `string` |  |  |  |
| `spec.keyProject` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpKmsAutokeyConfigScope`

The folder or project Autokey is configured on. Omit for the
provider's default project.

- rule: set at most one of project_id or folder_id (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

Project configuration: a literal project ID or a GcpProject
reference. Overrides any folder configuration above the project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

Folder configuration: the folder's numeric ID -- a literal or a
GcpFolder reference. Every project beneath it inherits the
configuration unless the project sets its own.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.keyProjectResolutionMode

`string`

How Autokey picks the project a new key is created in:
  "RESOURCE_PROJECT"      -- same-project storage: the key lives beside
                             the resource it protects
  "DEDICATED_KEY_PROJECT" -- dedicated-project storage: every key for
                             the folder lives in key_project (folder
                             configurations only)
  "DISABLED"              -- Autokey is off for the scope, overriding a
                             parent folder that has it on
Empty sends nothing; on a folder with key_project set Google then
treats the folder as dedicated-project storage.

- rule: key_project_resolution_mode must be RESOURCE_PROJECT, DEDICATED_KEY_PROJECT, or DISABLED

### spec.keyProject

`string | valueFrom`

The dedicated key project for a folder configuration: a literal
project ID or a GcpProject reference. The module sends Google's
projects/{id} form and enables the Cloud KMS API there. Its Cloud KMS
service agent needs roles/cloudkms.admin on it before the first key
handle is requested.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this block does:
  "" / "DELETE" -- clears the configuration: Autokey is off for the
                   scope (existing keys stay)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the block leaves management and the configuration
                   stays in force

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `key_project_needs_folder_scope`: key_project is set only on a folder-scoped configuration (a project configuration keeps keys in the project itself)
- `dedicated_mode_needs_folder_scope`: DEDICATED_KEY_PROJECT is available on a folder-scoped configuration only
- `dedicated_mode_needs_key_project`: DEDICATED_KEY_PROJECT needs a key_project to hold the keys

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpKmsAutokeyConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: folders/{id}/autokeyConfig or projects/{id}/autokeyConfig. |
| `status.outputs.parent` | `string` | The scope Autokey is configured on: folders/{id} or projects/{id}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.keyProject` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
