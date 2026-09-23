# GCP Vertex AI Search Data Connector

Connects a source -- Jira, Confluence, ServiceNow, SharePoint, OneDrive, Outlook, Salesforce, Slack, BigQuery, Google Drive, and more -- to Vertex AI Search and Gemini Enterprise, syncing it into data stores on a schedule or searching it live. Name the source, list the entities to bring in, point at the Secret Manager secrets that hold the credentials, and an engine can search the whole collection. Credentials never sit in the manifest.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `discoveryengine.googleapis.com` on the project
- **Data connector** -- a `discoveryengine.DataConnector`, which creates the collection and one data store per entity

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Discovery Engine admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Prerequisites Outside This Block

- **Secret Manager secrets** holding the source's credentials, readable by the Discovery Engine service agent.
- **Gemini Enterprise licenses** on the project for third-party sources.

### Optional Dependencies

- **GcpKmsKey** -- for customer-managed encryption of every store the connector creates, referenced by `kmsKeyName`.

## Deploy

### Console

Open the deployment store, find **GCP Vertex AI Search Data Connector**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Jira Federated with Actions** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiSearchDataConnector
metadata:
  name: jira-federated
  org: acme-corp
  env: prod
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

This creates a Jira collection with a data store for projects and one for issues, synced daily. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, reference the connector's `collection_id` output from a `GcpVertexAiSearchEngine`'s `collectionId`, and a `GcpKmsKey` from `kmsKeyName` when encryption must be customer-managed.

## Key Configuration

These are the most important decisions when configuring a connector. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Index or federate** -- `connectorModes` decides whether the source is indexed into the stores (`DATA_INGESTION`, fast ranked search) or searched live (`FEDERATED`, nothing copied); `ACTIONS` lets an assistant act on the source through a BAP connection.

**Which entities** -- each `entities[]` entry becomes a data store; `params` filter what is ingested and `keyPropertyMappings` say which source fields render as a result's title and description.

**Where the credentials live** -- every secret in `params` is a Secret Manager resource name; the Discovery Engine service agent fetches it. Nothing sensitive is written into the manifest or the state.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `collection_id` | The collection's id | A `GcpVertexAiSearchEngine`'s `collectionId` |
| `entity_data_stores` | The created stores' full names | An engine's controls |
| `state` | The connector's state | Dashboards |
| `static_ip_addresses` | The egress addresses | The source's allowlist |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Jira federated with actions** -- Live search over Jira with an assistant that can open and update issues, on static egress addresses. Start from the **Jira Federated with Actions** preset.

**ServiceNow periodic with CMEK** -- ServiceNow knowledge indexed daily in `us` under your own key. Start from the **ServiceNow Periodic with CMEK** preset.

## Works With

- [**GCP Vertex AI Search Engine**](/cloud-catalog/gcp-vertex-ai-search-engine) -- the app over the connector's stores
- [**GCP Vertex AI Search Data Store**](/cloud-catalog/gcp-vertex-ai-search-data-store) -- stores you fill yourself
- [**GCP Secret Manager Secret**](/cloud-catalog/gcp-secret-manager-secret) -- the source credentials
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- customer-managed encryption
