# Azure Observability Wave: Log Analytics + App Insights Depth, Azure Monitor Alerting Family

**Date**: 2026-07-10
**Type**: Feature — kind depth rework ×2 + new kind forge ×4
**Scope**: `apis/dev/planton/provider/azure/{azureloganalyticsworkspace,azureapplicationinsights,azuremonitordiagnosticsetting,azuremonitoractiongroup,azuremonitormetricalert,azuremonitorscheduledqueryalert}/v1`, kind registry, Azure E2E harness, `pkg/outputs` conformance, forge rule 009, `e2e/README.md`

## Summary

The Azure catalog gains its full observability pipeline. The two existing
telemetry-store kinds were rebuilt from thin specs to the complete azurerm
v4.80 surface, and four new Azure Monitor kinds route telemetry in and
alerts out:

- **`AzureLogAnalyticsWorkspace` (450, rework, breaking)** — the central
  Azure Monitor log store at full depth: the closed SKU vocabulary
  (pay-as-you-go and commitment tiers with the capacity pairing enforced in
  both directions), the security/network posture (Entra-only, private-link-
  only paths, resource-context access, forced CMK for query artifacts,
  immediate post-retention purge), a managed identity, the default
  data-collection-rule seam, and user tags. Outputs now separate the ARM
  resource ID (`workspace_id`, the FK seam) from the agent-facing customer
  GUID (`workspace_customer_id`) the provider confusingly also calls
  workspace_id.
- **`AzureApplicationInsights` (451, rework, breaking)** — workspace-based
  APM at full depth: the complete 8-value application-type vocabulary with
  case-exact wire mapping, cost guards (daily cap + notification, sampling),
  privacy and network posture, and the required workspace binding (classic
  mode was retired by Azure).
- **`AzureMonitorDiagnosticSetting` (452, new)** — the routing rule that
  makes any resource observable: a polymorphic target, category/category-
  group selections, and the four destination families (workspace, storage,
  event hub, partner solution), with Azure's at-least-one contracts
  front-loaded as validations.
- **`AzureMonitorActionGroup` (453, new)** — the global notification hub
  with all eleven receiver families (email/SMS/voice/push, Entra-
  authenticated webhooks, runbooks, Logic Apps, Functions, ARM-role
  fan-out, Event Hubs, ITSM).
- **`AzureMonitorMetricAlert` (454, new)** — static, dynamic
  (machine-learning), and web-test availability conditions over platform
  metrics, with per-family operator vocabularies enforced.
- **`AzureMonitorScheduledQueryAlert` (455, new)** — the KQL log alert
  (row-count and metric-measurement styles, dimension splits, failing
  periods, the auto-mitigation XOR mute contract, managed identity).

## Design Notes

- The scheduled-query kind drops the provider's "v2" naming artifact: ARM's
  type is `scheduledQueryRules`, and the provider's v1 resource is a
  superseded older-API shape (recorded skip).
- Polymorphic references (the diagnostic setting's target, the metric
  alert's scopes) deliberately carry no default kind — any resource can be
  a target and none dominates; references are explicit `valueFrom`.
- The diagnostic setting's `diagnostic_setting_id` output is the CONSTRUCTED
  ARM extension-resource ID; the provider's own state ID is a
  `{target}|{name}` composite no Azure API consumes.
- The offline plan gate caught a real service constraint before any live
  run: Log Analytics workspaces support only single identity models
  (SystemAssigned XOR UserAssigned) — the generic combined model was
  removed from the spec enum.

## Validation

- Offline: spec tests ×6 (119 cases, every CEL error path), targeted +
  release-equivalent builds ×6, `make build-go`, secret-coverage,
  validate-refs, `pkg/outputs` conformance ×6 (the two reworked kinds'
  first-ever cases), full `tofu plan` ×6 rendering every enum seam,
  33 manifests validated, audits ×6 at 100% (PARITY ✅ COVERAGE ✅) with
  apply-time validator source-diff sections.
- Live dual-engine E2E: **12/12 green** on the real test subscription —
  including the self-referential diagnostic setting (the fixture workspace
  routing its own audit logs into itself), the composed metric-alert chain
  (action-group fixture + scenario-local storage account), and the composed
  query-alert chain (workspace + action group). Zero orphans after the
  final sweep.

## Live-Caught Classes (fixed and folded back)

1. **Pulumi enum wire maps need an unspecified row when unspecified is
   legal** — a missing Go map key silently sends the empty string, which
   the provider rejects at deploy; the Terraform `locals.tf` null-guard
   idiom masks the class to one engine. Encoded in forge flow rule 009.
2. **Service-callback URLs are server-validated** — Azure Monitor action
   groups reject webhook receivers on placeholder domains
   (`WebhookServiceUriBlocked` on example.com). Fixtures and presets now
   carry domain-you-own URI shapes. Encoded in `e2e/README.md`.
