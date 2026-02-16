---
title: "AwsSagemakerDomain — Technical Reference"
description: "AwsSagemakerDomain — Technical Reference deployment documentation"
icon: "package"
order: 100
componentName: "awssagemakerdomain"
---

# AwsSagemakerDomain — Technical Reference

Comprehensive technical documentation for the AwsSagemakerDomain deployment component in OpenMCF. This document covers SageMaker Studio architecture, authentication modes, networking, storage, cost model, app types, idle management, Docker access, service limits, security, common patterns, and the v2 roadmap.

---

## Table of Contents

1. [SageMaker Studio Architecture](#sagemaker-studio-architecture)
2. [Authentication Modes](#authentication-modes)
3. [Network Modes](#network-modes)
4. [Storage Architecture](#storage-architecture)
5. [App Types](#app-types)
6. [Idle Timeout and Cost Management](#idle-timeout-and-cost-management)
7. [Docker Access](#docker-access)
8. [Cost Model](#cost-model)
9. [Service Limits](#service-limits)
10. [Security Model](#security-model)
11. [Common Patterns](#common-patterns)
12. [v2 Roadmap](#v2-roadmap)

---

## SageMaker Studio Architecture

Amazon SageMaker Studio is organized into a hierarchy of resources:

### Domain

The **Domain** is the top-level organizational unit. Each domain:

- Is associated with a single VPC and a set of subnets.
- Has exactly one authentication mode (IAM or SSO) — immutable after creation.
- Automatically provisions a dedicated Amazon EFS file system for user home directories.
- Defines default user settings (execution role, security groups, app settings) inherited by all user profiles.
- Creates an AWS-managed security group for the domain boundary (cross-app/cross-user traffic control).

A domain is a **ForceNew** resource for its core properties: changing `auth_mode`, `vpc_id`, `subnet_ids`, or `kms_key_id` requires domain replacement (delete + recreate).

### User Profiles

Each person who accesses the domain has a **User Profile** that:

- Inherits default settings from the domain (execution role, security groups, app configurations).
- Can override specific settings (e.g., different execution role, custom lifecycle configs).
- Owns a dedicated home directory on the domain's EFS file system at `/home/<user-profile-name>/`.
- Is associated with an IAM identity (IAM mode) or SSO identity (SSO mode).

User profiles are **not modeled in v1** of this OpenMCF component — they would be a separate resource kind.

### Spaces

**Spaces** are collaborative environments within a domain:

- A space can be **private** (single user) or **shared** (multiple users).
- Each space has its own EBS volume for working storage, separate from the EFS home directory.
- Spaces support JupyterLab, Code Editor, and custom applications.
- Space storage settings (default/max EBS size) are configured at the domain level.

Spaces are **not modeled in v1** — they would be a separate resource kind.

### Apps

**Apps** are the running compute instances within a user profile or space:

- Each app is backed by an EC2 instance with a specific instance type.
- Apps are created on-demand by users (or programmatically via API).
- App types include JupyterLab, KernelGateway, JupyterServer (legacy), Code Editor, and others.
- Apps are ephemeral — they can be created, stopped, and deleted independently.
- App lifecycle is managed by SageMaker, not by the domain provisioning layer.

---

## Authentication Modes

### IAM Authentication

| Aspect | Detail |
|--------|--------|
| **Identity source** | AWS IAM users, roles, or federated identities |
| **User mapping** | Each user profile maps to an IAM principal |
| **Access control** | IAM policies on the execution role |
| **Setup complexity** | Low — works with existing AWS credentials |
| **Best for** | Single-account teams, programmatic access, development environments |

With IAM authentication:
- Users authenticate using AWS credentials (access keys, CLI profiles, instance profiles, IRSA).
- Each user profile is associated with an IAM user or role ARN.
- Permissions are controlled by the execution role's IAM policies.
- No external identity provider required.

### SSO Authentication

| Aspect | Detail |
|--------|--------|
| **Identity source** | AWS IAM Identity Center (formerly AWS SSO) |
| **User mapping** | Each user profile maps to an SSO user or group |
| **Access control** | IAM policies on the execution role + SSO permission sets |
| **Setup complexity** | Medium — requires IAM Identity Center configuration |
| **Best for** | Enterprise teams, centralized identity, multi-account organizations |

With SSO authentication:
- Users authenticate via their corporate identity provider (Okta, Azure AD, SAML, etc.) through IAM Identity Center.
- SageMaker creates an IAM Identity Center application for the domain (ARN available in `single_sign_on_application_arn` output).
- User lifecycle (onboarding, offboarding) is managed through the identity provider.
- Users access Studio via a presigned URL that SSO generates.

### Tradeoffs

| Consideration | IAM | SSO |
|---|---|---|
| Setup time | Minutes | Hours (IdP integration) |
| User lifecycle | Manual (create IAM users) | Automated (synced from IdP) |
| MFA | IAM MFA (per-user) | IdP-managed (centralized) |
| Cross-account | Complex (cross-account roles) | Native (IAM Identity Center) |
| Audit trail | CloudTrail (IAM events) | CloudTrail + IdP logs |
| Domain migration | Cannot change after creation | Cannot change after creation |
| Recommendation | Dev/staging | Production/enterprise |

**Critical:** `auth_mode` is **ForceNew**. Changing from IAM to SSO (or vice versa) requires destroying and recreating the domain, including all user profiles, spaces, and apps.

---

## Network Modes

### PublicInternetOnly (Default)

- Domain ENIs have standard VPC routing — internet access via internet gateway or NAT.
- Notebooks can directly reach public endpoints: PyPI, GitHub, Docker Hub, public APIs.
- Data can egress the VPC to the internet.
- Simpler setup — no additional VPC endpoints required.

**Use case:** Development environments, workshops, prototyping.

### VpcOnly

- All traffic from notebooks and training jobs stays within the VPC.
- No direct internet access from domain ENIs.
- Internet access requires a NAT gateway in the VPC (for pip/conda installs, Git clones).
- Docker access is only available in VpcOnly mode.
- `vpc_only_trusted_accounts` (in Docker settings) restricts image pulls to approved AWS accounts.

**Required VPC endpoints for VpcOnly mode:**

| Endpoint | Service | Purpose |
|----------|---------|---------|
| `com.amazonaws.<region>.sagemaker.api` | SageMaker API | API calls from notebooks |
| `com.amazonaws.<region>.sagemaker.runtime` | SageMaker Runtime | Inference endpoint calls |
| `com.amazonaws.<region>.sts` | STS | Role assumption |
| `com.amazonaws.<region>.logs` | CloudWatch Logs | Log delivery |
| `com.amazonaws.<region>.s3` | S3 (gateway) | Data access, model artifacts |
| `com.amazonaws.<region>.ecr.api` | ECR API | Container image metadata |
| `com.amazonaws.<region>.ecr.dkr` | ECR Docker | Container image pulls |
| `com.amazonaws.<region>.servicecatalog` | Service Catalog | SageMaker Projects |

**Security implications:**
- Prevents data exfiltration via direct internet access.
- All outbound traffic is visible in VPC Flow Logs.
- Security groups provide fine-grained network access control.
- Required for compliance frameworks (HIPAA, SOC2, PCI-DSS, FedRAMP).

**Use case:** Production environments, regulated industries, sensitive data processing.

---

## Storage Architecture

### EFS Home Directories

Every SageMaker Domain automatically creates a dedicated Amazon EFS file system:

- **Mount point:** Each user profile gets `/home/<user-profile-name>/` on the shared EFS volume.
- **Lifecycle:** The EFS file system is created with the domain and deleted when the domain is deleted.
- **Encryption:** Encrypted at rest with the KMS key specified in `kms_key_id` (or the default `aws/elasticfilesystem` key).
- **Performance:** General Purpose mode by default. Burst throughput is sufficient for most notebook workloads.
- **Durability:** EFS stores data redundantly across multiple AZs within the region.
- **Backup:** Not automatically backed up — use AWS Backup to create EFS backup policies.

**What goes on EFS:**
- Jupyter notebooks (`.ipynb` files)
- Python scripts and code
- Configuration files (`.bashrc`, `.gitconfig`, conda environments)
- Small datasets (< tens of GB)

**What should NOT go on EFS:**
- Large datasets (use S3 instead)
- Docker images (use EBS space storage)
- Temporary build artifacts (use EBS space storage)

### EBS Working Storage (Spaces)

Spaces use EBS volumes for fast local storage:

- **Purpose:** Code checkouts, intermediate training data, model artifacts, Docker images.
- **Size:** Configured via `spaceStorageSettings` (default and maximum).
- **Performance:** gp3 volumes — 3,000 IOPS and 125 MiB/s baseline, burstable.
- **Lifecycle:** EBS volumes are attached to spaces and persist across app restarts within the same space.
- **Cost:** Per-GB-month for provisioned capacity.

### Storage Guidance

| Data Type | Recommended Storage | Reason |
|-----------|-------------------|--------|
| Notebooks & code | EFS (home dir) | Persistent, shared across apps |
| Large datasets | S3 | Scalable, cost-effective, decoupled from compute |
| Intermediate results | EBS (space) | Fast local I/O |
| Docker images | EBS (space) | Large, local-only |
| Model artifacts | S3 | Versioned, shareable |
| Conda environments | EFS (home dir) | Persistent across sessions |

---

## App Types

### JupyterLab (Recommended)

The primary IDE for SageMaker Studio. Modern web-based notebook and code editing environment.

| Feature | Detail |
|---------|--------|
| **Interface** | Tabbed notebooks, code editor, terminal, file browser |
| **Git integration** | Built-in Git GUI and CLI |
| **Extensions** | JupyterLab extensions for visualization, collaboration |
| **Auto-clone** | `codeRepositories` are cloned on startup |
| **Lifecycle configs** | Run scripts on startup (install packages, configure environment) |
| **Custom images** | Bring your own Docker images as kernels |
| **Idle shutdown** | Configurable via `idleSettings` |
| **Instance types** | `ml.t3.medium` to `ml.p3.16xlarge` |

### KernelGateway

Provides custom compute kernels backed by Docker images. Used when standard SageMaker kernels lack required libraries.

| Feature | Detail |
|---------|--------|
| **Purpose** | Custom ML framework environments |
| **Images** | Docker images registered via SageMaker Image + AppImageConfig |
| **Instance types** | CPU or GPU instances (`ml.t3.medium` to `ml.p3.16xlarge`) |
| **Lifecycle configs** | Run scripts on startup |
| **Use case** | Proprietary libraries, GPU-optimized stacks, specific framework versions |

### JupyterServer (Legacy)

The classic Jupyter Server interface. Being superseded by JupyterLab.

| Feature | Detail |
|---------|--------|
| **Interface** | Classic Jupyter Notebook interface |
| **Status** | Legacy — use JupyterLab for new domains |
| **Instance type** | Typically `system` (lightweight) |

### Code Editor

VS Code-based editor available in newer Studio versions.

| Feature | Detail |
|---------|--------|
| **Interface** | VS Code in the browser |
| **Status** | Not modeled in v1 — available in Studio but not configurable via domain settings |

---

## Idle Timeout and Cost Management

### How Idle Detection Works

SageMaker monitors user activity to determine if an app instance is idle:

- **Active signals:** Notebook cell execution, terminal commands, file saves, UI interactions, kernel activity.
- **Idle state:** No active signals for the configured timeout period.
- **Action:** Instance is automatically stopped. The user's EFS home directory and EBS space are preserved.
- **Restart:** Users relaunch the app from the Studio UI — cold start takes 1-5 minutes depending on instance type and lifecycle configs.

### Configuration Strategy

| Setting | Recommended Value | Purpose |
|---------|------------------|---------|
| `lifecycleManagement` | `ENABLED` | Always enable for production |
| `idleTimeoutInMinutes` | 120 (2 hours) | Balance cost savings vs restart friction |
| `minIdleTimeoutInMinutes` | 60 (1 hour) | Prevent disruptive short timeouts |
| `maxIdleTimeoutInMinutes` | 480 (8 hours) | Prevent users from disabling shutdown |

### Cost Impact

Without idle timeout, a single `ml.m5.large` instance running 24/7 costs approximately:
- **730 hours/month × ~$0.115/hour ≈ $84/month per user**

With a 2-hour idle timeout (assuming 8 hours active per day, 22 workdays/month):
- **176 hours/month × ~$0.115/hour ≈ $20/month per user**

For a 50-person data science team, idle timeout reduces compute costs from **~$4,200/month** to **~$1,000/month** — a 76% reduction.

---

## Docker Access

### Overview

Docker access in SageMaker Studio allows users to build, pull, and run Docker containers directly from notebooks and terminals. This is essential for:

- Building custom training containers.
- Testing inference images before deployment.
- Running containerized data processing pipelines.
- Creating reproducible ML environments.

### Requirements

| Requirement | Detail |
|-------------|--------|
| **Network mode** | `VpcOnly` only — Docker is not available in `PublicInternetOnly` mode |
| **Execution role** | Must have permissions for ECR, Docker socket access |
| **Trusted accounts** | `vpcOnlyTrustedAccounts` restricts which ECR registries can be used |

### Trusted Accounts

`vpcOnlyTrustedAccounts` is a critical security control:

- **Purpose:** Restricts Docker image pulls to images hosted in specified AWS accounts' ECR registries.
- **Common accounts to include:**
  - Your organization's account ID (for internal images).
  - Shared services account (for platform-wide base images).
  - `763104351884` — AWS Deep Learning Containers account (official PyTorch, TensorFlow, MXNet images).
  - `217643126080` — AWS SageMaker built-in algorithm images (some regions).
- **Security:** Prevents users from pulling arbitrary images from public registries or untrusted accounts.

### Security Implications

| Risk | Mitigation |
|------|-----------|
| Arbitrary image execution | `vpcOnlyTrustedAccounts` limits sources |
| Privilege escalation | SageMaker runs Docker in rootless mode |
| Network breakout | VpcOnly mode + security groups limit network access |
| Data exfiltration | No public internet; VPC Flow Logs for audit |
| Supply chain attacks | Pin image versions; use private ECR with vulnerability scanning |

---

## Cost Model

SageMaker Domain costs have several components:

### Compute (App Instances)

Charged per instance-hour while an app is running:

| Instance Type | vCPU | Memory | GPU | ~Price/hour |
|---------------|------|--------|-----|-------------|
| `ml.t3.medium` | 2 | 4 GB | — | $0.05 |
| `ml.m5.large` | 2 | 8 GB | — | $0.115 |
| `ml.m5.xlarge` | 4 | 16 GB | — | $0.23 |
| `ml.m5.2xlarge` | 8 | 32 GB | — | $0.461 |
| `ml.g4dn.xlarge` | 4 | 16 GB | 1× T4 | $0.736 |
| `ml.g5.xlarge` | 4 | 16 GB | 1× A10G | $1.408 |
| `ml.p3.2xlarge` | 8 | 61 GB | 1× V100 | $3.825 |

Prices are approximate US East (N. Virginia) pricing as of early 2026.

### EFS Storage (Home Directories)

- **Price:** ~$0.30/GB-month (EFS Standard), ~$0.016/GB-month (EFS Infrequent Access).
- **Lifecycle:** EFS Intelligent-Tiering moves infrequently accessed files automatically.
- **Typical usage:** 5-50 GB per user (notebooks, code, conda environments).

### EBS Storage (Spaces)

- **Price:** ~$0.08/GB-month (gp3).
- **Provisioned per space:** Charged for the entire provisioned volume, not just used capacity.
- **Typical usage:** 50-500 GB per space (datasets, Docker images, build artifacts).

### Data Transfer

- **Same AZ:** Free.
- **Cross AZ:** ~$0.01/GB each way.
- **To internet (via NAT):** Standard NAT gateway pricing (~$0.045/GB).
- **S3 access (via gateway endpoint):** Free.

### KMS

- **Customer-managed key:** $1/month per key + $0.03 per 10,000 API requests.
- **AWS-managed key:** Free (no per-request charges).

### Cost Optimization Strategies

1. **Enable idle timeout** — largest single cost saver (see [Idle Timeout section](#idle-timeout-and-cost-management)).
2. **Use `ml.t3.medium`** for development/exploration; reserve GPU instances for training.
3. **Set `maximumEbsVolumeSizeInGb`** to cap storage costs.
4. **Use S3** for large datasets instead of EFS or EBS.
5. **Use EFS Infrequent Access** for cold notebook archives.
6. **Use spot instances** for training jobs (not configurable at domain level, but relevant for SageMaker Training).

---

## Service Limits

Key AWS service limits for SageMaker Domains (as of early 2026):

| Limit | Default | Adjustable |
|-------|---------|------------|
| Domains per account per region | 5 | Yes |
| User profiles per domain | 1,000 | Yes |
| Apps per user profile | 5 per app type | Yes |
| Spaces per domain | 100 | Yes |
| Subnets per domain | 16 | No |
| Security groups per domain (domain-level) | 3 | No |
| Security groups per user profile | 5 | No |
| Custom images per app type | 200 | No |
| Code repositories per JupyterLab settings | 10 | No |
| Lifecycle config scripts per domain | 30 | Yes |
| Lifecycle config script size | 16 KB | No |
| EFS home directory per user | No hard limit (soft: storage cost) | — |
| EBS volume per space | Capped by `maximumEbsVolumeSizeInGb` | — |
| Docker trusted accounts | 20 | No |
| Idle timeout range | 60–525,600 minutes (1 hour to 365 days) | No |

---

## Security Model

### IAM Execution Role

The execution role is the primary security boundary for user workloads:

- **Trust policy:** Must allow `sagemaker.amazonaws.com` to assume the role.
- **Permissions:** S3 access (data, model artifacts), ECR access (custom images), Secrets Manager, CloudWatch Logs, etc.
- **Principle of least privilege:** Grant only the specific S3 buckets, ECR repos, and services needed.
- **Separate roles:** Consider separate execution roles for dev vs. prod domains.

**Minimal trust policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "sagemaker.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

### Network Security

| Layer | Control | Description |
|-------|---------|-------------|
| VPC | `vpc_id` | All domain traffic within a single VPC |
| Subnets | `subnet_ids` | ENI placement in specific subnets |
| Domain SGs | `domain_security_group_ids` | Domain-scoped shared resource access |
| User SGs | `security_group_ids` | Per-user notebook/app traffic |
| Domain boundary SG | Auto-created by AWS | Cross-app/cross-user traffic (output: `security_group_id_for_domain_boundary`) |
| Network mode | `app_network_access_type` | Public vs. VPC-only |

### Encryption

| Data | Encryption | Key |
|------|-----------|-----|
| EFS home directories | At rest (AES-256) | `kms_key_id` or `aws/elasticfilesystem` |
| EBS space volumes | At rest (AES-256) | AWS-managed |
| Shared notebook outputs | At rest | `sharing_settings.s3_kms_key_id` or bucket default |
| In transit | TLS 1.2+ | AWS-managed certificates |

### Compliance

VpcOnly mode + customer-managed KMS + security groups enables compliance with:
- HIPAA (BAA required with AWS)
- SOC 1, SOC 2, SOC 3
- PCI DSS
- FedRAMP
- GDPR (with appropriate data residency configuration)

---

## Common Patterns

### Data Science Team Workspace

A shared domain for a data science team with standardized environments:

- SSO authentication for corporate identity integration.
- JupyterLab as default IDE with team code repos pre-cloned.
- `ml.m5.large` for exploration; GPU instances available for training.
- 2-hour idle timeout for cost control.
- Custom images for approved ML frameworks.

### Secure ML Platform

Production ML platform with strict security controls:

- VpcOnly networking with all required VPC endpoints.
- Customer-managed KMS for EFS and S3 encryption.
- Domain and user-level security groups for defense in depth.
- Docker access restricted to approved ECR accounts.
- Notebook output sharing to encrypted S3 for audit trails.

### ML Education / Workshop

Temporary domain for training courses or hackathons:

- IAM authentication for simplicity.
- PublicInternetOnly for easy package installation.
- `ml.t3.medium` to minimize costs.
- Short idle timeout (60 minutes) to prevent forgotten instances.
- No custom images — use standard SageMaker kernels.

### Multi-Environment ML Platform

Separate domains for dev, staging, and prod:

- **Dev:** IAM auth, PublicInternetOnly, `ml.t3.medium`, liberal policies.
- **Staging:** IAM auth, VpcOnly, `ml.m5.large`, production-like policies.
- **Prod:** SSO auth, VpcOnly, customer KMS, strict SGs, Docker with trusted accounts, audit logging.

Each environment references its own VPC, security groups, and KMS keys via `valueFrom`.

---

## v2 Roadmap

Features under consideration for future versions of the AwsSagemakerDomain API:

### User Profiles (AwsSagemakerUserProfile)

- Per-user overrides of domain defaults (execution role, security groups, app settings).
- User-specific lifecycle configurations.
- Granular IAM policies per user.
- Modeled as a companion resource kind that references the domain.

### Spaces (AwsSagemakerSpace)

- Shared collaborative environments within a domain.
- Private spaces for individual work with dedicated EBS volumes.
- Space-level app configurations.
- Space ownership and sharing controls.

### RStudio Server Pro

- RStudio IDE integration for R-based data science teams.
- Requires RStudio license and separate package manager configuration.
- Different resource spec model (RStudio-specific instance types).

### SageMaker Canvas

- No-code ML interface for business analysts.
- AutoML model building without writing code.
- Separate app type with distinct settings.

### Code Editor (VS Code)

- VS Code in the browser as an alternative to JupyterLab.
- Configurable via domain settings (extensions, settings sync).

### Advanced Features Under Evaluation

- **Model Registry integration** — domain-level model governance policies.
- **Feature Store** — domain-level feature group access controls.
- **Pipeline defaults** — default pipeline execution roles and storage.
- **Custom file systems** — mount additional EFS or FSx volumes beyond the home directory.
- **Resource quotas** — per-user compute budgets and instance type restrictions.
- **Tagging policies** — auto-tagging of all domain-created resources for cost allocation.
