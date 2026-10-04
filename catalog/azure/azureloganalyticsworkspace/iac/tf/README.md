# AzureLogAnalyticsWorkspace - Terraform Module

Terraform implementation for the AzureLogAnalyticsWorkspace
kind.

## Resources Created

- `azurerm_log_analytics_workspace.main` -- the workspace, carrying the
  pricing tier, retention/quota dials, security and network posture, an
  optional managed identity, and merged governance tags

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.workspace_name` | 4-63 letters/digits/hyphens (spec-enforced); ForceNew -- renaming destroys the data |
| `spec.sku` | Proto enum value name mapped in `locals.tf` (PER_GB_2018 -> PerGB2018, ...); absent deploys PerGB2018. PerGB2018 <-> CapacityReservation transitions in place; other changes are ForceNew |
| `spec.reservation_capacity_in_gb_per_day` | Only with CAPACITY_RESERVATION (spec-enforced both directions) |
| `spec.daily_quota_gb` | -1 means unlimited -- the provider's own default, sent explicitly for parity |
| `spec.identity` | SystemAssigned XOR UserAssigned -- the workspace rejects the combined model |
| `spec.local_authentication_enabled` etc. | Presence-modeled true-default booleans; explicit false survives to the wire |
| `spec.data_collection_rule_id` | Applied by the provider via a follow-up update (ARM rejects a default DCR at create) |

## Outputs

`workspace_id` (ARM id -- the FK seam), `workspace_name`,
`workspace_customer_id` (the agent-facing GUID the provider calls
workspace_id), `resource_group_name`, `primary_shared_key` /
`secondary_shared_key` (sensitive), `identity_principal_id`.
