---
title: "App Runner Service"
description: "App Runner Service deployment documentation"
icon: "package"
order: 100
componentName: "awsapprunnerservice"
---

# AWS App Runner Service

Deploys an AWS App Runner Service from a container image or GitHub repository with automatic HTTPS, concurrency-based auto scaling, optional VPC egress via an inline VPC Connector, and optional customer-managed KMS encryption. The component supports two mutually exclusive source types: ECR image or GitHub code repository.

## What Gets Created

When you deploy an AwsAppRunnerService resource, OpenMCF provisions:

- **App Runner Service** — an `aws:apprunner:Service` resource with source configuration (image or code), instance sizing, health checks, networking, encryption, and observability settings
- **VPC Connector** — created only when `subnetIds` are provided and no existing `vpcConnectorArn` is referenced, allowing the service to reach resources in your VPC (databases, caches, internal APIs)
- **Auto Scaling Configuration Version** — created only when the `autoScaling` block is provided, controlling min/max instance counts and per-instance concurrency limits

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A container image in ECR or ECR Public** if using image-based deployment
- **An IAM access role** with ECR pull permissions if using a private ECR image (`image_repository_type: ECR`)
- **An App Runner Connection** (created via AWS Console or CLI with GitHub OAuth) if using code-based deployment
- **VPC subnets and security groups** if the service needs to access VPC resources
- **A customer-managed KMS key ARN** if encrypting the service with your own key (note: changing this value requires replacing the service)

## Quick Start

Create a file `app-runner.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsAppRunnerService.my-api
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: public.ecr.aws/nginx/nginx:latest
    imageRepositoryType: ECR_PUBLIC
```

Deploy:

```shell
openmcf apply -f app-runner.yaml
```

