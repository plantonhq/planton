# GcpSharedVpcServiceProject

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSharedVpcServiceProjectSpec attaches one SERVICE project to a Shared
VPC HOST project (GcpSharedVpcHost), so workloads in the service project
can be placed in the host's subnetworks: the application team keeps its
own project, IAM, quotas, and bill, while the network team keeps the one
set of networks every team shares.

One resource per attachment. A host may have many service projects; a
project can be a service project of at most one host. Both projects are
IMMUTABLE here -- moving a service project to another host is a detach
and an attach.

Attaching is not enough by itself: the service project's deployers (and
the Google-managed service agents that create its VMs, GKE nodes, and
Cloud SQL instances) still need `roles/compute.networkUser` on the
host's subnetworks -- or on the whole host project -- to actually use
them. Grant that with the catalog's IAM kinds (GcpProjectIamMember on
the host project, or subnetwork-level IAM); it is deliberately not part
of this resource, because those grants have their own lifecycle and are
usually per subnetwork, per team.

Permissions: the deploying identity needs `roles/compute.xpnAdmin` on the
organization (or the folder above both projects) to attach, plus
`roles/resourcemanager.projectIamAdmin`-class rights to read the service
project.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcServiceProject
metadata:
  name: planton-oss-e2e-gcpsvps
  id: planton-oss-e2e-gcpsvps
  org: planton-oss
  env: e2e
  labels:
    managed-by: planton-e2e
    e2e-component: gcpsharedvpcserviceproject
  annotations:
    planton.dev/e2e: "true"
    # A second project under the same organization, arranged by the owner
    # and exported to the harness; the lane skips honestly when it is unset.
    planton.dev/e2e-required-env: PLANTON_E2E_GCP_SECOND_PROJECT_ID
  tags:
    - planton-e2e
spec:
  # The host: the prerequisite GcpSharedVpcHost (the harness project), by
  # reference -- the host KIND, which orders host-enable -> attach.
  hostProjectId:
    valueFrom:
      kind: GcpSharedVpcHost
      name: planton-oss-e2e-gcpsvph-prereq
      fieldPath: status.outputs.host_project_id

  # The service project: a second project the harness identity may attach.
  serviceProjectId:
    value: ${E2E_ENV:PLANTON_E2E_GCP_SECOND_PROJECT_ID}

  # Empty detaches on destroy; ABANDON leaves the attachment. No other value.
  deletionPolicy: ""
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.hostProjectId` | `string \| valueFrom` | yes |  | GcpSharedVpcHost (`status.outputs.host_project_id`) |
| `spec.serviceProjectId` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.hostProjectId

`string | valueFrom` · required

The Shared VPC host to attach to: a reference to a GcpSharedVpcHost
(its host_project_id output) or the host's project ID as a literal.
Referencing the HOST KIND, not the project, is what orders a chart
correctly -- the project is enabled as a host before anything attaches
to it. Immutable.

- references: GcpSharedVpcHost (`status.outputs.host_project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSharedVpcHost, name: <that resource's name>, fieldPath: status.outputs.host_project_id}} -- a bare string does not parse

### spec.serviceProjectId

`string | valueFrom` · required

The project being attached as a service project: a reference to a
GcpProject (its project_id output) or the project ID as a literal.
Required -- Google's API takes both projects explicitly and the
provider has no ambient default for this one. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this resource does to the attachment in GCP. This
resource's own provider argument accepts ONE value, not the usual
DELETE/PREVENT/ABANDON trio:
  ""        -- the attachment is removed: the service project is
               detached from the host (the provider default). Fails
               while any resource in the service project still uses a
               host subnetwork.
  "ABANDON" -- the resource leaves management but the project stays
               attached with every workload intact -- for a service
               project whose VMs must outlive the chart that attached
               it (a shared platform project handed to another team).

- rule: deletion_policy must be empty (detach on destroy) or ABANDON -- this resource accepts no other value

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSharedVpcServiceProject, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_project_id` | `string` | The project ID now attached as a service project. |
| `status.outputs.host_project_id` | `string` | The host project ID it is attached to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.hostProjectId` | GcpSharedVpcHost | `status.outputs.host_project_id` |
| `spec.serviceProjectId` | GcpProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
