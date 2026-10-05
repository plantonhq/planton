# AzureFirewall -- Pulumi Module

Creates an Azure Firewall (`network.Firewall`, pulumi-azure classic v6) in the referenced resource group, deployed into the referenced subnet (or Virtual WAN hub), fronted by the referenced public IPs, enforcing the referenced policy. Behaviorally identical to the Terraform module for the same IaC input.

The entrypoint (`main.go`) loads the IaC input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain).

Key behaviors, documented inline in `module/main.go`:

- The sku pair is always sent explicitly (AZFW_VNet/Standard when unspecified); threat-intel mode only when specified (Azure defaults it to Alert).
- DNS servers implicitly force the DNS proxy ON in Azure's wire encoding -- both flags pass through verbatim so the coupling stays Azure's.
- Classic inline rule collections are deliberately not created -- policy-based management is the modeled path.
- Outputs export the data-path `private_ip_address` (the hub-spoke seam spoke route tables consume) plus the hub-model addressing Azure assigns.
- Azure Firewall provisions and deletes in 10-20+ minutes each way; the ForceNew surface is documented on the spec.
