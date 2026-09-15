# DigitalOceanProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanProjectSpec models the full digitalocean_project resource
surface: the account-level organizational container that groups droplets,
load balancers, domains, buckets, and most other DigitalOcean resources.

Membership is carried here on the project itself (the resources list);
DigitalOcean's standalone partial-ownership membership resource is
deliberately not modeled -- one project object owns its full membership
list, which is also how the API reports it back.

Destroying a project never destroys what is inside it: DigitalOcean
requires the project to be empty, so both provisioners relocate every
member resource to the account's default project first and retry the
delete while those moves settle server-side. The account's default
project itself cannot be deleted.

## Example

```yaml
# Reference manifests for DigitalOceanProject -- protovalidate-valid,
# embedded as the reference page's Example block, and the documents the
# offline tofu plans render. Two documents: a bare project, and a fully
# populated one with membership by literal URN.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanProject
metadata:
  name: web-team
spec:
  projectName: web-team
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanProject
metadata:
  name: web-production
spec:
  projectName: web-production
  description: Production web workloads
  purpose: Web Application
  environment: production
  resources:
    # Literal URNs; use valueFrom with an explicit kind to reference the
    # producing resource's urn output instead.
    - value: do:droplet:123456
    - value: do:space:web-assets
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.purpose` | `string` |  | `Web Application` |  |
| `spec.environment` | `string` |  |  |  |
| `spec.isDefault` | `bool` |  |  |  |
| `spec.resources` | `[]string \| valueFrom` |  |  |  |

## Field Details

### spec.projectName

`string` · required

Human-friendly name of the project, shown in the DigitalOcean console.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"175"}}

### spec.description

`string`

(Optional) Free-form description of the project.

- rule: {"string":{"maxLen":"255"}}

### spec.purpose

`string`

(Optional) The purpose of the project. DigitalOcean recognizes a set of
standard purposes (for example "Web Application", "Website or blog",
"Service or API") and stores anything else prefixed as "Other: <text>";
the provider strips that prefix on read, so any free text round-trips
cleanly. A value that itself starts with "Other:" is rejected here:
DigitalOcean keeps exactly one prefix and re-capitalizes the rest
("Other: probe" is stored as "Other: Probe"), the provider then strips
the prefix, and the read-back ("Probe") can never equal what was
written -- a permanent diff no provisioner can converge.

- default: `Web Application`
- rule: must not start with "Other:" -- DigitalOcean adds that prefix itself for non-standard purposes
- rule: {"string":{"maxLen":"255"}}

### spec.environment

`string`

(Optional) The environment of the project's resources. DigitalOcean
accepts these three values case-insensitively and reports the value back
capitalized (for example "Production"); declare it lowercase here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["development","staging","production"]}}

### spec.isDefault

`bool`

(Optional) Make this project the account's default project. Semi-
supported by DigitalOcean's API: it takes effect, but the account can
have only one default, so out-of-band changes to the default project
show up as drift here, and a project marked default cannot be deleted.
Leave unset unless the account's default is genuinely meant to be
managed as code.

### spec.resources

`[]string | valueFrom`

(Optional) Uniform resource names (URNs) of the resources this project
contains. Every DigitalOcean resource exposes a URN of the form
"do:<type>:<id>" -- for example "do:droplet:123456",
"do:dbaas:6ec9c684-...", "do:space:my-bucket", "do:domain:example.com".
Use a literal URN, or reference the producing resource's urn stack
output with an explicit valueFrom.kind -- the list is polymorphic
across kinds (droplets, load balancers, buckets, domains, ...), so no
single default kind applies and each reference names its own.
A resource can belong to exactly one project: listing it here moves it
from wherever it was, and removing it from the list moves it back to
the account's default project (nothing is ever destroyed by membership
changes). When the list is left empty entirely, membership is not
managed and out-of-band assignments are left untouched.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.project_id` | `string` | UUID of the project (the API identity, and the import id). |
| `status.outputs.owner_uuid` | `string` | UUID of the account or team that owns the project. |
| `status.outputs.owner_id` | `string` | Numeric id of the account or team that owns the project. |
| `status.outputs.resource_urns` | `[]string` | URNs of the resources DigitalOcean reports as members of the project after apply (for example "do:droplet:12345"), sorted. When the spec's resources list is set this is the membership the project holds; when it is left empty membership is unmanaged and the list reflects whatever the account has assigned out of band. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDropletAutoscalePool | `spec.dropletTemplate.projectId` | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
