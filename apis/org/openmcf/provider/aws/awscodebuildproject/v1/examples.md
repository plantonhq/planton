# AwsCodeBuildProject Examples

## 1. Minimal GitHub CI

The simplest setup: a GitHub-triggered CI project that runs tests on every push and pull request to main. No artifacts are produced — the project exists purely for validation (lint, test, type-check).

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodeBuildProject
metadata:
  name: my-app-ci
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
    reportBuildStatus: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-service-role
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED
          - type: HEAD_REF
            pattern: ^refs/heads/main$
```

## 2. Docker Build with ECR Push

A production Docker build project using privileged mode for Docker daemon access. Uses local Docker layer caching to speed up subsequent builds. The buildspec pushes the image to ECR — no S3 artifacts needed.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodeBuildProject
metadata:
  name: docker-builder
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-service.git
    gitCloneDepth: 1
    reportBuildStatus: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_LARGE
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    privilegedMode: true
    environmentVariables:
      - name: ECR_REPO
        value: 123456789012.dkr.ecr.us-east-1.amazonaws.com/my-service
      - name: DOCKER_BUILDKIT
        value: "1"
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-docker-role
  buildTimeout: 30
  cache:
    type: LOCAL
    modes:
      - LOCAL_DOCKER_LAYER_CACHE
      - LOCAL_SOURCE_CACHE
  logsConfig:
    cloudwatchLogs:
      status: ENABLED
      groupName:
        value: /aws/codebuild/docker-builder
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH
          - type: HEAD_REF
            pattern: ^refs/heads/(main|release/.*)$
```

## 3. CodePipeline Build Stage

A build project designed as a stage in AWS CodePipeline. Source and artifacts are both managed by the pipeline. Includes environment variables for deployment configuration stored in Secrets Manager.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodeBuildProject
metadata:
  name: pipeline-build
spec:
  region: us-west-2
  source:
    type: CODEPIPELINE
    buildspec: buildspec.yml
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_MEDIUM
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    environmentVariables:
      - name: STAGE
        value: production
      - name: DB_CONNECTION_STRING
        value: prod/db-connection-string
        type: SECRETS_MANAGER
  artifacts:
    type: CODEPIPELINE
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-pipeline-role
  buildTimeout: 20
  concurrentBuildLimit: 3
```

## 4. VPC Build (Private Resource Access)

A build project placed inside a VPC to access private resources like RDS databases or ElastiCache clusters during integration tests. Uses S3 cache for build dependencies.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodeBuildProject
metadata:
  name: integration-tests
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/backend-api.git
    reportBuildStatus: true
    fetchSubmodules: true
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_MEDIUM
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
    environmentVariables:
      - name: DB_HOST
        value: prod/rds-endpoint
        type: PARAMETER_STORE
      - name: REDIS_HOST
        value: prod/redis-endpoint
        type: PARAMETER_STORE
  artifacts:
    type: S3
    location:
      value: my-test-reports-bucket
    name: test-reports
    path: integration
    packaging: ZIP
    namespaceType: BUILD_ID
  serviceRole:
    value: arn:aws:iam::123456789012:role/codebuild-vpc-role
  buildTimeout: 45
  cache:
    type: S3
    location:
      value: my-cache-bucket/codebuild/integration-tests
  vpcConfig:
    vpcId:
      value: vpc-0a1b2c3d4e5f00001
    subnetIds:
      - value: subnet-0a1b2c3d4e5f00001
      - value: subnet-0a1b2c3d4e5f00002
    securityGroupIds:
      - value: sg-0a1b2c3d4e5f00001
  webhook:
    filterGroups:
      - filters:
          - type: EVENT
            pattern: PUSH
          - type: HEAD_REF
            pattern: ^refs/heads/main$

```

## 5. Cross-Resource Reference (valueFrom)

Using `valueFrom` to reference outputs from other OpenMCF resources instead of hardcoding values.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodeBuildProject
metadata:
  name: connected-build
spec:
  region: us-west-2
  source:
    type: GITHUB
    location: https://github.com/my-org/my-app.git
  environment:
    type: LINUX_CONTAINER
    computeType: BUILD_GENERAL1_SMALL
    image: aws/codebuild/amazonlinux2-x86_64-standard:5.0
  artifacts:
    type: NO_ARTIFACTS
  serviceRole:
    valueFrom:
      kind: AwsIamRole
      name: codebuild-role
      fieldPath: status.outputs.role_arn
  encryptionKey:
    valueFrom:
      kind: AwsKmsKey
      name: build-key
      fieldPath: status.outputs.key_arn
  vpcConfig:
    vpcId:
      valueFrom:
        kind: AwsVpc
        name: main-vpc
        fieldPath: status.outputs.vpc_id
    subnetIds:
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          fieldPath: status.outputs.private_subnets.0.id
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          fieldPath: status.outputs.private_subnets.1.id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: codebuild-sg
          fieldPath: status.outputs.security_group_id
  logsConfig:
    cloudwatchLogs:
      groupName:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: codebuild-logs
          fieldPath: status.outputs.log_group_name
```
