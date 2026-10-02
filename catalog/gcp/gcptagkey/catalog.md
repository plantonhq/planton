# GCP Tag Key

Creates a Google Cloud Resource Manager tag key: the NAME half of a tag such as `environment` or `data-classification`. A key owns a set of values (`GcpTagValue`: `prod`, `staging`), and a value is bound to a resource (`GcpTagBinding`) so organization policies, IAM conditions, and firewall policies can key on it. Tags are governed metadata -- created centrally, permissioned, immutable in name, evaluated by the platform's policy engines -- where labels are free-form and enforce nothing. A key is owned by the organization (the landing-zone default) or by one project; never by a folder.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Tag key** -- the `tags_tag_key` under the organization or project, with its short name, description, optional purpose, and optional allowed-values regex

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials that can administer tags at the owner. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Owner

- **IAM**: the deploying identity needs `roles/resourcemanager.tagAdmin` on the owner -- on the project for a project-owned key (the shape that needs no organization-level grant), on the organization otherwise.
- **Short names are unique per owner** and reserved for 30 days after deletion.
- **A folder cannot own a key** -- Google's rule; choose the organization or a project.

## Deploy

### Console

Open the deployment store, find **GCP Tag Key**, and click **Deploy**. The creation wizard walks you through the owner, the short name, and the optional purpose. Start from the **Environment Key (Organization)** preset in the [Presets](#presets) tab for the common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTagKey
metadata:
  name: environment
  org: acme-corp
  env: prod
spec:
  parent:
    organizationId: "123456789012"
  shortName: environment
  description: The environment a resource belongs to
```

```shell
planton apply -f tag-key.yaml
```

This creates the organization-wide key `environment`; its values are declared with `GcpTagValue` and bound with `GcpTagBinding`. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, a project-owned key references its `GcpProject` via ValueFromRef, and every `GcpTagValue` references the key's `name`:

```yaml
spec:
  parent:
    projectId:
      valueFrom:
        kind: GcpProject
        name: app-project
        fieldPath: status.outputs.project_id
  shortName: cost-center
```

The InfraPipeline deploys the project first, then the key, then every `GcpTagValue` that points at `status.outputs.name`.

## Key Configuration

These are the most important decisions when configuring a tag key. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Owner** -- exactly one of `organizationId` (values bindable anywhere in the organization) or `projectId` (values bindable only inside that project). Immutable.

**Short name** -- the name written in tag conditions (`resource.matchTag('{org}/{shortName}', ...)`); defaults to the manifest name. Any UTF-8 except `/`, `\`, `'`, `"`. Immutable: rename by creating a new key.

**Purpose** -- `GCE_FIREWALL` makes the key's values usable as targets and sources in network firewall policy rules (secure tags) and requires `purposeData.network`; `DATA_GOVERNANCE` classifies data for Sensitive Data Protection and BigQuery. Immutable once set.

**Allowed-values regex** -- an RE2 pattern every value's short name must match; setting it also makes the key DYNAMIC, so bindings may carry undeclared values that match (ticket numbers, team codes). Mutable.

**Deletion policy** -- `DELETE` (default) fails while any value still exists; `PREVENT` fails the destroy; `ABANDON` leaves the key and its values live.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `parent.projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | `tagKeys/{id}` | A `GcpTagValue`'s `tagKey`; a `GcpFolder`'s or `GcpProject`'s create-time `tags` key |
| `namespaced_name` | `{org_or_project}/{shortName}` | `resource.matchTag(...)` conditions in organization policies and IAM |
| `tag_key_id` | The bare numeric ID | Tooling |
| `create_time` | RFC 3339 creation timestamp | Auditing |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Environment key (organization)** -- the `environment` key every landing zone starts with. Start from the **Environment Key (Organization)** preset.

**Project-scoped key** -- a key one application manages for itself, inside its project. Start from the **Project-Scoped Key** preset.

**Firewall-purpose key** -- a `GCE_FIREWALL` key whose values are secure tags for network firewall policies. Start from the **Firewall Purpose Key** preset.

## Works With

- [**GCP Tag Value**](/cloud-catalog/gcp-tag-value) -- the values declared under this key
- [**GCP Tag Binding**](/cloud-catalog/gcp-tag-binding) -- attaches a value to a resource
- [**GCP Organization Policy**](/cloud-catalog/gcp-org-policy) -- rules conditioned on the tag
- [**GCP Project**](/cloud-catalog/gcp-project) -- the owner of a project-scoped key
