# AzureFirewallPolicy -- Pulumi Module

Creates an Azure Firewall Policy (`network.FirewallPolicy`, pulumi-azure classic v6) in the referenced resource group, with the full inspection/posture surface and merged governance tags. Behaviorally identical to the Terraform module for the same IaC input.

The entrypoint (`main.go`) loads the IaC input and delegates to `module.Resources`, which builds the Azure provider through the shared credential builder (static client secret, keyless web identity, or ambient chain). Rules are separate `AzureFirewallPolicyRuleCollectionGroup` resources -- this module deliberately creates only the policy.

Key behaviors, documented inline in `module/main.go`:

- The sku and threat-intelligence mode are always sent explicitly (Standard/Alert when unspecified) so both engines produce identical payloads.
- `auto_learn_private_ranges_enabled` is sent only when true -- Azure only records "Enabled" and disabling is by omission (sending false would read back as absent and churn state).
- Premium-only blocks (IDPS, TLS certificate) are spec-gated to the PREMIUM tier before the module ever runs.
- Outputs export the policy's ARM id (the join key for rule collection groups, firewalls, and child policies) and the system identity's principal id for Key Vault grants.
