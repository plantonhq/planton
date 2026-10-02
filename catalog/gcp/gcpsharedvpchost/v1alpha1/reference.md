# GcpSharedVpcHost

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSharedVpcHostSpec enables one project as a Shared VPC HOST: the
project whose VPC networks and subnetworks other projects (the SERVICE
projects, GcpSharedVpcServiceProject) attach to and place their workloads
in. Shared VPC is how an organization keeps one network team owning one
set of networks while many application teams deploy into them from their
own projects, with their own IAM and billing.

Enabling a host is a one-bit act on the project (Google's `enableXpnHost`)
with its own lifecycle, which is why it is a kind of its own rather than
a flag on GcpProject: the host role can be granted after the project
exists and revoked before it is deleted, and a chart places every
GcpSharedVpcServiceProject downstream of it by reference.

Destroy order matters: a host cannot be disabled while any service
project is still attached (Google refuses). A chart that declares the
attachments by reference destroys them first; a host disabled by hand
needs its attachments removed first.

The one permission this needs is organization-level: the deploying
identity holds `roles/compute.xpnAdmin` on the ORGANIZATION (or on the
folder above the project), never just on the project. Everything else
about Shared VPC -- which subnetworks a service project may use, and
which of its service accounts may use them -- is IAM on the host's
subnetworks (`roles/compute.networkUser`) and is declared with the
catalog's IAM kinds, not here.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcHost
metadata:
  name: planton-oss-e2e-gcpsvph
  id: planton-oss-e2e-gcpsvph
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcpsharedvpchost
  annotations:
    planton.dev/e2e: "true"
  tags:
    - planton-e2e
spec:
  # No project named: the provider's default project -- the harness project
  # exported as GOOGLE_PROJECT -- becomes the host. Enabling it needs
  # roles/compute.xpnAdmin on the ORGANIZATION, which the shared test identity
  # does not hold; the profile records the deferral. The reference and
  # literal forms are in e2e/scenarios/.

  # DELETE (default; refused while service projects are attached), PREVENT,
  # ABANDON.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project that becomes the Shared VPC host: a reference to a
GcpProject (its project_id output) or the project ID as a literal.
Empty means the provider's default project -- the project the
credentials are configured for -- so enabling the project you are
deploying into needs no configuration at all. Immutable: to host from
a different project, destroy this and declare a new one.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this resource does to the host role in GCP:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the project's host status is disabled; fails while any
               service project is still attached (destroy the
               GcpSharedVpcServiceProject resources first -- a chart's
               dependency order does this when they reference the host)
  "PREVENT" -- destroy FAILS; the guard for the host every service
               project in the organization depends on
  "ABANDON" -- the resource leaves management but the project stays a
               host with every attachment intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSharedVpcHost, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.host_project_id` | `string` | The project ID that is now the Shared VPC host -- the resolved value, so it is populated even when the spec left project_id empty and the provider's default project was used. What a GcpSharedVpcServiceProject's host_project_id references. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpSharedVpcServiceProject | `spec.hostProjectId` | `status.outputs.host_project_id` |

## See Also

- [Overview](../README.md)
