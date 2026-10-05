# AzureApplicationInsights - Terraform Module

Terraform implementation for the AzureApplicationInsights
kind.

## Resources Created

- `azurerm_application_insights.main` -- the workspace-based component,
  carrying the application type, workspace binding, retention/sampling/
  cap dials, privacy/auth/network posture, and merged governance tags

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.application_type` | Proto enum value name mapped in `locals.tf` to Azure's CASE-SENSITIVE strings ("Node.JS", "MobileCenter"); absent deploys "web". ForceNew |
| `spec.workspace_id` | The resolved workspace ARM id; repointable, never removable once set |
| `spec.daily_data_cap_in_gb` + notifications toggle | Applied by the provider through a separate billing API call after create |
| `spec.ip_masking_enabled` etc. | v5-positive presence-modeled booleans; explicit false survives to the wire |

## Outputs

`application_insights_id`, `application_insights_name`,
`instrumentation_key`, `connection_string` (the seam app kinds
reference), `app_id`. Microsoft documents both as non-secret; the
provider marks them sensitive, so the outputs unwrap them.
