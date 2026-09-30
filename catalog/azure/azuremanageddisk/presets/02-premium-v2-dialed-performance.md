# Premium SSD v2 with Dialed Performance

This preset creates a Premium SSD v2 data disk whose capacity, IOPS, and throughput are provisioned independently -- a small disk with big performance, impossible on the classic per-size tiers. It is the modern choice for performance-sensitive databases that do not need huge capacity.

## When to Use

- Databases and transactional workloads where IOPS matters more than capacity
- Replacing over-provisioned Premium SSDs bought purely for their performance tier
- Workloads whose performance needs change over time (both dials update in place)

## Key Configuration Choices

- **`storageAccountType: PREMIUM_V2_LRS`** -- data disks only (OS disks stay on classic Premium); zonal only, so the `zone` is required in zoned regions
- **`diskIopsReadWrite` / `diskMbpsReadWrite`** -- the independent dials; baseline is free (3000 IOPS / 125 MBps), beyond that bills per provisioned unit. Both update in place as needs change
- **`diskSizeGb: 128`** -- pay for the capacity the data actually needs; performance no longer forces the size
- **Ultra Disk instead?** -- `ULTRA_SSD_LRS` reaches a higher performance envelope but requires VM-family support and a zonal deployment; Premium v2 covers most needs with fewer constraints
- **Shared-disk clustering** -- add `maxShares` (and the read-only dials) to attach the disk to several VMs at once for failover clusters

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must support Premium SSD v2) | Azure regional availability docs |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `workload-data` (`name`) | A name describing the DATA, not the VM | Your naming convention |
| `"1"` (`zone`) | The availability zone matching the VM | Your zone layout |
