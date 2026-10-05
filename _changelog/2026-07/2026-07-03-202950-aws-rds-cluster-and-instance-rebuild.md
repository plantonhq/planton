# AWS RDS Cluster and Instance Rebuilt to the Full Provider Surface

**Date**: 2026-07-03
**Type**: Feature / Breaking Restructure
**Scope**: `aws-rds`

## Summary

`AwsRdsCluster` and `AwsRdsInstance` are rebuilt from the ground up against the
full Terraform AWS provider surface (`aws_rds_cluster`, `aws_rds_cluster_instance`,
`aws_db_instance`), with dual-engine parity, generator-owned Terraform contracts,
modern outputs, refreshed presets and docs, and first-time live E2E coverage
on both engines.

## AwsRdsCluster

The cluster spec now models every real cluster shape:

- **Aurora provisioned** — the new folded `instances` list is the cluster's
  compute: one writer plus any readers, each entry materialized as its own
  `aws_rds_cluster_instance` keyed by name in BOTH engines, so scaling readers
  is an in-place update that never touches the cluster. Previously the spec
  modeled no instances at all — a deployed cluster had zero compute.
- **Aurora Serverless v2** — `serverless_v2_scaling` (0–256 ACU bounds,
  min-capacity 0 enables auto-pause/scale-to-zero) with `db.serverless`
  instances.
- **Aurora Serverless v1** — `engine_mode: serverless` + `serverless_v1_scaling`
  (the legacy shape, modeled honestly as AWS-owned compute).
- **Multi-AZ RDS clusters** — community `mysql`/`postgres` engines with
  `db_cluster_instance_class`, provisioned storage, and io1/io2/gp3 storage
  types.

Depth added across the board: the AWS-managed master password
(`manage_master_user_password`, recommended default — the secret never touches
the manifest or IaC state; its ARN is exported), master-username honesty (AWS
has no default; required-unless-derived with CEL naming the derivation sources),
snapshot and point-in-time restore shapes (including Aurora copy-on-write fast
clones), Aurora Global Database membership + global/local write forwarding,
Aurora MySQL backtrack, IAM database authentication, engine `iam_roles`, the
RDS Data API, CloudWatch log exports, Performance Insights / Enhanced
Monitoring / Database Insights at the cluster level with per-instance
overrides, inline cluster parameters (module-managed parameter group with the
family derived from the pinned engine version), CA bundle, maintenance/backup
windows, and deletion safety (final-snapshot enforcement at validation time).

**Breaking (no users)**: the module-embedded security group is retired —
`allowed_cidr_blocks`, `associate_security_group_ids`, and `vpc_id` are gone.
Database ingress rules belong on referenced first-class `AwsSecurityGroup`
nodes; the cluster attaches groups by reference only.

## AwsRdsInstance

Rebuilt from a minimal 17-field spec to the full `aws_db_instance` surface:
managed master password (previously a plaintext password was *required*),
storage autoscaling (`max_allocated_storage_gb`) and gp3 IOPS/throughput
tuning, dedicated log volume, Multi-AZ standby vs AZ pinning (mutually
exclusive), read replicas with Oracle `replica_mode`, snapshot and
point-in-time restores, Active Directory domain join (AWS-managed and
self-managed shapes), blue/green deployment updates, IAM auth, log exports,
Performance Insights / Enhanced Monitoring / Database Insights, license model
and character-set/timezone create-time knobs, and the same deletion-safety
contract as the cluster.

## Both kinds

- Terraform `variables.tf` contracts are generator-owned under the drift guard;
  the cluster TF module was rewritten from scratch (the previous module had
  drifted beyond repair from the spec), and the instance TF module was
  conformed with its dead duplicate `resources/*.tf` copies removed.
- Outputs renamed to the catalog's semantic convention (`endpoint`,
  `reader_endpoint`, `arn`, `cluster_resource_id`, `master_user_secret_arn`,
  `instance_identifier`, `resource_id`, ...) with `pkg/outputs` conformance
  cases; zero PARITY-EXCEPTIONs across both engines.
- Registry prerequisites (`AwsSubnet`) drive composed E2E; presets refreshed
  (Aurora PostgreSQL / MySQL / Serverless v2; PostgreSQL / MySQL production and
  a new read-replica preset) and catalog pages/docs updated.

## E2E

New `aws-sdk-go-v2/service/rds` dependency; state-aware verifiers for both
kinds (a deleting/deleted resource counts as absent, the NAT-gateway lifecycle
class); scenarios covering an Aurora Serverless v2 cluster (0–1 ACU, one
`db.serverless` writer, managed password) and a single-AZ postgres
`db.t4g.micro` on gp3. Live dual-engine E2E green on all four lanes with a
clean zero-orphan account sweep.

The first live run caught a modeling gap every offline gate missed: the
provider marks `master_username` Optional+Computed, but CreateDBCluster hard-
requires it. The spec now encodes required-unless-derived CEL for the
credential fields, and the forge workflow gained a mandatory
create-time-required identification step so future specs verify requiredness
against the cloud API's Create call rather than provider schema flags. The
forge rule's E2E section also gained machine-level isolation guidance
(`TMPDIR`/`PULUMI_HOME`/`-count=1`) after concurrent suites sharing global
Pulumi state produced false `no stack named` teardown failures.
