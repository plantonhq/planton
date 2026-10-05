# AzureAiFoundry Guide

Judgment and internal conventions for the AI Foundry hub kind --
what the schema alone cannot carry.

## Parity accounting

Modeled from `azurerm_ai_foundry` at the pinned azurerm v5.0.0, full
surface, ZERO parity exceptions (the classic Pulumi SDK's
`aifoundry.Hub` carries every v5 argument -- verified field-by-field,
module compile is the standing verifier). Family renames only:
`region` -> `location`, `resource_group` -> `resource_group_name`
(recorded in `iac/provider-parity.yaml`).

## The versioned-key divergence (do not "fix" it)

The hub's `encryption.key_id` validator at v5 demands a VERSIONED Key
Vault key URI (`.../keys/{name}/{version}`) -- versionless URLs are
rejected. The spec's typed ref therefore points at AzureKeyVaultKey's
`key_id` output, NOT the `versionless_id` the classic ML workspace
recommends. Consequence a composer must know: key rotation does NOT
auto-propagate to the hub; rotating means re-pointing the field (and
the whole encryption block is ForceNew). This is the provider's hub
contract, not a modeling choice.

## Hub vs classic ML workspace

The hub IS an ML workspace at ARM (kind "Hub"), but its provider
surface differs in ways that matter to a composer:

- Application Insights is OPTIONAL here (required on the workspace)
  and updatable in place; the container registry is also updatable in
  place (ForceNew on the workspace).
- The hub has no feature-store / serverless-compute / outbound-rule
  arms -- hub outbound rules are managed from the Foundry portal/API;
  azurerm models no children for them.
- Soft-delete behaves exactly like the workspace: the ghost holds the
  name until purged. Both Planton modules enable the provider's
  `machine_learning.purge_soft_deleted_workspace_on_destroy` features
  flag, so a Planton destroy purges the ghost and the name frees
  immediately (proven live: a dual-engine cycle recreated the same
  fixed name minutes after the first engine's destroy). Know the
  listing boundary: NO CLI or REST API lists ghosts -- `az ml
  workspace list` has no soft-delete flag and Resource Graph indexes
  active resources only. The Azure portal's "Recently deleted" view
  (Azure Machine Learning service page, per region) is the one
  listing surface, and purging a ghost left by an outside-Planton
  delete happens there too.

## high_business_impact_enabled is sent only when true

The property is Optional+Computed and the SERVICE flips it true when
encryption is enabled. Both engines send it only when the spec says
true -- pinning false would fight the read-back with a perpetual
diff-and-replace (the flag is ForceNew). "False" in a manifest means
"leave it to the service", recorded on the field.

## Name regex vs error text

The provider's code regex `^[a-zA-Z0-9][\w-]{2,32}$` allows
underscores; its error message says "alphanumeric characters and '-'"
only. The spec mirrors the CODE (the established family precedent:
the regex is what ARM actually enforces).

## E2E shape

The smoke scenario is a system-identity hub on the shared fixture
companions (rg + kv + HNS-off storage). The CMK / user-assigned /
isolation arms are offline-proven only (they need pre-granted key
access the minimal lane deliberately does not compose). The hub's own
install profile (`e2e/prerequisite.yaml`) is load-bearing for the
AzureAiFoundryProject lane.
