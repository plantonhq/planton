# Azure Private Endpoint Rework + Five Verdict Forges (ASG, DES, SQL Failover Group, Activity Log Alert, Standard Web Test)

**Date**: July 12, 2026
**Type**: Feature (breaking rework + five new kinds + eleven reference retrofits)
**Components**: API Definitions, Azure Provider, IaC Modules (Terraform + Pulumi), E2E Framework, Catalog Kind Registry

## Summary

`AzurePrivateEndpoint` (414) -- the security-critical kind every private
PaaS topology composes through -- is reworked breaking from a 7-field
bundle to the full azurerm v4.80 surface, and five kinds are forged
around the catalog's remaining seams: `AzureApplicationSecurityGroup`
(428), `AzureDiskEncryptionSet` (429), `AzureMonitorActivityLogAlert`
(456), `AzureApplicationInsightsStandardWebTest` (457), and
`AzureMssqlFailoverGroup` (524). Eleven plain-string or annotation-less
reference seams across NSG rules, NICs, VM scale sets, managed disks,
VMs, AKS, and the metric alert now carry real `StringValueOrRef` foreign
keys with `default_kind`. Live dual-engine E2E: ten lanes green, zero
orphans; the disk encryption set's live lane is excluded with a recorded
technical reason.

## The AzurePrivateEndpoint rework (breaking)

The shipped spec hardcoded auto-approval, silently ignored static IPs,
supported one DNS zone, and built its Pulumi provider inline (the shape
that breaks keyless OIDC auth). Now:

