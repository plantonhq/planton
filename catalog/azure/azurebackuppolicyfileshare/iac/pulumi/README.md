# AzureBackupPolicyFileShare Pulumi Module

## Overview

Creates an Azure Backup policy for Azure Files shares -- the schedule and layered daily/weekly/monthly/yearly retention rules -- via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). The policy is a free configuration object, an ARM child of its Recovery Services vault.

## Resources Created

- `backup.PolicyFileShare` -- the policy (`.../vaults/{vault}/backupPolicies/{name}`)

## Outputs

- `backup_policy_id` -- the policy's full ARM ID (what protected file shares bind to)
- `backup_policy_name` -- unique on its vault

## Behavior Notes

- **Full engine parity**: the classic SDK carries the complete v5 policy surface (both schedule shapes, the hourly window, backup tier, snapshot retention, all retention forms) -- ZERO parity exceptions on this kind.
- **Daily or Hourly only** -- file-share policies have no Weekly schedule (the provider's own surface). The schedule SHAPE follows the frequency (spec CEL): Daily names its `time`; Hourly configures the `hourly` window instead.
- **`retention_daily` is always required** (the provider's own contract); counts run 1-200 (monthly 1-120, yearly 1-10 -- shorter than VM retention).
- **`backup_tier` picks where backups live**: `snapshot` (default) or `vault-standard`; `snapshot_retention_in_days` applies to vault-standard only and must be LESS THAN `retention_daily.count` (spec CEL mirrors the provider's CustomizeDiff).
- **No tags**: ARM backup policies carry no tags -- deliberately no tag map in this module, unlike the vault sibling.

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
