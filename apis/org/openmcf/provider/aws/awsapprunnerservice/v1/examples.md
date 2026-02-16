# AwsAppRunnerService -- Examples

## 1. Minimal ECR Public image deployment

The simplest possible App Runner service. Deploys a public Nginx image with all defaults (1 vCPU, 2 GB RAM, port 8080, TCP health check, auto scaling 1--25 instances). No IAM roles needed because ECR Public images require no authentication.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: hello-apprunner
spec:
  imageSource:
    imageIdentifier: "public.ecr.aws/nginx/nginx:latest"
    imageRepositoryType: "ECR_PUBLIC"
  port: "80"
```

## 2. Private ECR image with custom port and environment variables

Deploys a private container image from your ECR registry. Requires an `access_role_arn` for pulling the image and an `instance_role_arn` so the running application can call AWS APIs (e.g., read from DynamoDB). Environment variables configure the app; secrets are injected from Secrets Manager.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-api-service
spec:
  imageSource:
    imageIdentifier: "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-api:v2.1.0"
    imageRepositoryType: "ECR"
    accessRoleArn:
      value: "arn:aws:iam::123456789012:role/apprunner-ecr-access"
  port: "3000"
  startCommand: "node dist/server.js"
  cpu: "1024"
  memory: "2048"
  instanceRoleArn:
    value: "arn:aws:iam::123456789012:role/my-api-instance-role"
  environmentVariables:
    NODE_ENV: "production"
    LOG_LEVEL: "info"
    DB_HOST: "mydb.cluster-abc123.us-east-1.rds.amazonaws.com"
  environmentSecrets:
    DB_PASSWORD: "arn:aws:secretsmanager:us-east-1:123456789012:secret:mydb-password-AbCdEf"
    API_KEY: "arn:aws:ssm:us-east-1:123456789012:parameter/my-api/api-key"
  healthCheck:
    protocol: "HTTP"
    path: "/health"
    intervalSeconds: 10
    timeoutSeconds: 5
    healthyThreshold: 1
    unhealthyThreshold: 3
```

## 3. Production deployment with VPC egress, KMS encryption, and custom scaling

A hardened production configuration. The service runs in a VPC (inline VPC Connector) so it can reach RDS, ElastiCache, and internal APIs. Data at rest is encrypted with a customer-managed KMS key. Auto scaling is tuned for predictable API traffic with a minimum of 2 warm instances to eliminate cold starts.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-payments-api
  tags:
    team: payments
    env: production
    cost-center: CC-1234
spec:
  imageSource:
    imageIdentifier: "123456789012.dkr.ecr.us-east-1.amazonaws.com/payments-api:v5.0.3"
    imageRepositoryType: "ECR"
    accessRoleArn:
      value: "arn:aws:iam::123456789012:role/apprunner-ecr-access"
  port: "8080"
  cpu: "2048"
  memory: "4096"
  instanceRoleArn:
    value: "arn:aws:iam::123456789012:role/payments-api-instance-role"

  # Secrets from Secrets Manager and SSM Parameter Store
  environmentVariables:
    SERVICE_NAME: "payments-api"
    LOG_LEVEL: "warn"
  environmentSecrets:
    DB_CONNECTION_STRING: "arn:aws:secretsmanager:us-east-1:123456789012:secret:payments/db-conn-XyZ123"
    STRIPE_SECRET_KEY: "arn:aws:secretsmanager:us-east-1:123456789012:secret:payments/stripe-key-AbC456"

  # VPC egress -- inline VPC Connector (2 AZs for HA)
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
  securityGroupIds:
    - value: "sg-0123456789abcdef0"

  # Customer-managed encryption (ForceNew)
  kmsKeyArn:
    value: "arn:aws:kms:us-east-1:123456789012:key/mrk-abcdef1234567890abcdef1234567890"

  # Tuned auto scaling
  autoScaling:
    minSize: 2
    maxSize: 10
    maxConcurrency: 50

  # HTTP health check
  healthCheck:
    protocol: "HTTP"
    path: "/healthz"
    intervalSeconds: 5
    timeoutSeconds: 3
    healthyThreshold: 1
    unhealthyThreshold: 3

  # Observability (X-Ray)
  observabilityEnabled: true
  observabilityConfigurationArn:
    value: "arn:aws:apprunner:us-east-1:123456789012:observabilityconfiguration/xray-config/1/abc123"

  ipAddressType: "DUAL_STACK"
  autoDeploymentsEnabled: false
