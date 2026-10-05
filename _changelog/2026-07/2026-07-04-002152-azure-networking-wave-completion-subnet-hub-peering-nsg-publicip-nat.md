# Azure Networking Wave Completion: Subnet Attach Hub, VNet Peering, Public IP Prefix, and the NSG/PublicIp/NAT Reworks

**Date**: July 4, 2026
**Type**: Feature + Breaking Change
**Components**: Azure Provider, API Definitions, IAC Modules, E2E Framework

## Summary

The Azure networking wave closes with six kinds in one coherent story: the
**subnet attach model**. `AzureSubnet` becomes the composition hub Azure itself
designed it to be -- route tables, network security groups, and NAT gateways now
attach *to the subnet* through first-class foreign keys realized as association
resources on both engines. Two new kinds join the catalog:
`AzureVirtualNetworkPeering` (420, one resource per peering direction) and
`AzurePublicIpPrefix` (421, contiguous reserved address ranges). Three existing
kinds are reworked to the full `azurerm` v4.80 floor: `AzureNetworkSecurityGroup`
(closed enums, port/address XOR, application security groups),
`AzurePublicIp` (SKU/tier/version/DNS-scope/DDoS/prefix allocation), and
`AzureNatGateway` -- whose bundled create-an-IP-internally black box is dissolved
into explicit references to first-class addresses. The shared E2E framework gains
**scenario-declared extra fixtures**, the mechanism that lets a scenario prove
OPTIONAL composition seams without corrupting registry prerequisites.

## Problem Statement / Motivation

The wave opener (virtual network, route table, DNS zone links) shipped the
building blocks; this session ships the joints:

