# Azure Chart Catalog Completed: Cosmos API Backend, Data Lakehouse, Zonal Web Tier + Composition Seam Retrofits

**Date**: July 16, 2026
**Type**: Feature
**Components**: Azure Infra Charts, Azure Provider, API Definitions, Reference Validation, Provider Framework

## Summary

Three new Azure Infra Charts — `azure/cosmos-db-api-backend`, `azure/data-lakehouse-storage`, and `azure/zonal-web-tier-vmss` — complete the Azure chart catalog at twelve. Building them surfaced three composition gaps, each fixed at the root: the ADLS filesystem's POSIX-ACL principal fields became references, the storage account gained per-service diagnostic-target outputs, and the offline reference validator learned to resolve map-typed outputs addressed by entry key (the load balancer's name-keyed pool ids), which the deploy-time resolver already supported.

## The Three Charts

### `azure/cosmos-db-api-backend` — keyless, private, planet-scale

The RBAC-only Cosmos posture assembled end to end (~15 resources at defaults): a SQL-API account with `localAuthenticationEnabled: false` (key-based data access rejected outright) and no public data plane; a workload `AzureUserAssignedIdentity` holding a CUSTOM least-privilege data-plane role (the full read surface + create/replace/upsert — `items/*` would silently include delete, so the verbs are spelled out); a private endpoint on the `Sql` subresource with the `privatelink.documents.azure.com` zone; continuous backup; DEDICATED-mode diagnostics; and a `TotalRequests` alert dimension-filtered to `StatusCode = 429` — the RU-ceiling signal. Toggles: zone redundancy, a second region with automatic failover.

### `azure/data-lakehouse-storage` — the governed medallion foundation

An HNS account (ZRS) carrying the three medallion zones as first-class filesystems, each with its own encryption scope (CMK on by default: purge-protected RBAC vault + RSA-3072 versionless key + Crypto Service Encryption User granted to the ACCOUNT's system identity — the scope-CMK contract), root POSIX ACLs whose named entries reference the engineering/analyst identities BY PRINCIPAL ID, and five `AzureRoleAssignment`s at **container-proxy scope** (`filesystem_id`) — "analysts read curated and nothing else" as architecture. Lifecycle rules encode the zone economics (raw hot→cool→archive; sandbox auto-expiry; curated stays hot). The data-protection story is stated honestly: HNS has no versioning/PITR — soft delete + lifecycle + immutable-landing discipline is the real story. The audit trail targets the account's **blob service** id so `StorageRead/Write/Delete` logs land in the workspace.

### `azure/zonal-web-tier-vmss` — the honest IaaS baseline

A FLEXIBLE-orchestration scale set across zones 1–3 (`platformFaultDomainCount: 1`, user-assigned identity — the Flexible contract) behind a zone-redundant Standard LB. Pool membership is wired from the member side through the LB's name-keyed map output (`status.outputs.backend_pool_ids.web`); egress is an explicit outbound rule with `disableOutboundSnat` on the traffic rules (Azure rejects the overlap); the NSG admits 80/443 and nothing else — no management port exists (Bastion is the day-2 answer; NIC-side NAT-rule attachment is Uniform-only). Default cloud-init installs nginx so the tier serves its own health probe out of the box; the SSH key defaults to a discarded throwaway that boots fine and admits no one.

## Composition Seam Retrofits (the charts' proof-of-composition duty)

1. **`AzureStorageDataLakeGen2Filesystem`: `owner`, `group`, and `aces[].object_id` → `StringValueOrRef`** (default kind `AzureUserAssignedIdentity` → `status.outputs.principal_id`). A lake chart granting an in-graph identity zone access could not feed the ACL by reference. All three fields carry the same value class, so all three moved together; the GUID-format rules live in the field comments now (wrappers cannot carry value-format constraints), `$superuser` stays first-class as a literal, and the mask/other-take-no-qualifier contract became a presence-form message CEL. Spec tests gained literal + reference forms; hack manifest, preset, and docs migrated.

2. **`AzureStorageAccount`: four service-level diagnostic-target outputs** — `blob_service_id`, `file_service_id`, `queue_service_id`, `table_service_id` (`{account}/blobServices/default` etc., constructed identically on both engines; ARM materializes the services implicitly, so there is nothing to read back). Data-access logs live on these sub-resources — the account-level id exposes only account metrics, so before this, no chart could wire a storage audit trail by reference. `pkg/outputs` conformance covers all four.

3. **`pkg/refcheck`: map outputs addressed by entry key now resolve.** The offline validator rejected `status.outputs.backend_pool_ids.web` ("no field named 'web' on …BackendPoolIdsEntry") — it descended into the synthetic map-entry message instead of treating the segment as a runtime KEY, while the deploy-time resolver and the platform's output flattener both support exactly this form. `ResolveValueFromPath` now consumes the segment after a map field as the entry key (string map values terminate the path; message map values continue through it), with a regression suite in `pkg/refcheck/resolve_test.go`. This closes a validator-vs-runtime split that would have falsely failed any chart using the LB/App-Gateway name-keyed output seams.

## Catalog assets

The `azurevirtualmachinescaleset` and `azurefirewall` logos (both absent from the CDN) were published from the platform repo's asset tree (official Azure architecture icons, R2 upload + cache purge, curl-200 verified), and `azure/hub-spoke-network-foundation`'s icon moved off its VNet-family stand-in onto the firewall logo — closing that recorded follow-up.

## Validation

Offline gate (live E2E is closed for this phase, by design):

- `planton chart validate --all charts/azure` with the working-tree CLI: **12/12 green** (defaults + every bool toggle flipped per chart); chart structure guard green.
- Spec tests green for both retrofitted kinds; `pkg/outputs`, `pkg/refcheck` (new suite), and `pkg/infrachart` suites green.
- Targeted builds + release-equivalent Pulumi builds for both retrofitted modules; `make build-go` green (re-run after the resolver fix).
- `planton tofu plan` green on both retrofitted kinds' hack manifests (the filesystem plan renders the wrapper forms; the account plan renders all four service ids).
- `secret-coverage --check` and `validate-refs --check` green.
- Icon URLs curl-200 ×4 (the three new charts + hub-spoke).

## Workflow uplift

- `_rules/charts/forge-planton-infra-chart.mdc`: map-typed outputs are addressed by entry key in `valueFrom` fieldPaths (and dots in sub-resource names break dot-path addressing — name accordingly).
- `_rules/catalog-kind/forge/flow/004-outputs.mdc`: export the ids of implicit sub-resources other kinds target (the diagnostic-scope class), constructed from the parent id when the provider offers no readback.

## Impact

The Azure chart catalog is complete: twelve production-shaped environments, every cross-resource seam wired by reference, validated offline against the compiled-in schemas. The three retrofits ship to every consumer of the two kinds, and the resolver fix unblocks the map-output seam for every provider's charts.

---

**Status**: ✅ Production Ready
