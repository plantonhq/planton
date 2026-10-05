# GcpAlloydbUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpAlloydbUserSpec defines a database user (`google_alloydb_user`) on an
AlloyDB cluster.

Users are first-class nodes: one per application with its own credential
(ALLOYDB_BUILT_IN) or passwordless IAM authentication (ALLOYDB_IAM_USER).

## Example

```yaml
# Exercises a BUILT_IN application user offline, referencing its cluster by
# literal resource path.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbUser
metadata:
  name: hack-orders-app
spec:
  cluster:
    value: projects/my-project/locations/us-central1/clusters/hack-orders
  userId: orders-app
  password: HackAppPassword123!  # replace before applying anywhere real
  userType: ALLOYDB_BUILT_IN
  databaseRoles:
    - alloydbiamuser
  # Drop the user on destroy (GCP's default, made explicit). PREVENT guards
  # a credential apps still use; ABANDON leaves the user on the cluster.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.cluster` | `string \| valueFrom` | yes |  | GcpAlloydbCluster (`status.outputs.cluster_id`) |
| `spec.userId` | `string` | yes |  |  |
| `spec.userType` | `string` |  | `ALLOYDB_BUILT_IN` |  |
| `spec.password` | `string` (sensitive) | yes |  |  |
| `spec.databaseRoles` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the AlloyDB cluster.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.cluster

`string | valueFrom` · required

The AlloyDB cluster this user lives on. Accepts the full cluster resource
path or a reference to a GcpAlloydbCluster resource. Immutable.

- references: GcpAlloydbCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAlloydbCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.userId

`string` · required

The database role name of the user. Immutable.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.userType

`string` · optional (explicit presence)

User authentication type. ALLOYDB_BUILT_IN (default) uses a password;
ALLOYDB_IAM_USER authenticates through IAM without a stored password.
Immutable.

- default: `ALLOYDB_BUILT_IN`
- rule: user_type must be ALLOYDB_BUILT_IN or ALLOYDB_IAM_USER

### spec.password

`string` · required · sensitive

Password for ALLOYDB_BUILT_IN users. Mutable — updating rotates in place.
Never set for ALLOYDB_IAM_USER.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"1"}}

### spec.databaseRoles

`[]string`

Database roles granted to this user (e.g. "alloydbiamuser", "alloydbsuperuser").

### spec.deletionPolicy

`string`

What happens to the database user in GCP when this resource is destroyed.
  "DELETE"  -- (GCP's default when unset) the user is dropped from the
               cluster; objects it owns inside PostgreSQL keep their
               ownership rows, so reassign ownership first
  "PREVENT" -- destroy FAILS; protects a credential applications still
               authenticate with
  "ABANDON" -- the user is removed from management but keeps existing
               (and authenticating) on the cluster

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `iam_user_must_not_set_password`: ALLOYDB_IAM_USER must not set a password — authentication goes through IAM

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpAlloydbUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Fully qualified user resource name. Format: projects/{project}/locations/{location}/clusters/{cluster}/users/{user} |
| `status.outputs.user_id` | `string` | The user_id as stored by AlloyDB. |
| `status.outputs.cluster_id` | `string` | Fully qualified cluster resource name this user belongs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cluster` | GcpAlloydbCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
