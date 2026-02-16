# Add AwsCodePipeline Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added AwsCodePipeline as a new deployment component (enum 331, id_prefix `awscp`), enabling declarative management of AWS CodePipeline continuous delivery pipelines through OpenMCF. The component supports V1 and V2 pipelines with stages, actions, artifact stores, git-based triggers, pipeline variables, and advanced execution modes.

## Problem Statement / Motivation

AWS CodePipeline is the standard CI/CD orchestration service in the AWS ecosystem, connecting source providers (GitHub, Bitbucket, CodeCommit), build services (CodeBuild), and deployment targets (ECS, Lambda, S3, CloudFormation) into automated release pipelines. Without an OpenMCF component, teams managing AWS infrastructure through OpenMCF had no way to declaratively define their delivery pipelines alongside the infrastructure they deploy to.

### Pain Points

- No declarative pipeline-as-code option within the OpenMCF framework
- CodeBuild projects existed in OpenMCF but had no orchestrator to chain them with source and deploy stages
- Teams needed to manage pipelines outside of OpenMCF, breaking the single-pane-of-glass model

## Solution / What's New

A complete AwsCodePipeline deployment component with:

- **Proto API** — 4 proto files with 11 message types covering pipeline type (V1/V2), execution modes (SUPERSEDED/QUEUED/PARALLEL), artifact stores (single-region and cross-region with KMS encryption), stages with polymorphic actions (Source/Build/Test/Deploy/Approval/Invoke/Compute), git-based triggers (push and pull request with branch/file path/tag filtering), and pipeline-level variables
- **46 validation tests** — all passing, covering valid configurations, required fields, enum validations, range validations, and 4 cross-field CEL rules
- **Pulumi module (Go)** — 4 files: main.go, locals.go, outputs.go, pipeline.go with full trigger and variable support
- **Terraform module (HCL)** — 4 files with dynamic blocks for stages, actions, triggers (push/PR with branch/file path/tag filters), and variables
- **3 presets** — github-source-codebuild, ecr-ecs-deploy, s3-lambda-deploy
- **Production documentation** — README.md, examples.md (6 examples), docs/README.md (technical reference), catalog-page.md

### Key Design Decisions

- **V2 default**: New pipelines default to V2, which supports triggers, variables, and advanced execution modes. V1 is supported for backward compatibility.
- **Action configuration as `map<string, string>`**: Matches the Terraform/Pulumi/AWS API pattern. Each action provider (20+) has unique configuration keys — strongly typing all of them would be enormous and fragile. Presets and documentation compensate.
- **Stage conditions excluded from v1**: The before_entry, on_success, on_failure condition blocks are advanced V2 features with deep nesting. Deferred to v2 following the 80/20 rule.
- **Webhooks excluded**: Legacy V1 mechanism. V2 native triggers via CodeStar Connections are superior and recommended.
- **Custom action types excluded**: Account-level resource with independent lifecycle, not 1:1 with a pipeline.
- **4 StringValueOrRef references**: role_arn (-> AwsIamRole), artifact store location (-> AwsS3Bucket), artifact store encryption key (-> AwsKmsKey), action role_arn (-> AwsIamRole)

## Implementation Details

### Proto Structure

```
AwsCodePipelineSpec
├── pipeline_type (optional, default "V2")
├── execution_mode (optional, default "SUPERSEDED")
├── role_arn (StringValueOrRef -> AwsIamRole)
├── artifact_stores[] (min 1)
│   ├── location (StringValueOrRef -> AwsS3Bucket)
│   ├── region (cross-region support)
│   └── encryption_key_id (StringValueOrRef -> AwsKmsKey)
├── stages[] (min 2)
│   ├── name
│   └── actions[] (min 1)
│       ├── name, category, owner, provider, version
│       ├── configuration (map<string, string>)
│       ├── input_artifacts[], output_artifacts[]
│       ├── namespace, region, run_order, timeout
│       └── role_arn (StringValueOrRef -> AwsIamRole)
├── triggers[] (V2 only, max 50)
│   ├── provider_type ("CodeStarSourceConnection")
│   └── git_configuration
│       ├── source_action_name
│       ├── push[] (branches, file_paths, tags filters)
│       └── pull_request[] (branches, file_paths, events)
└── variables[] (V2 only)
    ├── name
    ├── default_value
    └── description
```

### CEL Cross-Field Validations

1. Triggers require pipeline_type = "V2"
2. Variables require pipeline_type = "V2"
3. Execution modes QUEUED/PARALLEL require pipeline_type = "V2"

## Benefits

- **Declarative pipeline management** — Define CI/CD pipelines as YAML alongside infrastructure
- **Cross-resource composability** — StringValueOrRef enables pipelines to reference IAM roles, S3 buckets, and KMS keys from other OpenMCF components
- **V2-first design** — Modern defaults with triggers and variables out of the box
- **Infra chart ready** — Pipelines can be composed into infra charts with dependency-aware deployment ordering

## Impact

- **34 files**, ~3,850 lines of non-generated code
- **46 validation tests** — all passing
- Registered as enum 331 in cloud_resource_kind.proto
- Complements the existing AwsCodeBuildProject component (enum 330) — together they provide complete CI/CD coverage in the AWS provider

## Related Work

- **AwsCodeBuildProject** (R31, enum 330) — Build projects that CodePipeline orchestrates
- **AwsIamRole** — Pipeline execution roles
- **AwsS3Bucket** — Artifact storage
- Part of the AWS resource expansion project (20260215.02.sp.aws-resource-expansion)

---

**Status**: Production Ready
