# GitHub CI (Linux)

This preset creates a CI-only CodeBuild project that triggers on pushes and pull requests to the main branch of a GitHub repository. Build status is reported back to GitHub as commit status checks. No artifacts are produced — the project exists for linting, testing, and validation.

## When to Use

- Standard CI pipeline for GitHub-hosted repositories
- Teams that need commit status checks (green/red builds on PRs)
- Projects where the build output is pass/fail (not a deployable artifact)
- Quick setup for new repositories that need automated testing

## Key Configuration Choices

- **GITHUB** (`source.type`) — GitHub.com repository via CodeStar Connections
- **BUILD_GENERAL1_SMALL** (`computeType`) — 3 GB memory, 2 vCPUs; sufficient for most test suites
- **amazonlinux2-x86_64-standard:5.0** (`image`) — AWS managed image with common runtimes pre-installed
- **NO_ARTIFACTS** (`artifacts.type`) — CI-only, no build output stored
- **Webhook** — Triggers on PUSH, PULL_REQUEST_CREATED, and PULL_REQUEST_UPDATED to main

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<github-repo-https-url>` | GitHub repository HTTPS URL (e.g., `https://github.com/org/repo.git`) | GitHub repository settings |
| `<codebuild-service-role-arn>` | IAM role ARN granting CodeBuild access to source and logs | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **02-docker-build-ecr** — Use instead when building Docker images and pushing to ECR
- **03-codepipeline-stage** — Use instead when CodeBuild is a stage in CodePipeline
