# AWS OpenSearch Domain

Deploys a managed Amazon OpenSearch Service domain with configurable cluster topology (data nodes, dedicated masters, coordinator node pools, UltraWarm, cold storage), EBS volume configuration, VPC or public deployment modes, fine-grained access control with JWT bearer authentication, Cognito and IAM Identity Center sign-in for Dashboards, KMS encryption, Auto-Tune with maintenance scheduling, AI/ML capabilities (natural-language query generation, the S3 vectors engine, GPU vector acceleration), and CloudWatch log publishing. Two decisions are one-way doors worth making deliberately: VPC placement is ForceNew (adding or removing it recreates the domain), and fine-grained access control cannot be disabled once enabled.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **OpenSearch Domain** -- a managed search and analytics domain running the specified engine version (OpenSearch or Elasticsearch), with configurable data node count, instance type, and EBS storage
- **Dedicated Master Nodes** -- created only when `clusterConfig.dedicatedMasterEnabled` is `true`; separate nodes that handle cluster management without competing with data workloads
- **Coordinator Node Pools** -- created only when `clusterConfig.nodeOptions` entries are configured; dedicated nodes that take over request routing, query fan-out, and response aggregation for high-concurrency workloads
- **UltraWarm Storage Tier** -- created only when `clusterConfig.warmEnabled` is `true`; S3-backed nodes for infrequently accessed read-only data at lower cost
- **Cold Storage Tier** -- created only when `clusterConfig.coldStorageEnabled` is `true`; lowest-cost S3-backed tier for rarely queried data (requires UltraWarm)
- **VPC Endpoints** -- created only when `vpcOptions` is configured; places the domain into VPC subnets with security group controls instead of public internet access
- **Fine-Grained Access Control** -- configured only when `advancedSecurityOptions.enabled` is `true`; enables internal user database or IAM master-user authentication, optional anonymous access, JWT bearer authentication (`jwtOptions`, OpenSearch 2.11+), and role-based index permissions
- **Cognito Dashboards Sign-in** -- configured only when `cognitoOptions.enabled` is `true`; users sign in to OpenSearch Dashboards through a Cognito user pool and are authorized through a Cognito identity pool
- **SAML Dashboards Sign-in** -- created only when `samlOptions` is configured (requires fine-grained access control); its own provider resource wiring an external SAML 2.0 identity provider, so removing the block disables SAML without touching the domain
- **Cross-Account VPC Endpoint Grants** -- created only when `authorizedVpcEndpointAccessAccounts` entries are configured; one grant per listed account authorizing it to create OpenSearch-managed VPC endpoints against this domain
- **IAM Identity Center Integration** -- configured only when `identityCenterOptions.enabledApiAccess` is `true`; workforce users reach Dashboards and the domain APIs with their Identity Center identity
- **Auto-Tune** -- configured through `autoTuneOptions`; AWS optimizes JVM heap, disk I/O, and queue sizes from live metrics, with blue/green optimizations scheduled through explicit maintenance windows or the domain's off-peak window (`offPeakWindowOptions`)
- **AI/ML Capabilities** -- configured through `aimlOptions`; natural-language query generation in Dashboards (OpenSearch 2.13+), the S3 vectors engine, and GPU-accelerated vector search
- **Log Publishing Pipelines** -- created only when `logPublishingOptions` entries are configured; streams index slow logs, search slow logs, application logs, or audit logs (requires fine-grained access control) to CloudWatch Logs
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Subnets** (optional, for VPC mode) in 2 or 3 Availability Zones matching the cluster's `availabilityZoneCount`. Private subnets are recommended. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **Security groups** (optional, for VPC mode) allowing HTTPS (port 443) inbound from clients that need to access the domain. Provide security group IDs directly or reference an AwsSecurityGroup Infra Component.
- **A KMS key** (optional) for at-rest encryption beyond the default AWS-managed `aws/es` key. The KMS key choice is ForceNew and cannot be changed after domain creation. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **CloudWatch log groups** (optional) for publishing domain logs. Each log type requires a dedicated log group. Provide log group ARNs directly or reference AwsCloudwatchLogGroup Infra Components.

## Deploy

### Console

