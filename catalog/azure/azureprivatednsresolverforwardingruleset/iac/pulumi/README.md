# AzurePrivateDnsResolverForwardingRuleset Pulumi Module

## Overview

Creates a DNS forwarding ruleset -- the rule book that steers DNS queries for chosen domains out of Azure through a DNS Private Resolver's outbound endpoints -- and its forwarding rules, on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module.

## Resources Created

- `privatedns.ResolverDnsForwardingRuleset` -- the ruleset, bound to the resolver's outbound endpoint(s)
- `privatedns.ResolverForwardingRule` -- one per `spec.forwarding_rules` entry

## Outputs

- `dns_forwarding_ruleset_id` -- the ruleset's ARM resource ID (what virtual network links reference)
- `dns_forwarding_ruleset_name` -- the ruleset's name

## Behavior Notes

- **A ruleset binds at most 2 outbound endpoints, both from the SAME resolver** -- Azure enforces it at deploy time.
- **Domain names are fully qualified WITH the trailing dot** (`corp.contoso.com.`) -- ARM stores them that way; write them that way in the spec.
- **Everything on a rule updates in place EXCEPT `domain_name`**, which replaces that rule (Pulumi resource names key on the rule name, so siblings are untouched).
- **The port is always sent explicitly** -- 53 when the spec leaves it unset -- so both engines produce identical wire shapes.
- **Rules carry NO tags** -- ARM gives them a free-form `metadata` map instead, modeled per rule.
- **Rulesets and rules are free at rest** -- only endpoint hours and query volume bill.

## Engine Parity

ZERO parity exceptions: the classic SDK v6.38.0 carries the complete v5 argument surface for both resources, verified field-by-field at planning.
