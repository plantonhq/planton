# AzureDnsRecord -- Terraform Module

Creates one DNS record set in an Azure public DNS zone. The record type is whichever typed payload the spec carries (spec validation guarantees exactly one), so exactly one of the module's nine count-gated `azurerm_dns_*_record` resources materializes.

The module receives its inputs from the Planton iac-input contract (`metadata` + `spec` variables); `StringValueOrRef` fields (resource group, zone name, alias targets) arrive pre-resolved as strings.

Key behaviors, documented inline in `main.tf` and `locals.tf`:

- Typed payloads carry DNS's real value shapes: MX entries are (preference, exchange) pairs, SRV entries are priority/weight/port/target, CAA entries are flags/tag/value -- every field user-declared, never synthesized.
- A/AAAA/CNAME support Azure alias records (`target_resource_id` XOR literal values); empty collections pass as null so the provider never sees an empty-but-present argument.
- The CAA tag enum maps to Azure's lowercase wire vocabulary; the provider's string-typed MX preference is converted from the spec's integer.
- Outputs coalesce per attribute across the nine variants (`record_id`, `fqdn`).
