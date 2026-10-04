# AzureApplicationSecurityGroup - Pulumi Module

Pulumi implementation for the AzureApplicationSecurityGroup
kind.

## Architecture

```
network.ApplicationSecurityGroup (one empty, named NIC grouping)
```

## Key Design Decisions

- **The ASG is an empty anchor by design** -- membership is declared
  from the NIC / scale-set IP-configuration side and rule usage from
  the NSG side, matching ARM's own model; the module creates only the
  named group.
- **Everything except tags is ForceNew** -- rename or region change
  replaces the group and breaks every referrer, so the module keeps the
  update surface honest rather than papering over it.
- **User tags merge over metadata tags** (user wins), with the family's
  documented PARITY-EXCEPTION on `resource_kind` / `resource_id` tag
  shape versus the Terraform module -- output-neutral.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
