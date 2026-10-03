# AzureMachineLearningDatastore Pulumi Module

## Overview

Registers a datastore on an Azure Machine Learning workspace using the classic `pulumi-azure` (azurerm-bridged) SDK, from the kind's typed IaC input. The spec's variant block selects which of the three SDK resources is created.

## Design Decisions

- **Wire map identical to the Terraform module**: the same variant switch (blob / data-lake / file-share), the same service-data identity enum map, and the same omit-when-unset semantics.
- **The identity-argument rename**: the SDK mirrors the provider's naming split (`ServiceDataAuthIdentity` on blob, `ServiceDataIdentity` on the other two) -- ONE spec field feeds both, recorded in the parity manifest.
- **Write-only credentials**: keys, SAS tokens, and client secrets are sensitive on the SDK schema and never returned by ARM.
- **Provider builder**: credentials resolve through the shared `pulumiazureprovider` builder (static client secret, keyless web identity, or ambient chain).

## Inputs

The module consumes `AzureMachineLearningDatastoreIacInput`: the target resource (metadata + spec) and the Azure provider configuration. The workspace and storage-target references arrive pre-resolved; `GetValue()` returns the literal ARM ID.

## Outputs

- `datastore_id`, `datastore_name`
- `is_default` -- settable on the blob variant, read back on the others

## Local Development

```shell
# compile the module
go build ./...
```
