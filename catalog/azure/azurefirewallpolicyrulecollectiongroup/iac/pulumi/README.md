# AzureFirewallPolicyRuleCollectionGroup -- Pulumi Module

Creates a firewall policy rule collection group (`network.FirewallPolicyRuleCollectionGroup`, pulumi-azure classic v6) nested under the referenced policy, carrying the spec's application, network, and DNAT collections. Behaviorally identical to the Terraform module for the same IaC input.

The entrypoint (`main.go`) loads the IaC input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain).

Key behaviors, documented inline in `module/main.go`:

- Enum vocabularies are translated through explicit wire maps (ALLOW -> "Allow", ANY -> "Any", HTTPS -> "Https"); the DNAT collection's action is the one-value constant "Dnat", sent unconditionally.
- IP Group references (source/destination) arrive as resolved ARM ids and pass through verbatim -- the reusable alternative to literal address lists.
- The NAT rule's destination_ports list is capped at one entry by the spec (ARM's own cap); the bridge models it as a singular string and the module sends the single entry.
- No tags: ARM does not support tags on rule collection groups.
