---
title: "CI/CD Deploy Pipeline"
description: "This preset creates a custom RAM policy combining the minimal permissions a CI/CD pipeline needs to build container images, deploy to ACK clusters, and write build logs to SLS. No single system..."
type: "preset"
rank: "02"
presetSlug: "02-cicd-deploy-pipeline"
componentSlug: "ram-policy"
componentTitle: "RAM Policy"
provider: "alicloud"
icon: "package"
order: 2
---

# CI/CD Deploy Pipeline

This preset creates a custom RAM policy combining the minimal permissions a CI/CD pipeline needs to build container images, deploy to ACK clusters, and write build logs to SLS. No single system policy covers this cross-service workflow, making it one of the most common reasons to create a custom policy.

## When to Use

- CI/CD pipelines (Jenkins, GitLab CI, GitHub Actions runners) that push images to Container Registry and deploy to ACK
- Build automation roles that need scoped access across CR, ACK, and SLS without granting full admin
- Platform teams creating standardized pipeline roles for multiple application teams
- Attaching to an `AliCloudRamRole` via `policyAttachments` with `policyType: Custom`

## Key Configuration Choices

- **Container Registry scoped to namespace** (`acs:cr:*:*:repository/<namespace>/*`) -- Pipelines can push and pull images only within their designated registry namespace. This prevents one team's pipeline from overwriting another team's images.
- **ACK read + kubeconfig access** (`DescribeClusterDetail`, `GetClusterKubeconfig`, `DescribeClusterNodes`, `DescribeClusterResources`) -- The minimal set needed to retrieve a kubeconfig and deploy workloads. Actual Kubernetes RBAC controls what happens inside the cluster; this policy only governs the Alibaba Cloud API layer.
- **SLS scoped to project** (`acs:log:*:*:project/<project>/*`) -- Build logs are written to a specific SLS project rather than granting account-wide log access. `PostLogStoreLogs` is the write action; `GetLogStore` and `ListLogStores` allow the pipeline to verify store existence before writing.
- **Force delete enabled** (`force: true`) -- Pipeline policies are frequently iterated as capabilities expand. Force-delete ensures clean teardown without manual detachment from roles.
- **Automatic version rotation** (`rotateStrategy: DeleteOldestNonDefaultVersionWhenLimitExceeded`) -- CI/CD policies evolve frequently as new services are integrated. Automatic rotation prevents the 5-version limit from blocking updates.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-policy-name>` | RAM policy name, unique per account (1-128 chars: letters, digits, hyphens) | Choose a name following your naming convention (e.g., `cicd-pipeline-deploy`) |
| `<your-registry-namespace>` | Container Registry namespace the pipeline pushes to | ACR console or `AliCloudContainerRegistryEnterprise` stack outputs |
| `<your-log-project-name>` | SLS project name for pipeline build logs | SLS console or `AliCloudLogProject` stack outputs |

## Related Presets

- **01-scoped-oss-access** -- Use instead when you need single-service bucket access rather than cross-service pipeline permissions
