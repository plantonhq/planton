# Examples for AwsAppRunnerService Terraform Module

## Minimal manifest (YAML) - Image source

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: my-web-app
  org: my-org
spec:
  image_source:
    image_identifier: public.ecr.aws/nginx/nginx:latest
    image_repository_type: ECR_PUBLIC
  port: "80"
  cpu: "1024"
  memory: "2048"
```

## Private ECR image with VPC and auto scaling

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: api-backend
  org: my-org
  tags:
    app: api
    env: prod
spec:
  image_source:
    image_identifier: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.2.3
    image_repository_type: ECR
    access_role_arn:
      value: arn:aws:iam::123456789012:role/apprunner-ecr-access-role
  port: "8080"
  start_command: "java -jar app.jar"
  cpu: "2048"
  memory: "4096"
  instance_role_arn:
    value: arn:aws:iam::123456789012:role/apprunner-instance-role
  environment_variables:
    LOG_LEVEL: info
  environment_secrets:
    DB_PASSWORD: arn:aws:secretsmanager:us-east-1:123456789012:secret:db-password
  health_check:
    protocol: HTTP
    path: /health
    interval_seconds: 10
    timeout_seconds: 5
    healthy_threshold: 2
    unhealthy_threshold: 3
  auto_scaling:
    min_size: 2
    max_size: 10
    max_concurrency: 50
  subnet_ids:
    - value: subnet-aaaa1111
    - value: subnet-bbbb2222
  security_group_ids:
    - value: sg-0123456789abcdef0
  kms_key_arn:
    value: arn:aws:kms:us-east-1:123456789012:key/abcde-12345
  is_publicly_accessible: true
  ip_address_type: IPV4
  auto_deployments_enabled: true
```

## GitHub code source with API configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAppRunnerService
metadata:
  name: nodejs-api
  org: my-org
spec:
  code_source:
    repository_url: https://github.com/my-org/my-api
    branch: main
    source_directory: api
    connection_arn:
      value: arn:aws:apprunner:us-east-1:123456789012:connection/github/my-org/abc123
    configuration_source: API
    runtime: NODEJS_18
    build_command: npm ci && npm run build
  port: "3000"
  start_command: "node server.js"
  cpu: "1024"
  memory: "2048"
  environment_variables:
    NODE_ENV: production
  auto_scaling:
    min_size: 1
    max_size: 5
    max_concurrency: 100
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./apprunner.yaml
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./apprunner.yaml --auto-approve
```
