# AzureDnsZone -- Terraform Module

Creates an Azure public DNS zone (`azurerm_dns_zone`) in the referenced resource group, with optional Start of Authority customization and merged governance tags.

The module receives its inputs from the Planton iac-input contract (`metadata` + `spec` variables); `StringValueOrRef` fields arrive pre-resolved as strings. Records are separate `AzureDnsRecord` resources -- this module deliberately creates only the zone.

Key behaviors, documented inline in `main.tf`:

- Renaming the zone replaces it, every record in it, and the assigned name-server set (breaking registrar delegation until updated).
- The SOA block is only sent when the spec customizes it; unset timers fall back to Azure's defaults, and the SOA host name is never sent (Azure owns it).
- Outputs export the delegation handoff (`nameservers`) and the record-addressing join key (`zone_name` + `resource_group_name`).
