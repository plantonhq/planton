# AWS Security Group Enrichment + MWAA Rebuild + MSK Serverless Forge

**Date**: July 4, 2026
**Type**: Feature (breaking on MWAA, zero users)
**Components**: API Definitions, AWS Provider, Terraform Modules, Pulumi Modules, E2E Framework, Site Catalog

## Summary

Three kinds in one session (owner-approved override of the one-per-session rhythm): `AwsSecurityGroup` gains prefix-list rule support, ICMP/-1 CEL honesty, revoke-on-delete, and first-class E2E citizenship with new `security_group_arn`/`owner_id` outputs; `AwsMwaaEnvironment` retires the catalog's **last embedded shadow security group** (DD-004 convergence complete — zero embedded SGs remain); `AwsMskServerlessCluster` is forged as enum 351 with the honest tiny AWS surface (subnets + optional SG refs, SASL/IAM hardcoded). Live dual-engine E2E ran only where creds allowed: SG profile is `deferred` (ambient `InvalidClientTokenId`); MWAA and MSK Serverless stay preview-only per the session decision (`status: deferred`, 120-min ceilings).

## Problem Statement / Motivation

- **SG** had inline rules but missed managed prefix lists (S3/DynamoDB gateway PLs), revoke-on-delete, ICMP port validation, conformance enrollment, and any first-class E2E lane — it was only a prerequisite fixture for other kinds.
- **MWAA** was the sole remaining embedded-SG carrier: five module-managed `aws_security_group*` resources, CIDR/associate/vpc_id fields, and a `security_group_id` output that violated the attach-list doctrine settled across the data wave.
- **MSK Serverless** had no Planton kind despite being a distinct provider resource with a minimal immutable surface — organizations using serverless Kafka had no leaf node.

## Solution / What's New

### AwsSecurityGroup (additive enrichment)

- `SecurityGroupRule.prefix_list_ids` for managed prefix list references in ingress/egress.
- ICMP/-1 CEL on `from_port`/`to_port` (type/code overloading documented; all-protocol `-1` ⇒ ports must be 0).
- `revoke_rules_on_delete` top-level bool (default false).
- Outputs: `security_group_arn`, `owner_id` (additive alongside existing `security_group_id`; no redundant `name`/`vpc_id` exports).
- No `name_prefix` — verified absent across all AWS protos; cloud names derive from `metadata.name` catalog-wide.
- First-class E2E: `e2e/profile.yaml`, `rules-rich` scenario, dual-engine test entries, EC2 verifier already existed.
- Outputs-conformance case added (first for this kind).

### AwsMwaaEnvironment (breaking rebuild)

- Removed `allowed_cidr_blocks`, `associate_security_group_ids`, `vpc_id` and their CEL rules.
- `security_group_ids` redefined as **required** attach-list of `AwsSecurityGroup` refs (`min_items: 1`).
- Dropped `security_group_id` output; added `created_at`, `database_vpc_endpoint_service`, `webserver_vpc_endpoint_service`.
- Deleted five TF SG resources and Pulumi `security_group.go`; both engines pass refs through directly.
- Closed Pulumi parity gap: `WorkerReplacementStrategy` wired (SDK v7).
- Naming basis converged: `metadata.id` → `metadata.name` on both engines.
- Registry: `prerequisites: [AwsSubnet, AwsSecurityGroup]`.
- Generator-owned `variables.tf` + drift allowlist entry; Pulumi entrypoint anatomy completed (Makefile, iac-input.yaml).
- E2E scaffold (deferred profile, scenarios, `mwaaEnvironmentVerifier`, dual-engine test stubs).
- Presets, spec tests, docs rewritten.

### AwsMskServerlessCluster (new kind, enum 351)

- Package `awsmskserverlesscluster/v1/` with full anatomy: 4 protos, TF + Pulumi modules, presets, docs, catalog-page, hack manifest.
- Spec: `region`, `subnet_ids` (FK AwsSubnet, min 1), `security_group_ids` (FK AwsSecurityGroup, max 5, optional). SASL/IAM auth **not** modeled as a decorative bool — both modules hardcode `client_authentication.sasl.iam.enabled = true`; message header documents mandatory IAM auth.
- Outputs: `cluster_arn`, `cluster_name`, `cluster_uuid`, `bootstrap_brokers_sasl_iam`.
- Enum `351`, `id_prefix: "awsmsksl"`, `prerequisites: [AwsSubnet]`; catalogkindreflect map regenerated.
- Verifier, conformance case, drift allowlist, E2E profile (deferred), site catalog mirror.

### E2E discover fix

- `pkg/e2e/profile/discover.go`: PascalCase entries for `awssecuritygroup`, `awsmwaaenvironment`, `awsmskserverlesscluster` so CI profile discovery resolves correctly.

## Validation

- Spec/CEL tests green for all three kinds.
- `pkg/outputs` conformance (2 new cases: SG, MWAA; MSK Serverless enrolled).
- TF drift guard green (`AwsMwaaEnvironment` allowlisted; MSK Serverless enrolled).
- `planton validate-refs --check`, `validate-outputs`, `secret-coverage --check` green.
- `tofu validate` green on all three TF modules; Pulumi `go build` green on all three entrypoints.
- E2E test package compiles with `-tags=e2e`.
- **Live SG dual-engine E2E green** (run once credentials became available): deploy → verify → destroy → verify-clean on both provisioners against the rules-rich scenario (Pulumi 4m06s, Terraform 1m33s, incl. the AwsVpc prerequisite chain); zero-orphan sweep clean; profile flipped to `green`.
- **NOT run**: live MWAA / MSK Serverless — deferred by session decision; profiles `deferred`, artifacts live-ready.

## Breaking Changes

Zero users; no migration.

**MWAA**: `allowed_cidr_blocks`, `associate_security_group_ids`, `vpc_id` removed; `security_group_ids` changed semantics and became required; `security_group_id` output dropped.

**Chart note** (end-of-phase charts wave): MWAA and MSK Serverless manifest shape changes break charts — not touched this session per RoE §4.

## Impact

DD-004 embedded-shadow-SG convergence is **complete** — the entire AWS catalog now composes network ingress through first-class `AwsSecurityGroup` nodes. SG is a first-class E2E citizen and richer rule surface. MWAA and MSK Serverless are at doctrine depth with dual-engine offline proof; live lanes await creds (SG) or owner re-run (MWAA/MSK Serverless).
