# GcpDeployCustomTargetType

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDeployCustomTargetTypeSpec declares a Cloud Deploy custom target type
(`google_clouddeploy_custom_target_type`): how Cloud Deploy renders and
deploys a release to a system it does not deploy to natively -- a
vendor API, an internal platform, Terraform, a non-Kubernetes runtime.

Cloud Deploy deploys natively to GKE, Cloud Run, and fleet clusters.
For anything else, a custom target type names the container or Skaffold
custom action that does the work; each Cloud Deploy target that deploys
that way points at the type (custom_target.custom_target_type on a
GcpDeployTarget). Many targets share one type, and a type has no link to
any pipeline, so it is its own block.

The work is defined one of two ways (at most one):

  - tasks: a deploy container (and optionally a render container) Cloud
    Deploy runs in Cloud Build -- the simple path, no Skaffold
    knowledge needed;
  - custom_actions: Skaffold custom actions by name, defined in the
    release's skaffold.yaml or in remote Skaffold modules this type
    includes (from Git, a Cloud Build repository, or Cloud Storage).

Important behavioral notes:

  - location and custom_target_type_id are create-time decisions;
    everything else updates in place and applies to the next rollout.
  - The containers run in the execution environment of each target
    that uses the type (its service account and worker pool), with
    Cloud Deploy's CLOUD_DEPLOY_* environment variables describing the
    release, target, and output location.
  - Delete the targets that use a type before the type; a chart orders
    them by their reference.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployCustomTargetType
