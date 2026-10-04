# GcpCloudBuildRepository

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudBuildRepositorySpec links one code-host repository into Cloud
Build through an existing connection (`google_cloudbuildv2_repository`).

The repository is what builds consume: a GcpCloudBuildTrigger's
repository_event_config, source_to_build, or git_file_source, and a
GcpDeployCustomTargetType's Skaffold modules, reference its name output.
A connection holds many repositories; each is its own block so the team
that owns a repository can link it without touching the connection.

The repository lives in its connection's project and region: both
modules take the project and location from parent_connection's full
name, so they can never disagree with it.

Important behavioral notes:

  - Every field is a create-time decision: changing any of them
    replaces the link (the code on the host is untouched).
  - The connection must have finished its installation (its
    installation_stage output is COMPLETE) before Google can link a
    repository through it.
  - Destroy removes the link from Cloud Build, never the repository on
    the code host.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildRepository
metadata:
  name: orders
spec:
  parentConnection:
    value: projects/my-gcp-project/locations/us-central1/connections/acme-github
  repositoryId: acme-orders
  remoteUri: https://github.com/acme/orders.git
  annotations:
    team: orders
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.parentConnection` | `string \| valueFrom` | yes |  | GcpCloudBuildConnection (`status.outputs.name`) |
| `spec.repositoryId` | `string` |  |  |  |
| `spec.remoteUri` | `string` | yes |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.parentConnection

`string | valueFrom` · required

The connection the repository is linked through, by full name
(projects/{project}/locations/{location}/connections/{connection}): a
GcpCloudBuildConnection reference (its name output) or the literal
name. Google's provider parses the project and location from this full
form, so a short name is refused. Required. Immutable.

- references: GcpCloudBuildConnection (`status.outputs.name`)
- rule: parent_connection must be a full connection name: projects/{project}/locations/{location}/connections/{connection}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildConnection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.repositoryId

`string`

The repository's ID in Cloud Build, unique in the connection: letters,
digits, and any of -._~%!$&'()*+,;=@ (Google's rule). Usually the
repository's name on the host. Defaults to metadata.name. Immutable.

- rule: repository_id may contain only letters, digits, and -._~%!$&'()*+,;=@

### spec.remoteUri

`string` · required

The repository's HTTPS clone URI on the code host, e.g.
"https://github.com/acme/orders.git". Required. Immutable.

- rule: remote_uri must be the repository's https:// clone URI
- rule: {"required":true}

### spec.annotations

`map<string, string>`

Annotations on the repository link (AIP-128 key/value metadata; a
repository has no labels). Immutable: a change replaces the link.

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the link is removed from Cloud Build
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the link leaves management and stays in Cloud Build

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudBuildRepository, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/connections/{connection}/repositories/{repository_id}. The value a trigger's repository fields and a custom target type's Cloud Build repository take. |
| `status.outputs.repository_id` | `string` | The repository's ID in Cloud Build. |
| `status.outputs.remote_uri` | `string` | The repository's HTTPS clone URI on the code host. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.parentConnection` | GcpCloudBuildConnection | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudBuildTrigger | `spec.repositoryEventConfig.repository` | `status.outputs.name` |
| GcpCloudBuildTrigger | `spec.sourceToBuild.repository` | `status.outputs.name` |
| GcpCloudBuildTrigger | `spec.gitFileSource.repository` | `status.outputs.name` |
| GcpDeployCustomTargetType | `spec.customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
