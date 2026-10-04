# AzureAiFoundryProject Guide

Judgment and internal conventions for the AI Foundry project
kind -- what the schema alone cannot carry.

## Parity accounting

Modeled from `azurerm_ai_foundry_project` at the pinned azurerm
v5.0.0, full surface, ZERO parity exceptions (the classic Pulumi
SDK's `aifoundry.Project` carries every v5 argument -- verified
field-by-field, module compile is the standing verifier). One family
rename: `region` -> `location` (recorded in
`iac/provider-parity.yaml`).

## Why there is no resource group field

The provider derives the project's resource group from the hub
reference (create parses the hub ID and places the project in the
hub's group). Modeling a group field would invite a value the
provider ignores -- the absence IS the contract, recorded on the spec
message.

## The inheritance hack in the provider's update path

Hubs and projects share one ARM API. When updating a project, the
provider NILS the inherited properties (vault, storage, insights,
registry, network, encryption) before patching -- ARM errors when a
project update carries them. Nothing for a composer to do, but it
explains why the project's surface is deliberately thin: adding
"convenience" fields for inherited posture would fight the API.

## Update surface

`description`, `friendly_name`, `identity`,
`primary_user_assigned_identity`, and `tags` update in place; `name`,
`region`, the hub linkage, and `high_business_impact_enabled` are
ForceNew. The provider's RequiredWith between the primary identity
and the identity block is spec CEL (reference-safe presence check on
the StringValueOrRef).

## high_business_impact_enabled is sent only when true

Same class as the hub: Optional+Computed, service-flipped when hub
encryption applies. Both engines send only when true; a pinned false
would fight the read-back on a ForceNew flag.

## Deletion has no purge seam -- and measured live, that is fine

Projects are ML workspaces at ARM, so deletion is the family's SOFT
delete -- but unlike the hub resource, the provider's project delete
never consults the `machine_learning.
purge_soft_deleted_workspace_on_destroy` features flag (it sends the
default delete options; verified in the provider source at the pinned
v5.0.0). Setting the flag in this kind's provider config would be a
silent no-op, so the modules deliberately do not. Measured live:
destroying a project and recreating the SAME name in the same
resource group minutes later succeeded -- a project ghost does not
block recreation the way a standalone hub/workspace ghost does (the
hub's own purge-on-destroy sweeps its children's ghosts with it). If
a standalone project ghost ever does surface, the Azure portal's
"Recently deleted" view (per region) is the only listing and purge
surface -- no CLI or REST API lists ML ghosts.

## E2E shape

The smoke scenario is a system-identity project on the fixture hub
(`planton-oss-e2e-azure-fixture-aif`, the hub's own install profile)
-- the lane's proof is the cross-kind wiring: hub chain deploys, the
project's `aiServicesHubId` resolves from the hub fixture's outputs,
and ARM returns kind "Project". The user-assigned arm is
offline-proven only (tenant-scoped identity values).
