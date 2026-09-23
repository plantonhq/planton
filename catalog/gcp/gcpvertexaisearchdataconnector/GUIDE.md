# GcpVertexAiSearchDataConnector Guide

This guide explains how a Vertex AI Search data connector is shaped and the decisions behind the spec. For field-by-field detail see the README; for ready-made configurations see the presets.

The product has worn three names. The API is **Discovery Engine** (`discoveryengine.googleapis.com`), the search product is **Vertex AI Search**, and the console now groups it under **AI Applications**, with **Gemini Enterprise** as the assistant that third-party connectors serve. This kind is the connector in all three vocabularies.

## The connector is a collection

A `GcpVertexAiSearchDataStore` is a store you fill yourself, and it lives in Google's `default_collection`. A data connector is a different root: setting it up creates a **collection** of its own and one data store inside it per entity you list, and Google fills those stores from the source on a schedule. That is why the connector is its own kind rather than an arm of the data store -- the two share nothing but a location -- and why an engine over connector-synced stores names the connector's `collection_id` rather than `default_collection`.

## Sources, entities, and modes

`dataSource` names the source as Google does: first-party `bigquery`, `google_drive`, `google_mail`, `google_calendar`, `google_chat`, `gcp_fhir`; third-party `jira`, `confluence`, `servicenow`, `sharepoint`, `onedrive`, `outlook`, `salesforce`, `slack`, `notion`, `github`, `gitlab`, `zendesk`, `box`, `dropbox`, `workday`, and more. Google's connector documentation is the full list, so the spec validates the shape of the name, not a fixed set.

Each `entities[]` entry is one object type of the source (a Jira `issue`, a Confluence `Space`, a Salesforce `Account`) and becomes one data store; `params` holds the entity's inclusion filters as compact JSON and `keyPropertyMappings` says which source fields render as a result's title and description. Entity names are immutable; add or remove entities by editing the list.

`connectorModes` decides what the connector does: `DATA_INGESTION` indexes the source into the stores; `FEDERATED` searches the source live and indexes nothing; `ACTIONS` lets an assistant act on the source (create an issue, post a comment) through a Business Application Platform connection configured by `actionConfig` and `bapConfig`; `EUA` and `FEDERATED_AND_EUA` carry the searching user's own identity to the source.

## The schedule

`refreshInterval` is the full-sync cadence (a duration string such as `86400s`; 30 minutes to 7 days) and `incrementalRefreshInterval` the incremental one (third-party sources; Google defaults to three hours). Setting them equal disables incremental sync. `autoRunDisabled` and `incrementalSyncDisabled` pause either kind without deleting anything; `syncMode` is `PERIODIC` or `STREAMING`.

## Credentials never appear here

Google's connectors read their credentials from Secret Manager: the `params` values that carry a secret (`client_secret`, `password`, `refresh_token`, ...) are Secret Manager secret **resource names**, and the Discovery Engine service agent needs `secretmanager.secretAccessor` on each. The spec therefore holds no sensitive material, and neither does either engine's state. `params` is string pairs; `jsonParams` is the same as one JSON string for sources whose parameters nest -- exactly one of the two.

## Reaching the source

`staticIpEnabled` gives the connector fixed egress addresses (the `static_ip_addresses` output) for a source behind an allowlist; `destinationConfigs` route through private destinations, and `private_connectivity_project_id` names the tenant project the source must allowlist. Both are immutable in effect: the static-IP switch replaces the connector.

## Encryption and destroy

`kmsKeyName` encrypts every data store the connector creates under a Cloud KMS key in the collection's location; it is immutable. `deletionPolicy` is `DELETE` (default; the connector, the collection, and every created store go, indexed data included), `PREVENT` (destroy fails), or `ABANDON` (everything leaves management and keeps syncing).
