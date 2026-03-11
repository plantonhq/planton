# AWS Batch Compute Environment Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added the `AwsBatchComputeEnvironment` resource kind (enum 321, id_prefix `awsbat`) to OpenMCF, enabling declarative management of AWS Batch compute environments with bundled job queues and optional fair-share scheduling policies. This is the 28th new AWS resource kind in the cloud provider expansion effort.

## Problem Statement / Motivation

AWS Batch is a foundational service for running batch processing workloads — data pipelines, ETL, ML training, and scientific computing. Before this component, teams deploying AWS Batch had to manually manage the three-way relationship between compute environments, job queues, and scheduling policies. There was no declarative, version-controlled way to provision this infrastructure through OpenMCF.

### Pain Points

- No OpenMCF support for AWS Batch workloads
- Manual orchestration of compute environments and job queues as separate Terraform/Pulumi resources
- No built-in validation for type-specific field requirements (e.g., instance_role for EC2, spot_iam_fleet_role for SPOT)
- No cross-resource references via StringValueOrRef for VPC, security group, and IAM role dependencies

## Solution / What's New

A complete `AwsBatchComputeEnvironment` deployment component covering:

- **4 compute types**: EC2, SPOT, FARGATE, FARGATE_SPOT
- **Bundled job queues**: At least one required, with priority routing and automatic job-state time-limit actions
- **Fair-share scheduling**: Optional policy for multi-team capacity distribution
- **Full IaC**: Both Pulumi (Go) and Terraform (HCL) implementations with feature parity

### Bundling Design

The component bundles compute environments + job queues + scheduling policy because a compute environment without a queue is incomplete infrastructure. Job definitions are excluded — they have independent lifecycles (versioned, application-level) and should be managed separately.

## Implementation Details

### Proto API (4 files)

- `spec.proto` — 8 message types with CEL cross-field validations (instance_role required for EC2/SPOT, spot_fleet_role required for SPOT, launch template id-or-name exclusivity)
- `stack_outputs.proto` — 6 outputs including per-queue ARN map
- `api.proto` — KRM wiring with `aws.openmcf.org/v1` api_version
- `stack_input.proto` — Standard stack input with provider config

### Pulumi Module (6 files)

- `main.go` — Orchestrates creation in dependency order: scheduling policy → compute environment → job queues
- `compute_environment.go` — Handles all 4 resource types with conditional field mapping
- `job_queue.go` — Creates queues with compute environment and scheduling policy references
- `scheduling_policy.go` — Optional fair-share policy creation
- `locals.go` / `outputs.go` — Standard patterns

### Terraform Module (4 files)

- `main.tf` — All three resources with `dynamic` blocks for conditional configuration
- `variables.tf` / `locals.tf` / `provider.tf` — Standard patterns

### Validation Tests

27 test cases covering valid configurations (Fargate, EC2, SPOT, FARGATE_SPOT, scheduling policy, update policy, launch template, time-limit actions) and invalid inputs (missing queues, missing instance_role, bid_percentage out of range, launch template conflicts, etc.).

### Presets

3 deployment patterns: serverless Fargate, EC2 managed with multi-queue, and Spot cost-optimized with fair-share scheduling.

## Benefits

- **Declarative batch infrastructure**: Compute environments, queues, and scheduling policies in a single YAML manifest
- **Cross-resource references**: StringValueOrRef for subnets, security groups, IAM roles, and KMS keys
- **Type-safe validation**: CEL validations catch misconfiguration before deployment
- **Production presets**: Ready-to-deploy patterns for common batch architectures

## Impact

- Expands AWS coverage from 30 to 31 new resource kinds in the expansion project
- Enables future infra charts for batch processing, data pipelines, and ML training workloads
- Fills a key gap in the AWS compute story alongside ECS, EKS, Lambda, and App Runner

## Related Work

- Part of project 20260215.02.sp.aws-resource-expansion (R28)
- Parent project: 20260212.01.openmcf-cloud-provider-expansion
- Follows patterns established by AwsNeptuneCluster (R26) and AwsMemorydbCluster (R27)

---

**Status**: Production Ready
