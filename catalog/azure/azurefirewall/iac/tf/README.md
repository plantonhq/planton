# AzureFirewall -- Terraform/OpenTofu Module

Creates an Azure Firewall (`azurerm_firewall`, azurerm ~> 5.0) in the referenced resource group, deployed into the referenced subnet (or Virtual WAN hub), fronted by the referenced public IPs, enforcing the referenced policy. Behaviorally identical to the Pulumi module for the same IaC input.

Credentials are injected by the runtime as `ARM_*` environment variables (the provider block is deliberately empty -- that is what enables keyless/OIDC auth).

Key behaviors, documented inline in `main.tf` and `locals.tf`:

- Enum values arrive as proto value names and are translated through explicit wire maps (`AZFW_VNET` -> `AZFW_VNet`, `STANDARD` -> `Standard`); the sku pair is always sent explicitly, threat-intel mode only when specified.
- DNS servers implicitly force the DNS proxy ON in Azure's wire encoding -- both flags pass through verbatim so the coupling stays Azure's.
- Classic inline rule collections are deliberately not created -- policy-based management is the modeled path.
- Outputs export the data-path `private_ip_address` (the hub-spoke seam spoke route tables consume) plus the hub-model addressing Azure assigns.
- The provider validates the subnet name segments (`AzureFirewallSubnet` / `AzureFirewallManagementSubnet`) at plan time after references resolve.
