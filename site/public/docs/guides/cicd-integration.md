---
title: "CI/CD Integration"
description: "Patterns for running OpenMCF in CI/CD pipelines — GitHub Actions, GitLab CI, non-interactive flags, and credential injection"
icon: "integration"
order: 100
---

# CI/CD Integration

OpenMCF is designed to run unattended in CI/CD pipelines. This guide covers patterns for GitHub Actions, GitLab CI, and general automation, including credential injection, non-interactive execution, manifest validation, and Kustomize overlay selection.

## Non-Interactive Flags

By default, OpenMCF prompts for confirmation before making changes. In CI/CD, use `--yes` or `--auto-approve` to skip the prompt:

```bash
# Either flag works — they are equivalent
openmcf pulumi up -f manifest.yaml --yes
openmcf apply -f manifest.yaml --auto-approve
```

Both flags are available on deployment commands (`apply`, `pulumi up`, `tofu apply`, `terraform apply`, `destroy`).

## Validation in Pipelines

Run manifest validation as an early step in your pipeline. Validation catches schema errors, type mismatches, and constraint violations in seconds — before any cloud API call:

```bash
openmcf validate -f manifest.yaml
```

Exit code 0 means valid. Any non-zero exit code indicates validation failure, which should stop the pipeline.

## GitHub Actions

### Basic Deployment (AWS)

```yaml
name: Deploy Infrastructure

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install OpenMCF
        run: |
          curl -sSL https://get.openmcf.org | bash

      - name: Validate manifest
        run: openmcf validate -f ops/aws/database.yaml

      - name: Deploy
        run: openmcf pulumi up -f ops/aws/database.yaml --yes
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: us-west-2
```

### Kustomize with Branch-Based Overlays

```yaml
name: Deploy API

on:
  push:
    branches: [main, develop]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Determine overlay
        id: env
        run: |
          if [ "${{ github.ref }}" == "refs/heads/main" ]; then
            echo "overlay=prod" >> $GITHUB_OUTPUT
          else
            echo "overlay=dev" >> $GITHUB_OUTPUT
          fi

      - name: Deploy
        run: |
          openmcf pulumi up \
            --kustomize-dir services/api/kustomize \
            --overlay ${{ steps.env.outputs.overlay }} \
            --yes
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

### GCP with Service Account

```yaml
name: Deploy to GCP

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy
        run: openmcf pulumi up -f ops/gcp/database.yaml --yes
        env:
          GOOGLE_APPLICATION_CREDENTIALS: ${{ runner.temp }}/gcp-key.json

      # Write the service account key to a temp file
      - name: Setup GCP credentials
        run: echo '${{ secrets.GCP_SERVICE_ACCOUNT_KEY }}' > ${{ runner.temp }}/gcp-key.json
```

Alternatively, use the `-p` flag with a base64-encoded key stored as a secret:

```yaml
      - name: Create provider config
        run: |
          echo "service_account_key_base64: ${{ secrets.GCP_SA_KEY_BASE64 }}" > /tmp/gcp-cred.yaml

      - name: Deploy
        run: openmcf pulumi up -f ops/gcp/database.yaml -p /tmp/gcp-cred.yaml --yes
```

### Azure with Service Principal

```yaml
name: Deploy to Azure

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy
        run: openmcf pulumi up -f ops/azure/cluster.yaml --yes
        env:
          ARM_CLIENT_ID: ${{ secrets.ARM_CLIENT_ID }}
          ARM_CLIENT_SECRET: ${{ secrets.ARM_CLIENT_SECRET }}
          ARM_TENANT_ID: ${{ secrets.ARM_TENANT_ID }}
          ARM_SUBSCRIPTION_ID: ${{ secrets.ARM_SUBSCRIPTION_ID }}
```

### Dynamic Image Tags

Use `--set` to inject build-time values like image tags:

```yaml
      - name: Deploy with commit SHA
        run: |
          openmcf pulumi up \
            -f ops/k8s/api.yaml \
            --set spec.container.image.tag=${{ github.sha }} \
            --yes
```

## GitLab CI

### Basic Deployment

```yaml
stages:
  - validate
  - deploy

validate:
  stage: validate
  script:
    - openmcf validate -f ops/aws/database.yaml

deploy:
  stage: deploy
  script:
    - openmcf pulumi up -f ops/aws/database.yaml --yes
  variables:
    AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY
    AWS_REGION: us-west-2
  only:
    - main
```

### Kustomize with Branch-Based Overlays

```yaml
deploy:
  stage: deploy
  script:
    - |
      if [ "$CI_COMMIT_BRANCH" == "main" ]; then
        OVERLAY="prod"
      elif [ "$CI_COMMIT_BRANCH" == "staging" ]; then
        OVERLAY="staging"
      else
        OVERLAY="dev"
      fi
    - openmcf pulumi up --kustomize-dir services/api/kustomize --overlay $OVERLAY --yes
  only:
    - main
    - staging
    - develop
```

### Protected Environments

Use GitLab's environment feature for deployment approvals:

```yaml
deploy-prod:
  stage: deploy
  script:
    - openmcf pulumi up -f ops/database.yaml --yes
  environment:
    name: production
  when: manual
  only:
    - main
```

## Credential Injection Patterns

### Store Credentials as CI/CD Secrets

| Platform | Where to Store | Notes |
|----------|---------------|-------|
| GitHub Actions | Settings > Secrets and variables > Actions | Use repository or organization secrets |
| GitLab CI | Settings > CI/CD > Variables | Mark as "Protected" and "Masked" |
| Jenkins | Credentials Manager | Use "Secret text" or "Secret file" types |

### Multi-Environment Credentials

Use environment-specific secrets for different deployment targets:

```yaml
# GitHub Actions with environments
jobs:
  deploy-prod:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy
        run: openmcf pulumi up -f ops/database.yaml --yes
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

GitHub Actions environments can require reviewers for production deployments, providing an approval gate before infrastructure changes are applied.

## Pipeline Best Practices

**Validate before deploying.** Run `openmcf validate` as an early pipeline step. Failed validation should block deployment.

**Use preview/plan before apply.** In production pipelines, run `openmcf pulumi preview` or `openmcf plan` first, then apply in a separate step or with manual approval:

```bash
openmcf pulumi preview -f manifest.yaml
# Review output
openmcf pulumi up -f manifest.yaml --yes
```

**Pin manifest versions.** If manifests are in a separate repository, reference a specific commit or tag rather than a branch to ensure reproducible deployments.

**Use `--set` for build-time values only.** Inject dynamic values like image tags and commit SHAs via `--set`. Keep permanent configuration in the manifest file and commit it to version control.

**Separate validate and deploy stages.** Validation runs in seconds and can run on every pull request. Deployment should only run on merge to the target branch.

## What's Next

- [Credentials](./credentials) — Quick reference for all provider credentials
- [AWS Provider Setup](./aws-provider-setup) — AWS IAM for CI/CD
- [GCP Provider Setup](./gcp-provider-setup) — GCP service accounts for CI/CD
- [Azure Provider Setup](./azure-provider-setup) — Azure service principals for CI/CD
- [Kustomize Integration](./kustomize) — Multi-environment overlay workflows
- [Advanced Usage](./advanced-usage) — Runtime overrides and combining techniques
