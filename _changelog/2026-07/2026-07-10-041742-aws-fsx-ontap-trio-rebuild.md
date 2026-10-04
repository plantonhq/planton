# AWS FSx for NetApp ONTAP Trio at the Full Provider Surface

**Date**: July 10, 2026
**Type**: Feature (with breaking spec changes on the file system and volume)
**Components**: AWS Provider, API Definitions, Pulumi CLI Integration, IAC Stack Runner, Testing Framework

## Summary

The FSx for NetApp ONTAP trio — `AwsFsxOntapFileSystem`, `AwsFsxOntapStorageVirtualMachine`, and `AwsFsxOntapVolume` — now models the complete provider surface with contracts that match AWS's real API instead of the provider's loosest validators. The file system gains the missing whole-file-system throughput arm as a proper exactly-one-of contract, sheds two fields the FSx API does not accept for ONTAP, and drops an output AWS never populates; the volume gains the byte-precise `size_in_bytes` arm (the only way past 2 PiB) and `final_backup_tags`, and its delete-time controls become presence-honest in both engines. All three Terraform contracts moved to the generator under the drift guard, both engines converged on one tag identity, and the family shipped its first E2E artifacts with live lanes recorded as deferred.

## Problem Statement / Motivation

The ONTAP trio predated the current kind anatomy: hand-written Terraform contracts (all three carried a vestigial `type = any` spec variable that nothing read), `= 5.82.0`/`~> 5.0` provider pins, `Pulumi.yaml` files pointing at prebuilt binaries no checkout contains, no E2E artifacts, and no drift or outputs-conformance enrollment.

### Pain Points

- **Two spec fields did not exist on the resource.** The file-system spec carried `copy_tags_to_backups` and `skip_final_backup`, but `aws_fsx_ontap_file_system` has no such arguments — ONTAP backups are volume-scoped. Both engines silently ignored the fields, so a user's explicit backup decision changed nothing.
- **Half the throughput contract was missing.** AWS sizes ONTAP throughput through exactly one of `ThroughputCapacity` (whole file system) or `ThroughputCapacityPerHAPair`; the spec required the per-pair arm and had no whole-system arm at all. The per-pair value set was also validated as one flat list when AWS's real tiers differ per deployment generation — a first-generation tier on a SINGLE_AZ_2 file system passed validation and failed at create.
- **Rules looser than AWS.** The spec allowed `HDD` storage (ONTAP is SSD-only — HDD belongs to Windows/Lustre and Intelligent-Tiering to OpenZFS/Lustre), allowed 12 HA pairs on SINGLE_AZ_1 (scale-out is SINGLE_AZ_2-only), validated capacity as a flat range instead of AWS's per-HA-pair formula (1024–524288 GiB per pair, 192 TiB first-generation ceiling), and had no subnet-count contract (single-AZ takes exactly one, multi-AZ exactly two).
- **A dead output.** The file system exported `dns_name`, which the FSx API never populates for ONTAP — data access is via SVM endpoints. Anything composing on it would read an empty string.
- **Volumes past 2 PiB were unrepresentable — and the spec said to leave the platform.** The volume spec's own comment directed users to "use the AWS console or CLI with size_in_bytes" for large volumes instead of modeling the field.
- **Delete-time controls silently dropped.** The volume's Pulumi module sent `skip_final_backup`, `copy_tags_to_backups`, `bypass_snaplock_enterprise_retention`, `storage_efficiency_enabled`, and the SnapLock booleans only when `true` — an explicit `false` never reached AWS, so "take a final backup" or "disable efficiency" could not be expressed. `final_backup_tags` was missing entirely.
- **Tag identity diverged three ways.** Terraform emitted only a `Name` tag plus labels; Pulumi emitted the `planton.ai/*` identity set without `Name`; neither matched the converged family convention.

## Solution / What's New

### AwsFsxOntapFileSystem (breaking)

