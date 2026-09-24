# A Container App and a Job name the registry they pull from

## What changed

- **`AzureContainerAppRegistry.server` and `AzureContainerAppJobRegistry.server` are typed references to `AzureContainerRegistry`** (default wiring: the registry's `login_server` output), each returning at a new field number with the old one reserved. The Container Instance's identical field, `AzureContainerInstanceImageRegistryCredential.server`, has carried this shape from the start; an app or a job pulling the same image from the same registry took a plain hostname.
- The `string.min_len` rule leaves with the retype; the reference message's own rule already refuses an empty literal. The two auth-mode rules read `username` and `password_secret_name` and never `server`, so they stand unchanged.
- Both Pulumi modules read the literal through the reference; the four presets and the two e2e manifests that set a server carry it under `value:`; both catalog pages' Consumes tables name the registry.

## Why

Three Azure kinds pull private images and name their registry; two of them could not say which Planton-managed registry it was. The console offered no picker, the set lane could not order the registry before the workload, and the diagram drew the container group's line and not the apps'.

## How to check

```bash
go test ./catalog/azure/azurecontainerapp/v1alpha1/ ./catalog/azure/azurecontainerappjob/v1alpha1/
grep -n 'reserved 1;' catalog/azure/azurecontainerapp/v1alpha1/spec.proto catalog/azure/azurecontainerappjob/v1alpha1/spec.proto
```
