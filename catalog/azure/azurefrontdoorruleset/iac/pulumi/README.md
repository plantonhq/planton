# AzureFrontDoorRuleSet -- Pulumi Module

Creates a `cdn.FrontdoorRuleSet` on the referenced Front Door profile
plus one `cdn.FrontdoorRule` per entry in `spec.rules` (pulumi-azure
classic v6), through the shared Azure provider builder (static client
secret, keyless web identity, or ambient chain).

## Behavior notes

- **Evaluation order is the rule's own `order` field** -- resource
  creation order carries no meaning; the parent reference gives the
  dependency edge.
- **Two protocol dialects**: the shared forwarding-protocol enum maps
  to `Http`/`Https` on the redirect action but `HttpOnly`/`HttpsOnly`
  on the route-configuration override -- both maps live in `locals.go`
  with the reason.
- **Address-condition operators default to `IPMatch`** when
  unspecified, and optional enums are sent only when chosen (IaC
  inputs never materialize proto defaults).
- **Header DELETE actions omit the value** -- the provider rejects an
  empty value on Append/Overwrite and any value on Delete; the spec's
  CEL guarantees the right combination arrives.
- **No Azure tags** -- ARM does not support tags on rule sets or rules.

## Outputs

- `rule_set_id` -- what routes reference in `rule_set_ids`
- `rule_set_name`

## Build

```shell
make build
```
