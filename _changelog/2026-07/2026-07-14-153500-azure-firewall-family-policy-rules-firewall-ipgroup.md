# Azure Firewall Family: Policy, Rule Collection Groups, Firewall, and IP Groups

**Date**: 2026-07-14
**Type**: Feature (new catalog kinds)
**Scope**: `apis/dev/planton/provider/azure/{azurefirewallpolicy,azurefirewallpolicyrulecollectiongroup,azurefirewall,azureipgroup}/v1`, kind registry, Azure E2E harness, outputs conformance

## Summary

Four new Azure catalog kinds bring centralized network security to
the catalog, modeled on Azure's own separation of concerns: a reusable
**Firewall Policy** (the rule-and-inspection document), independently
deployable **Rule Collection Groups** (the rules, one group per team or
application), the **Azure Firewall** instance (the data plane in a
dedicated subnet or Virtual WAN hub), and **IP Groups** (reusable address
sets that rules reference instead of repeating literal CIDR lists). They
open the network-security enum sub-band 530-539.

Both engines (OpenTofu on `azurerm ~> 4.0`, Pulumi on the shared
`pulumiazureprovider` builder with pulumi-azure v6) implement the same
contract at 100% behavioral parity, live-proven end to end.

## What Users Can Do Now

- **Author security policy once, enforce it everywhere**: a policy carries
  threat intelligence (alert/deny + allowlist), DNS proxying, SNAT
  posture, SQL redirect, explicit proxy, policy analytics into Log
  Analytics, and -- on the PREMIUM tier -- TLS inspection (the CA
  referenced from an `AzureKeyVaultCertificate`'s versionless secret face,
  read through a user-assigned identity) and IDPS with per-signature
  overrides and bypass lists. Policies inherit via `base_policy_id`.
- **Deploy rules independently of the policy and each other**: groups
  carry application rules (FQDN/URL/category/TLS-terminating, header
  injection), network rules (L3/L4 incl. FQDN destinations), and DNAT
  rules (publish internal services through the firewall's public IP) --
  with source/destination IP Groups as first-class references.
- **Run the hub-spoke shape end to end**: the firewall deploys into the
  `AzureFirewallSubnet` (name and /26 contracts documented and
  fixture-proven), fronted by referenced Standard static public IPs; its
  `private_ip_address` output is exactly what a spoke route table's
  VIRTUAL_APPLIANCE next hop consumes. Forced tunneling (management IP
  configuration) and Virtual WAN hub deployment are fully modeled.
- **Curate address sets by intent**: IP Groups update in place and every
  referencing rule follows -- no more copy-pasted CIDR lists across rule
  collections.

## Design Decisions

- **Policy-based rule management only.** The classic in-firewall rule
  collection trio is a recorded skip: ARM rejects mixing it with an
  attached policy, and a greenfield catalog carries one rule surface.
- **Rules fold inside the group** (an ordered document; nothing references
  an individual rule); the group is the deployment unit, exactly as ARM
  models it.
- **The DNAT collection action is a one-value constant** ("Dnat") -- not
  modeled; both engines send it unconditionally.
- **Front-loaded contracts.** The azurerm firewall service ships NO
  CustomizeDiff, so the ARM contracts users would otherwise discover at
  apply time are authoring-time validations here: Premium gating of
  IDPS/TLS on the policy; the firewall's deployment-model pairing
  (hub ⇔ virtual_hub), exactly-one-subnet, public-IP-required-unless-
  management-path, management-name collision, and the live-discovered
  policy-owns-DNS exclusion; the DNAT translated-target XOR and
  TCP/UDP-only vocabulary on rules.

## Validation

- 86 spec tests across the four kinds (every CEL error path); outputs
  conformance cases ×4; `secret-coverage --check` and
  `validate-refs --check` green; full `planton tofu plan` renders for all
  four hack manifests; targeted + release-equivalent + `make build-go` +
  Bazel builds green; per-kind audits at 100% (PARITY ✅ COVERAGE ✅) with
  apply-time validator source-diff sections.
- **Live dual-engine E2E: 8/8 lanes green** against the test
  subscription -- IP Group 148s/179s; policy 147s/166s; rule collection
  group through the composed policy chain with a live IP-Group reference
  345s/319s; the policy-attached firewall through the five-fixture chain
  (RG → VNet → AzureFirewallSubnet → public IP + policy) 20m07s/21m50s.
  Zero orphans (subscription-wide sweep clean).
- One ARM-only contract was caught live and closed in-session: firewall-
  level DNS parameters on a policy-attached firewall
  (`AzureFirewallDNSConfigNotAllowedForVhubOrVnetWithPolicy`) -- now a
  spec validation with both error paths tested, and the failure class is
  documented in the E2E guide.
