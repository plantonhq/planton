# AwsCodePipeline Examples

## 1. Minimal CI Pipeline (GitHub Source + CodeBuild)

The simplest pipeline: fetch source from GitHub via CodeStar Connection and run a CodeBuild build project. This is the standard CI pipeline for running tests and producing build artifacts on every push to main.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: my-app-ci
spec:
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location:
        value: my-app-pipeline-artifacts
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-app
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: BuildAndTest
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-app-build
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput
```

## 2. CI/CD with Manual Approval (Source → Build → Approve → ECS Deploy)

A four-stage pipeline for deploying containerized applications to ECS with a manual approval gate before production. The approval stage sends a notification and waits for human confirmation before proceeding with the deployment.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: my-service-cicd
spec:
  pipelineType: V2
  executionMode: QUEUED
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-cicd-role
  artifactStores:
    - location:
        value: my-service-pipeline-artifacts
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-service
            BranchName: main
            OutputArtifactFormat: CODE_ZIP
          outputArtifacts:
            - SourceOutput
          namespace: SourceVariables

    - name: Build
      actions:
        - name: DockerBuild
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-service-docker-build
            EnvironmentVariables: '[{"name":"IMAGE_TAG","value":"#{SourceVariables.CommitId}","type":"PLAINTEXT"}]'
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput

    - name: Approval
      actions:
        - name: ManualApproval
          category: Approval
          owner: AWS
          provider: Manual
          version: "1"
          configuration:
            CustomData: "Review build #{SourceVariables.CommitId} before deploying to production"
            ExternalEntityLink: "https://github.com/my-org/my-service/commit/#{SourceVariables.CommitId}"
            NotificationArn: arn:aws:sns:us-east-1:123456789012:pipeline-approvals

    - name: Deploy
      actions:
        - name: ECSDeployProd
          category: Deploy
          owner: AWS
          provider: ECS
          version: "1"
          configuration:
            ClusterName: production-cluster
            ServiceName: my-service
            FileName: imagedefinitions.json
          inputArtifacts:
            - BuildOutput

  triggers:
    - providerType: CodeStarSourceConnection
      gitConfiguration:
        sourceActionName: GitHubSource
        push:
          - branches:
              includes:
                - main
```

## 3. V2 Pipeline with Git Triggers (Push + File Path Filtering)

A V2 pipeline with sophisticated trigger filtering. The pipeline only runs when code changes are pushed to main or release branches, and only when application source files change — ignoring documentation-only changes.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: my-api-ci
spec:
  pipelineType: V2
  executionMode: QUEUED
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-service-role
  artifactStores:
    - location:
        value: my-api-pipeline-artifacts
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-api
            BranchName: main
            DetectChanges: "false"
          outputArtifacts:
            - SourceOutput
          namespace: SourceVariables

    - name: Build
      actions:
        - name: UnitTests
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          runOrder: 1
          configuration:
            ProjectName: my-api-unit-tests
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - TestResults

        - name: LintAndFormat
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          runOrder: 1
          configuration:
            ProjectName: my-api-lint
          inputArtifacts:
            - SourceOutput

    - name: IntegrationTest
      actions:
        - name: IntegrationTests
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-api-integration-tests
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - IntegrationResults

  variables:
    - name: Environment
      defaultValue: staging
      description: Target deployment environment
    - name: SkipIntegrationTests
      defaultValue: "false"
      description: Set to true to skip integration tests

  triggers:
    - providerType: CodeStarSourceConnection
      gitConfiguration:
        sourceActionName: GitHubSource
        push:
          - branches:
              includes:
                - main
                - "release/*"
            filePaths:
              includes:
                - "src/**"
                - "tests/**"
                - "buildspec*.yml"
                - "Dockerfile"
              excludes:
                - "docs/**"
                - "*.md"
                - ".github/**"
                - "LICENSE"
        pullRequest:
          - branches:
              includes:
                - main
            events:
              - OPEN
              - UPDATE
```

## 4. S3 Source to Lambda Deploy (Serverless Deployment)

A pipeline that triggers when a deployment package is uploaded to S3, then deploys it to a Lambda function. Ideal for serverless applications where the build artifact is a zip file uploaded to S3 by an external CI system.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: my-lambda-deploy
spec:
  pipelineType: V1
  roleArn:
    value: arn:aws:iam::123456789012:role/codepipeline-lambda-role
  artifactStores:
    - location:
        value: my-lambda-pipeline-artifacts
  stages:
    - name: Source
      actions:
        - name: S3Source
          category: Source
          owner: AWS
          provider: S3
          version: "1"
          configuration:
            S3Bucket: my-lambda-packages
            S3ObjectKey: deployments/my-function/latest.zip
            PollForSourceChanges: "true"
          outputArtifacts:
            - LambdaPackage

    - name: Deploy
      actions:
        - name: DeployFunction
          category: Invoke
          owner: AWS
          provider: Lambda
          version: "1"
          configuration:
            FunctionName: my-deployment-function
            UserParameters: '{"targetFunction":"my-api-handler","alias":"live"}'
          inputArtifacts:
            - LambdaPackage
```

