# AzureVpnServerConfiguration Pulumi Module

## Overview

This Pulumi module provisions a VPN Server Configuration using the
Azure Classic provider (`pulumi-azure`). It creates one
`network.VpnServerConfiguration` -- the reusable point-to-site
authentication policy -- plus one
`network.VpnServerConfigurationPolicyGroup` per `policy_groups` entry,
parented to the configuration.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- Policy groups are created as children (`pulumi.Parent`) of the
  configuration and their ARM IDs are republished keyed by the group's
  NAME (`policy_group_ids`) -- the family's name-keyed map convention.
- The spec's CEL contracts guarantee each enabled authentication type
  brings its block; the module wires blocks only when configured
  (never sends empty objects).
- `vpn_protocols` is omitted when the spec leaves it empty so ARM's
  default selection applies without read drift (Optional+Computed on
  the provider).
- The RADIUS server `secret` is a sensitive StringValueOrRef -- the
  platform resolves the reference before the module runs, and the
  provider schema masks the value in state and preview. ARM never
  returns it on reads.

## Inputs

The module receives an `AzureVpnServerConfigurationIacInput` containing:

- `target.spec.vpn_authentication_types` -- "AAD" / "Certificate" / "Radius"
- `target.spec.aad_authentication` / `client_root_certificates` / `radius` -- each type's parameters
- `target.spec.ipsec_policy` -- optional pinned client proposal
- `target.spec.policy_groups` -- named member-matching rules
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `vpn_server_configuration_id` | Full ARM ID -- what a point-to-site gateway's `vpn_server_configuration_id` references |
| `vpn_server_configuration_name` | The configuration's name |
| `policy_group_ids` | Each policy group's ARM ID keyed by group name |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
