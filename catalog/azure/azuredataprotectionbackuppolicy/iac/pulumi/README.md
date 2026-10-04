# AzureDataProtectionBackupPolicy Pulumi Module

## Overview

Creates a Data Protection backup policy -- the schedule and retention rules for one datasource type, expressed as one spec with six variant blocks -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module. Exactly one variant is set (validated at admission); the module creates exactly one provider resource.

## Resources Created (exactly one, by variant)

- `dataprotection.BackupPolicyBlobStorage` -- the `blob_storage` variant
- `dataprotection.BackupPolicyDisk` -- the `disk` variant
- `dataprotection.BackupPolicyKubernetesCluster` -- the `kubernetes_cluster` variant
- `dataprotection.BackupPolicyMysqlFlexibleServer` -- the `mysql_flexible_server` variant
- `dataprotection.BackupPolicyPostgresqlFlexibleServer` -- the `postgresql_flexible_server` variant
- `dataprotection.BackupPolicyDataLakeStorage` -- the `data_lake_storage` variant

## Outputs

- `backup_policy_id` -- the policy's full ARM ID, whichever variant ran; what backup instances bind their policy by
- `backup_policy_name` -- the policy's name, unique on its vault

## Behavior Notes

- **Every variant is immutable after create**: the provider ships no update path (near-total ForceNew) -- changing anything replaces the policy, and bound backup instances re-bind to the replacement.
- **One vault reference, three provider addressing styles**: most variants take the vault's ARM ID directly; `BackupPolicyKubernetesCluster` takes vault NAME + resource group (both derived from the ID in `locals.go` -- ARM IDs are structured); `BackupPolicyDataLakeStorage` takes the same ID under its own argument name.
- **Data Lake priorities are order-derived**: the provider stamps rule N with priority N+1 (there is no priority argument on that resource) -- list rules in priority order.
- **Stores are pinned per datasource** (the service's own surface today): disk and Kubernetes retain on `OperationalStore`; the flexible servers, Data Lake, and blob's named rules on `VaultStore`; blob's dual default tiers select their stores by which duration field is set. The spec vocabularies mirror these exactly and widen if the service does.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for all six variants -- zero parity exceptions.

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).
