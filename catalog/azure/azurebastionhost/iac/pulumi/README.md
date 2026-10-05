# AzureBastionHost Pulumi Module

## Overview

Creates an Azure Bastion host -- the managed jump service that opens RDP/SSH sessions to virtual machines over their private addresses -- on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module.

## Resources Created

- `compute.BastionHost` -- the Bastion host. The classic SDK's bridge parks this resource in the COMPUTE package (token `azure:compute/bastionHost:BastionHost`), not network where the ARM resource lives -- the import path is correct as written.

## Outputs

- `bastion_host_id` -- the host's ARM resource ID
- `bastion_host_name` -- the host's name
- `dns_name` -- the DNS name sessions connect through
- `private_only_enabled` -- whether the host deployed private-only (a Premium host without a public IP)

## Behavior Notes

- **The subnet must be named EXACTLY `AzureBastionSubnet`** and sized /26 or larger -- ARM validates the name at deploy time (the requirement lives on the referenced subnet, so no offline gate can check it).
- **The host binds its public IP EXCLUSIVELY** (Standard SKU, static allocation) -- never share the address with another consumer.
- **SKU upgrades are in-place; downgrades REPLACE the host** (the provider forces a new resource -- Azure has no downgrade path).
- **`kerberos_enabled` is applied at CREATE only** -- the provider has no update path for it and silently ignores later changes; plan Kerberos up front or replace the host to change it.
- **A Premium host without a public IP deploys PRIVATE-ONLY** (reachable only from connected networks) -- surfaced in the `private_only_enabled` output.
- **Billing starts at provisioning** (~10-minute creates): hourly per SKU, plus per-scale-unit on Standard/Premium. Developer is free.
- **Engine parity**: the classic SDK v6.38.0 carries the FULL azurerm v5 surface for this kind -- zero parity exceptions.
