# GcpVertexAiSearchDataConnector

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiSearchDataConnectorSpec defines a Vertex AI Search data
connector (`google_discovery_engine_data_connector`, the Discovery Engine
API behind the console's AI Applications / Gemini Enterprise): a
COLLECTION of data stores that Google syncs from a source -- Jira,
Confluence, ServiceNow, SharePoint, OneDrive, Outlook, Salesforce,
Slack, BigQuery, Google Drive, and the rest -- on a schedule. The
connector is the collection: setting it up creates the collection and
one data store per entity below. An engine searches those stores by
naming this collection (GcpVertexAiSearchEngine.collection_id) and the
created stores (the outputs' entity_data_stores).

Credentials never appear in this spec. Google's connectors read them
from Secret Manager: the params values that carry a secret (client
secrets, passwords, refresh tokens) are Secret Manager secret RESOURCE
NAMES, and the Discovery Engine service agent needs accessor rights on
those secrets.

Immutable: location, collection_id, collection_display_name,
data_source, kms_key_name, static_ip_enabled, and each entity's name.
Everything else updates in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataConnector
metadata:
  name: jira-federated
spec:
  projectId:
    value: my-gcp-project
  # Discovery Engine multi-regions: global, us, or eu. Engines over the
  # connector's stores must match.
  location: global
  collectionDisplayName: Jira Federated
  dataSource: jira
  dataSourceVersion: 3
  # Connection parameters: the secrets are Secret Manager resource names,
  # never the secret material. The Discovery Engine service agent needs
  # accessor rights on them.
  params:
    instance_uri: https://example.atlassian.net
    instance_id: projects/my-gcp-project/secrets/jira-instance-id
    client_id: projects/my-gcp-project/secrets/jira-client-id
    client_secret: projects/my-gcp-project/secrets/jira-client-secret
    refresh_token: projects/my-gcp-project/secrets/jira-refresh-token
    auth_type: OAUTH
  # Daily full syncs, incremental syncs every six hours.
  refreshInterval: 86400s
  incrementalRefreshInterval: 21600s
  syncMode: PERIODIC
  connectorModes:
    - DATA_INGESTION
  # One data store per entity; issues filtered to one project key.
  entities:
    - entityName: project
    - entityName: issue
      params: '{"inclusion_filters":{"projectKey":["OPS"]}}'
    - entityName: comment
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.collectionId` | `string` |  |  |  |
| `spec.collectionDisplayName` | `string` |  |  |  |
| `spec.dataSource` | `string` | yes |  |  |
| `spec.dataSourceVersion` | `int32` |  |  |  |
| `spec.params` | `map<string, string>` |  |  |  |
| `spec.jsonParams` | `string` |  |  |  |
| `spec.refreshInterval` | `string` | yes |  |  |
| `spec.incrementalRefreshInterval` | `string` |  |  |  |
| `spec.incrementalSyncDisabled` | `bool` |  |  |  |
| `spec.autoRunDisabled` | `bool` |  |  |  |
| `spec.syncMode` | `string` |  |  |  |
| `spec.connectorModes` | `[]string` |  |  |  |
| `spec.staticIpEnabled` | `bool` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.entities` | `[]GcpVertexAiSearchDataConnectorEntity` |  |  |  |
| `spec.entities[].entityName` | `string` | yes |  |  |
| `spec.entities[].params` | `string` |  |  |  |
| `spec.entities[].keyPropertyMappings` | `map<string, string>` |  |  |  |
| `spec.destinationConfigs` | `[]GcpVertexAiSearchDataConnectorDestinationConfig` |  |  |  |
| `spec.destinationConfigs[].key` | `string` |  |  |  |
| `spec.destinationConfigs[].destinations` | `[]GcpVertexAiSearchDataConnectorDestination` |  |  |  |
| `spec.destinationConfigs[].destinations[].host` | `string` |  |  |  |
| `spec.destinationConfigs[].destinations[].port` | `int32` |  |  |  |
| `spec.destinationConfigs[].params` | `string` |  |  |  |
| `spec.actionConfig` | `GcpVertexAiSearchDataConnectorActionConfig` |  |  |  |
| `spec.actionConfig.actionParams` | `map<string, string>` |  |  |  |
| `spec.actionConfig.createBapConnection` | `bool` |  |  |  |
| `spec.bapConfig` | `GcpVertexAiSearchDataConnectorBapConfig` |  |  |  |
| `spec.bapConfig.supportedConnectorModes` | `[]string` |  |  |  |
| `spec.bapConfig.enabledActions` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the collection lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

Where the collection lives: "global", "us", or "eu". Engines over its
stores must match. Immutable.

- rule: {"required":true,"string":{"in":["global","us","eu"]}}

### spec.collectionId

`string`

The collection's id -- 1-63 characters, RFC 1034 (lowercase letters,
digits, hyphens; starts with a letter). Defaults to metadata.name.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$"}}

### spec.collectionDisplayName

`string`

The collection's name in the console (up to 1024 characters). Defaults
to metadata.name. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.dataSource

`string` · required

The source, as Google names it: first-party "bigquery", "gcp_fhir",
"google_mail", "google_drive", "google_calendar", "google_chat";
third-party "jira", "confluence", "servicenow", "sharepoint",
"onedrive", "outlook", "salesforce", "slack", "notion", "github",
"gitlab", "zendesk", "box", "dropbox", "workday", and more (Google's
connector documentation is the full list). Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z][a-z0-9_]*$"}}

### spec.dataSourceVersion

`int32` · optional (explicit presence)

The source's API version where the connector supports several, e.g.
3 for Jira v3. Sent only when set.

- rule: {"int32":{"gte":1}}

