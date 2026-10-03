# AzureFrontDoorFirewallPolicy -- Pulumi Module

Creates a `cdn.FrontdoorFirewallPolicy` (pulumi-azure classic v6) in the
referenced resource group, through the shared Azure provider builder
(static client secret, keyless web identity, or ambient chain).

## Behavior notes

- **Global resource** -- ARM fixes the location; no region is sent.
- **Challenge lifetimes are Premium-conditional**: on Premium the
  module always sends the JS-challenge and CAPTCHA lifetimes
  (spec value or the documented default of 30) because Azure enables
  both policies there unconditionally; on Standard nothing is sent
  (Azure rejects the fields -- the spec's CELs keep values out).
- **Provider defaults pass through**: enabled, request-body check,
  custom-rule enabled/priority and the rate-limit pair are sent only
  on an explicit spec choice (IaC inputs never materialize proto
  defaults).
- **Enum prefixes are proto-local**: `RULE_SET_*` / `OVERRIDE_*` /
  `SELECTOR_*` / `EXCLUDE_*` / `SCRUB_*` values map to ARM's bare
  names in `locals.go`.
- **Managed-rule override `enabled` defaults FALSE** on the provider --
  listing a rule is the disable gesture; the module lets that default
  through.

## Outputs

- `firewall_policy_id` -- what AzureFrontDoorSecurityPolicy references
  in `firewall_policy_id`
- `firewall_policy_name`

## Build

```shell
make build
```
