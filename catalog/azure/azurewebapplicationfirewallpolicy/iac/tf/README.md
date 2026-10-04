# AzureWebApplicationFirewallPolicy - Terraform Module

Terraform implementation for the AzureWebApplicationFirewallPolicy
kind.

## Resources Created

- `azurerm_web_application_firewall_policy.main` -- the policy: custom
  rules, managed-rule configuration (rule sets, overrides, exclusions),
  policy settings (mode, body dials, log scrubbing), and tags

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.custom_rules[].rule_type` / `action` | Spec enum name strings (`MATCH_RULE`/`RATE_LIMIT_RULE`, `ALLOW`/`BLOCK`/`LOG`/`JS_CHALLENGE`) mapped through exhaustive lookups in `locals.tf` |
| `spec.custom_rules[].match_conditions` | Variable/operator/transform enums as name strings; unknown values fail the plan loudly |
| `spec.managed_rules.managed_rule_sets[].type` | Unset applies `OWASP` (azurerm's default, materialized identically on both engines) |
| `spec.managed_rules.exclusions[].selector_match_operator` | `SELECTOR_*` name strings mapped to ARM's `Equals`/`Contains`/... |
| `spec.policy_settings.mode` | Unset applies `Prevention` |
| `spec.policy_settings.log_scrubbing.rules[].match_variable` | `SCRUB_*` name strings mapped to ARM's scrubbing variables |

## Usage

```hcl
module "waf_policy" {
  source = "./path/to/module"

  metadata = {
    name = "waf-baseline"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "network-rg"
    policy_name    = "org-waf-baseline"

    managed_rules = {
      managed_rule_sets = [
        { version = "3.2" }
      ]
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
custom match and rate-limit rules, managed rule sets with per-rule
overrides and scoped exclusions, the full policy-settings surface including
log scrubbing, and user tags. `file_upload_enforcement` is forwarded only
when explicitly set on both engines (it is only honored with OWASP 3.2).