metadata:
  name: vendor-deployer
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  customTargetTypeId: vendor-deployer
  description: Deploys releases through the vendor's API
  labels:
    team: release-eng
  annotations:
    owner: release-eng
  tasks:
    render:
      container:
        image: us-docker.pkg.dev/my-gcp-project/deploy/vendor-renderer:1.4
    deploy:
      container:
        image: us-docker.pkg.dev/my-gcp-project/deploy/vendor-deployer:1.4
        command:
          - /bin/deploy
        args:
          - --wait
        env:
          VENDOR_REGION: us
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.customTargetTypeId` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.customActions` | `GcpDeployCustomTargetTypeCustomActions` |  |  |  |
| `spec.customActions.deployAction` | `string` | yes |  |  |
| `spec.customActions.renderAction` | `string` |  |  |  |
| `spec.customActions.includeSkaffoldModules` | `[]GcpDeployCustomTargetTypeSkaffoldModule` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].configs` | `[]string` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].git` | `GcpDeployCustomTargetTypeGitSource` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].git.repo` | `string` | yes |  |  |
| `spec.customActions.includeSkaffoldModules[].git.path` | `string` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].git.ref` | `string` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo` | `GcpDeployCustomTargetTypeCloudBuildRepoSource` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository` | `string \| valueFrom` | yes |  | GcpCloudBuildRepository (`status.outputs.name`) |
| `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.path` | `string` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.ref` | `string` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudStorage` | `GcpDeployCustomTargetTypeCloudStorageSource` |  |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudStorage.source` | `string` | yes |  |  |
| `spec.customActions.includeSkaffoldModules[].googleCloudStorage.path` | `string` |  |  |  |
| `spec.tasks` | `GcpDeployCustomTargetTypeTasks` |  |  |  |
| `spec.tasks.deploy` | `GcpDeployCustomTargetTypeTask` | yes |  |  |
| `spec.tasks.deploy.container` | `GcpDeployCustomTargetTypeContainer` |  |  |  |
| `spec.tasks.deploy.container.image` | `string` | yes |  |  |
| `spec.tasks.deploy.container.command` | `[]string` |  |  |  |
| `spec.tasks.deploy.container.args` | `[]string` |  |  |  |
| `spec.tasks.deploy.container.env` | `map<string, string>` |  |  |  |
| `spec.tasks.render` | `GcpDeployCustomTargetTypeTask` |  |  |  |
| `spec.tasks.render.container` | `GcpDeployCustomTargetTypeContainer` |  |  |  |
| `spec.tasks.render.container.image` | `string` | yes |  |  |
| `spec.tasks.render.container.command` | `[]string` |  |  |  |
| `spec.tasks.render.container.args` | `[]string` |  |  |  |
| `spec.tasks.render.container.env` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the type lives in -- the project of the targets that use
it: a literal project ID or a GcpProject reference. Empty means the
provider's default project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the type lives in, e.g. "us-central1" -- the region of the
targets that use it. Required. Immutable.

- rule: {"required":true}

### spec.customTargetTypeId

`string`

The type's ID, unique in the project and location: 1-63 lowercase
letters, digits, and hyphens, starting with a letter and not ending
with a hyphen. Defaults to metadata.name. Immutable.

- rule: custom_target_type_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen

### spec.description

`string`

A description of the type, up to 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.labels

`map<string, string>`

Labels on the type. The platform attribution labels are added on top
and win on key conflicts.

### spec.annotations

`map<string, string>`

Annotations on the type (user metadata Cloud Deploy never reads).
Only the keys declared here are managed.

### spec.customActions

`GcpDeployCustomTargetTypeCustomActions`

Render and deploy through Skaffold custom actions. Mutually exclusive
with tasks.

### spec.customActions.deployAction

`string` · required

The name of the Skaffold custom action that deploys, as defined in
the release's skaffold.yaml or an included module. Required.

- rule: {"required":true}

### spec.customActions.renderAction

`string`

The name of the Skaffold custom action that renders. Empty means
Cloud Deploy renders with `skaffold render`.

### spec.customActions.includeSkaffoldModules

`[]GcpDeployCustomTargetTypeSkaffoldModule`

Remote Skaffold modules Cloud Deploy adds to the release's Skaffold
config, so the custom actions can live in a shared repository or
bucket instead of every application's skaffold.yaml.

- rule: a Skaffold module names exactly one source: git, google_cloud_build_repo, or google_cloud_storage

### spec.customActions.includeSkaffoldModules[].configs

`[]string`

The Skaffold config modules (the `metadata.name` of each config) to
use from the source. Empty uses every config in the file.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.customActions.includeSkaffoldModules[].git

`GcpDeployCustomTargetTypeGitSource`

A Git repository Cloud Deploy clones.

### spec.customActions.includeSkaffoldModules[].git.repo

`string` · required

The repository to clone, e.g. "https://github.com/acme/deploy-actions.git".
Required.

- rule: {"required":true}

### spec.customActions.includeSkaffoldModules[].git.path

`string`

The Skaffold file's path from the repository root. Empty means
skaffold.yaml at the root.

### spec.customActions.includeSkaffoldModules[].git.ref

`string`

The branch or tag to clone. Empty means the default branch.

### spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo

`GcpDeployCustomTargetTypeCloudBuildRepoSource`

A repository linked through a Cloud Build connection.

### spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository

`string | valueFrom` · required

The repository, by full name
(projects/{p}/locations/{l}/connections/{c}/repositories/{r}): a
GcpCloudBuildRepository reference or the literal name. Required.

- references: GcpCloudBuildRepository (`status.outputs.name`)
- rule: repository must be a full repository name: projects/{project}/locations/{location}/connections/{connection}/repositories/{repository}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildRepository, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.path

`string`

The Skaffold file's path from the repository root. Empty means
skaffold.yaml at the root.

### spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.ref

`string`

The branch or tag to clone. Empty means the default branch.

### spec.customActions.includeSkaffoldModules[].googleCloudStorage

`GcpDeployCustomTargetTypeCloudStorageSource`

A Cloud Storage location Cloud Deploy copies.

### spec.customActions.includeSkaffoldModules[].googleCloudStorage.source

`string` · required

The objects to copy, recursively: "gs://my-bucket/dir/configs/*"
copies everything under dir/configs. Required.

- rule: source must be a Cloud Storage path starting with gs://
- rule: {"required":true}

### spec.customActions.includeSkaffoldModules[].googleCloudStorage.path

`string`

The Skaffold file's path relative to source. Empty means skaffold.yaml.

### spec.tasks

`GcpDeployCustomTargetTypeTasks`

Render and deploy through containers Cloud Deploy runs. Mutually
exclusive with custom_actions.

### spec.tasks.deploy

`GcpDeployCustomTargetTypeTask` · required

The task that deploys. Required.

- rule: {"required":true}

### spec.tasks.deploy.container

`GcpDeployCustomTargetTypeContainer`

The container Cloud Deploy runs in Cloud Build for this task -- set it,
since it is the task's whole definition.

### spec.tasks.deploy.container.image

`string` · required

The container image, e.g.
"us-docker.pkg.dev/acme/deploy/vendor-deployer:1.4". Required.

- rule: {"required":true}

### spec.tasks.deploy.container.command

`[]string`

The entrypoint, replacing the image's own.

### spec.tasks.deploy.container.args

`[]string`

The arguments, replacing the image's default arguments.

### spec.tasks.deploy.container.env

`map<string, string>`

Environment variables set in the container, alongside the
CLOUD_DEPLOY_* variables Cloud Deploy sets.

### spec.tasks.render

`GcpDeployCustomTargetTypeTask`

The task that renders. Unset means Cloud Deploy's default rendering.

### spec.tasks.render.container

`GcpDeployCustomTargetTypeContainer`

The container Cloud Deploy runs in Cloud Build for this task -- set it,
since it is the task's whole definition.

### spec.tasks.render.container.image

`string` · required

The container image, e.g.
"us-docker.pkg.dev/acme/deploy/vendor-deployer:1.4". Required.

- rule: {"required":true}

### spec.tasks.render.container.command

`[]string`

The entrypoint, replacing the image's own.

### spec.tasks.render.container.args

`[]string`

The arguments, replacing the image's default arguments.

### spec.tasks.render.container.env

`map<string, string>`

Environment variables set in the container, alongside the
CLOUD_DEPLOY_* variables Cloud Deploy sets.

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the type is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the type leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.one_definition`: set at most one of custom_actions or tasks

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDeployCustomTargetType, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/customTargetTypes/{custom_target_type_id}. |
| `status.outputs.custom_target_type_id` | `string` | The custom target type's ID. |
| `status.outputs.uid` | `string` | Google's unique identifier for the custom target type. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository` | GcpCloudBuildRepository | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDeployTarget | `spec.customTarget.customTargetType` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
