# Azure Container Apps Family: Depth Rework + Decomposition

**Date**: 2026-07-10
**Scope**: `apis/dev/planton/provider/azure/azurecontainerappenvironment/v1`,
`.../azurecontainerapp/v1`, `.../azurecontainerappjob/v1` (new),
`.../azurecontainerappenvironmentstorage/v1` (new),
`.../azurecontainerappenvironmentdaprcomponent/v1` (new),
`catalog_kind.proto`, `pkg/catalogkindreflect`, `pkg/outputs`,
`aa_e2e` + `e2e/azure`, forge rule 009, `e2e/README.md`

## Summary

The Azure Container Apps family reaches the full azurerm v4.80 surface:
the environment (440) and app (441) reworked breaking, and three new kinds
forged — `AzureContainerAppJob` (445), `AzureContainerAppEnvironmentStorage`
(446), and `AzureContainerAppEnvironmentDaprComponent` (447) — completing
the app-hosting sub-band (440–447). Live dual-engine E2E green 10/10 with
zero orphans.

## What Changed

### AzureContainerAppEnvironment (440) — breaking rework

- `name` → `environment_name`; full v4.80 surface: `logs_destination`
  closed enum with the workspace pairing contracts as CELs,
  `dapr_application_insights_connection_string` (sensitive, ForceNew),
  `infrastructure_resource_group_name` (requires workload profiles),
  `public_network_access` enum (ENABLED conflicts with ILB — CEL),
  `mutual_tls_enabled`, managed identity (UAI FKs), user tags, and the
  workload-profile type promoted to a closed 14-value SKU enum
  (Consumption + serverless GPU + D/E/NC families).
- The environment custom DNS suffix FOLDS into the spec (ARM models it as
  a singleton patch on the environment; the association resource realizes
  it on both engines).
- Outputs gain `environment_name`, `docker_bridge_cidr`,
  `custom_domain_verification_id`, `identity_principal_id`.

### AzureContainerApp (441) — breaking rework

- The shipped ingress parity defect is closed: BOTH engines now wire BOTH
  `cors` and `client_certificate_mode` (previously each engine carried one
  and dropped the other).
- Volume `storage_type` completes to the provider's real 4-value
  vocabulary (EmptyDir / AzureFile / NfsAzureFile / Secret); `storage_name`
  becomes a real FK to the new environment-storage kind.
- `name` → `container_app_name`; user tags; custom scale rules gain
  `identity_id` and the provider's exact KEDA scaler allowlist as a CEL;
  the provider's cross-field contracts front-loaded (registry auth XOR,
  secret Key-Vault/identity pairing, traffic-weight targeting, per-probe-
  type threshold ceilings 30/48/240 with readiness-only success threshold);
  every string vocabulary now a closed enum with per-engine wire maps
  (including the lowercase `accept/require/ignore` client-cert wire values,
  verified against the vendored SDK constants).

### New kinds (445–447)

- **AzureContainerAppJob** (`azcaj`): run-to-completion workloads — the
  full template subset (containers/init/volumes, no scale block), required
  `replica_timeout_in_seconds`, retry limit, and exactly-one-of
  manual/schedule/event triggers (CEL) with the event trigger's execution
  fan-out contract (min/max executions, polling, KEDA rules with
  `identity_id`). Outputs include `event_stream_endpoint` and
  `outbound_ip_addresses`.
- **AzureContainerAppEnvironmentStorage** (`azcaes`): registers an Azure
  Files share on the environment — SMB (account + sensitive access key,
  FK-defaulted to `AzureStorageAccount` outputs) XOR NFS (CEL), share FK
  to `AzureStorageShare`; closes the app/job share-backed volume seam.
- **AzureContainerAppEnvironmentDaprComponent** (`azcadapr`): Dapr
  state stores/brokers/bindings on the environment — metadata entries with
  value-XOR-secret CELs, sensitive Dapr secrets, app scoping.

### Cross-cutting

- All five Pulumi modules on the shared `pulumiazureprovider.Get` builder;
  empty Terraform provider blocks; registry prerequisites wired
  (env→ResourceGroup; app/job/storage/dapr→Environment).
- `pkg/outputs` conformance cases ×5; verifiers ×5 (generic ARM GetByID at
  the Microsoft.App `2025-01-01` pin); E2E scenarios/profiles/entrypoints
  for the whole family.
- Live-caught and fixed: the environment modules sent
  `internal_load_balancer_enabled`/`zone_redundancy_enabled` without the
  subnet — the provider rejects a SPECIFIED false without its RequiredWith
  pairing (manifest-driven IaC inputs materialize proto defaults). Both
  engines now gate the pair on subnet presence.

## Validation

- Offline gate fully green: spec tests ×5 (every CEL error path), targeted
  + release-equivalent builds ×5, `make build-go`, secret-coverage,
  validate-refs, `pkg/outputs` ×5, `validate-outputs` ×5, full
  `planton tofu plan` ×5 hack manifests, 12 presets + 8 E2E manifests
  validate, audits ×5 at 100% Fully Complete PARITY ✅ COVERAGE ✅.
- **Live dual-engine E2E green 10/10** (~118-minute suite): environment
  662s/671s, app (composed RG → env → app with ingress + HTTP scaling)
  728s/704s, job 664s/682s, storage (scenario-local account + share chain)
  848s/854s, Dapr component 640s/643s. Zero orphans (resource groups and
  storage accounts both empty after the sweep).
- One-time subscription bootstrap performed: `Microsoft.App` namespace
  registered (the harness's registration opt-out surfaces this once per
  new resource-provider family).

## Learnings folded back (learn-once)

- Forge rule 009: the RequiredWith-paired-optionals class — presence
  guards are insufficient when a proto default materializes; gate on the
  pairing field on BOTH engines.
- `e2e/README.md`: the `MissingSubscriptionRegistration` bootstrap class
  and its one-time `az provider register` fix.