This creates a publicly accessible App Runner service running an ECR Public image with default settings: 1 vCPU, 2 GB memory, port 8080, and auto scaling from 1 to 25 instances.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the App Runner service will be created (e.g., `us-east-1`, `eu-west-1`). | Required; non-empty |
| `imageSource` or `codeSource` | `object` | Deployment source. Exactly one must be provided. | CEL: `exactly_one_source` |
| `imageSource.imageIdentifier` | `string` | Full container image URI including tag or digest (e.g., `ACCOUNT.dkr.ecr.REGION.amazonaws.com/REPO:TAG`). | Min length 1 |
| `imageSource.imageRepositoryType` | `string` | Type of image repository: `ECR` (private) or `ECR_PUBLIC`. | Must be `ECR` or `ECR_PUBLIC` |
| `imageSource.accessRoleArn` | `string` | IAM role ARN that grants App Runner permission to pull from private ECR. Required when `imageRepositoryType` is `ECR`. Can reference AwsIamRole via `valueFrom`. | Required for `ECR` |
| `codeSource.repositoryUrl` | `string` | GitHub repository URL (e.g., `https://github.com/owner/repo`). | Min length 1 |
| `codeSource.branch` | `string` | Branch name to deploy from (e.g., `main`). | Min length 1 |
| `codeSource.connectionArn` | `string` | ARN of an App Runner Connection for GitHub access. | Required |
| `codeSource.configurationSource` | `string` | Where build/runtime config comes from: `API` (inline in spec) or `REPOSITORY` (apprunner.yaml in repo). | Must be `API` or `REPOSITORY` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `port` | `string` | `8080` | Port the application listens on. App Runner routes incoming HTTPS traffic to this port. |
| `startCommand` | `string` | — | Override the container start command. For image source, overrides ENTRYPOINT/CMD. |
| `environmentVariables` | `map<string, string>` | `{}` | Plaintext environment variables injected at runtime. Keys prefixed with `AWSAPPRUNNER` are reserved. |
| `environmentSecrets` | `map<string, string>` | `{}` | Environment secrets as ARNs of Secrets Manager secrets or SSM Parameter Store parameters. The `instanceRoleArn` must have read permissions. |
| `cpu` | `string` | `1024` | CPU per instance. Accepts `256`, `512`, `1024`, `2048`, `4096` or `0.25 vCPU`, `0.5 vCPU`, `1 vCPU`, `2 vCPU`, `4 vCPU`. |
| `memory` | `string` | `2048` | Memory per instance. Accepts `512`–`12288` (MB) or `0.5 GB`–`12 GB`. Not all CPU/memory combinations are valid. |
| `instanceRoleArn` | `string` | — | IAM role instances assume at runtime to call AWS APIs (S3, DynamoDB, etc.). Can reference AwsIamRole via `valueFrom`. |
| `healthCheck.protocol` | `string` | `TCP` | Health check protocol: `TCP` (port open) or `HTTP` (GET request expecting 200). |
| `healthCheck.path` | `string` | `/` | URL path for HTTP health checks. Ignored when protocol is `TCP`. |
| `healthCheck.intervalSeconds` | `int` | `5` | Seconds between health checks. Range: 1–20. |
| `healthCheck.timeoutSeconds` | `int` | `2` | Seconds to wait for a health check response. Range: 1–20. |
| `healthCheck.healthyThreshold` | `int` | `1` | Consecutive successes to mark healthy. Range: 1–20. |
| `healthCheck.unhealthyThreshold` | `int` | `5` | Consecutive failures to mark unhealthy and trigger replacement. Range: 1–20. |
| `autoScaling.minSize` | `int` | `1` | Minimum warm instances kept running at all times. Range: 1–25. |
| `autoScaling.maxSize` | `int` | `25` | Maximum instances during traffic spikes. Range: 1–25. Must be >= `minSize`. |
| `autoScaling.maxConcurrency` | `int` | `100` | Concurrent requests per instance before scaling out. Range: 1–200. |
| `vpcConnectorArn` | `string` | — | ARN of an existing VPC Connector. Mutually exclusive with `subnetIds`. |
| `subnetIds` | `string[]` | `[]` | VPC subnet IDs for creating an inline VPC Connector. Mutually exclusive with `vpcConnectorArn`. Can reference AwsVpc via `valueFrom`. |
| `securityGroupIds` | `string[]` | `[]` | Security group IDs for the inline VPC Connector. Only used when `subnetIds` is provided. Can reference AwsSecurityGroup via `valueFrom`. |
| `isPubliclyAccessible` | `bool` | `true` | When `false`, the service is only reachable via a VPC Ingress Connection (not managed by this component). |
| `ipAddressType` | `string` | `IPV4` | IP address type for the endpoint: `IPV4` or `DUAL_STACK` (IPv4 + IPv6). |
| `kmsKeyArn` | `string` | — | Customer-managed KMS key ARN for encrypting the service's image and data logs. **ForceNew**: changing this replaces the service. Can reference AwsKmsKey via `valueFrom`. |
| `observabilityEnabled` | `bool` | `false` | Enables AWS X-Ray tracing. Requires `observabilityConfigurationArn`. |
| `observabilityConfigurationArn` | `string` | — | ARN of an App Runner Observability Configuration. Required when `observabilityEnabled` is `true`. |
| `autoDeploymentsEnabled` | `bool` | `true` | Automatically redeploy when the source image tag is pushed or the code branch receives a commit. |
| `codeSource.runtime` | `string` | — | Runtime for code builds. Required when `configurationSource` is `API`. Values: `PYTHON_3`, `NODEJS_12`–`NODEJS_18`, `CORRETTO_8`, `CORRETTO_11`, `GO_1`, `DOTNET_6`, `PHP_81`, `RUBY_31`. |
| `codeSource.buildCommand` | `string` | — | Shell command to build the application (e.g., `npm ci && npm run build`). Only used when `configurationSource` is `API`. |
| `codeSource.sourceDirectory` | `string` | — | Subdirectory containing the application source. Defaults to repository root. Useful for monorepos. |

## Examples

### Image from ECR Public with Custom Port

Deploy a public container image with a non-default port and reduced instance size:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: lightweight-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsAppRunnerService.lightweight-api
spec:
  imageSource:
    imageIdentifier: public.ecr.aws/myalias/my-app:latest
    imageRepositoryType: ECR_PUBLIC
  port: "3000"
  cpu: "512"
  memory: "1024"
