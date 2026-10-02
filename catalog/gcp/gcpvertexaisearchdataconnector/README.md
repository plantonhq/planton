# GCP Vertex AI Search Data Connector

A Vertex AI Search data connector -- a collection of data stores Google syncs from a source (Jira, Confluence, ServiceNow, SharePoint, OneDrive, Outlook, Salesforce, Slack, BigQuery, Google Drive, and the rest) on a schedule -- on the Discovery Engine API (the console calls the product AI Applications / Gemini Enterprise). Setting the connector up creates the collection and one data store per entity you list; engines search those stores by naming the collection. Credentials never appear in the spec: the parameters carry Secret Manager resource names, and Google's service agent reads the secrets.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project (never disabled on destroy)
- **Data connector** -- a `discovery_engine_data_connector`, which creates the collection and one data store per `entities[]` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Prerequisites Outside This Block

- **Secret Manager secrets** holding the source credentials, with the Discovery Engine service agent granted `secretmanager.secretAccessor` on each.
- **Gemini Enterprise licenses** on the project for third-party connectors.

### Optional Dependencies

- **`GcpKmsKey`** -- a key in the collection's location encrypting every store the connector creates (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataConnector
metadata:
  name: jira-federated
spec:
  location: global
  dataSource: jira
  dataSourceVersion: 3
  params:
    instance_uri: https://example.atlassian.net
    client_id: projects/my-gcp-project/secrets/jira-client-id
    client_secret: projects/my-gcp-project/secrets/jira-client-secret
    refresh_token: projects/my-gcp-project/secrets/jira-refresh-token
    auth_type: OAUTH
  refreshInterval: 86400s
  entities:
    - entityName: project
    - entityName: issue
```

```shell
planton apply -f vertex-ai-search-data-connector.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | `global`, `us`, or `eu`; engines over the collection must match. Immutable. |
| `dataSource` | `string` | The source as Google names it (`jira`, `confluence`, `servicenow`, `sharepoint`, `bigquery`, `google_drive`, ...). Immutable. |
| `refreshInterval` | `string` | Full-sync interval as a duration (`86400s`), 30 minutes to 7 days. |
| `params` or `jsonParams` | `map` / `string` | Exactly one: the source's connection parameters as string pairs, or as one JSON string. Secrets are Secret Manager resource names. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `collectionId` | `string` | `metadata.name` | RFC 1034 id of the collection the connector creates. Immutable. |
| `collectionDisplayName` | `string` | `metadata.name` | Console name. Immutable. |
| `dataSourceVersion` | `int` | source default | The source's API version (3 for Jira v3). |
| `incrementalRefreshInterval` | `string` | 3 hours | Incremental-sync interval (third-party sources). |
| `incrementalSyncDisabled`, `autoRunDisabled` | `bool` | `false` | Pause incremental or full syncs. |
| `syncMode` | `string` | `PERIODIC` | `PERIODIC` or `STREAMING`. |
| `connectorModes` | `[]string` | Google's default | `DATA_INGESTION`, `FEDERATED`, `ACTIONS`, `EUA`, `FEDERATED_AND_EUA`. |
| `staticIpEnabled` | `bool` | `false` | Fixed egress addresses (exported) for the source to allowlist. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path. Immutable. |
| `entities[]` | `[]object` | none | `entityName` (immutable), `params` (JSON), `keyPropertyMappings` (field -> `title` / `description`). One data store per entity. |
| `destinationConfigs[]` | `[]object` | none | `key`, `destinations[] { host, port }`, `params` (JSON) -- private destinations. |
| `actionConfig`, `bapConfig` | `object` | none | The action side of an ACTIONS-mode connector: `actionParams`, `createBapConnection`; `supportedConnectorModes`, `enabledActions`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one of `params` or `jsonParams`.
- Durations are `NNNs` strings; modes and sync modes are Google's values; ports are 1-65535.
- Collection ids are RFC 1034; entity names are non-empty.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/collections/{collection}/dataConnector` |
| `collection_id` | `string` | The collection's id -- what an engine's `collectionId` names |
| `location` | `string` | The collection's location |
| `state` | `string` | `CREATING`, `ACTIVE`, `RUNNING`, `WARNING`, `FAILED`, ... |
| `entity_data_stores` | `[]string` | The data stores Google created, one per entity, in manifest order |
| `static_ip_addresses` | `[]string` | The egress addresses when `staticIpEnabled` |
| `private_connectivity_project_id` | `string` | The tenant project behind a private-connectivity connector |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **The connector is the collection.** It is a different root from a `GcpVertexAiSearchDataStore`: it creates its own collection and its own stores. Engines over those stores reference this kind's `collection_id`.
- **Credentials live in Secret Manager.** Parameter values that carry a secret are secret resource names; the Discovery Engine service agent needs accessor rights on them.
- **Immutable core.** The collection ids, the source, the location, the key, the static-IP switch, and each entity's name replace the connector when changed; the schedule, parameters, modes, and the action side update in place.
- **Third-party connectors need Gemini Enterprise licenses** on the project.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpVertexAiSearchEngine** -- the app that searches the connector's stores (`collectionId`)
- **GcpVertexAiSearchDataStore** -- a store you fill yourself, in `default_collection`
- **GcpSecretManagerSecret** -- the secrets the connector's parameters name
- **GcpKmsKey** -- customer-managed encryption for the created stores

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
