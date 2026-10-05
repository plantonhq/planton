# AzureAvailabilitySet Pulumi Module

## Overview

Creates an availability set -- the classic pre-zones placement grouping that spreads VMs across separate fault domains and update domains so one hardware failure or maintenance window cannot take them all down.

## Resources Created

- `compute.AvailabilitySet` -- the availability set (domain counts, managed alignment, optional proximity placement group, tags)

## Outputs

- `availability_set_id` -- the set's ARM resource ID (what VMs reference to join it)
- `availability_set_name` -- the set's name

## Behavior Notes

- **The whole configuration is fixed at creation** -- only tags update in place; everything else forces replacement.
- **Unset optionals ride the provider defaults**: 5 update domains, 3 fault domains, `managed = true`. The module sends fields only when set.
- **`managed = true` aligns fault domains with managed-disk storage** -- the right value for every managed-disk VM. Some regions support fewer than 3 fault domains; Azure rejects a count the region cannot provide.
- **An availability set is free** -- it bills nothing itself.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureAvailabilitySet` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.