## 5. Cross-Account Deployment Pipeline

A pipeline that builds in the CI account and deploys to a production account using cross-account IAM role assumption. The deploy action assumes a role in the target account to perform CloudFormation stack operations.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: cross-account-deploy
spec:
  pipelineType: V2
  executionMode: QUEUED
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: codepipeline-orchestrator-role
      fieldPath: status.outputs.role_arn
  artifactStores:
    - location:
        valueFrom:
          kind: AwsS3Bucket
          name: pipeline-artifacts
          fieldPath: status.outputs.bucket_name
      encryptionKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: pipeline-key
          fieldPath: status.outputs.key_arn
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:111111111111:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/infra-templates
            BranchName: main
          outputArtifacts:
            - SourceOutput
          namespace: SourceVariables

    - name: Build
      actions:
        - name: PackageTemplates
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: infra-template-packager
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - PackagedTemplates

    - name: DeployStaging
      actions:
        - name: StagingCloudFormation
          category: Deploy
          owner: AWS
          provider: CloudFormation
          version: "1"
          configuration:
            ActionMode: CREATE_UPDATE
            StackName: my-app-staging
            TemplatePath: PackagedTemplates::template-output.yaml
            RoleArn: arn:aws:iam::222222222222:role/cloudformation-deploy-role
            Capabilities: CAPABILITY_IAM,CAPABILITY_NAMED_IAM
          inputArtifacts:
            - PackagedTemplates
          roleArn:
            value: arn:aws:iam::222222222222:role/codepipeline-cross-account-role

    - name: ApproveProduction
      actions:
        - name: ManualApproval
          category: Approval
          owner: AWS
          provider: Manual
          version: "1"
          configuration:
            CustomData: "Staging deployment complete. Approve production deployment?"
            NotificationArn: arn:aws:sns:us-east-1:111111111111:prod-approvals

    - name: DeployProduction
      actions:
        - name: ProductionCloudFormation
          category: Deploy
          owner: AWS
          provider: CloudFormation
          version: "1"
          configuration:
            ActionMode: CREATE_UPDATE
            StackName: my-app-production
            TemplatePath: PackagedTemplates::template-output.yaml
            RoleArn: arn:aws:iam::333333333333:role/cloudformation-deploy-role
            Capabilities: CAPABILITY_IAM,CAPABILITY_NAMED_IAM
          inputArtifacts:
            - PackagedTemplates
          roleArn:
            value: arn:aws:iam::333333333333:role/codepipeline-cross-account-role

  triggers:
    - providerType: CodeStarSourceConnection
      gitConfiguration:
        sourceActionName: GitHubSource
        push:
          - branches:
              includes:
                - main
            tags:
              includes:
                - "v*"
```

## 6. Cross-Resource References (valueFrom)

Using `valueFrom` to reference outputs from other OpenMCF resources instead of hardcoding ARNs, bucket names, and project names.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCodePipeline
metadata:
  name: connected-pipeline
spec:
  pipelineType: V2
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: pipeline-role
      fieldPath: status.outputs.role_arn
  artifactStores:
    - location:
        valueFrom:
          kind: AwsS3Bucket
          name: pipeline-artifacts
          fieldPath: status.outputs.bucket_name
      encryptionKeyId:
        valueFrom:
          kind: AwsKmsKey
          name: pipeline-encryption-key
          fieldPath: status.outputs.key_arn
  stages:
    - name: Source
      actions:
        - name: GitHubSource
          category: Source
          owner: AWS
          provider: CodeStarSourceConnection
          version: "1"
          configuration:
            ConnectionArn: arn:aws:codeconnections:us-east-1:123456789012:connection/abc12345-def6-7890-ghij-klmnopqrstuv
            FullRepositoryId: my-org/my-app
            BranchName: main
          outputArtifacts:
            - SourceOutput
    - name: Build
      actions:
        - name: BuildAndTest
          category: Build
          owner: AWS
          provider: CodeBuild
          version: "1"
          configuration:
            ProjectName: my-app-build
          inputArtifacts:
            - SourceOutput
          outputArtifacts:
            - BuildOutput
    - name: Deploy
      actions:
        - name: DeployToECS
          category: Deploy
          owner: AWS
          provider: ECS
          version: "1"
          configuration:
            ClusterName: production-cluster
            ServiceName: my-app-service
            FileName: imagedefinitions.json
          inputArtifacts:
            - BuildOutput
          roleArn:
            valueFrom:
              kind: AwsIamRole
              name: ecs-deploy-role
              fieldPath: status.outputs.role_arn
```
