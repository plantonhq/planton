# AzureFirewallPolicyRuleCollectionGroup -- Terraform/OpenTofu Module

Creates a firewall policy rule collection group (`azurerm_firewall_policy_rule_collection_group`, azurerm ~> 5.0) nested under the referenced policy, carrying the spec's application, network, and DNAT collections. Behaviorally identical to the Pulumi module for the same IaC input.

Credentials are injected by the runtime as `ARM_*` environment variables (the provider block is deliberately empty -- that is what enables keyless/OIDC auth).

Key behaviors, documented inline in `main.tf` and `locals.tf`:

- Enum values arrive as proto value names and are translated through explicit wire maps (ALLOW -> "Allow", ANY -> "Any", HTTPS -> "Https"); the DNAT collection's action is the one-value constant "Dnat", sent unconditionally.
- IP Group references arrive as resolved ARM ids and pass through verbatim.
- The provider locks the parent policy for every write, so concurrent group deployments against one policy queue rather than conflict.
- No tags: ARM does not support tags on rule collection groups.
