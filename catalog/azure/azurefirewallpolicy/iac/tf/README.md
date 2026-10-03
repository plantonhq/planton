# AzureFirewallPolicy -- Terraform/OpenTofu Module

Creates an Azure Firewall Policy (`azurerm_firewall_policy`, azurerm ~> 5.0) in the referenced resource group, with the full inspection/posture surface and merged governance tags. Behaviorally identical to the Pulumi module for the same IaC input.

Credentials are injected by the runtime as `ARM_*` environment variables (the provider block is deliberately empty -- that is what enables keyless/OIDC auth).

Key behaviors, documented inline in `main.tf` and `locals.tf`:

- Enum values arrive as proto value names and are translated through explicit wire maps (`STANDARD` -> `Standard`, `IDPS_ALERT` -> `Alert`, ...); the sku and threat-intelligence mode fall back to Azure's own defaults, sent explicitly.
- `auto_learn_private_ranges_enabled` is sent only when true -- Azure's one-way "Enabled" semantics.
- Premium-only blocks (IDPS, TLS certificate) are spec-gated to the PREMIUM tier before the module ever runs.
- Outputs export the policy's ARM id (the join key for rule collection groups, firewalls, and child policies) and the system identity's principal id for Key Vault grants.
