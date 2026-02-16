# GitHub Code Source (Node.js)

This preset creates an App Runner service that deploys directly from a GitHub repository using the Node.js 18 managed runtime. App Runner clones the repository, runs the build command, and starts the application -- no container image or CI/CD pipeline required. Build and runtime configuration is provided inline via `configurationSource: API`.

## When to Use

- Deploying Node.js web applications or APIs directly from GitHub
- Teams that want a fully managed build + deploy pipeline without maintaining CI/CD infrastructure
- Rapid prototyping where you want code-to-URL in minutes
- Small teams or solo developers who don't want to manage Docker builds

## Key Configuration Choices

- **GitHub code source** (`codeSource`) -- App Runner clones the repo and manages the entire build-to-deploy lifecycle.
- **API configuration** (`configurationSource: API`) -- Build and runtime settings are defined in this manifest. Use `REPOSITORY` instead if you prefer an `apprunner.yaml` file in your repo.
- **Node.js 18 runtime** (`runtime: NODEJS_18`) -- Change to `PYTHON_3`, `CORRETTO_11`, `GO_1`, `DOTNET_6`, `PHP_81`, or `RUBY_31` for other languages.
- **Build command** (`buildCommand: npm ci && npm run build`) -- Installs dependencies and compiles the application. Adjust for your build tool (e.g., `yarn install && yarn build`).
- **Auto-deploy enabled** (`autoDeploymentsEnabled: true`) -- Every push to the configured branch triggers an automatic build and deployment. Disable for production environments where deployments should be deliberate.
- **Default auto scaling and health check** -- Uses App Runner defaults (1--25 instances, TCP health check). Customize for production.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<github-repository-url>` | GitHub repo URL (e.g., `https://github.com/my-org/my-app`) | Your GitHub repository page |
| `<branch-name>` | Branch to deploy from (e.g., `main`, `production`) | Your Git branching strategy |
| `<apprunner-connection-arn>` | ARN of the App Runner Connection for GitHub access | AWS Console → App Runner → GitHub connections |
| `<application-port>` | Port your app listens on (e.g., `3000`, `8080`) | Your `package.json` start script or app config |

## Prerequisites

An **App Runner Connection** must be created before using this preset:

1. Go to **AWS Console → App Runner → GitHub connections**
2. Click **Add new** and complete the GitHub OAuth handshake
3. Wait for the connection status to become `AVAILABLE`
4. Copy the connection ARN and use it as `<apprunner-connection-arn>`

A single connection can be shared across multiple App Runner services.

## Related Presets

- **01-basic-public-image** -- Use when deploying from a pre-built container image instead of source code.
- **02-production-vpc-encrypted** -- Use when you need VPC egress, encryption, and production-grade scaling.
