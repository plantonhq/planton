# GcpTagBinding

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTagBindingSpec attaches ONE tag value to ONE resource: it is the act
of tagging. The value (GcpTagValue, e.g. `environment/prod`) already
exists; the binding says "this project / this folder / this VM carries
it", and from that moment every organization policy conditioned on the
tag (`resource.matchTag`), every IAM condition, and every firewall
policy rule that targets it applies to the resource -- and, for
projects and folders, to everything beneath them, because tags are
inherited down the hierarchy.

A binding is REPLACED, never edited: every field is immutable, so a
change destroys the old binding and creates the new one. A resource can
carry at most one value per key -- binding `environment/staging` to a
project that already has `environment/prod` is rejected by Google, not
swapped; destroy the old binding first. Deleting a binding is instant
and has no soft-delete window.

Two Google resources sit behind this kind. Organizations, folders,
projects, and other GLOBAL resources use the global binding; a REGIONAL
or ZONAL resource (a Compute instance, a Cloud SQL instance, a GKE
cluster) must be bound through the location-scoped binding, which is why
`location` exists: set it and the module switches resource.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagBinding
metadata:
  name: project-environment-prod
spec:
  # The value to bind: a GcpTagValue reference (its name output,
  # tagValues/{id}), or that literal, or the namespaced form
  # {org}/{key}/{value}. Immutable.
  tagValue:
    value: tagValues/281476102962987

  # The tagged resource: at most one of projectId (a GcpProject reference,
  # or a literal ID or NUMBER), folderId, organizationId, resourceName (any
  # other taggable resource's full resource name). Omit the whole block to
  # tag the provider's default project. Immutable.
  parent:
    projectId:
      value: "123456789012"

  # For a regional or zonal resourceName only: its region or zone.
  # location: us-central1

  # DELETE (default; immediate, no recovery window), PREVENT, ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.tagValue` | `string \| valueFrom` | yes |  | GcpTagValue (`status.outputs.name`) |
| `spec.parent` | `GcpTagBindingParent` |  |  |  |
| `spec.parent.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_number`) |
| `spec.parent.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.parent.organizationId` | `string` |  |  |  |
| `spec.parent.resourceName` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.tagValue

`string | valueFrom` · required

The tag value to bind: a reference to a GcpTagValue resource (its
`name` output, `tagValues/{id}`) or, as a literal, that name or the
namespaced form `{org_id}/{key_short_name}/{value_short_name}`.
Immutable.

- references: GcpTagValue (`status.outputs.name`)
- rule: tag_value must be tagValues/{numeric_id} (a GcpTagValue's name output) or the namespaced form {parent}/{key}/{value}, or a reference to a GcpTagValue
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTagValue, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.parent

`GcpTagBindingParent`

The resource that carries the tag. At most one arm; all empty means
the provider's default project (the project the deploying credentials
are configured for) -- the common "tag the project I am deploying
into" needs no parent at all. Immutable.

- rule: set at most one of project_id, folder_id, organization_id, or resource_name (empty means the provider's default project)

### spec.parent.projectId

`string | valueFrom`

A project: a reference to a GcpProject resource (its project_number
output) or, as a literal, the project's NUMBER or ID. Google requires
the number in a binding's parent; when a literal is not numeric the
module looks the number up once at apply time (a read of the project,
which the deploying identity can already see).

- references: GcpProject (`status.outputs.project_number`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_number}} -- a bare string does not parse

### spec.parent.folderId

`string | valueFrom`

A folder: the folder's numeric ID -- a literal, or a reference to a
GcpFolder resource (its folder_id output). Every project and folder
beneath it inherits the tag.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.parent.organizationId

`string`

The organization: the numeric organization ID, without the
`organizations/` prefix. Everything in the estate inherits the tag.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.parent.resourceName

`string`

Any other taggable resource, by its full resource name -- the
`//{service}.googleapis.com/...` form Google's "full resource name"
convention defines. Common shapes:
  //compute.googleapis.com/projects/{project}/zones/{zone}/instances/{name}
    (set location to the zone)
  //compute.googleapis.com/projects/{project}/global/networks/{name}
  //sqladmin.googleapis.com/projects/{project}/instances/{name}
    (set location to the region)
  //container.googleapis.com/projects/{project}/locations/{location}/clusters/{name}
    (set location to the cluster's location)
  //storage.googleapis.com/projects/_/buckets/{name}
  //bigquery.googleapis.com/projects/{project}/datasets/{name}
Set `location` for every regional or zonal resource.

- rule: resource_name must be a full resource name beginning with //{service}.googleapis.com/ -- e.g. //compute.googleapis.com/projects/p/zones/us-central1-a/instances/vm1

### spec.location

`string`

For a REGIONAL or ZONAL resource named in parent.resource_name: its
region (`us-central1`) or zone (`us-central1-a`). Required for such
resources -- Google serves their bindings from a regional endpoint --
and must stay empty for organizations, folders, projects, and other
global resources. Immutable.

- rule: location must be a region such as us-central1 or a zone such as us-central1-a

### spec.deletionPolicy

`string`

What destroying this resource does to the binding in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the binding is removed; the resource no longer carries
               the tag (immediate, no recovery window)
  "PREVENT" -- destroy FAILS; the guard for a tag a policy depends on
  "ABANDON" -- the binding is removed from management but the resource
               keeps the tag

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `location_needs_resource_name`: location applies only to a regional or zonal resource named in parent.resource_name -- organizations, folders, and projects are global

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTagBinding, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | The binding's resource name, `tagBindings/{url-encoded full resource name}/{tagValues/id}` -- the handle Google's Tag Bindings API deletes and lists the binding by. |
| `status.outputs.parent` | `string` | The full resource name the tag is bound to, exactly as sent to Google (`//cloudresourcemanager.googleapis.com/projects/{number}` for a project) -- the resolved form of whichever parent arm the spec set, useful when the spec named a project by ID and the module resolved its number. |
| `status.outputs.tag_value` | `string` | The bound tag value's resource name, `tagValues/{numeric_id}`. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.tagValue` | GcpTagValue | `status.outputs.name` |
| `spec.parent.projectId` | GcpProject | `status.outputs.project_number` |
| `spec.parent.folderId` | GcpFolder | `status.outputs.folder_id` |

## See Also

- [Overview](../README.md)
