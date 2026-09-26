# AzureFunctionAppFlexConsumption Terraform Module

## Overview

Creates an Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, always-ready instance pools, and explicit blob-container deployment storage.

## Resources Created

- `azurerm_function_app_flex_consumption` -- the app (runtime declaration, deployment-storage binding, scale configuration, site config, app settings, connection strings, sticky settings, identity, Easy Auth v2, tags)

## Variables

The generated `variables.tf` mirrors the proto contract:

- `metadata` -- Planton resource metadata (name, org, env, labels, tags)
- `spec` -- the AzureFunctionAppFlexConsumptionSpec fields; the resource-group, service-plan, storage-key, identity, App Insights, and subnet references arrive as resolved literals

## Outputs

- `function_app_id` -- the app's ARM resource ID
- `default_hostname` -- the app's FQDN ({name}.azurewebsites.net)
- `outbound_ip_addresses` / `possible_outbound_ip_addresses` -- the active and superset outbound IP lists (use the superset for durable firewall allowlists)
- `identity_principal_id` / `identity_tenant_id` -- the system-assigned identity (empty unless enabled)
- `custom_domain_verification_id` -- the DNS TXT value for custom-domain binding
- `kind` -- Azure's kind string ("functionapp,linux")
- `site_credential_name` / `site_credential_password` -- the Kudu/SCM publishing credential (the password is sensitive; live only while basic-auth publishing is enabled)

## Behavior Notes

- **The FC1 gate**: the provider verifies the referenced plan's SKU is FC1 at create time and rejects anything else -- the plan reference must point at a Flex Consumption plan.
- **storage_container_type is a constant**: `blobContainer` is the argument's single legal value, so the spec omits it and this module sends the constant.
- **The storage-auth triangle**: connection-string auth requires the access key, user-assigned-identity auth requires the identity id (spec-enforced, mirroring the provider's create-time checks); system-assigned needs neither. Empty optionals become null so the provider's not-empty validators never see a present-but-empty value.
- **Write-only fields**: `storage_access_key` (returned only inside the AzureWebJobsStorage app setting in key mode), `zip_deploy_file` (never returned), and `app_service_logs` (update-only apply, never read) do not round-trip -- re-supplying them is expected, not drift. `elastic_instance_minimum` is accepted but never read back on this hosting model.
- **App settings the provider manages**: `AzureWebJobsStorage`, `DEPLOYMENT_STORAGE_CONNECTION_STRING`, the App Insights keys, and `WEBSITE_HEALTHCHECK_MAXPINGFAILURES` are derived from their dedicated fields and filtered from `app_settings` reads.
- **Presence-fires conflicts**: the Entra ID credential pair (secret setting name vs certificate thumbprint) and the login token-store backings are ConflictsWith partners that fire on argument PRESENCE -- this module nulls the empty member so a populated-plus-empty pair never reaches the provider.
- **Unset optionals ride the platform defaults**: instance memory 2048 MB, max instances 100, HTTP concurrency Azure-assigned, TLS floors 1.2, client-cert mode Optional, load balancing LeastRequests.
