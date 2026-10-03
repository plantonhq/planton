# AzureDnsRecord -- Pulumi Module

Creates one DNS record set in an Azure public DNS zone (pulumi-azure classic v6, `dns.*Record`). The record type is whichever typed payload the spec carries (spec validation guarantees exactly one), so exactly one branch of the module's type dispatch runs. Behaviorally identical to the Terraform module for the same IaC input.

The entrypoint (`main.go`) loads the IaC input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain).

Key behaviors, documented inline in `module/main.go`:

- Typed payloads carry DNS's real value shapes: MX entries are (preference, exchange) pairs, SRV entries are priority/weight/port/target, CAA entries are flags/tag/value -- every field user-declared, never synthesized.
- A/AAAA/CNAME support Azure alias records (`target_resource_id` XOR literal values); unused arguments stay nil so the provider's exactly-one contract is satisfied.
- The CAA tag enum maps to Azure's lowercase wire vocabulary; the SDK's string-typed MX preference is converted from the spec's integer.
- Outputs export the record set's ARM id and fqdn, identical to the Terraform module.
