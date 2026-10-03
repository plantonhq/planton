# AzureBackupPolicyVm Pulumi Module

## Overview

Creates an Azure Backup policy for IaaS virtual machines -- the schedule and layered daily/weekly/monthly/yearly retention rules -- via the classic Pulumi Azure provider (`pulumi-azure/sdk/v6`, bridged from azurerm). The policy is a free configuration object, an ARM child of its Recovery Services vault.

## Resources Created

- `backup.PolicyVM` -- the policy (`.../vaults/{vault}/backupPolicies/{name}`)

## Outputs

- `backup_policy_id` -- the policy's full ARM ID (what protected VMs bind to)
- `backup_policy_name` -- unique on its vault

## Behavior Notes

- **Full engine parity**: the classic SDK carries the complete v5 policy surface (backup schedule, hourly dials, tiering, both retention forms, consistency type) -- ZERO parity exceptions on this kind.
- **Frequency decides the retention layers** (spec CEL mirrors the provider's CustomizeDiff): Hourly/Daily schedules require `retention_daily`; Weekly requires `retention_weekly` and forbids `retention_daily`. Hourly needs `policy_type` V2.
- **Azure rejects 2-6 days of daily retention** at create -- front-loaded in the spec: count is 1, or 7+.
- **`policy_type` is ForceNew** -- changing the policy generation replaces the policy.
- **No tags**: ARM backup policies carry no tags -- deliberately no tag map in this module, unlike the vault sibling.

## Development

```bash
go build ./...
```

The module entrypoint is `main.go` at this directory's root (the release contract); the implementation lives in `module/`.