- **Folded private service connection**: target resource id (polymorphic
  bare reference) XOR Private Link Service alias -- the provider's
  ExactlyOneOf as a message CEL, plus the alias's
  `.azure.privatelinkservice` suffix contract as a field CEL; repeated
  `subresource_names` (the data-exfiltration boundary); manual
  connections with the request-message pairing front-loaded in BOTH
  directions (the provider's CustomizeDiff).
- **DNS zone group** over a repeated `AzurePrivateDnsZone` FK list --
  folded, not a separate node: an endpoint without DNS registration
  resolves to the service's PUBLIC ip, silently defeating the private
  link.
- **Static `ip_configurations`**, `custom_network_interface_name`, user
  tags.
- **ASG membership as the association resource** on both engines
  (Azure's member-side model).
- Shared Pulumi provider builder; registry gains
  `prerequisites: [AzureSubnet]`.
- Outputs: `private_endpoint_id`/`_name`, `private_ip_address`,
  `network_interface_id`.

## The five forged kinds

- **AzureApplicationSecurityGroup (428, `azasg`)**: the workload-identity
  grouping NSG rules target instead of CIDRs. Deliberately empty --
  membership is member-side; the kind exists to open the FK seam.
  **Retrofits**: NSG rule `source/destination_application_security_group_ids`
  ×2, NIC `application_security_group_ids`, VMSS ip-configuration ASG ids
  (all three orchestration-mode files) -- plain strings →
  `StringValueOrRef` FKs.
- **AzureDiskEncryptionSet (429, `azdes`)**: customer-managed-key
  encryption for managed disks. Required identity block (system / user /
  both with the ids-match-type CEL), key FK defaulting to
  `AzureKeyVaultKey.versionless_id` (rotation-on posture; the
  versioned-vs-versionless pairing documented -- CELs cannot dereference
  resolved references), three encryption types with the provider's exact
  irregular `ConfidentialVm` wire casing, cross-tenant
  `federated_client_id`. **Retrofits**: managed disk ×2, VM ×2, VMSS ×3
  gained `default_kind`; AKS's plain-string DES seam converted to a real
  FK.
- **AzureMssqlFailoverGroup (524, `azmsqlfog`)**: the SQL logical-server
  DR grouping with the failover-following listener. AUTOMATIC (grace ≥
  60) XOR MANUAL (no grace) mirroring the provider's CustomizeDiff;
  primary/partner/database FKs; the listener endpoints DNS-composed as
  outputs (azurerm exports only the id).
- **AzureMonitorActivityLogAlert (456, `azactalert`)**: the only path to
  alerting on control-plane operations, service-health incidents, and
  resource-health transitions. Seven-category criteria with plural-only
  narrowing (azurerm's redundant singular forms folded), three
  exclusivity contracts as CELs, the four-value location allowlist
  mirroring the provider's CustomizeDiff, action-group FK actions.
- **AzureApplicationInsightsStandardWebTest (457, `azwebtest`)**:
  synthetic availability monitoring. Request block (seven-verb
  vocabulary, headers, body, redirect/dependent dials), validation rules,
  300/600/900 frequency vocabulary, geo-locations. **Retrofit**: the
  metric alert's `web_test_id` plain string → FK on the new kind, closing
  the availability-alert composition loop.

## Audit-caught defects (fixed before ship)

The formal per-kind audit against the azurerm source caught four defects
after the initial implementation:

1. **Web test -- silently dropped SSL lifetime** (user-facing): the
   provider's expand sends `ssl_cert_remaining_lifetime` only when
   `ssl_check_enabled` is true; a lifetime without the check was
   accepted and silently dropped -- cert-expiry monitoring OFF while the
   user believes it is on. Front-loaded as a message CEL; the spec
   comment that claimed the opposite corrected.
2. **Web test -- missed https gate**: the provider's CustomizeDiff
   rejects SSL checks on non-https URLs; now a root-message CEL.
3. **Failover group -- wrong default documentation**: five artifacts
   claimed the read-only listener failover defaults to enabled; the
   provider sends Disabled when unset. Both engines already behaved
   identically -- the docs lied. All corrected.
4. **Disk encryption set -- unrecorded skip**: Managed-HSM-backed keys
   (azurerm 4.x's deprecated `managed_hsm_key_id` alias) now a recorded
   skip on the key field.

Audit reports at each kind's `docs/audit/2026-07-12-*.md` -- all six at
100% PARITY ✅ COVERAGE ✅ with apply-time validator source-diff sections.

## Validation

- **Offline gate**: spec tests ×6 kinds (incl. new CEL error paths);
  `make build-go`; `secret-coverage --check`; `validate-refs --check`
  (all new FK edges resolve); `pkg/outputs` conformance ×6; `tofu
  validate` + full `planton tofu plan` ×6 hack manifests; presets ×12
  validate; site catalog regen.
- **Live dual-engine E2E, ten lanes green, zero orphans**: ASG
  (130s/147s), activity log alert (156s/173s), standard web test
  (277s/288s), private endpoint through the composed
  subnet + storage-target + privatelink-zone chain (464s/567s), failover
  group through the composed two-region server pair + S0 database chain
  (Pulumi 749s; Terraform re-run after a transient Azure
  `OperationTimedOut` rollback on a fixture server -- environment, not
  module).
- **Disk encryption set live E2E excluded (recorded)**: the service
  requires a purge-protected vault, which cannot be purged at teardown
  (7+ day hold) -- the zero-orphan gate is unreachable on the shared
  test subscription. Profile `status: deferred`; the composed scenario +
  four fixtures ship ready-to-run.

## Environment classes recorded (e2e/README + build doc)

- Chunked `buf generate --path` runs must cover EVERY edited proto
  directory including shared registries -- a missed one keeps a stale
  `.pb.go` that fails silently at runtime (a kind's prerequisites
  resolve empty); verify via `git status` after the stub copy.
- Microsoft.Sql: `eastus` is offer-restricted (`ProvisioningDisabled`)
  on the test subscription; and logical-server creates can return a
  transient `OperationTimedOut` whose asynchronous rollback leaves a
  phantom server blocking RG deletion for several minutes.

## Impact

- Private PaaS connectivity is now first-class: any catalog service with
  a Private Link surface composes with `AzurePrivateEndpoint` +
  `AzurePrivateDnsZone` for DNS-correct private access, governed by NSG
  rules through ASG membership.
- CMK disk encryption (DES) closes the disk/VM/VMSS/AKS encryption
  story; SQL DR (failover group) closes the MSSQL family; activity-log
  alerting + synthetic web tests close the planned observability
  surface.
- Shared-builder migration: 57 of ~72 Azure Pulumi modules.

---

**Status**: ✅ Production Ready
