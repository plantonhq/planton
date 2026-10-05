# Azure Event Hubs Family: Namespace Depth Rework + Seven-Kind Decomposition

**Date**: 2026-07-12
**Type**: Feature (breaking rework + seven new kinds + two reference retrofits)
**Scope**: `apis/dev/planton/provider/azure/azureeventhub*`, `azuremonitoractiongroup`, `azuremonitordiagnosticsetting`, `catalogkind`, `pkg/catalogkindreflect`, `pkg/outputs`, `aa_e2e`, `e2e/azure`

## Summary

The Azure Event Hubs service is now modeled as a complete family of eight
first-class, composable kinds. `AzureEventHubNamespace` (471) was reworked
breaking from a bundled composite (namespace + inline event hubs + inline
consumer groups) to the full azurerm v4.80 namespace surface, and seven
kinds were forged around it: `AzureEventHub` (477, the partitioned stream
with its capture-to-storage block), `AzureEventHubConsumerGroup` (478),
`AzureEventHubAuthorizationRule` (479, one polymorphic kind over Azure's
two byte-identical SAS-rule ARM types), `AzureEventHubDisasterRecoveryConfig`
(520), `AzureEventHubSchemaGroup` (521), `AzureEventHubCluster` (522, the
dedicated single-tenant tier), and
`AzureEventHubNamespaceCustomerManagedKey` (523, BYOK encryption as the
second-step configuration Azure itself models it as).

Live dual-engine E2E: 12 lanes green (~67-minute suite) including the
geo-DR pairing proven end to end on Standard-tier namespaces across two
regions; the cluster and CMK kinds are offline-gated with the technical
reasons recorded in their E2E profiles (Azure's 4-hour cluster-deletion
moratorium; CMK's add-only lifecycle makes a destroy phase unverifiable).
Zero orphans after the sweep.

## The namespace rework (breaking)

- `name` → `namespace_name` (provider-exact regex; the name is the public
  DNS identity and Kafka bootstrap host — Event Hubs shares the Service
  Bus DNS zone).
- Bundled `event_hubs[]`/`consumer_groups[]` DISSOLVED into the
  first-class kinds; the `event_hub_ids` map output retired with them
  (verified consumer-free).
- Closed BASIC/STANDARD/PREMIUM sku enum (STANDARD default; the
  Premium-boundary ForceNew CustomizeDiff documented) with tier-dependent
  capacity semantics: throughput units (1-40) on multi-tenant tiers,
  processing units (1/2/4/8/16) on PREMIUM — both directions as CELs.
- Auto-inflate + `maximum_throughput_units` ceiling. Deliberately NO
  pairing CEL: azurerm enforces no schema pairing (its only guard is
  zeroing the ceiling on a Basic downgrade, which the provider itself
  performs); ARM validates the combination at apply time — documented,
  not invented.
- `dedicated_cluster_id` FK → `AzureEventHubCluster` (ForceNew
  placement).
- All three managed-identity models; `local_authentication_enabled`
  keyless dial (false = Entra-only; pair with AzureRoleAssignment
  data-plane roles).
- Network rule set FOLDED (inline in azurerm; no independent lifecycle):
  required explicit default_action, DENY-requires-admitted-sources CEL,
  the block-vs-namespace public-access pairing CEL (default-aware in both
  directions), trusted-services bypass, IP rules (per-rule action is a
  one-value constant — modeled as masks), `AzureSubnet` FK rules.
- `zone_redundant` REMOVED — not part of the azurerm v4 surface; the old
  field silently deployed nothing on Pulumi (a shipped parity gap, now
  closed). `minimum_tls_version` a recorded one-value-constant skip.
- Outputs: `namespace_id`/`namespace_name`/`identity_principal_id` + the
  root SAS rule's SIX credential faces (primary/secondary key, connection
  string, and geo-DR alias connection string) — all sensitive.
- Registry gains `prerequisites: [AzureResourceGroup]`; the Pulumi module
  migrated from inline `azure.NewProvider` to the shared
  `pulumiazureprovider.Get` builder (keyless auth now works).

## The seven forged kinds

- **AzureEventHub (477, `azehub`)**: parent by ARM id; partition_count
  1-1024 (tier caps + never-decrease as documented apply-time contracts);
  `message_retention` XOR `retention_description` (hour-granular DELETE
  windows, Kafka-style COMPACT with tombstone retention — the
  policy-to-field pairing front-loaded because the provider silently
  drops the mismatched field); ACTIVE/DISABLED/SEND_DISABLED gate;
  capture FOLDED (encoding, 60-900s/10-500MB cadence, the
  all-nine-tokens archive format CEL, `AzureStorageAccount` +
  `AzureStorageContainer` FKs, three storage-auth modes with the
  UserAssigned-requires-identity pairing CEL). The `enabled` flag is
  `optional bool` + required so an explicit false (pause archival, keep
  config) validates.