```

### Private ECR Image with VPC Access

Deploy from a private ECR registry with VPC egress so the service can reach an RDS database:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: backend-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsAppRunnerService.backend-api
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: 123456789012.dkr.ecr.us-east-1.amazonaws.com/backend:v1.2.0
    imageRepositoryType: ECR
    accessRoleArn: arn:aws:iam::123456789012:role/apprunner-ecr-access
  port: "8080"
  cpu: "1024"
  memory: "2048"
  instanceRoleArn: arn:aws:iam::123456789012:role/backend-instance-role
  environmentVariables:
    DB_HOST: mydb.cluster-abc123.us-east-1.rds.amazonaws.com
    DB_PORT: "5432"
  environmentSecrets:
    DB_PASSWORD: arn:aws:secretsmanager:us-east-1:123456789012:secret:backend/db-password
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
  securityGroupIds:
    - sg-backend-egress
  healthCheck:
    protocol: HTTP
    path: /health
    intervalSeconds: 10
    timeoutSeconds: 5
    healthyThreshold: 1
    unhealthyThreshold: 3
```

### GitHub Code Source

Deploy directly from a GitHub repository using a managed Node.js runtime:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: web-frontend
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAppRunnerService.web-frontend
spec:
  region: us-east-1
  codeSource:
    repositoryUrl: https://github.com/my-org/web-frontend
    branch: main
    connectionArn: arn:aws:apprunner:us-east-1:123456789012:connection/github-conn/abc123
    configurationSource: API
    runtime: NODEJS_18
    buildCommand: npm ci && npm run build
  port: "3000"
  startCommand: npm start
  cpu: "1024"
  memory: "2048"
  autoScaling:
    minSize: 2
    maxSize: 10
    maxConcurrency: 80
```

### Full-Featured Production Deployment

Production configuration with private ECR, VPC networking, KMS encryption, X-Ray observability, tuned auto scaling, and controlled deployments:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAppRunnerService.prod-api
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: 123456789012.dkr.ecr.us-east-1.amazonaws.com/prod-api:v3.1.0
    imageRepositoryType: ECR
    accessRoleArn: arn:aws:iam::123456789012:role/apprunner-ecr-access
  port: "8080"
  cpu: "2048"
  memory: "4096"
  instanceRoleArn: arn:aws:iam::123456789012:role/prod-api-instance-role
  environmentVariables:
    LOG_LEVEL: info
    SERVICE_NAME: prod-api
  environmentSecrets:
    DATABASE_URL: arn:aws:secretsmanager:us-east-1:123456789012:secret:prod/database-url
    API_KEY: arn:aws:ssm:us-east-1:123456789012:parameter/prod/api-key
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-prod-api-egress
  healthCheck:
    protocol: HTTP
    path: /readyz
    intervalSeconds: 5
    timeoutSeconds: 2
    healthyThreshold: 1
    unhealthyThreshold: 3
  autoScaling:
    minSize: 3
    maxSize: 20
    maxConcurrency: 50
  kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/mrk-prod-encryption-key
  observabilityEnabled: true
  observabilityConfigurationArn: arn:aws:apprunner:us-east-1:123456789012:observabilityconfiguration/xray-config/1/abc123
  autoDeploymentsEnabled: false
  isPubliclyAccessible: true
  ipAddressType: DUAL_STACK
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: ref-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsAppRunnerService.ref-api
spec:
  region: us-east-1
  imageSource:
    imageIdentifier: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-api:latest
    imageRepositoryType: ECR
    accessRoleArn:
      valueFrom:
        kind: AwsIamRole
        name: ecr-access-role
        field: status.outputs.role_arn
  instanceRoleArn:
    valueFrom:
      kind: AwsIamRole
      name: api-instance-role
      field: status.outputs.role_arn
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: api-egress-sg
        field: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: service-encryption-key
      field: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `service_arn` | `string` | Full ARN of the App Runner Service |
| `service_id` | `string` | Unique identifier assigned by App Runner |
| `service_url` | `string` | Public HTTPS URL for the service (e.g., `abc123.us-east-1.awsapprunner.com`) |
| `service_name` | `string` | Computed name of the service (derived from `metadata.name`) |
| `service_status` | `string` | Current operational status (e.g., `RUNNING`, `CREATE_FAILED`, `OPERATION_IN_PROGRESS`) |
| `vpc_connector_arn` | `string` | ARN of the VPC Connector. Only set when VPC egress is configured via `subnetIds`. |
| `auto_scaling_configuration_arn` | `string` | ARN of the Auto Scaling Configuration Version. Only set when `autoScaling` is provided. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for VPC Connector egress
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls outbound traffic from the VPC Connector
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides the ECR access role and instance runtime role
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides the customer-managed encryption key
- [AwsEcrRepo](/docs/catalog/aws/ecr-repo) — hosts private container images for image-based deployments
