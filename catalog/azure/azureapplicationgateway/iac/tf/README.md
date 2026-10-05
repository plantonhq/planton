# AzureApplicationGateway - Terraform Module

Terraform implementation for the AzureApplicationGateway
kind.

## Resources Created

- `azurerm_application_gateway.main` -- the gateway with every sub-object
  inline: SKU/autoscale, the derived gateway IP configuration, frontends,
  ports, pools, backend settings, listeners (L7 + L4), routing rules
  (L7 + L4), URL path maps, probes, certificates (SSL / trusted root /
  trusted client), SSL profiles and policies, redirects, rewrite rule
  sets, Private Link configurations, custom error pages, buffering, and
  tags

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.sku` | Spec enum name string (`BASIC`/`STANDARD_V2`/`WAF_V2`); name and tier both derive from it |
| `spec.*.protocol` | One shared vocabulary (`HTTP`/`HTTPS`/`TCP`/`TLS`) mapped in `locals.tf`; spec validation restricts each block to its layer |
| `spec.request_routing_rules[].rule_type` | `BASIC_ROUTING`/`PATH_BASED_ROUTING` → ARM's `Basic`/`PathBasedRouting` |
| `spec.backend_http_settings[].cookie_based_affinity_enabled` | Boolean → ARM's `Enabled`/`Disabled` string |
| `spec.ssl_policy` / `ssl_profiles[].ssl_policy` | Policy type + TLS versions as enum name strings, mapped through exhaustive lookups |
| `spec.frontend_ip_configurations[]` | Public XOR private (spec-validated); the gateway IP configuration name is derived (`{name}-gateway-ip`), matching the Pulumi module |
| `spec.routing_rules[].backend_settings_name` | Maps to azurerm's `routing_rule.backend_name` (the L4 settings reference) |

## Outputs

`application_gateway_id`, `application_gateway_name`, name-keyed
`backend_address_pool_ids` and `frontend_ip_configuration_ids` maps, and
the private frontends' `private_ip_address`/`private_ip_addresses`.

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation
across the full surface: L7 and L4 routing, TLS/mTLS, rewrites,
redirects, WAF policy attachment at all three levels, Private Link,
buffering, and the map outputs. Three azurerm backend-settings fields
(`sni_name`, `sni_validation_enabled`,
`certificate_chain_validation_enabled`) are deliberately unmodeled -- the
pulumi SDK cannot express them yet (recorded in `../../GUIDE.md`).
