# Overview

The **AzureFunctionAppFlexConsumption** component deploys an Azure Function App on the Flex Consumption plan -- Azure's newest serverless Functions hosting model: per-instance memory selection, a configurable scale-out ceiling, and always-ready instance pools that eliminate cold starts, billed per execution.

## Purpose

- **Serverless functions without cold-start pain**: name the functions that must respond instantly and keep warm instances just for them -- the rest of the app still scales to zero.
- **Scale you actually control**: choose the memory size each instance gets (512/2048/4096 MB), cap the fan-out (1-1000 instances), and tune per-instance HTTP concurrency.
- **Deployment storage you own**: the app's code package lives in YOUR blob container, and the app authenticates to it by connection string or -- credential-free -- by managed identity.

## Key Features

- Full azurerm v5 surface: the flat runtime declaration (Node, .NET isolated, Java, PowerShell, Python, custom handler), all three deployment-storage authentication modes, always-ready pools, HTTP concurrency, the complete flex site_config (App Insights wiring, IP restrictions with Front Door header filters, CORS, TLS floors, health checks), connection strings, sticky settings, managed identity, client certificates, VNet integration, and Easy Auth v2 with every identity provider.
- The deployment-storage contract is validated before anything reaches Azure: connection-string auth requires the access key, user-assigned-identity auth requires the identity id -- exactly the checks Azure runs at create time, front-loaded.
- Chart-ready: `resource_group` defaults its reference to AzureResourceGroup, `service_plan_id` to AzureServicePlan (the FC1 tier), `storage_access_key` to AzureStorageAccount's primary key, `storage_user_assigned_identity_id` and identity ids to AzureUserAssignedIdentity, `application_insights_connection_string` to AzureApplicationInsights, and subnet references to AzureSubnet; `default_hostname` is what DNS records and upstream proxies consume.
- Secure by default: the storage access key, App Insights credentials, and connection-string values are marked sensitive; Easy Auth provider secrets are referenced by app-setting NAME, never inline; `https_only` deploys true.
- Legacy `auth_settings` (v1): superseded by `auth_settings_v2`, which this kind models fully.

## Use Cases

- **Event-driven APIs and webhooks**: HTTP-triggered functions with an always-ready pool for the hot path.
- **Queue and event processing**: Service Bus, Event Grid, and blob triggers with the fan-out ceiling as the cost lever.
- **Enterprise-authenticated function APIs**: Easy Auth v2 validates Entra ID (or any OIDC) tokens before requests reach function code.
- **Locked-down serverless**: identity-based deployment storage, closed basic-auth publishing, IP restrictions pinned to a Front Door profile.

## Relationship to AzureFunctionApp

Azure models Flex Consumption apps as their own resource type with a distinct configuration surface. Classic Consumption (Y1), Elastic Premium (EP*), and Dedicated-plan function apps are the `AzureFunctionApp` kind; this kind is only for FC1 plans.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
