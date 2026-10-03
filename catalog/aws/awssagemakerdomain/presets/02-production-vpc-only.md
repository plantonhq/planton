# Production VPC-Only Domain

A security-hardened SageMaker Domain for production ML teams with SSO authentication,
VPC-only networking, KMS encryption, and cost management via idle shutdown.

## When to Use

- Production ML platforms with compliance requirements
- Enterprise teams with centralized identity via AWS IAM Identity Center
- Environments where data exfiltration prevention is mandatory
- Teams that need cost guardrails for compute resources

## Configuration Highlights

- **Auth mode**: SSO (centralized identity management via IAM Identity Center)
- **Network**: VpcOnly (all traffic stays within VPC, requires NAT for internet)
- **Encryption**: Customer-managed KMS key for EFS home directories
- **Security**: Domain-level and user-level security groups for layered isolation
- **Auditability**: `executionRoleIdentityConfig: USER_PROFILE_NAME` attributes CloudTrail events to the acting user, not just the shared role
- **Identity propagation**: Identity Center user identity forwarded into Athena/Redshift/Lake Formation (`trustedIdentityPropagationStatus: ENABLED`)
- **Cost allocation**: domain tags propagate to apps/spaces/profiles (`tagPropagation: ENABLED`)
- **IDE**: JupyterLab with `ml.t3.medium` default instance
- **Cost control**: 2-hour idle timeout (saves ~70% on compute vs always-on) plus the priciest GPU tiers hidden from the instance picker
- **Storage**: 20 GB default / 200 GB max EBS per space
- **Landing page**: JupyterLab opens by default

## Cost Estimate

Domain infrastructure: EFS home-directory storage, billed per GB-month.
Per-user compute is the main driver — the `ml.t3.medium` default instance bills hourly while running, and the 2-hour idle timeout is what keeps that line small by shutting instances down outside working hours. EBS space storage bills per GB-month.

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awssagemakerdomain.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Add `sharingSettings` to enable notebook output sharing to S3
- Add `dockerSettings` to enable custom container workflows
- Add `kernelGatewayAppSettings` for custom ML framework images
- Add `jupyterLabAppSettings.codeRepositories` for auto-cloned Git repos
