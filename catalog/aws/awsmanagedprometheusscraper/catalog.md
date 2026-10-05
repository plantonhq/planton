# AWS Managed Prometheus Scraper

Deploys an Amazon Managed Prometheus (AMP) scraper — AWS's agentless Prometheus collector that discovers and scrapes an EKS cluster's pods (or static endpoints in a VPC) and remote-writes the metrics to an AMP workspace or a CloudWatch dataset. No DaemonSet to install, no agent upgrades, no collector capacity planning. The spec enforces exactly one source arm and exactly one destination arm; the whole source replaces on change, while alias, destination, role configuration, and the scrape configuration update in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Prometheus Scraper** — the AWS-managed collector, placed into your subnets, with its source (an EKS cluster via `sourceEks`, or a bare VPC placement via `sourceVpc`), its destination (`ampWorkspaceArn` or `cloudwatchDatasetArn`), the Prometheus scrape configuration, and the optional cross-account role pair. Creates run long — AWS provisions collector infrastructure for up to 30 minutes, and deletes drain for up to 20
- **Scraper Logging Configuration** — created only when `logging` is set; sends the scraper's component logs (SERVICE_DISCOVERY, COLLECTOR, EXPORTER) to a CloudWatch log group

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with AMP scraper permissions, plus EKS describe and access-entry permissions for EKS sources. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The destination** — an AMP workspace (reference an AwsManagedPrometheus's `workspace_arn` output) or a CloudWatch dataset ARN.
- **At least two subnets** — CreateScraper rejects fewer ("Number of subnets must be at least 2"); for EKS sources use the cluster's subnets.
- **EKS arm only** — the cluster itself, referenced by ARN. AWS also needs the cluster's access entries to allow the scraper's role, which AWS manages automatically for same-account setups.
- **VPC arm only** — at least one security group whose rules let the collectors reach your scrape targets' ports.

## Deploy

### Console

Open the deployment store, find **AWS Managed Prometheus Scraper**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: the source arm (EKS cluster or VPC placement), the destination, and the optional scrape configuration, role pair, and logging sections. Start from the **EKS Cluster Scraper (AWS Default Configuration)** preset in the [Presets](#presets) tab for the day-1 shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsManagedPrometheusScraper
metadata:
  name: platform-scraper
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  alias: platform-scraper
  sourceEks:
    clusterArn:
      valueFrom:
        kind: AwsEksCluster
        name: platform-eks
        fieldPath: status.outputs.cluster_arn
    subnetIds:
      - value: subnet-0a1b2c3d4e5f60001
      - value: subnet-0a1b2c3d4e5f60002
  ampWorkspaceArn:
    valueFrom:
      kind: AwsManagedPrometheus
      name: platform-metrics
      fieldPath: status.outputs.workspace_arn
```

```shell
planton apply -f aws-managed-prometheus-scraper.yaml
```

This provisions collectors into the two subnets, resolves AWS's published default scrape configuration for EKS (kubelet, cAdvisor, pod service discovery — `scrapeConfiguration` was left unset), and remote-writes the cluster's metrics into the referenced workspace. An Infra Job tracks the provisioning in real time.

### InfraChart

When the scraper deploys alongside its workspace in one chart, wire the destination via ValueFromRef:

```yaml
spec:
  region: us-east-1
  sourceEks:
    clusterArn:
      valueFrom:
        kind: AwsEksCluster
        name: platform-eks
        fieldPath: status.outputs.cluster_arn
    subnetIds:
      - value: subnet-0a1b2c3d4e5f60001
      - value: subnet-0a1b2c3d4e5f60002
  ampWorkspaceArn:
    valueFrom:
      kind: AwsManagedPrometheus
      name: platform-metrics
      fieldPath: status.outputs.workspace_arn
```

The InfraPipeline resolves the dependency graph, provisions the workspace first, then points the scraper's collectors at it.

## Key Configuration

These are the most important decisions when configuring a scraper. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One source, one destination** — the spec enforces exactly one of `sourceEks` / `sourceVpc` and exactly one of `ampWorkspaceArn` / `cloudwatchDatasetArn`. A scraper needs no AMP workspace to exist (it can deliver to CloudWatch), which is why it is its own kind rather than a workspace satellite.

**The source is a replacement boundary** — changing anything under the source (cluster, subnets, security groups) re-provisions the collector, and re-provisioning means a scrape gap of many minutes. When continuity matters, roll a second scraper against the new source first, then retire the old one.

**Budget the lifecycle, not just the resource** — creates run up to 30 minutes and deletes drain up to 20. A scraper in a tight CI window will look hung when it is merely provisioning; both engines pin the provider's long timeouts, so give pipelines the same headroom.

**The default scrape configuration is EKS-only** — leaving `scrapeConfiguration` unset resolves AWS's published default at deploy: a sensible kubelet, cAdvisor, and pod-service-discovery baseline. VPC sources have no default (nothing to discover), so the spec requires your own `scrape_configs` there. Scrape-configuration edits apply in place — iterating on relabeling rules costs no replacement.

**Cross-account scraping is a role pair** — `roleConfiguration` takes both halves (the scraped account's role and the destination account's role) or neither. Even then, the AWS-managed `scraper_role_arn` still needs remote-write granted on the destination workspace's resource policy.

**Cost follows collectors and volume** — the drivers are the collector infrastructure AWS runs for you (per scraper, for as long as it exists) and the metric volume it pushes into the destination's ingestion. A generous scrape configuration against a large cluster is an ingestion decision, not just a discovery one.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEksCluster** | `sourceEks.clusterArn` | `status.outputs.cluster_arn` |
| **AwsSubnet** | `sourceEks.subnetIds[]` / `sourceVpc.subnetIds[]` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `sourceEks.securityGroupIds[]` / `sourceVpc.securityGroupIds[]` | `status.outputs.security_group_id` |
| **AwsManagedPrometheus** | `ampWorkspaceArn` | `status.outputs.workspace_arn` |
| **AwsIamRole** | `roleConfiguration.sourceRoleArn` / `roleConfiguration.targetRoleArn` | `status.outputs.role_arn` |
| **AwsCloudwatchLogGroup** | `logging.logGroupArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `scraper_role_arn` | The AWS-managed role the scraper writes to its destination with | Granting `aps:RemoteWrite` in a cross-account destination workspace's resource policy |

`scraper_id` and `scraper_arn` are also present — the AWS-generated identity, echoed for audit and import addressing rather than composition.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EKS cluster with the AWS default configuration** — point the scraper at the cluster and workspace and leave `scrapeConfiguration` unset; metrics flow with zero agents installed. The trade is control: the default scrapes AWS's chosen baseline, so bring your own configuration when you need custom relabeling or job selection. Start from the **EKS Cluster Scraper (AWS Default Configuration)** preset.

**VPC scraper for static targets** — collectors placed into private subnets scrape node exporters or any Prometheus endpoint by address, no EKS required. The security group must allow egress to the target ports, and the scrape configuration is yours to write. Start from the **VPC Scraper for Static Targets** preset.

**Platform smoke test** — a VPC-sourced scraper against a static target proves the whole lifecycle without an EKS cluster: collectors provision, and an empty scrape is a healthy scraper. Useful for validating the pipeline before pointing at production clusters.

## Works With

- [**AWS Managed Prometheus**](/infra-catalog/aws-managed-prometheus) — the usual destination, wired via `ampWorkspaceArn` from the workspace's `workspace_arn`
- [**AWS EKS Cluster**](/infra-catalog/aws-eks-cluster) — the scraped source for the EKS arm, wired via `clusterArn`
- [**AWS Subnet**](/infra-catalog/aws-subnet) — where the collectors place their network interfaces (at least two)
- [**AWS Security Group**](/infra-catalog/aws-security-group) — controls the collectors' reach to scrape targets; required for VPC sources
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the source and target halves of cross-account scraping
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) — where the scraper's component logs land