```

## 4. GitHub code source with API configuration (Node.js)

Deploys directly from a GitHub repository. App Runner clones the repo, runs the build command, and starts the application. The build/runtime configuration is provided inline via `configuration_source: API`, so no `apprunner.yaml` file is needed in the repository.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-frontend-app
spec:
  codeSource:
    repositoryUrl: "https://github.com/my-org/frontend-app"
    branch: "main"
    connectionArn:
      value: "arn:aws:apprunner:us-east-1:123456789012:connection/github-connection/abc123"
    configurationSource: "API"
    runtime: "NODEJS_18"
    buildCommand: "npm ci && npm run build"
  port: "3000"
  startCommand: "npm start"
  cpu: "1024"
  memory: "2048"
  environmentVariables:
    NODE_ENV: "production"
    API_BASE_URL: "https://api.example.com"
  autoDeploymentsEnabled: true
```

## 5. GitHub code source with REPOSITORY configuration

When `configuration_source` is `REPOSITORY`, App Runner reads build and runtime settings from an `apprunner.yaml` file at the root of the repository (or in `source_directory`). This keeps your infrastructure configuration co-located with the application code.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-python-api
spec:
  codeSource:
    repositoryUrl: "https://github.com/my-org/python-api"
    branch: "production"
    sourceDirectory: "services/api"
    connectionArn:
      value: "arn:aws:apprunner:us-east-1:123456789012:connection/github-connection/abc123"
    configurationSource: "REPOSITORY"
  # port, startCommand, and build settings come from apprunner.yaml in the repo
  instanceRoleArn:
    value: "arn:aws:iam::123456789012:role/python-api-instance-role"
  autoDeploymentsEnabled: true
  autoScaling:
    minSize: 1
    maxSize: 5
    maxConcurrency: 80
```

The corresponding `apprunner.yaml` in the repo (`services/api/apprunner.yaml`):

```yaml
version: 1.0
runtime: python3
build:
  commands:
    build:
      - pip install -r requirements.txt
run:
  command: gunicorn -w 4 -b 0.0.0.0:8080 app:app
  network:
    port: 8080
  env:
    - name: FLASK_ENV
      value: production
```

## 6. Infra chart pattern with `valueFrom` references

In an OpenMCF infra chart, you can wire outputs from other components into this service using `valueFrom` references. This avoids hard-coding ARNs and makes your infrastructure composable.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-order-service
spec:
  imageSource:
    imageIdentifier: "123456789012.dkr.ecr.us-east-1.amazonaws.com/order-service:v1.0.0"
    imageRepositoryType: "ECR"
    accessRoleArn:
      # Reference the ECR access role from an AwsIamRole component in the same chart
      valueFrom:
        kind: AwsIamRole
        name: apprunner-ecr-role
        fieldPath: "status.outputs.role_arn"
  port: "8080"
  instanceRoleArn:
    # Reference the instance role from another AwsIamRole component
    valueFrom:
      kind: AwsIamRole
      name: order-service-role
      fieldPath: "status.outputs.role_arn"

  # Reference subnets from an AwsVpc component
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        fieldPath: "status.outputs.private_subnets.0.id"
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        fieldPath: "status.outputs.private_subnets.1.id"

  # Reference security group from an AwsSecurityGroup component
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: apprunner-egress-sg
        fieldPath: "status.outputs.security_group_id"

  # Reference KMS key from an AwsKmsKey component
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: "status.outputs.key_arn"

  autoScaling:
    minSize: 2
    maxSize: 10
    maxConcurrency: 50

  healthCheck:
    protocol: "HTTP"
    path: "/health"
```

## CLI flows

Validate manifest:

```bash
openmcf validate --manifest ./apprunner.yaml
```

Pulumi deploy:

```bash
openmcf pulumi update \
  --manifest ./apprunner.yaml \
  --stack my-org/project/dev \
  --module-dir apis/org/openmcf/provider/aws/awsapprunnerservice/v1/iac/pulumi
```

Note: Provider credentials are supplied via stack input, not in the spec.
