# Jira Federated with Actions

## Use Case

Connect Jira Cloud to Gemini Enterprise so users search issues live (federated, nothing indexed) and the assistant can act -- create and update issues, change their status, comment -- through a Business Application Platform connection. Egress uses static addresses Jira can allowlist.

## When to Use

- Bringing Jira into an enterprise search or assistant app without copying its data
- Letting an assistant open and update tickets on a user's behalf
- Sources that need the connector's IP addresses allowlisted

## What This Creates

- A data connector in `global` for the `jira` source (API v3) that creates a collection with one data store per entity: projects, issues, comments, attachments
- `FEDERATED` and `ACTIONS` modes, daily refresh, a BAP connection with four enabled actions, static egress addresses

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `params` | OAuth to `example.atlassian.net` | Your instance and the Secret Manager secrets holding its OAuth client secret and refresh token. |
| `connectorModes` | federated + actions | `DATA_INGESTION` to index Jira into the stores instead of searching it live (then engines can search it like any store). |
| `bapConfig.enabledActions` | four | Add `update_comment` and `upload_attachment`, or trim to read-mostly actions. |
| `entities[].params` | none | Per-entity inclusion filters as compact JSON, e.g. a list of project keys. |
| `staticIpEnabled` | `true` | `false` when Jira needs no allowlist; the addresses appear in the `static_ip_addresses` output. |

The Discovery Engine service agent needs `secretmanager.secretAccessor` on every secret named here, and the project needs Gemini Enterprise licenses for third-party connectors. Engines search this collection by referencing the connector's `collection_id`.
