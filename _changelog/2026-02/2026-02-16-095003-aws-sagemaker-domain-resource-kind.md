# AWS SageMaker Domain Resource Kind

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, AWS Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added AwsSagemakerDomain as the twenty-sixth new AWS resource kind in the cloud provider expansion project. This component deploys SageMaker Studio domains — the most deeply nested AWS resource in the expansion — with IAM/SSO authentication, VPC networking, JupyterLab and KernelGateway IDE configurations, Docker access for custom containers, idle timeout for cost management, and notebook output sharing.

## Problem Statement / Motivation

ML teams need managed JupyterLab environments integrated with their VPC, IAM roles, and security boundaries. SageMaker Domain is the foundational resource for Amazon SageMaker Studio, but its Terraform provider schema has 4 levels of nesting with ~17 optional nested blocks inside `default_user_settings` alone and 10 different app settings types. Without a well-scoped Planton component, teams face raw Terraform complexity that obscures the essential configuration choices.

### Pain Points

- SageMaker Domain is one of the most deeply nested AWS resources, requiring careful 80/20 scoping
- 7 ForceNew fields (auth_mode, domain_name, vpc_id, subnet_ids, kms_key_id, retention_policy, domain_settings.security_group_ids) make design decisions consequential
- 10 different app settings blocks (JupyterLab, KernelGateway, Canvas, CodeEditor, RStudio, etc.) require opinionated selection for v1
- Cost management via idle timeout is critical but buried in app_lifecycle_management.idle_settings nesting
- Docker access is increasingly important for custom training containers but lives inside domain_settings

## Solution / What's New

A focused v1 covering ~85% of production SageMaker Domain deployments through careful feature selection:

### Proto API (spec.proto)

- **10 nested messages**, **44 fields**, **9 CEL validations**
- Top-level: auth_mode, vpc_id, subnet_ids, kms_key_id, app_network_access_type, default_user_settings, domain_security_group_ids, docker_settings
- DefaultUserSettings: execution_role_arn, security_group_ids, landing URI, studio web portal, JupyterLab, KernelGateway, sharing, EBS storage
- JupyterLab: resource spec, lifecycle configs, custom images (max 200), code repos (max 10), idle settings
- Shared ResourceSpec: instance_type, lifecycle_config_arn, sagemaker_image_arn, image_version_alias/arn
- All cross-resource references use StringValueOrRef with default_kind annotations

### Design Decisions

- **Selective domain_settings flattening**: domain_security_group_ids and docker_settings promoted to top level. RStudio and execution_role_identity_config deferred to v2.
- **Idle settings flattened**: app_lifecycle_management.idle_settings collapsed to idle_settings directly in JupyterLab settings. The extra nesting adds no semantic value.
- **No managed security group pattern**: Unlike database resources (RDS, Redshift, MSK), SageMaker creates its own domain boundary SG. No allowed_cidr_blocks/vpc_id managed SG needed.
- **JupyterLab + KernelGateway only**: Deferred Canvas (~10%), RStudio (~5%), Code Editor (~15%), R Session (~3%), TensorBoard (~5%), JupyterServer (~10%) to v2.
- **Docker at top level**: Promoted from domain_settings because Docker is a first-class concern for ML teams building custom containers.

### Deliberately Deferred to v2

Canvas app settings, RStudio Server Pro, Code Editor, R Session, TensorBoard, JupyterServer (classic), default_space_settings, custom_file_system_config, custom_posix_user_config, studio_web_portal_settings, retention_policy, tag_propagation, auto_mount_home_efs.

## Implementation Details

### Files Created

- **Proto API**: spec.proto (10 messages, 44 fields, 9 CEL validations), stack_outputs.proto (6 outputs), api.proto, stack_input.proto
- **Tests**: spec_test.go (48 tests: 23 happy path, 20 failure, 5 envelope)
- **Pulumi module**: main.go, locals.go, outputs.go, domain.go (single sagemaker.NewDomain with nested builder functions)
- **Terraform module**: main.tf (dynamic blocks for all optional settings), variables.tf, locals.tf, outputs.tf, provider.tf
- **Documentation**: README.md, examples.md (6 examples), docs/README.md (architecture, cost model, security)
- **Presets**: 01-basic-jupyter-domain, 02-production-vpc-only, 03-ml-team-with-custom-images
- **Catalog page**: sagemaker-domain.md + AWS index update
- **Registration**: AwsSagemakerDomain = 270 in cloud_resource_kind.proto (id_prefix: sgmkd)

### Key Technical Details

- Pulumi module creates a single `sagemaker.NewDomain` resource — all complexity is in building nested `DomainDefaultUserSettingsArgs`
- Builder functions (buildJupyterLabAppSettings, buildKernelGatewayAppSettings, etc.) handle nil-checking and conditional construction
- Terraform module uses dynamic blocks for all optional nested configurations
- JupyterLab idle settings nested under `app_lifecycle_management.idle_settings` in TF (flattened in proto)
- Space storage settings wrapped in `default_ebs_storage_settings` in TF provider

## Benefits

- ML teams can deploy production SageMaker Studio domains with a single YAML manifest
- VpcOnly + KMS + idle timeout preset provides security-hardened, cost-managed default
- Custom image support enables standardized ML frameworks across teams
- Docker access enables custom container workflows without leaving Studio
- 6 stack outputs enable downstream resource composition (user profiles, IAM policies, monitoring)

## Impact

- **Users**: ML platform teams, data scientists, ML engineers deploying SageMaker Studio
- **Infra charts**: Enables future ml-notebook-environment infra chart composition
- **Resource count**: AWS coverage grows from 50 to 51 resource kinds

## Related Work

- Part of [20260215.02.sp.aws-resource-expansion](/docs/projects/aws-resource-expansion) — R22 of ~32 AWS resource kinds
- Follows patterns established by AwsMskCluster (R21), AwsRedshiftCluster (R20), and AwsEksCluster
- Enables future AwsSagemakerUserProfile and AwsSagemakerSpace components

---

**Status**: Production Ready
**Timeline**: Single session implementation
