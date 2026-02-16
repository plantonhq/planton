# AwsSagemakerDomain — Pulumi Module Architecture

This document describes the architecture of the Pulumi IaC module that provisions Amazon SageMaker Domains from the `AwsSagemakerDomainSpec`.

---

## File Structure

```
iac/pulumi/
├── Pulumi.yaml              # Pulumi project metadata
├── main.go                  # Entry point — loads stack input, calls module.Resources()
└── module/
    ├── main.go              # Orchestrator — creates resources in order, exports outputs
    ├── locals.go            # Locals struct — holds resolved AwsSagemakerDomain + labels map
    ├── domain.go            # SageMaker Domain resource with all configuration blocks
    └── outputs.go           # Output key constants (6 total)
```

---

## Resource Creation Flow

The `module.Resources()` function orchestrates resource creation in a strict dependency order:

```
1. initializeLocals()          → Locals struct (labels, resolved target)
2. domain()                    → *sagemaker.Domain
3. ctx.Export(...)             → 6 stack outputs
```

### Step 1: Initialize Locals

`locals.go` creates a `Locals` struct containing:
- `AwsSagemakerDomain` — the target resource from stack input.
- `Labels` — a standard tag map applied to all created resources.

### Step 2: SageMaker Domain

`domain.go` is the main resource creation file. It constructs the `sagemaker.DomainArgs` by mapping every spec field to the Pulumi AWS provider's `sagemaker.Domain` arguments.

**Key mapping logic:**

| Spec Section | Domain Argument | Notes |
|---|---|---|
| `authMode` | `AuthMode` | Direct string mapping: `IAM` or `SSO` |
| `vpcId` | `VpcId` | Resolved from `StringValueOrRef` |
| `subnetIds` | `SubnetIds` | Resolved list from `StringValueOrRef` |
| `kmsKeyId` | `KmsKeyId` | Optional; resolved from `StringValueOrRef` |
| `appNetworkAccessType` | `AppNetworkAccessType` | Default: `PublicInternetOnly` |
| `domainSecurityGroupIds` | `DomainSettings.SecurityGroupIds` | Optional list |
| `dockerSettings` | `DomainSettings.DockerSettings` | Conditional block |
| `defaultUserSettings.executionRoleArn` | `DefaultUserSettings.ExecutionRole` | Resolved from `StringValueOrRef` |
| `defaultUserSettings.securityGroupIds` | `DefaultUserSettings.SecurityGroups` | Optional resolved list |
| `defaultUserSettings.defaultLandingUri` | `DefaultUserSettings.DefaultLandingUri` | Optional string |
| `defaultUserSettings.studioWebPortal` | `DefaultUserSettings.StudioWebPortal` | Default: `ENABLED` |
| `defaultUserSettings.jupyterLabAppSettings` | `DefaultUserSettings.JupyterLabAppSettings` | Conditional block; see below |
| `defaultUserSettings.kernelGatewayAppSettings` | `DefaultUserSettings.KernelGatewayAppSettings` | Conditional block; see below |
| `defaultUserSettings.sharingSettings` | `DefaultUserSettings.SharingSettings` | Conditional block |
| `defaultUserSettings.spaceStorageSettings` | `DefaultUserSettings.SpaceStorageSettings` | Conditional block |

**JupyterLab settings mapping:**

| Spec Field | Pulumi Argument | Notes |
|---|---|---|
| `defaultResourceSpec` | `DefaultResourceSpec` | Instance type, lifecycle config, image ARN |
| `lifecycleConfigArns` | `LifecycleConfigArns` | List of ARNs |
| `customImages` | `CustomImages` | List of `{AppImageConfigName, ImageName, ImageVersionNumber}` |
| `codeRepositories` | `CodeRepositories` | List of `{RepositoryUrl}` |
| `idleSettings` | `EmrSettings` / custom | `LifecycleManagement`, `IdleTimeoutInMinutes`, min/max bounds |

**KernelGateway settings mapping:**

| Spec Field | Pulumi Argument | Notes |
|---|---|---|
| `defaultResourceSpec` | `DefaultResourceSpec` | Instance type, lifecycle config |
| `lifecycleConfigArns` | `LifecycleConfigArns` | List of ARNs |
| `customImages` | `CustomImages` | List of custom image definitions |

**Docker settings mapping:**

| Spec Field | Pulumi Argument | Notes |
|---|---|---|
| `enableDockerAccess` | `EnableDockerAccess` | `ENABLED` or `DISABLED` |
| `vpcOnlyTrustedAccounts` | `VpcOnlyTrustedAccounts` | List of AWS account IDs |

### Step 3: Export Outputs

`main.go` exports 6 outputs using the constants defined in `outputs.go`:

| Output Key | Source | Conditional |
|---|---|---|
| `domain_id` | `domain.Id` | No |
| `domain_arn` | `domain.Arn` | No |
| `domain_url` | `domain.Url` | No |
| `home_efs_file_system_id` | `domain.HomeEfsFileSystemId` | No |
| `security_group_id_for_domain_boundary` | `domain.SecurityGroupIdForDomainBoundary` | No |
| `single_sign_on_application_arn` | `domain.SingleSignOnApplicationArn` | **Yes** — only when `authMode` is `SSO` |

The 5 core outputs are always exported. The `single_sign_on_application_arn` output is only populated when SSO authentication is used; it returns an empty string for IAM-authenticated domains.

---

## AWS Provider Configuration

The entry point (`main.go`) loads stack input via `stackinput.LoadStackInput()` and passes it to `module.Resources()`.

The provider is configured in `module/main.go`:
- If `ProviderConfig` is nil → default AWS provider (ambient credentials).
- If `ProviderConfig` is set → explicit provider with `AccessKey`, `SecretKey`, `Region`, and optional `SessionToken`.

The provider instance is passed to every resource via `pulumi.Provider(provider)`.

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `pulumi/pulumi/sdk` | v3 | Pulumi SDK |
| `pulumi/pulumi-aws/sdk` | v7 | AWS Classic provider |
| `pkg/errors` | — | Error wrapping |
| `openmcf/.../awssagemakerdomain/v1` | — | Generated protobuf types |
| `openmcf/pkg/iac/pulumi/pulumimodule/stackinput` | — | Stack input loader |
