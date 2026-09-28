# AzureDnsZone -- Pulumi Module

Creates an Azure public DNS zone (`dns.Zone`, pulumi-azure classic v6) in the referenced resource group, with optional Start of Authority customization and merged governance tags. Behaviorally identical to the Terraform module for the same stack input.

The entrypoint (`main.go`) loads the stack input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain). Records are separate `AzureDnsRecord` resources -- this module deliberately creates only the zone.

Key behaviors, documented inline in `module/main.go`:

- Renaming the zone replaces it, every record in it, and the assigned name-server set (breaking registrar delegation until updated).
- The SOA block is only sent when the spec customizes it; unset timers fall back to Azure's defaults, and the SOA host name is never sent (Azure owns it).
- Outputs export the delegation handoff (`nameservers`) and the record-addressing join key (`zone_name` + `resource_group_name`).
