# AWS MSK Cluster + OpenSearch Domain: Data-Wave Tail Rebuild to the Full Provider Surface

**Date**: July 4, 2026
**Type**: Feature (breaking, zero users)
**Components**: API Definitions, AWS Provider, Terraform Modules, Pulumi Modules, Testing Framework

## Summary

`AwsMskCluster` and `AwsOpenSearchDomain` — the last two kinds of the data wave — are rebuilt to their full provider surfaces. MSK retires the embedded shadow security group (ingress now lives on referenced first-class `AwsSecurityGroup` nodes, which the brokers attach directly), gains multi-VPC PrivateLink connectivity, dual-stack networking, Express-broker rebalancing, and two folded satellites (SCRAM secret associations, cluster policy). OpenSearch grows from a mid-depth spec to the complete `aws_opensearch_domain` surface — Cognito Dashboards auth, JWT and anonymous FGAC arms, IAM Identity Center, AI/ML options, per-node-type topology, full Auto-Tune, off-peak windows, and the dual-stack V2 endpoint outputs. Both kinds ship their first-ever E2E artifacts; per the session decision, validation ran as `tofu plan` + `pulumi preview` on both engines (no live provision lane), and both profiles record the deferral.

## Problem Statement / Motivation

### Pain Points

- **MSK carried the retired embedded-SG pattern** — `allowed_cidr_blocks`, `associate_security_group_ids`, and a `vpc_id` that existed only to feed a module-managed shadow security group. Every other data kind had already converged to referenced first-class security groups.
- **MSK was missing real provider surface**: no PrivateLink (`vpc_connectivity`), no dual-stack (`network_type`), no Express-broker rebalancing, no SCRAM secret association, no cluster policy — the mechanism behind cross-account PrivateLink access.
- **OpenSearch was ~10 blocks short of the provider**: no Cognito, no JWT/anonymous FGAC, no Identity Center, no AI/ML options, no `node_options`, no off-peak window, an Auto-Tune boolean instead of the real options block, and no V2 endpoint outputs.
- **Structural debt on both**: OpenSearch's TF contract was a hand-written `type = any`; both kinds keyed cloud names off `metadata.id` on one engine (the cross-engine identity class); MSK's Pulumi entrypoint lacked its Makefile and iac-input template; neither kind had any E2E coverage.

## Solution / What's New

### AwsMskCluster