- `AzureSubnet` was an 8-field spec with a single CIDR, a single delegation, and
  NO attach seams -- the route table's `route_table_id` output had nothing to
  consume it, NSGs could not be associated, and NAT gateways attached themselves
  *to* subnets (the wrong direction; Azure's model is subnet-side).
- `AzureNatGateway` silently created up to SIX resources from one spec: an
  internal public IP prefix, an internal public IP, the gateway, two IP
  associations, and a subnet association. The addresses were invisible to the
  resource graph, unallowlistable, and unreusable.
- `AzureNetworkSecurityGroup` invented a "plural takes precedence" port contract
  azurerm actually rejects as a conflict, and its protocol set was missing `Ah`
  and `Esp`.
- `AzurePublicIp` modeled a fraction of the real surface (no SKU, no IP version,
  no DNS-label scoping, no DDoS stance, no prefix allocation).
- VNet peering and public IP prefixes did not exist at all -- no hub-spoke
  story, no contiguous-SNAT-range story.
- The E2E framework could only deploy fixtures from registry `prerequisites`,
  which correctly means "required parents." Optional seams -- the entire point
  of this wave -- had no honest live-proof mechanism.

## Solution / What's New

### Framework: scenario-declared extra E2E fixtures

A scenario manifest may now carry the `planton.dev/e2e-extra-prerequisites`
annotation listing kind names (resolved through each kind's install profile)
and/or repo-relative manifest paths (for extra instances, e.g. a peering's
second network). Extra fixtures deploy after registry prerequisites, bring their
own transitive prerequisites (deduplicated against the chain), join the same
reference resolution, and tear down in reverse. Registry `prerequisites` keep
their strict deploy-ordering semantics. Unit-tested in
`e2e/framework/runner/dependencies_test.go`; documented in `e2e/README.md` and
the forge rule.

### AzureSubnet (rework, enum 411, id_prefix `azsub`)

The full `azurerm_subnet` + association surface:

- `address_prefixes` (repeated, dual-stack capable) XOR `ip_address_pool`
  (Network Manager IPAM), enforced by message CEL
- Repeated `delegations`; `service_endpoints` + `service_endpoint_policy_ids`
- `private_endpoint_network_policies` as a closed enum;
  `private_link_service_network_policies_enabled`,
  `default_outbound_access_enabled` (optional bools defaulting true),
  `sharing_scope` with the ARM constraint (requires explicit outbound-false) as CEL
- **The attach seams**: `route_table_id`, `network_security_group_id`, and
  `nat_gateway_id` FKs; both engines create the corresponding
  `azurerm_subnet_*_association` / `Subnet*Association` resources
- Parent modeled as a single `virtual_network_id` FK -- resource group and
  network name derived from the ARM id on both engines
- Outputs: `subnet_id`, `subnet_name`, repeated `address_prefixes` (actual
  ranges, IPAM-visible), `virtual_network_name`, `resource_group_name`

### AzureVirtualNetworkPeering (new, enum 420, id_prefix `azpeer`)

One resource = one peering direction (Azure's model; presets teach the
hub-spoke pair): local `virtual_network_id` parent FK (RG/name derived from the
ARM id), `remote_virtual_network_id` FK, the four connectivity flags with
Azure's defaults, subnet-scoped peering (`peer_complete_virtual_networks_enabled`
+ local/remote subnet names), `only_ipv6_peering_enabled`. No tags -- ARM does
not support them on peerings.

### AzurePublicIpPrefix (new, enum 421, id_prefix `azpippfx`)

The thin contiguous-range kind: `prefix_length` (28-31), `ip_version`, `sku`,
`sku_tier`, `zones`, `tags`. Outputs the allocated `ip_prefix` CIDR -- the value
partners allowlist once. Referenced by both `AzurePublicIp.public_ip_prefix_id`
(draw an address from the range) and `AzureNatGateway.public_ip_prefix_ids`
(SNAT through the range).

### AzureNetworkSecurityGroup (rework, enum 412)

`direction`/`access`/`protocol` become closed enums (protocol gains `AH` and
`ESP` -- azurerm allows six values, not four); singular vs plural ports and
address prefixes are hard XORs via CEL (azurerm's actual contract);
`source/destination_application_security_group_ids` added; user `tags` added;
outputs renamed to `network_security_group_id`/`network_security_group_name`.

### AzurePublicIp (rework, enum 413)

The v4.80 floor: `sku` (STANDARD/STANDARD_V2 -- Basic is retired and not
modeled), `sku_tier` (REGIONAL/GLOBAL with the Standard-only CEL), `ip_version`,
`domain_name_label` + `domain_name_label_scope` (hashed-label takeover defense),
`reverse_fqdn`, `ip_tags`, `public_ip_prefix_id` FK, `ddos_protection_mode` +
`ddos_protection_plan_id` (paired via CEL), `edge_zone`, user `tags`. Allocation
stays Static -- every current SKU requires it.

### AzureNatGateway (rework, enum 407)

The black box is dissolved: `subnet_id` REMOVED (subnets attach themselves),
`public_ip_prefix_length` REMOVED (the prefix is a first-class kind). The
gateway now references `public_ip_ids` and `public_ip_prefix_ids` (repeated
FKs); modules create only the gateway plus the two association resource types.
Added `name`, `sku_name` (STANDARD / STANDARD_V2 with the zones-must-be-empty
CEL), `zones`; `idle_timeout_in_minutes` kept; `tags` converged to the
merge-over-metadata grain.

## Implementation Details

- **Both engines at parity** on the shared keyless-capable provider builder
  (`pulumiazureprovider.Get`) -- subnet, NSG, public IP, and NAT migrated off
  inline `NewProvider` (10 of ~42 Azure modules now on the builder).
- **Presence-guarded optional fields in Pulumi**: unset optional ints/bools
  (`idle_timeout_in_minutes`, the subnet's two true-default bools, the peering's
  true-default flags) now fall back to the proto default explicitly instead of
  the getter's zero value, matching the Terraform modules' `optional(...)`
  encodings on every input path.
- **Documented tag-shape exception**: the Terraform modules' `resource_kind`
  snake-case literal and `resource_id` name-fallback vs the Pulumi modules'
  lowered enum string and id-only emission is recorded as an output-neutral
  `PARITY-EXCEPTION` in both modules of all four tag-bearing kinds; aligning the
  shapes is a family-wide convention decision.
- **FK graph**: subnet's three attach FKs point at the route table, NSG, and NAT
  outputs; NAT points at the public IP and prefix outputs; public IP points at
  the prefix output. `validate-refs --check` green.
- **E2E**: six verifiers on the generic ARM GetByID pattern; the subnet scenario
  is the wave's composed showcase (RG → VNet → route table + NSG + NAT →
  subnet with all three seams attached); the NAT scenario associates a real
  public IP AND a prefix (zone-matched fixtures); the peering scenario peers the
  fixture network to a second, path-declared network.

## Validation

- Offline: spec tests ×6 green; targeted + release-equivalent builds;
  `make build-go`; gazelle regen; `secret-coverage --check` (Azure 100%; none of
  the six carries secret material); `validate-refs --check`; `pkg/outputs`
  conformance cases ×6; `tofu init/validate/fmt` + full `planton tofu plan` on
  all six hack manifests; every preset validates against its reworked spec;
  audits PARITY/COVERAGE-gated for all six kinds.
- Live (test subscription): all 12 scenario runs green (6 kinds × Pulumi +
  OpenTofu), including the composed subnet showcase and the NAT
  address-association proof. One real finding fixed mid-run: ARM rejects
  associating non-zonal addresses with a zonal Standard NAT gateway
  (`PublicIPOrPrefixAndStandardSkuNatGatewayZoneDoNotMatch`) -- the address
  fixtures are now zone-pinned to match. Zero orphans after the suite.

## Impact

Azure networking on Planton now composes the way Azure itself composes: networks
partition into subnets; subnets declare their routing, filtering, and egress by
reference; gateways SNAT through visible, allowlistable, reusable addresses;
networks peer explicitly per direction. Every seam is graphable, and every seam
is live-proven on both engines.

## Related Work

- Builds on the networking wave opener (virtual network rework + route table +
  private DNS zone links) -- the route table's `route_table_id` output shipped
  there specifically for this session's subnet seam.
- The scenario-declared extra fixtures mechanism benefits every provider's E2E
  suite, not just Azure.

---

**Status**: ✅ Production Ready
