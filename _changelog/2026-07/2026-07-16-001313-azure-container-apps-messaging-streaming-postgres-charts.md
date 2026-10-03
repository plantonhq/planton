# Four Azure Infra Charts: Container Apps, Service Bus Messaging, Event Streaming, HA PostgreSQL — plus the Keyless Dapr/KEDA Seams

**Date**: July 16, 2026
**Type**: Feature
**Components**: Azure Provider, Infra Charts, API Definitions

## Summary

Four production-shaped Azure infra charts joined the catalog — `azure/container-apps-environment`, `azure/service-bus-messaging`, `azure/event-streaming-platform`, and `azure/ha-postgres` — bringing the Azure catalog to nine charts. Building the Container Apps chart surfaced two kind seams that could not be wired by reference, so two retrofits shipped with it: the Dapr component's metadata values and the KEDA scale-rule identity became `StringValueOrRef` fields, which is what lets the chart compose fully keyless messaging (no connection string anywhere in the deployment).

## Problem Statement / Motivation

The Azure chart catalog covered networking, AKS, observability, and the web/static-site edge, but not the serverless-container, messaging, streaming, or database environments teams reach for next. Each is tedious to assemble correctly by hand:

- Container Apps with Dapr pub/sub is usually wired with a Service Bus connection string in a secret — a SAS credential where Azure offers Entra auth — and Dapr's default entity management creates queues behind IaC's back.
- Correct Service Bus topology (DLQ posture, SQL-filtered fan-out, per-entity least-privilege credentials) is subtle enough that most estates learn it from their first incident.
- Streaming with schema governance and a durable archive path is a week of wiring; capture is usually configured with storage keys.
- The HA + VNet-injection + CMK + Entra-admin PostgreSQL checklist requires a vault, key, identity, and grant assembled in exactly the right order.

### Pain Points

- `AzureContainerAppEnvironmentDaprComponent.metadata[].value` was a plain string, so a Dapr component could not reference a managed identity's `client_id` — keyless auth required hand-copying a GUID.
- The KEDA scale-rule `identity_id` (on `AzureContainerAppCustomScaleRule` and `AzureContainerAppJobEventScaleRule`) was a plain string, so a scale rule could not track the identity resource it authenticates as.

## Solution / What's New

### Kind retrofits (the seam-gap discipline)

Both fields became foreign-key-capable, following the established update workflow:

- **Dapr metadata `value`** → bare `StringValueOrRef` (no default kind — metadata entries are kind-type specific). The `dapr_metadata_value_xor_secret` CEL moved to presence form. The canonical use is an `azureClientId` entry tracking an `AzureUserAssignedIdentity`'s `client_id` output.
- **Scale-rule `identity_id`** → `StringValueOrRef` with default kind `AzureUserAssignedIdentity` → `status.outputs.identity_id`; the literal `"System"` stays first-class. Both siblings moved together (identical message shape).

Modules unwrap with `GetValue()` on the Pulumi side; the Terraform variable shapes are unchanged (references arrive flattened by the tfvars converter). Spec tests cover literal and reference forms; hack manifests, the E2E scenario, presets, README, catalog page, and docs moved to the wrapper form. The `02-servicebus-pubsub` preset was rewritten to the keyless shape.

### The four charts

**`azure/container-apps-environment`** (18 documents at defaults) — a VNet-injected (/21, `Microsoft.App/environments`-delegated), zone-redundant environment running a public API, an ingress-less worker, and a cron job. The messaging spine is fully keyless: one user-assigned identity, Azure Service Bus Data Sender + Data Receiver grants, a Dapr pub/sub kind (`pubsub.azure.servicebus.queues`) whose `namespaceName` is render-composed and whose `azureClientId` rides the new reference seam, with `disableEntityManagement: "true"` so the topic's backing queue stays a first-class IaC resource with real DLQ posture. The worker scales 0→N on queue depth through a KEDA `azure-servicebus` rule that authenticates as the same identity (the other new seam), and mounts an SMB Azure Files volume registered with the account key by reference (the one seam Azure offers no identity path for — stated honestly in the template). Toggles: `zone_redundancy_enabled`, `internal_only_enabled`.

**`azure/service-bus-messaging`** (15 at defaults) — the integration backbone: per-service command queues (looped over a list param) each with DLQ posture and its own send-only/listen-only SAS pair; an events topic with a SQL-filtered subscription plus a `1=1` audit catch-all (the `$Default` rule cannot be declared — the always-true filter states the intent); a namespace-wide `DeadletteredMessages > 0` alert. The `premium_enabled` toggle renders the sku + capacity + partition trio together (the PREMIUM CELs demand all three).

**`azure/event-streaming-platform`** (15 at defaults) — separate ingest and processed hubs (different data, not different cursors), per-application consumer groups, an AVRO/BACKWARD schema group, and keyless capture: the namespace's system-assigned identity writes Avro archives into an HNS (Data Lake Gen2) storage account through a Storage Blob Data Contributor grant — no storage key exists for archival. Alerts: `ThrottledRequests > 0` (severity 1) and `CaptureBacklog > 500MB`. Toggles: `auto_inflate_enabled` (renders the TU ceiling only alongside the flag, mirroring Azure's pairing), `capture_enabled` (drops the whole archive arm), `keyless_only_enabled` (flips `local_authentication_enabled` off and suppresses the SAS rules in the same branch).

**`azure/ha-postgres`** (13 at defaults) — zone-redundant HA, the CEL-enforced VNet-injection trio (delegated subnet + `privatelink.postgres.database.azure.com` zone + public access off), Entra administration through a user-assigned identity, and CMK ON by default: purge-protected RBAC vault, RSA-3072 wrap/unwrap key referenced versionless, and the `Key Vault Crypto Service Encryption User` grant. Curated `server_parameters` (`log_min_duration_statement`, `log_lock_waits`) with the reasoning inline. Toggles: `ha_enabled`, `cmk_enabled`, `geo_redundant_backup_enabled`, `replica_enabled` (a `create_mode: REPLICA` server tracking the primary by self-reference, inheriting SKU/storage).

## Validation (offline gate — live E2E is closed by owner directive)

- `planton chart validate` (CLI built from this tree) green for each chart and for the full nine-chart Azure catalog (defaults + every bool toggle flipped once).
- Chart structure guard green; all four iconUrls verified live (curl 200).
- Retrofits: spec tests green ×3 (literal + reference forms), `planton tofu plan` green on all three hack manifests, Pulumi module release builds green ×3, `make build-go` green, `validate-refs --check` and `secret-coverage --check` green.
- One defect the toggle-flip variant caught in-session: the PREMIUM Service Bus branch initially omitted `premium_messaging_partitions`, which the namespace's own CEL rejected — fixed by rendering the sku/capacity/partitions trio together.

## Impact

Teams can deploy a working event-driven container estate, a correct messaging topology, a schema-governed streaming platform with a lakehouse landing zone, and an auditor-ready PostgreSQL — each in minutes, each secure by default. The two retrofitted seams benefit every future composition: any chart or manifest can now wire Dapr components and KEDA scalers to managed identities by reference.

## Related Work

- The authoring contract (`_rules/charts/forge-planton-infra-chart.mdc`) gained two secure-by-default clauses distilled from this work: prefer the identity seam over an existing credential output, and disable a service's entity management when IaC owns the topology.
- Completes DD-006 sessions 038-039; remaining: the final three charts, the shared-builder sweep, full-catalog validation, and the Phase-1 release.

---

**Status**: ✅ Production Ready
