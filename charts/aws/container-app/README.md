# AWS Container App

The simplest container-to-URL deployment on AWS. Provisions an App Runner service with ECR for container images and optional DynamoDB for data persistence. Zero infrastructure management -- push a container image and get a public HTTPS URL.

This chart is the AWS equivalent of deploying to [GCP Cloud Run](../../gcp/cloud-run-environment/). No VPC, no load balancer, no cluster management -- just containers.

## Architecture

```
                       Users
                         │
                         ▼
              ┌─────────────────────┐
              │ AwsAppRunnerService │
              │  (auto-scaled)      │
              │  HTTPS endpoint     │
              └───┬─────────┬───────┘
                  │         │
       ┌──────────┘         └──────────┐
       ▼                               ▼
  ┌──────────────┐            ┌──────────────┐
  │  AwsEcrRepo  │            │ AwsDynamodb  │
  │ (images)     │            │ (database)   │
  └──────────────┘            └──────────────┘

  ┌──────────────────────┐
  │     AwsIamRole       │
  │  (instance role)     │
  └──────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsEcrRepo, AwsIamRole, AwsDynamodb
Layer 1 (dep IAM):   AwsAppRunnerService
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| ECR Repository | `AwsEcrRepo` | compute | Always | Container image registry |
| IAM Role | `AwsIamRole` | identity | Always | App Runner instance role with scoped permissions |
| App Runner Service | `AwsAppRunnerService` | compute | Always | Auto-scaled container service with HTTPS |
| DynamoDB Table | `AwsDynamodb` | database | `databaseEnabled` | NoSQL data persistence |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| **Service** | | | |
| `service_name` | App Runner service name | `my-app` | Yes |
| **Container** | | | |
| `ecr_repo_name` | ECR repository name | `my-app` | Yes |
| `image_uri` | Full ECR image URI with tag | `""` | Yes |
| `container_port` | Port the container listens on | `8080` | Yes |
| **Sizing** | | | |
| `cpu` | CPU units (256, 512, 1024, 2048, 4096) | `1024` | Yes |
| `memory` | Memory in MB (512-12288) | `2048` | Yes |
| **Database** | | | |
| `databaseEnabled` | Create DynamoDB table | `false` | No |
| `dynamodb_table_name` | Table name | `app-data` | No |
| `dynamodb_hash_key` | Partition key attribute | `id` | No |

## Common Configurations

### Minimal (just the container)

```yaml
databaseEnabled: false
image_uri: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest
container_port: "8080"
```

### With DynamoDB

```yaml
databaseEnabled: true
dynamodb_table_name: my-app-data
dynamodb_hash_key: pk
image_uri: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:latest
```

### Larger instance

```yaml
cpu: "4096"
memory: "12288"
```

## Important Notes

- **Container image is required.** The App Runner service will not deploy until a valid image exists at the `image_uri`. Build and push your image to ECR before deploying this chart.
- App Runner provides a **public HTTPS URL** automatically. No load balancer, certificate, or DNS configuration is needed for basic usage.
- App Runner auto-scales based on concurrent requests. There is no `replicas` parameter -- scaling is fully managed.
- DynamoDB uses **on-demand billing** (PAY_PER_REQUEST) for zero capacity planning. Switch to provisioned capacity after deployment if cost optimization is needed.
- The IAM instance role grants DynamoDB access only when `databaseEnabled` is true. Permissions are scoped to the specific table, not wildcarded.
- The `image_uri` must be the full ECR URI including the tag (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.0`). The ECR repository is created by this chart, but the image must be pushed separately.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
