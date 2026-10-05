# AzureIpGroup -- Pulumi Module

Creates an Azure IP Group (`network.IPGroup`, pulumi-azure classic v6) in the referenced resource group, carrying the spec's address set and merged governance tags. Behaviorally identical to the Terraform module for the same IaC input.

The entrypoint (`main.go`) loads the IaC input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain).

Key behaviors, documented inline in `module/main.go`:

- The group is a passive address set; consumption is declared from the rule's side (firewall policy rules reference `ip_group_id`), so the module creates exactly one resource.
- `cidrs` updates in place -- an address change retargets every referencing rule without recreating anything.
- Renaming or moving the group replaces it, and every rule that referenced it must be re-pointed.