- **Embedded SG retired (breaking)**: `allowed_cidr_blocks` / `associate_security_group_ids` / `vpc_id` removed; `security_group_ids` flips to the attach shape and becomes **required** (`min_items: 1`) — the provider marks `broker_node_group_info.security_groups` Required+ForceNew, and requiredness-honesty is the standing convention. The `security_group_id` output is gone; both modules lose the shadow SG and its five rule resources.
- **New provider surface**: `vpc_connectivity` (PrivateLink auth schemes, CEL-coupled to the cluster's own `authentication`), `network_type` (IPV4/DUAL), `rebalancing_status` (Express brokers), `public_access_type` hardened by CEL (authenticated TLS only), and the three `bootstrap_brokers_vpc_connectivity_*` outputs.
- **Folded satellites** (the cluster-keyed settings class): `scram_secret_arns` — SCRAM secret associations, materialized per-ARN in both engines, `AmazonMSK_`-prefix + uniqueness validated, `sensitive_exempt_reason` recorded (ARNs are references, never secret material) — and `cluster_policy` as a JSON string (one `aws_msk_cluster_policy` per cluster).
- **Seven spec-level CEL rules** covering provisioned throughput, configuration mutual exclusion and revision, SCRAM coupling, the three PrivateLink scheme couplings, and the public-access posture.

### AwsOpenSearchDomain

- **Full FGAC**: anonymous auth, JWT bearer tokens (`jwks_url` XOR `public_key` CEL), plus the existing internal-database and IAM master-user arms.
- **New blocks**: `cognito_options` (Dashboards auth, all-three-fields CEL), `identity_center_options`, `aiml_options` (natural-language query generation, S3 vectors engine, serverless vector acceleration), `node_options` (per-node-type topology), full `auto_tune_options` (desired state, maintenance schedules, rollback, off-peak alignment — replacing the old boolean), `off_peak_window_options`, `automated_snapshot_start_hour`, `deployment_strategy`.
- **V2 endpoint outputs**: `endpoint_v2`, `dashboard_endpoint_v2`, `domain_endpoint_v2_hosted_zone_id`.

### Both kinds, one contract

- Generator-owned `variables.tf` under the drift guard (OpenSearch's `type = any` legacy is gone); provider floors on the v6 line (MSK `>= 6.41.0`, OpenSearch `>= 6.31.0`); naming basis `metadata.name` on both engines with the cloud name argument set explicitly; Pulumi entrypoint anatomy completed (Makefile, iac-input template); presets, catalog pages, READMEs, and deep docs rewritten to the new shapes (the stale `autoTuneEnabled` and embedded-SG narratives are gone everywhere).
- Outputs-conformance enrollment for both kinds (all 17 MSK outputs, all 8 OpenSearch outputs flatten onto their protos).

### E2E (first-ever for both; preview/plan lanes per session decision)

- **`AwsSecurityGroup` becomes an E2E citizen**: first prerequisite install profile (Kafka 9092-9098 + ZooKeeper 2181-2182 ingress from the E2E VPC), an EC2 `DescribeSecurityGroups` verifier, and `prerequisites: [AwsVpc]` in the kind registry — MSK's registry entry declares `prerequisites: [AwsSubnet, AwsSecurityGroup]`.
- **Verifiers**: Kafka `DescribeClusterV2` keyed on the cluster ARN (DELETING = absent) and OpenSearch `DescribeDomain` keyed on the domain name (`Deleted` flag = absent) — the RDS lifecycle class. Two new AWS SDK modules (`service/kafka`, `service/opensearch`) wired through go.mod, MODULE.bazel, and the verify BUILD file.
- **Scenarios**: MSK — two kafka.t3.small brokers over the two-AZ subnet prerequisites with the SG prerequisite attached, SASL/IAM, folded server-properties configuration; OpenSearch — a public single-node t3.small.search leaf (encryption at rest, node-to-node TLS, HTTPS-only; Auto-Tune excluded — unsupported on t3). Dual-engine test entrypoints registered.
- **Validation lanes run**: `tofu init` + `validate` + `plan` and `pulumi preview` green on both kinds against the hack manifests (MSK: 2 TF resources / 4 Pulumi URNs with all outputs wired; OpenSearch: 1 TF resource / 3 Pulumi URNs). Live provision was deliberately skipped (MSK runs 25-40 minutes each way per engine); both profiles carry `status: deferred` with the reason and stay re-runnable.

## Validation

- Offline gate green: spec/CEL tests for both kinds, outputs conformance (2 new cases), TF drift guard, `validate-refs`, `secret-coverage`, kind-map regeneration, catalogkindreflect + E2E-runner test suites, `go vet` on the e2e-tagged test package, and all 11 touched manifests (presets, scenarios, prerequisites, hack manifests) CLI-validated.
- `tofu validate`/`plan` + `pulumi preview` green on both kinds (the tofu validate pass caught and fixed one block-name defect: the broker-log S3 destination is `s3`, not `s3_logs`).
- NOT run (recorded): live deploy → verify → destroy lanes for both kinds — deferred by owner decision; the artifacts are live-ready.

## Breaking Changes

Zero users; no migration. For the record — MSK: `allowed_cidr_blocks`, `associate_security_group_ids`, and `vpc_id` removed; `security_group_ids` changed semantics (attach shape) and became required; `security_group_id` output dropped. OpenSearch: `auto_tune_enabled` became the `auto_tune_options` message. Chart note for the end-of-phase charts wave: `charts/aws/kafka-streaming/templates/streaming.yaml` still renders `vpcId` + shadow-SG semantics and breaks against the new MSK spec.

## Impact

The data wave is complete: every data kind now composes network ingress through first-class security-group nodes, and the catalog covers managed Kafka and OpenSearch at the depth advanced organizations reach — PrivateLink multi-VPC Kafka with cross-account policies and SCRAM credential rotation; OpenSearch with Cognito/JWT/Identity-Center auth arms, per-node-type topologies, and dual-stack endpoints — with both engines proven equivalent by one generator-owned contract and a conformance bar per kind.
