# Premium Data Disk

This preset creates an empty zonal Premium SSD data disk -- the production default for database volumes and any data that must outlive its virtual machine. The VM attaches it through its `dataDiskAttachments` (referencing this disk's `disk_id` output with a LUN and caching mode), so replacing the VM never touches the data.

## When to Use

- Database data and log volumes
- Any stateful workload whose data must survive VM replacement, resizing, or rebuild
- Volumes that need Premium SSD's consistent per-size performance and credit-based bursting

## Key Configuration Choices

- **`storageAccountType: PREMIUM_LRS`** -- the production tier; drop to `STANDARD_SSD_LRS` for light workloads or lift to `PREMIUM_V2_LRS` when IOPS/throughput must be dialed independently of size
- **`createOption: EMPTY` + `diskSizeGb`** -- a blank volume; size can only ever increase, so start realistic rather than generous
- **`zone`** -- a zonal disk only attaches to VMs in the same zone; align it with the VM's zone (or use a ZRS SKU without a zone for zone-redundant data)
- **Attachment lives on the VM** -- declare which VM mounts this disk in the VM's `dataDiskAttachments`; the disk spec deliberately knows nothing about its consumers
- **`tier`** -- optionally buy a bigger size's performance (e.g. `P40` on a 256 GiB disk) for bursty workloads without over-provisioning capacity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the attaching VM's region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `workload-data` (`name`) | A name describing the DATA, not the VM | Your naming convention |
| `"1"` (`zone`) | The availability zone ("1", "2", or "3") matching the VM | Your zone layout |
