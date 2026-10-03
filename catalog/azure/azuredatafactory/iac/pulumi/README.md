# AzureDataFactory Pulumi Module

## Overview

Creates an Azure Data Factory -- the workspace every other Data Factory resource lives inside -- with its managed identity, optional git repository binding, workspace-wide global parameters, customer-managed-key encryption, named credentials, managed virtual network, and managed private endpoints.

## Resources Created

- `datafactory.Factory` -- the factory (identity, repo binding, global parameters, network posture, inline customer-managed-key encryption)
- `datafactory.CredentialUserManagedIdentity` -- one per named entry in `user_managed_identity_credentials`, keyed by name
- `datafactory.CredentialServicePrincipal` -- one per named entry in `service_principal_credentials`, keyed by name
- `datafactory.ManagedPrivateEndpoint` -- one per named entry in `managed_private_endpoints`, keyed by name

## Outputs

- `data_factory_id` -- the factory's ARM resource ID (the target an AzureDataFactoryPipeline's `data_factory_id` references)
- `data_factory_name` -- the factory's name
- `identity_principal_id` -- the system-assigned identity's principal (empty when no system-assigned identity)
- `credential_ids` -- name-keyed ARM IDs of the composed credentials (both flavors share one namespace)
- `managed_private_endpoint_ids` -- name-keyed ARM IDs of the composed managed private endpoints

## Behavior Notes

- **The repository binding is a side-channel call**: the provider configures the repo AFTER the factory exists, and removing the block does NOT detach the repository (the provider calls no repo-clear API) -- detach in the Data Factory Studio.
- **Customer-managed-key encryption rides the factory's inline fields** (the provider's standalone CMK resource writes the same encryption object; it demands a VERSIONED key where the inline fields accept versionless too). Once enabled, the key cannot be removed without recreating the factory.
- **The managed virtual network is one-way**: enabling it updates in place (the provider creates the managed network named "default"); disabling it REPLACES the factory.
- **Managed private endpoints are create-only**: the provider has no Update -- every field change replaces that endpoint (siblings untouched; entries keyed by name). The TARGET side must approve each connection; the endpoint provisions to Succeeded while approval is still Pending.
- **Factory names are globally unique across Azure** -- a taken name fails at deploy time.
- **Platform defaults always sent**: public network enabled, repo publishing enabled, managed VNet false -- the rendered plan states every value.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureDataFactory` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.