- **AzureEventHubConsumerGroup (478, `azehcg`)**: single `event_hub_id`
  parent reference — both engines parse azurerm's legacy discrete-name
  addressing from the resolved ARM id with identical anchored semantics;
  `$Default` is CEL-reserved (the service-created catch-all cannot be
  adopted declaratively).
- **AzureEventHubAuthorizationRule (479, `azehauth`)**: exactly-one-of
  `namespace_id` XOR `event_hub_id` dispatching to the matching provider
  resource on both engines; at-least-one-right + manage-requires-both
  CELs (the provider's shared CustomizeDiff); `RootManageSharedAccessKey`
  reserved; six sensitive credential faces + the geo-DR alias pair; the
  TF module coalesces outputs PER ATTRIBUTE across the count-gated
  variants.
- **AzureEventHubDisasterRecoveryConfig (520, `azehdr`)**: alias +
  primary (ForceNew, name-parsed) + partner (updatable — the provider
  break-pairs and re-pairs); the break-pair → delete → 404-wait →
  alias-name-release destroy choreography documented in both modules;
  faithful to Event Hubs' own surface (no credential outputs — alias
  connection strings live on the namespace/rule kinds).
- **AzureEventHubSchemaGroup (521, `azehsg`)**: the schema registry —
  NONE/BACKWARD/FORWARD compatibility, AVRO/JSON formats, every field
  ForceNew (the resource has no update surface).
- **AzureEventHubCluster (522, `azehclu`)**: dedicated capacity units;
  both modules compose the `Dedicated_{n}` ARM sku from the count (the
  tier name is a one-value constant); the 4-hour deletion moratorium
  and per-CU billing documented prominently.
- **AzureEventHubNamespaceCustomerManagedKey (523, `azehcmk`)**: BYOK as
  its own kind following the service's own grain — system-assigned-
  identity CMK is only expressible as a second step (create → grant →
  apply); 1-10 key references defaulting to versionless ids; the
  ADD-ONLY lifecycle (provider delete is a deliberate no-op) taught
  everywhere it matters.

## Reference retrofits

- `AzureMonitorActionGroup`'s event-hub receiver `event_hub_name`: plain
  string → `StringValueOrRef` → `AzureEventHub.event_hub_name`.
- `AzureMonitorDiagnosticSetting`'s event-hub destination:
  `eventhub_authorization_rule_id` and `eventhub_name` plain strings →
  FKs on the new rule and hub kinds; the stream-to-siem preset now
  composes through them. These close the catalog's last Event Hub
  plain-string seams.

## Validation

- Offline: 131 spec tests across the 8 kinds (every CEL error path);
  chunked `buf generate` (persistent remote-plugin degradation; the
  documented `--path` workaround) + the full-tree Java compile gate;
  kind-map + gazelle regen; targeted + release-equivalent builds ×8;
  `make build-go`; Bazel builds of all 8 kind trees;
  `secret-coverage --check` (fourteen new sensitive credential faces);
  `validate-refs --check` (13 new FK edges); `pkg/outputs` conformance
  cases ×8; full `planton tofu plan` on all 8 hack manifests (every
  enum/CEL seam rendered); 13 presets + all E2E manifests validate;
  audits ×8 at 100% Fully Complete, PARITY ✅ COVERAGE ✅, each with an
  apply-time validator source-diff section; retrofit audit addendums ×2;
  site catalog regenerated (8 event-hub slugs).
- Live (test subscription): **12 lanes green + 4 profile-bound skips**
  in one 4015-second suite — namespace 162s/182s (DENY firewall +
  auto-inflate + root-SAS outputs), hub 367s/439s (hour-granular
  retention + full capture through the scenario-local storage chain),
  consumer group 236s/266s, authorization rule 437s/493s (BOTH dispatch
  paths per engine), schema group 195s/212s, geo-DR 535s/469s (the
  twin-region Standard-tier pairing, break-pair destroy, alias-name
  release). Cluster ×2 + CMK ×2 skipped by their binding E2E profiles
  with recorded reasons. Zero orphans (resource groups, namespaces,
  storage accounts, clusters all empty).

## Learnings folded back

- Forge flow rule 001: a REQUIRED bool whose false is meaningful must be
  `optional bool` + required (presence-tracked) — a plain bool + required
  rejects the explicit false it exists to express.
- e2e/README: background suite lanes must be smoke-checked after launch
  (PID alive + non-empty log) — a nohup child of a short-lived shell can
  be reaped before the suite starts, leaving a zero-byte log and no
  error.
