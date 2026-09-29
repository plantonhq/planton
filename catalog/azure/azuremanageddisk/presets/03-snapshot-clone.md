# Snapshot Clone

This preset creates a disk by cloning an existing snapshot (or another managed disk) -- the restore, environment-duplication, and troubleshooting workhorse. The clone is a full, independent disk: writes to it never touch the source.

## When to Use

- Restoring a data volume from a nightly snapshot
- Duplicating a production dataset into a staging environment
- Forensic copies of a disk under investigation, attached to a clean VM

## Key Configuration Choices

- **`createOption: COPY` + `sourceResourceId`** -- accepts a snapshot ARM ID (the usual case) or a managed disk's ARM ID for a direct disk-to-disk clone
- **`diskSizeGb` omitted** -- the clone inherits the source's size (the actual value surfaces in the `disk_size_gb` output); set it larger to grow at creation, never smaller
- **SKU is free to differ** -- clone a Standard snapshot onto Premium for a performance upgrade during restore
- **`zone`** -- clones do not inherit placement; pin the zone to match the VM that will attach it
- **RESTORE is different** -- materializing an Azure Backup recovery point uses `createOption: RESTORE` with the recovery point's ID; COPY is for snapshots and disks

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the attaching VM's region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<snapshot-or-disk-arm-id>` | The snapshot or disk to clone | `az snapshot list -o table` or the portal |
| `"1"` (`zone`) | The availability zone matching the VM | Your zone layout |
