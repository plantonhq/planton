# AWS App Runner Service Component (R23)

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, AWS Provider, Pulumi Module, Terraform Module

## Summary

Added AwsAppRunnerService as the 27th new AWS resource kind in the Planton cloud provider expansion project. This component provides a fully managed container application service supporting deployment from ECR images or GitHub code repositories, with inline VPC Connector creation, concurrency-based auto scaling, KMS encryption, health checks, and X-Ray observability.

## Problem Statement / Motivation

App Runner is AWS's simplest path from container image to production HTTPS endpoint -- an increasingly popular choice for web APIs, microservices, and internal tools. Without an Planton component, teams deploying App Runner services had to fall back to raw Terraform or Pulumi, losing the declarative YAML workflow, cross-resource references via `StringValueOrRef`, and infra chart composability that Planton provides for other AWS services.

### Pain Points

- No declarative way to deploy App Runner services through Planton
- VPC Connectors and Auto Scaling Configurations are separate TF/Pulumi resources that users must manage independently
- No standard pattern for wiring App Runner services into infra charts alongside VPCs, security groups, and KMS keys

## Solution / What's New

A complete AwsAppRunnerService deployment component with 41 files covering proto API, Pulumi/Terraform IaC modules, comprehensive validation tests, documentation, presets, and a catalog page.

### Key Design Decisions

- **Dual source types**: Both ECR image and GitHub code source supported in v1 (App Runner's core differentiator)
- **Inline VPC Connector bundling**: Users provide `subnet_ids` and `security_group_ids`; the module creates the VPC Connector automatically. Also supports referencing an existing `vpc_connector_arn` for shared connector scenarios
- **Inline Auto Scaling Configuration**: Users specify `min_size`, `max_size`, `max_concurrency` directly; the module creates the AWS Auto Scaling Configuration Version resource
- **Flattened runtime config**: Environment variables, secrets, port, and start_command are top-level fields (shared across both source types) rather than nested inside source-specific blocks
- **Both CPU/memory formats accepted**: Numeric ("1024", "2048") and human-readable ("1 vCPU", "2 GB") via CEL validation
- **Observability included**: Two simple fields (`observability_enabled` + `observability_configuration_arn`) for X-Ray tracing

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 20 top-level fields, 4 nested messages (`ImageSource`, `CodeSource`, `HealthCheck`, `AutoScaling`), 7 message-level CEL validations + 5 nested-message CEL validations
- **stack_outputs.proto**: 7 outputs (service_arn, service_id, service_url, service_name, service_status, vpc_connector_arn, auto_scaling_configuration_arn)
- `StringValueOrRef` for: `instance_role_arn`, `access_role_arn`, `connection_arn`, `vpc_connector_arn`, `subnet_ids`, `security_group_ids`, `kms_key_arn`, `observability_configuration_arn`

### Validation Tests (52 specs)

- 22 happy path tests (minimal, ECR/ECR_PUBLIC, code source API/REPOSITORY, VPC egress, auto scaling, health checks, observability, human-readable CPU/memory, valueFrom references, production-ready)
- 25 failure tests (mutual exclusions, invalid enums, range violations, missing required fields)
- 5 API envelope tests

### Pulumi Module (6 files)

- `main.go` + `locals.go` + `outputs.go` (standard orchestration pattern)
- `vpc_connector.go` -- Conditional VPC Connector creation from subnets + security groups
- `auto_scaling.go` -- Auto Scaling Configuration Version creation
- `service.go` -- Main service with `buildSourceConfiguration()` handling both image and code paths

### Terraform Module (5 files)

- `main.tf` with 3 conditional resources (vpc_connector, auto_scaling_config, service)
- Dynamic blocks for all optional configuration sections
- Feature parity with Pulumi module

### Documentation

- `README.md` -- User-facing spec reference with quick start
- `examples.md` -- 6 examples from minimal to infra-chart patterns
- `docs/README.md` -- 370+ lines covering architecture, cost model, security, comparisons with ECS/Lambda/EKS

### Presets

- `01-basic-public-image` -- ECR Public, defaults, no VPC
- `02-production-vpc-encrypted` -- Private ECR, VPC egress, KMS, tuned scaling
- `03-github-code-source` -- GitHub repo, Node.js, API configuration

### Enum Registration

- `AwsAppRunnerService = 320` in `cloud_resource_kind.proto` (Containers category, id_prefix: `awsar`)

## Benefits

- **Zero to HTTPS in one YAML manifest**: Simplest possible deployment experience for containerized web apps
- **Infra chart composable**: Rich `StringValueOrRef` outputs enable wiring into serverless-api and containerized-web-app charts
- **Two deployment models**: Teams can choose between pre-built container images (CI/CD pipeline) or build-from-source (GitHub direct)
- **Bundled sub-resources**: VPC Connector and Auto Scaling Config managed inline -- users don't need to understand AWS resource topology

## Impact

- AWS resource coverage: 27 of ~32 new components complete
- Phase 2 progress: 9 of 10 components done (AwsMwaaEnvironment and AwsTransitGateway remaining)
- Enables future "containerized web app" infra chart pattern combining App Runner + VPC + RDS/DynamoDB

## Related Work

- Part of project `20260215.02.sp.aws-resource-expansion`
- Follows patterns from AwsLambda (serverless compute), AwsEcsService (container deployment)
- VPC Connector bundling follows DD03 pattern from Azure sub-project

---

**Status**: Production Ready
**Timeline**: Single session (2026-02-16)
