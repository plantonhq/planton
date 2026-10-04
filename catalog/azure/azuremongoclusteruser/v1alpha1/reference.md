# AzureMongoClusterUser

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureMongoClusterUserSpec** grants a Microsoft Entra ID principal
access to an Azure Cosmos DB for MongoDB vCore cluster
(AzureMongoCluster). This is an ACCESS GRANT, not a password user:
the principal (a user, a service principal, or a managed identity's
service principal) authenticates to MongoDB with its Entra token
and receives the granted database roles. Native username/password
administration lives on the cluster itself
(administrator_username/administrator_password).

The target cluster must list "MicrosoftEntraID" in its
authentication_methods -- Azure rejects the grant at deploy time
otherwise. Azure pins the identity provider to "MicrosoftEntraID"
(the only value the service accepts today); both engines send it
explicitly.

EVERY field is create-only: the provider ships no update path --
changing anything replaces the grant (a harmless replace: dropping
and re-adding an access binding).

## Example

```yaml
# Example for docs and offline validation: a managed identity's
# service principal granted root on the admin database (cluster-wide
# access). References are literal values so the manifest validates
# standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMongoClusterUser
metadata:
  name: test-mongo-cluster-user
  id: test-mongo-cluster-user
  org: test-org
  env: test
spec:
  mongoClusterId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DocumentDB/mongoClusters/acme-orders-db
  objectId:
    value: 11111111-2222-3333-4444-555555555555
  principalType: servicePrincipal
  roles:
    - database: admin
      role: root
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.mongoClusterId` | `string \| valueFrom` | yes |  | AzureMongoCluster (`status.outputs.mongo_cluster_id`) |
| `spec.objectId` | `string \| valueFrom` | yes |  |  |
| `spec.principalType` | `string` | yes |  |  |
| `spec.roles` | `[]AzureMongoClusterUserRole` | yes |  |  |
| `spec.roles[].database` | `string` | yes |  |  |
| `spec.roles[].role` | `string` | yes |  |  |

## Field Details

### spec.mongoClusterId

`string | valueFrom` · required

The cluster the principal is granted access to, by ARM ID.
Reference an AzureMongoCluster output or pass a literal ID.

**ForceNew**: changing this destroys and recreates the grant.

- references: AzureMongoCluster (`status.outputs.mongo_cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMongoCluster, name: <that resource's name>, fieldPath: status.outputs.mongo_cluster_id}} -- a bare string does not parse

### spec.objectId

`string | valueFrom` · required

The Entra principal being granted access, by OBJECT ID (a UUID).
No single kind dominates -- reference an
AzureUserAssignedIdentity's principal id output with an explicit
valueFrom for workload identities, or pass a human user's or
service principal's object id as a literal.

**ForceNew**: changing this destroys and recreates the grant.

- rule: object_id must be the principal's object id -- a UUID like 00000000-0000-0000-0000-000000000000
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.principalType

`string` · required

What kind of Entra principal the object id names: "user" (a
person) or "servicePrincipal" (an application or a managed
identity). Managed identities are granted through their service
principal -- use "servicePrincipal" for them.

**ForceNew**: changing this destroys and recreates the grant.

- rule: {"required":true,"string":{"in":["user","servicePrincipal"]}}

### spec.roles

`[]AzureMongoClusterUserRole` · required

The database roles granted to the principal, at least one. Azure
currently accepts exactly one role name, "root" (full access),
scoped to a database -- grant it on "admin" for cluster-wide
access. Azure owns the role vocabulary and will widen it over
time.

**ForceNew**: changing the roles destroys and recreates the grant.

- rule: {"repeated":{"minItems":"1"}}

### spec.roles[].database

`string` · required

The database the role is scoped to, e.g. "admin".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.roles[].role

`string` · required

The role's name. Azure's Mongo vCore service accepts "root"
today (the provider rejects anything else).

- rule: {"required":true,"string":{"in":["root"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMongoClusterUser, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.mongo_cluster_user_id` | `string` | The grant's Azure Resource Manager ID ({cluster_id}/users/{object_id}). |
| `status.outputs.mongo_cluster_user_name` | `string` | The grant's ARM name -- the granted principal's object id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.mongoClusterId` | AzureMongoCluster | `status.outputs.mongo_cluster_id` |

## See Also

- [Overview](../README.md)
