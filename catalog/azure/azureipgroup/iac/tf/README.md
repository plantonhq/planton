# AzureIpGroup -- Terraform/OpenTofu Module

Creates an Azure IP Group (`azurerm_ip_group`, azurerm ~> 5.0) in the referenced resource group, carrying the spec's address set and merged governance tags. Behaviorally identical to the Pulumi module for the same IaC input.

Credentials are injected by the runtime as `ARM_*` environment variables (the provider block is deliberately empty -- that is what enables keyless/OIDC auth).

Key behaviors, documented inline in `main.tf`:

- The group is a passive address set; consumption is declared from the rule's side (firewall policy rules reference `ip_group_id`), so the module creates exactly one resource.
- `cidrs` updates in place -- an address change retargets every referencing rule without recreating anything.
- Renaming or moving the group replaces it, and every rule that referenced it must be re-pointed.
