# CodePipeline Build Stage

This preset creates a CodeBuild project designed to run as a build or test stage within an AWS CodePipeline. Source code and build artifacts are managed by the pipeline — CodeBuild receives input from the previous stage and passes output to the next. No webhook is configured because CodePipeline handles triggering.

## When to Use

- CodeBuild is a step in a multi-stage CodePipeline
- The pipeline manages source, build, and deploy stages
- Teams using the full AWS-native CI/CD stack (CodeCommit/GitHub → CodeBuild → CodeDeploy)
- Projects where build triggers are controlled by the pipeline, not by repository events

## Key Configuration Choices

- **CODEPIPELINE** (`source.type` and `artifacts.type`) — Both must be CODEPIPELINE when used as a pipeline stage
- **BUILD_GENERAL1_MEDIUM** (`computeType`) — 7 GB memory, 4 vCPUs; balanced for typical build workloads
- **buildspec.yml** (`buildspec`) — Explicit buildspec path (CodePipeline can also override this)
- **buildTimeout: 20** — Pipeline stages should be fast; 20 minutes is generous for most builds
- **concurrentBuildLimit: 3** — Prevents runaway costs from parallel pipeline executions
- **No webhook** — CodePipeline handles triggering; webhooks would conflict

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<deployment-stage>` | Stage name (e.g., `staging`, `production`) | Your pipeline configuration |
| `<codebuild-service-role-arn>` | IAM role ARN with appropriate build permissions | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-github-ci-linux** — Use instead for standalone CI without CodePipeline
- **02-docker-build-ecr** — Use instead for standalone Docker builds triggered by webhooks