- **Throughput as exactly one arm**: new `throughput_capacity` (whole-file-system: 128–4096) XOR `throughput_capacity_per_ha_pair`, with per-generation value sets enforced at validation — first generation 128–4096; SINGLE_AZ_2/MULTI_AZ_2 single-pair 384/768/1536/3072/6144; scale-out 1536/3072/6144 per pair.
- **API-honest contracts**: storage_type tightened to SSD (with the field comment pointing at volume tiering policies, ONTAP's real cost-tiering mechanism); `ha_pairs > 1` gated to SINGLE_AZ_2; capacity validated by the per-HA-pair formula plus the 192 TiB first-generation ceiling; subnet counts per deployment type; `preferred_subnet_id` required for multi-AZ and invalid for single-AZ (single-AZ derives it from the only subnet in both engines — the provider requires the argument unconditionally); CIDR format on `endpoint_ip_address_range`; HH:MM and d:HH:MM window formats; disk IOPS bounded 0–2,400,000.
- **Removed**: `copy_tags_to_backups` and `skip_final_backup` (not part of this resource), and the never-populated `dns_name` output (data access is via SVM endpoints — now stated on the outputs).
- `route_table_ids` gained its foreign-key annotation (`AwsSubnet.route_table_id`), and both engines now send `ha_pairs` and the resolved backup retention whenever present (zero is a real value).

### AwsFsxOntapStorageVirtualMachine

- Validation brought to the provider's real bounds: AD domain name ≤255, username/password ≤256, administrators group ≤256, organizational unit ≤2000, and per-entry IPv4 format on `dns_ips`.
- Update-in-place honesty documented: only `file_system_id`, `name`, and `root_volume_security_style` replace the SVM; the admin password and the entire Active Directory block (domain join included) update in place.

### AwsFsxOntapVolume (breaking)

- **Size as exactly one arm**: `size_in_megabytes` (now optional) XOR the new `size_in_bytes` int64 arm (20 MiB to ~20 PiB) — byte-precise sizing and the only route past 2 PiB, closing the "use the AWS console" gap. The Terraform module converts through the provider's string-typed nullable integer; Pulumi formats the int64 directly.
- **New**: `final_backup_tags` (delete-time tags on the final backup).
- **Presence honesty**: `storage_efficiency_enabled` is now a tri-state optional (true enables, false disables, unset keeps ONTAP's default), and every delete-time/backup boolean is sent whenever present in BOTH engines — an explicit `false` finally reaches AWS. SnapLock retention values now transmit correctly for unit types including a meaningful zero (e.g. a 0-day minimum retention), eliding only INFINITE/UNSPECIFIED.
- Tightened bounds: junction path ≤255, snapshot policy ≤255, `aggr[0-9]{1,2}` aggregate-name format, retention/autocommit values ≤65535.
- The delete-time semantics are documented where users will see them: the controls are read from state at destroy time and must be applied before the deletion.

### Cross-cutting

- **Generator-owned Terraform contracts** for all three kinds (drift-guard enrolled); modules rewritten onto `var.spec.*`/`var.metadata.*` with authoring comments, `backend.tf`, and `>= 6.0.0` provider floors (the surface predates the v6 line; floors keep the FSx family on one provider major).
- **Tag convergence**: `Name` + the `planton.ai/*` identity set in both engines, with user labels merging in on the Terraform side; the SVM and volume comments spell out the two-name model (ONTAP-internal `spec.name` vs the Infra Component's `metadata.name`).
- **Pulumi entrypoint hygiene**: the `runtime.options.binary` residue removed from the file-system and SVM `Pulumi.yaml` (project-named binaries no checkout contains — fails only at deploy/preview), the volume's placeholder project name and commented `debug.sh` replaced, and `iac-input.yaml` worked examples added for all three.
- **Registry**: prerequisites now encode the composition chain — file system → `[AwsSubnet]`, SVM → `[AwsFsxOntapFileSystem]`, volume → `[AwsFsxOntapStorageVirtualMachine]`.
- **First E2E artifacts**: two new lifecycle-aware verifiers (`DescribeStorageVirtualMachines`, `DescribeVolumes`; DELETING counts as absent) beside the reused file-system verifier; file-system and SVM prerequisite fixtures (the file-system fixture deliberately uses the cheapest creatable shape — SINGLE_AZ_1 with the whole-system 128 MB/s arm — which also live-proves the first-generation arm); four scenarios (single-AZ minimal, multi-AZ full-surface, UNIX SVM, mounted FlexVol); six test entrypoints; outputs-conformance cases for all three kinds. Profiles record `status: deferred` (FSx provisions in tens of minutes and bills accordingly) with per-profile exclusion ledgers: Active Directory (no directory exists in the test account), SnapLock (a COMPLIANCE volume is undeletable until retention expires), FLEXGROUP/DP arms, and multi-AZ route tables.
- **Docs and presets**: the file system's missing root README created; catalog pages and architecture docs corrected (SSD-only storage, throughput arms, the dead output, delete-time semantics); presets moved off now-invalid first-generation tiers onto real second-generation values, plus a new scale-out preset (4 HA pairs, 6 GB/s aggregate); hack manifests extended to full-surface shapes so the offline plan proof covers the arms the deferred lanes exclude.

## Verification

- Spec/CEL tests for all three kinds (new cases for every added rule and both XOR contracts), the variables.tf drift guard, and the outputs-conformance suite — all green.
- Reference-integrity gate green; secret-coverage reports no findings for the FSx kinds.
- Manifest validation across all 19 ONTAP manifests (presets, hack manifests, fixtures, scenarios).
- Offline engine proofs with live credentials: `tofu init` + `plan` per kind ("Plan: 1 to add" each, against the full-surface hack manifests) and `pulumi preview` per kind ("+3 to create" each) — the proof line for the family's deferred live lanes.
- Repo-wide `make build-go` green; site catalog mirror regenerated (the new scale-out preset publishes).

## Impact

- **Users can finally express what AWS actually accepts** — and are stopped at validation when they express what it doesn't. Misconfigurations that previously failed minutes into a create (first-generation tiers on second-generation deployments, HDD storage, oversized single-pair capacity, wrong subnet counts) now fail before any cloud call.
- **Explicit `false` means false.** Backup, efficiency, and SnapLock decisions reach AWS in both engines, identically.
- **The FSx family is complete**: all six FSx kinds plus the data repository association now share one anatomy — generator-owned contracts, converged tags, first-class E2E artifacts, and provider floors on the v6 line.