### spec.params

`map<string, string>`

Source connection parameters as string pairs -- instance URI, auth
type, client id, and the Secret Manager resource names of the secrets
(client_secret, password, refresh_token). Exactly one of params or
json_params.

### spec.jsonParams

`string`

The same parameters as one compact JSON string, for sources whose
parameters nest. Exactly one of params or json_params.

### spec.refreshInterval

`string` · required

How often a full sync runs, as a duration string ("86400s" for daily);
30 minutes to 7 days. Equal to incremental_refresh_interval disables
incremental sync.

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.incrementalRefreshInterval

`string`

How often an incremental sync runs (third-party sources), as a
duration string; 30 minutes to 7 days. Empty lets Google default
(3 hours).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.incrementalSyncDisabled

`bool`

Pause incremental syncs.

### spec.autoRunDisabled

`bool`

Pause full syncs.

### spec.syncMode

`string`

PERIODIC (Google's default) or STREAMING. Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PERIODIC","STREAMING"]}}

### spec.connectorModes

`[]string`

What the connector does: DATA_INGESTION (index the source), FEDERATED
(search the source live without indexing), ACTIONS (act on the source),
EUA (end-user authentication), FEDERATED_AND_EUA. Empty lets Google
default.

- rule: {"repeated":{"items":{"string":{"in":["DATA_INGESTION","ACTIONS","FEDERATED","EUA","FEDERATED_AND_EUA"]}}}}

### spec.staticIpEnabled

`bool`

Give the connector static egress IP addresses (for sources behind an
allowlist); the addresses appear in the outputs. Immutable.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting every data store the
connector creates: a GcpKmsKey reference or a literal key path. Omit
for Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.entities

`[]GcpVertexAiSearchDataConnectorEntity`

The source entities to ingest; Google creates one data store per
entity. Add or remove an entity by editing the list; an entity's name
is immutable.

### spec.entities[].entityName

`string` · required

The entity's name as the source names it. Supported values depend on
the source: Jira -- project, issue, attachment, comment, worklog;
Confluence -- Content, Space; Salesforce -- Lead, Opportunity, Contact,
Account, Case, Contract, Campaign; ServiceNow -- catalog, incident,
knowledge_base. Immutable.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.entities[].params

`string`

Entity-specific ingestion parameters as a compact JSON string (Google
normalizes it), e.g. {"inclusion_filters":{"knowledgeBaseSysId":["123"]}}.

### spec.entities[].keyPropertyMappings

`map<string, string>`

Source field -> key property (title, description, ...) so results
render a title and description from the right fields.

### spec.destinationConfigs

`[]GcpVertexAiSearchDataConnectorDestinationConfig`

Where the connector reaches or serves data, keyed by destination.

### spec.destinationConfigs[].key

`string`

The configuration's key, e.g. "url".

### spec.destinationConfigs[].destinations

`[]GcpVertexAiSearchDataConnectorDestination`

The destinations for this key.

### spec.destinationConfigs[].destinations[].host

`string`

The destination host, e.g. "https://example.atlassian.net".

### spec.destinationConfigs[].destinations[].port

`int32` · optional (explicit presence)

The port the destination accepts.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.destinationConfigs[].params

`string`

Destination parameters as a compact JSON string, e.g.
{"destination_type":"private"}.

### spec.actionConfig

`GcpVertexAiSearchDataConnectorActionConfig`

The action side of an ACTIONS-mode connector.

### spec.actionConfig.actionParams

`map<string, string>`

Connection parameters for the action side, as string pairs. Credentials
are named as Secret Manager secret resource names, never pasted.

### spec.actionConfig.createBapConnection

`bool`

Create a Business Application Platform connection for the actions.

### spec.bapConfig

`GcpVertexAiSearchDataConnectorBapConfig`

Which actions an ACTIONS-mode connector exposes.

### spec.bapConfig.supportedConnectorModes

`[]string`

Connector modes the BAP connection supports; Google offers ACTIONS.

- rule: {"repeated":{"items":{"string":{"in":["ACTIONS"]}}}}

### spec.bapConfig.enabledActions

`[]string`

Actions enabled on the source, e.g. create_issue, update_issue,
change_issue_status, create_comment, update_comment, upload_attachment.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.deletionPolicy

`string`

What happens to the connector, its collection, and the data stores it
created when this resource is destroyed:
  "" / "DELETE" -- deleted, indexed data included
  "PREVENT"     -- destroy fails
  "ABANDON"     -- everything leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `connector.params_xor_json_params`: a connector's source parameters are exactly one of params (string pairs) or json_params (a JSON string)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiSearchDataConnector, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name of the connector: projects/{project}/locations/{location}/collections/{collection_id}/dataConnector. |
| `status.outputs.collection_id` | `string` | The collection's id -- what an engine's collection_id names to search the connector's stores. |
| `status.outputs.location` | `string` | The collection's location (global, us, or eu). |
| `status.outputs.state` | `string` | The connector's state: CREATING, ACTIVE, RUNNING, WARNING, FAILED, INITIALIZATION_FAILED, or UPDATING. |
| `status.outputs.entity_data_stores` | `[]string` | Full resource names of the data stores Google created, one per entity, in manifest order -- what an engine's data_store_ids (by their last segment) or a control's data_store names. |
| `status.outputs.static_ip_addresses` | `[]string` | The static egress IP addresses when static_ip_enabled; allowlist them at the source. |
| `status.outputs.private_connectivity_project_id` | `string` | The tenant project behind a private-connectivity connector, to be allowlisted at the source. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpVertexAiSearchEngine | `spec.collectionId` | `status.outputs.collection_id` |

## See Also

- [Overview](../README.md)
