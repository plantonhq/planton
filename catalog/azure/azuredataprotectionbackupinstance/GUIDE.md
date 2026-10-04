# Azure Data Protection Backup Instance -- Operational Guide

Judgment calls that matter when you run backup instances in production.

## Grants first, instance second -- always

Azure validates the vault identity's datasource permissions at create time, and role assignments propagate asynchronously (minutes, occasionally longer). Deploy the AzureRoleAssignment resources BEFORE the instance -- in a chart, the reference wiring orders them; standalone, wait for the grants to land. An authorization-class create failure ("appropriate permissions", "does not have authorization") means missing or still-propagating grants, not a broken configuration. Retrying after a few minutes is the first move.

## The policy's variant must match the instance's

A disk instance binds a DISK policy; a blob instance a BLOB policy -- both on the same vault. Azure rejects mismatches at create. When you replace a policy (policies are immutable -- every change replaces them), update the instance's `backupPolicyId` in the same change: it is the instance's only in-place-updatable field, except on the Kubernetes variant where it replaces the instance.

## Protection configured is not a restore point

Creating the instance registers protection; the FIRST backup runs on the policy's schedule. Until it completes there is nothing to restore. For anything you are about to change destructively, check the vault's restore points first -- do not assume day-zero coverage.

## Deletion deletes the backups -- plan the exit

Destroying the instance stops protection AND removes the backup data. With the vault's soft delete on (the default), the data lingers as a soft-deleted item for 14 days -- recoverable, but it also HOLDS the vault's deletion for that window, and re-protecting the same datasource inside the window collides with the ghost. For environments that create and destroy protection frequently (test estates), a vault with soft delete Off trades the safety net for clean teardown.

## The Kubernetes variant is a bigger commitment

The AKS variant needs the Backup extension installed on the cluster and a trusted-access role binding to the vault before the instance can be created -- cluster-side setup this kind deliberately does not own. It is also immutable end to end: every change, including the policy binding, replaces the instance. Model AKS backup as part of cluster provisioning, not an afterthought.
