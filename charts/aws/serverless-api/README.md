# AWS Serverless API

Provisions a production-ready serverless API backend on Lambda with HTTP API Gateway, optional DynamoDB database, SQS async processing, SNS notifications, and Cognito authentication. This is a composable serverless microservices pattern -- enable only the components your API needs.

This chart is the AWS equivalent of the [GCP Serverless API Backend](../../gcp/serverless-api-backend/). Instead of Cloud Run + Cloud SQL + Pub/Sub, it uses Lambda + DynamoDB + SQS -- the canonical AWS serverless stack.

## Architecture

```
                        Clients
                           │
                           ▼
                 ┌───────────────────┐
                 │ AwsHttpApiGateway │
                 │ (HTTP API + CORS) │
                 │  $default stage   │
                 └────────┬──────────┘
                          │ AWS_PROXY
                          ▼
                 ┌───────────────────┐
                 │    AwsLambda      │
                 │  (API handler)    │
                 │                   │
                 └───┬────┬────┬────┘
                     │    │    │
          ┌──────────┘    │    └──────────┐
          ▼               ▼               ▼
  ┌──────────────┐ ┌────────────┐ ┌─────────────┐
  │ AwsDynamodb  │ │AwsSqsQueue │ │AwsSnsTopic  │
  │  (database)  │ │  (async)   │ │(fan-out)    │
  └──────────────┘ └────────────┘ └─────────────┘

  ┌──────────────────────┐  ┌──────────────────────────┐
  │  AwsIamRole          │  │  AwsCloudwatchLogGroup   │
  │  (execution role)    │  │  (API access logs)       │
  └──────────────────────┘  └──────────────────────────┘

  ┌──────────────────────┐
  │ AwsCognitoUserPool   │
  │ (authentication)     │
  └──────────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsIamRole, AwsCloudwatchLogGroup, AwsDynamodb,
                     AwsSqsQueue, AwsSnsTopic, AwsCognitoUserPool
Layer 1 (dep IAM):   AwsLambda
Layer 2 (dep Lambda): AwsHttpApiGateway
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| IAM Role | `AwsIamRole` | identity | Always | Lambda execution role with scoped permissions |
| CloudWatch Log Group | `AwsCloudwatchLogGroup` | monitoring | Always | API Gateway access logs |
| Lambda Function | `AwsLambda` | compute | Always | API request handler |
| HTTP API Gateway | `AwsHttpApiGateway` | compute | Always | HTTP routing with CORS and access logging |
| DynamoDB Table | `AwsDynamodb` | database | `databaseEnabled` | NoSQL data persistence |
| SQS Queue | `AwsSqsQueue` | messaging | `messagingEnabled` | Asynchronous task processing |
| SNS Topic | `AwsSnsTopic` | messaging | `notificationsEnabled` | Fan-out event notifications |
| Cognito User Pool | `AwsCognitoUserPool` | security | `authEnabled` | User authentication and JWT tokens |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| **Lambda** | | | |
| `lambda_function_name` | Function name | `api-handler` | Yes |
| `lambda_code_source_type` | `CODE_SOURCE_TYPE_IMAGE` or `CODE_SOURCE_TYPE_S3` | `CODE_SOURCE_TYPE_IMAGE` | Yes |
| `lambda_image_uri` | ECR image URI (for IMAGE type) | `""` | When IMAGE |
| `lambda_s3_bucket` | S3 bucket with code zip (for S3 type) | `""` | When S3 |
| `lambda_s3_key` | S3 object key for code zip (for S3 type) | `""` | When S3 |
| `lambda_runtime` | Runtime (for S3 type, e.g., `nodejs20.x`) | `nodejs20.x` | When S3 |
| `lambda_handler` | Handler (for S3 type, e.g., `index.handler`) | `index.handler` | When S3 |
| `lambda_memory_mb` | Memory in MB (128-10240) | `256` | Yes |
| `lambda_timeout_seconds` | Timeout in seconds (1-900) | `30` | Yes |
| **API Gateway** | | | |
| `api_name` | HTTP API name | `api-gateway` | Yes |
| `api_description` | API description | `Serverless API backend` | No |
| **Database** | | | |
| `databaseEnabled` | Create DynamoDB table | `true` | No |
| `dynamodb_table_name` | Table name | `api-data` | No |
| `dynamodb_hash_key` | Partition key attribute | `id` | No |
| **Messaging** | | | |
| `messagingEnabled` | Create SQS queue | `false` | No |
| `sqs_queue_name` | Queue name | `api-tasks` | No |
| **Notifications** | | | |
| `notificationsEnabled` | Create SNS topic | `false` | No |
| `sns_topic_name` | Topic name | `api-notifications` | No |
| **Authentication** | | | |
| `authEnabled` | Create Cognito User Pool | `false` | No |
| `cognito_pool_name` | User pool name | `api-users` | No |
| `cognito_client_name` | App client name | `api-client` | No |

## Common Configurations

### Minimal (Lambda + API Gateway + DynamoDB)

```yaml
databaseEnabled: true
messagingEnabled: false
notificationsEnabled: false
authEnabled: false
lambda_code_source_type: CODE_SOURCE_TYPE_IMAGE
lambda_image_uri: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-api:latest
```

### Full Serverless Stack

```yaml
databaseEnabled: true
messagingEnabled: true
notificationsEnabled: true
authEnabled: true
lambda_code_source_type: CODE_SOURCE_TYPE_IMAGE
lambda_image_uri: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-api:latest
```

### S3 Zip Deployment

```yaml
lambda_code_source_type: CODE_SOURCE_TYPE_S3
lambda_s3_bucket: my-deployments-bucket
lambda_s3_key: api-handler/v1.0.0.zip
lambda_runtime: python3.12
lambda_handler: app.lambda_handler
```

## Lambda Environment Variables

The Lambda function receives these environment variables automatically based on enabled components:

| Variable | Condition | Value |
|----------|-----------|-------|
| `API_NAME` | Always | API Gateway name |
| `TABLE_NAME` | `databaseEnabled` | DynamoDB table name |
| `QUEUE_NAME` | `messagingEnabled` | SQS queue name |
| `TOPIC_NAME` | `notificationsEnabled` | SNS topic name |

Use the AWS SDK to interact with these resources by name. The SDK resolves full ARNs and URLs automatically.

## IAM Permissions

The execution role dynamically includes permissions based on enabled components:

| Permission | Condition |
|------------|-----------|
| `AWSLambdaBasicExecutionRole` (CloudWatch Logs) | Always |
| DynamoDB: GetItem, PutItem, UpdateItem, DeleteItem, Query, Scan | `databaseEnabled` |
| SQS: SendMessage, ReceiveMessage, DeleteMessage | `messagingEnabled` |
| SNS: Publish | `notificationsEnabled` |

Permissions are scoped to the specific resources created by this chart (not wildcarded to all resources).

## Adding JWT Authentication

When `authEnabled` is true, the chart creates a Cognito User Pool with an app client. To protect your API routes with JWT authentication, update the API Gateway after deployment:

1. Note the Cognito outputs from the deployment:
   - `user_pool_endpoint` -- used as the JWT issuer
   - `client_ids.<cognito_client_name>` -- used as the JWT audience

2. Update the API Gateway resource to add a JWT authorizer and secure routes:
   ```yaml
   authorizers:
     - name: cognito-auth
       authorizerType: JWT
       jwtConfiguration:
         issuer: "<user_pool_endpoint from Cognito outputs>"
         audiences:
           - "<client_id from Cognito outputs>"
       identitySources:
         - "$request.header.Authorization"
   routes:
     - routeKey: "$default"
       authorizationType: JWT
       authorizerName: cognito-auth
       integration:
         ...
   ```

This two-step process is required because the API Gateway JWT authorizer configuration uses plain strings (not cross-resource references) for the issuer and audience fields.

## Important Notes

- **Lambda code is required.** The Lambda function will not deploy until valid code is provided via `lambda_image_uri` (ECR image) or `lambda_s3_bucket` + `lambda_s3_key` (S3 zip). Build and push your image or upload your zip before deploying this chart.
- The API Gateway creates a `$default` catch-all route that proxies all requests to the Lambda function. Add specific routes (e.g., `GET /users`, `POST /orders`) by updating the API Gateway resource after deployment.
- DynamoDB uses **on-demand billing** (PAY_PER_REQUEST) for zero management and automatic scaling. Switch to provisioned capacity after deployment if cost optimization is needed.
- The SQS queue uses **Standard** type (at-least-once delivery, best-effort ordering). Switch to FIFO by updating the queue resource if exactly-once processing is required.
- CORS is configured to allow all origins (`*`). Restrict `allowOrigins` to your specific domain for production use.
