# AzureMssqlFailoverGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureMssqlFailoverGroupSpec** defines an Azure SQL Failover Group: a
declarative disaster-recovery grouping that replicates a set of databases
from a primary logical server to one or more partner servers in other
regions and provides a single read-write (and optional read-only)
listener endpoint that follows the primary through a failover.

Applications connect to the failover group's listener
(`{group}.database.windows.net`) instead of a specific server, so a
planned or automatic failover redirects traffic to the new primary
without a connection-string change. The group is a first-class node: the
primary server, the partner servers, and the databases each have their
own lifecycle and are referenced here.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMssqlFailoverGroup
metadata:
  name: test-failover-group
spec:
  name: test-fog-planton
  serverId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Sql/servers/primary-server
  partnerServers:
    - serverId:
        value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg-dr/providers/Microsoft.Sql/servers/partner-server
  databaseIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Sql/servers/primary-server/databases/app-db
  readWriteEndpointFailoverPolicy:
    mode: AUTOMATIC
    graceMinutes: 60
  tags:
    purpose: hack-test
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.serverId` | `string \| valueFrom` | yes |  | AzureMssqlServer (`status.outputs.server_id`) |
| `spec.partnerServers` | `[]AzureMssqlFailoverGroupPartnerServer` | yes |  |  |
| `spec.partnerServers[].serverId` | `string \| valueFrom` | yes |  | AzureMssqlServer (`status.outputs.server_id`) |
| `spec.databaseIds` | `[]string \| valueFrom` |  |  | AzureMssqlDatabase (`status.outputs.database_id`) |
| `spec.readWriteEndpointFailoverPolicy` | `AzureMssqlFailoverGroupReadWritePolicy` | yes |  |  |
| `spec.readWriteEndpointFailoverPolicy.mode` | `enum` | yes |  |  |
| `spec.readWriteEndpointFailoverPolicy.graceMinutes` | `int32` |  |  |  |
| `spec.readonlyEndpointFailoverPolicyEnabled` | `bool` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.name

`string` · required

The name of the failover group. It becomes the listener DNS label
(`{name}.database.windows.net`), so it must be globally unique across
Azure SQL. 1-63 characters, lowercase letters, numbers, and hyphens;
cannot start or end with a hyphen. Fixed at creation.

- rule: Failover group names are 1-63 lowercase letters, numbers, and hyphens, and cannot start or end with a hyphen
- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.serverId

`string | valueFrom` · required

The primary logical server whose databases this group replicates.
Defaults to referencing an AzureMssqlServer's server_id output. Fixed
at creation -- the group is anchored to its primary.

- references: AzureMssqlServer (`status.outputs.server_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMssqlServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.partnerServers

`[]AzureMssqlFailoverGroupPartnerServer` · required

The partner (secondary) servers that receive the replicated databases,
each in a DIFFERENT region than the primary and each other. At least
one is required. In practice a group has a single partner; Azure allows
more for multi-region topologies.

- rule: {"repeated":{"minItems":"1"}}

### spec.partnerServers[].serverId

`string | valueFrom` · required

The partner logical server, in a different region than the primary.
Defaults to referencing an AzureMssqlServer's server_id output. The
group is created ON its primary and replicates TO this server, so on a
diagram this reference is access, not placement.

- references: AzureMssqlServer (`status.outputs.server_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMssqlServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.databaseIds

`[]string | valueFrom`

The databases on the primary server to include in the group. Each is
replicated to every partner server. Databases must live on the primary
server. Empty is allowed (an empty group whose databases are added
later), but a DR group with no databases protects nothing. Each entry
defaults to referencing an AzureMssqlDatabase's database_id output.

- references: AzureMssqlDatabase (`status.outputs.database_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureMssqlDatabase, name: <that resource's name>, fieldPath: status.outputs.database_id}} -- a bare string does not parse

### spec.readWriteEndpointFailoverPolicy

`AzureMssqlFailoverGroupReadWritePolicy` · required

The read-write listener's failover policy -- how the primary role moves
to a partner when the primary is unavailable. Required: the group's
whole purpose is this policy.

- rule: {"required":true}
- rule: grace_minutes must be at least 60 for AUTOMATIC failover, and must be omitted for MANUAL failover

### spec.readWriteEndpointFailoverPolicy.mode

`enum` · required

How failover is triggered. AUTOMATIC fails the group over on its own
after the grace period when Azure detects an outage; MANUAL requires an
operator to initiate every failover.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `azure_mssql_failover_group_failover_mode_unspecified` -- Not specified -- invalid; a mode must be chosen.
- `AUTOMATIC` -- Azure fails the group over automatically after the grace period.
- `MANUAL` -- An operator initiates every failover.

### spec.readWriteEndpointFailoverPolicy.graceMinutes

`int32`

For AUTOMATIC mode: how many minutes Azure waits, after detecting an
outage, before failing over -- the window in which the outage might
recover on its own without data loss. Must be at least 60. Must be
omitted (0) for MANUAL mode.

- rule: {"int32":{"gte":0}}

### spec.readonlyEndpointFailoverPolicyEnabled

`bool` · optional (explicit presence)

Whether the read-only listener (`{name}.secondary.database.windows.net`)
also fails over with the primary. Unspecified leaves it DISABLED (the
provider sends Disabled when unset): after a failover the read-only
listener keeps pointing at the old primary until it recovers. Set true
for the read-only endpoint to follow the failover.

### spec.tags

`map<string, string>`

Free-form tags applied to the group, merged over the Planton-derived
resource tags (organization, environment, resource id); a user tag with
the same key wins. Updatable in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureMssqlFailoverGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.failover_group_id` | `string` | The Azure Resource Manager ID of the failover group. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Sql/servers/{server}/failoverGroups/{name} |
| `status.outputs.failover_group_name` | `string` | The failover group's name -- also the DNS label of its listener endpoints. |
| `status.outputs.read_write_listener_endpoint` | `string` | The read-write listener FQDN applications connect to; it always points at the current primary, so it survives a failover unchanged. Format: {name}.database.windows.net |
| `status.outputs.read_only_listener_endpoint` | `string` | The read-only listener FQDN for read-only workloads routed to a secondary. Format: {name}.secondary.database.windows.net |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.serverId` | AzureMssqlServer | `status.outputs.server_id` |
| `spec.partnerServers[].serverId` | AzureMssqlServer | `status.outputs.server_id` |
| `spec.databaseIds` | AzureMssqlDatabase | `status.outputs.database_id` |

## See Also

- [Overview](../README.md)