Open the deployment store, find **AWS OpenSearch Domain**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-Node Development Domain** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsOpenSearchDomain
metadata:
  name: app-search
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 2
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
```

```shell
planton apply -f opensearch-domain.yaml
```

This creates a two-node OpenSearch domain with gp3 EBS volumes, encryption at rest and in transit using default AWS-managed keys, and no VPC placement (publicly accessible). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the OpenSearch domain to a VPC and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: search-encryption-key
      fieldPath: status.outputs.key_arn
  vpcOptions:
    subnetIds:
      - valueFrom:
          kind: AwsSubnet
          name: private-az1
          fieldPath: status.outputs.subnet_id
      - valueFrom:
          kind: AwsSubnet
          name: private-az2
          fieldPath: status.outputs.subnet_id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: search-sg
          fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the OpenSearch domain with the resolved values.

## Key Configuration

These are the most important decisions when configuring an OpenSearch domain. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Public vs. VPC deployment** -- By default, the domain is publicly accessible (no `vpcOptions`). Set `vpcOptions` with subnet and security group IDs to deploy into a VPC for production workloads. VPC configuration is ForceNew -- adding or removing VPC options destroys and recreates the domain.

**Cluster topology** -- Configure `clusterConfig` with `instanceType` and `instanceCount` for data nodes. Enable `dedicatedMasterEnabled` with 3 master nodes for production cluster stability. Enable `zoneAwarenessEnabled` with 2 or 3 AZs for resilience. Use `multiAzWithStandbyEnabled` for 99.99% availability SLA (requires 3 AZs).

**Storage tiers** -- EBS is the primary storage (`ebsOptions`). Enable `warmEnabled` for UltraWarm (S3-backed, lower cost for infrequently accessed data). Enable `coldStorageEnabled` for the lowest-cost tier (requires UltraWarm). This tiered approach optimizes cost for mixed hot/warm/cold data patterns.

**Encryption and access control** -- Enable `encryptAtRestEnabled` and `nodeToNodeEncryptionEnabled` for production. Enable `advancedSecurityOptions` for fine-grained access control with the internal user database or an IAM master user (mutually exclusive), optionally adding JWT bearer authentication for token-based service access. Once enabled, fine-grained access control cannot be disabled without recreating the domain. `accessPolicies` carries the domain's resource-based IAM policy -- the primary perimeter for public domains.

**Dashboards sign-in** -- Choose the sign-in story by audience: `cognitoOptions` puts a Cognito user pool login in front of Dashboards (all three integration inputs are required together), while `identityCenterOptions` extends workforce single sign-on to Dashboards and the domain APIs. Either can combine with fine-grained access control -- sign-in authenticates, FGAC authorizes.

**Self-maintenance** -- `autoTuneOptions` lets AWS continuously optimize the cluster (not supported on t2/t3 instance types); its blue/green optimizations run inside explicit maintenance schedules or the domain's daily 10-hour off-peak window (`offPeakWindowOptions`), never both. `autoSoftwareUpdateEnabled` applies service software updates in the same window, and `deploymentStrategy` controls how AWS provisions capacity for blue/green migrations.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsSubnet** (optional) | `vpcOptions.subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `vpcOptions.securityGroupIds` | `status.outputs.security_group_id` |
| **AwsIamRole** (optional) | `advancedSecurityOptions.masterUserArn` | `status.outputs.role_arn` |
| **AwsCognitoUserPool** (optional) | `cognitoOptions.userPoolId` | `status.outputs.user_pool_id` |
| **AwsIamRole** (optional) | `cognitoOptions.roleArn` | `status.outputs.role_arn` |
| **AwsCloudwatchLogGroup** (optional) | `logPublishingOptions[*].cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |
| **AwsCertManagerCert** (optional) | `domainEndpointOptions.customEndpointCertificateArn` | `status.outputs.cert_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `domain_id` | AWS-assigned domain identifier | Monitoring dashboards, operational scripts |
| `domain_name` | Domain name matching the metadata | API calls, connection configuration |
| `domain_arn` | Amazon Resource Name of the domain | IAM policies, cross-service permissions |
| `endpoint` | Domain endpoint for index and search requests | Application connection strings, API gateway targets |
| `dashboard_endpoint` | OpenSearch Dashboards URL | Browser access for visualization and management |
| `endpoint_v2` | Dual-stack (IPv4 + IPv6) V2 domain endpoint | Clients on IPv6 networks; works with both `ipAddressType` settings |
| `dashboard_endpoint_v2` | Dashboards URL on the dual-stack V2 endpoint | Browser access over IPv6 |
| `domain_endpoint_v2_hosted_zone_id` | Route 53 hosted zone ID for the V2 endpoint | Route53 alias records pointing DNS at the domain |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-node development** -- A single t3.small.search node with gp3 storage for development and testing. No VPC, no dedicated masters; encryption at rest and node-to-node stay on. Start from the **Single-Node Development Domain** preset.

**Production VPC** -- Multi-node domain deployed into a VPC with dedicated master nodes, zone awareness across 2 AZs, encryption at rest and in transit, and fine-grained access control enabled. Start from the **Production VPC Domain** preset.

**Analytics with warm and cold storage** -- Production domain with UltraWarm and cold storage tiers enabled for cost-optimized retention of time-series or log data. Suitable for observability and analytics workloads with mixed hot/warm/cold access patterns. Start from the **Analytics Domain with Warm + Cold Storage** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides subnets for VPC-mode domain placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides network access control for VPC-mode domains
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for at-rest encryption
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the FGAC master user and the Cognito service role
- [**AWS Cognito User Pool**](/infra-catalog/aws-cognito-user-pool) -- provides the sign-in directory for Dashboards
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- provides log destinations for domain log publishing
- [**AWS ACM Certificate**](/infra-catalog/aws-cert-manager-cert) -- provides a TLS certificate for custom domain endpoints