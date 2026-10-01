---
title: "How to Deploy Your First Service with Zero-Config CI/CD"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "service-hub"
  - "ci-cd"
  - "buildpacks"
  - "getting-started"
  - "pipeline"
  - "kustomize"
category: "service-hub"
excerpt: "Push code to GitHub and get a running deployment on Kubernetes -- no Dockerfile, no pipeline YAML, no CI/CD configuration."
---

# How to Deploy Your First Service with Zero-Config CI/CD

This tutorial takes you from a GitHub repository containing application code to a running deployment on Kubernetes. You will not write a Dockerfile. You will not author pipeline YAML. You will not configure a CI/CD system. Planton handles all of that.

You will do two things: tell Planton where your code lives and where you want it deployed. The platform auto-detects your application's language using Cloud Native Buildpacks, builds a container image, and deploys it to Kubernetes through a fully managed pipeline. When you push code to your default branch, the entire cycle repeats automatically.

> **Note**: The Planton web console provides a guided creation wizard for Services.
> This tutorial uses the CLI/YAML approach for stability and reproducibility.
> The console UI evolves frequently -- always check it for the latest experience.

> **A note on field casing**: Planton manifests accept both camelCase and snake_case
> field names (standard protobuf JSON mapping). This tutorial uses camelCase, matching
> the convention used by Planton presets and ops manifests. Either style works.

## What You Will Learn

- Create a Service resource that links your GitHub repository to Planton's CI/CD system
- Structure kustomize overlays that define where and how your application deploys
- Monitor a platform-managed pipeline from image build through deployment
- Verify your deployed service is running on Kubernetes
- Trigger a new deployment by pushing a code change

## Prerequisites

- [ ] A Planton organization with at least one environment configured
- [ ] A GitHub connection (see [How to Connect Your GitHub Account to Planton](/tutorials/how-to-connect-your-github-account-to-planton))
- [ ] A container registry connection (see [How to Connect a Container Registry to Planton](/tutorials/how-to-connect-a-container-registry-to-planton))
- [ ] A Kubernetes cluster accessible from your Planton organization
- [ ] A Git repository with a [Cloud Native Buildpacks](https://paketo.io/)-compatible application (Node.js, Go, Java, Python, Ruby, .NET, PHP, or other Paketo-supported language)
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)

## How It Works

A Planton Service connects your Git repository to a CI/CD pipeline that builds container images and deploys them to Kubernetes using kustomize overlays. You define the Service in a YAML manifest pointing to your repo, and Planton handles the rest -- building on every push, exporting manifests, and applying them to your target cluster. For more on how Services, pipelines, and kustomize overlays work together, see the [CI/CD documentation](/docs/ci-cd/what-is-a-service).

## Step 1: Prepare Your Repository

Your application repository needs a `_kustomize/` directory that tells Planton how to deploy your service. You can scaffold this automatically with the CLI or create it by hand.

### Option A: Scaffold with the CLI

If you have the `planton` CLI and are inside your Git repository, run:

```bash
planton service kustomize init --envs dev
```

It writes one empty overlay per environment you name, plus the merge schema:

```text
_kustomize/
├── planton-schema.json
└── overlays/
    └── dev/
        └── kustomization.yaml
```

Add your service's manifest (a KubernetesDeployment) to the overlay and list it under the overlay's kustomization `resources`, as in Option B, before continuing to Step 2.

### Option B: Create the Structure Manually

Create the following directory structure in your repository root:

```text
_kustomize/
├── base/
│   ├── kustomization.yaml
│   └── service.yaml
└── overlays/
    └── dev/
        ├── kustomization.yaml
        └── service.yaml
```

**`_kustomize/base/kustomization.yaml`**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - service.yaml
```

**`_kustomize/base/service.yaml`**

This is a `KubernetesDeployment` manifest -- a Planton Resource that defines how your application runs on Kubernetes. The base contains configuration shared across all environments.

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: my-service
  org: your-org
spec:
  version: main
  container:
    app:
      image:
        repo: ghcr.io/your-github-org/your-repo
      resources:
        requests:
          cpu: 50m
          memory: 100Mi
        limits:
          cpu: 1000m
          memory: 1Gi
      ports:
        - name: rest
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
  availability:
    minReplicas: 1
```

A few things to note about this manifest:

- **`image.repo`** is the container registry path where your built image will be stored. Set this to match the registry and path you configured in your container registry connection. The image **tag** is not specified here -- the pipeline automatically sets it to the Git commit SHA when deploying.
- **`containerPort`** should match the port your application listens on. The example uses 8080; adjust to match your application.
- **`resources`** sets the CPU and memory requests and limits for your container. Start with conservative values and adjust based on actual usage.

**`_kustomize/overlays/dev/kustomization.yaml`**

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: service.yaml
```

**`_kustomize/overlays/dev/service.yaml`**

The overlay patches only what differs from the base. At minimum, it sets `metadata.env` to identify the target Planton environment:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: my-service
  env: dev
```

The `env` field tells Planton which environment this deployment targets. The platform resolves the Kubernetes cluster and namespace from your environment's configuration -- you do not need to specify a target cluster in the manifest.

### Verify Locally

Before proceeding, confirm that your kustomize structure is valid:

```bash
kubectl kustomize _kustomize/overlays/dev/
```

This command merges the overlay with the base and prints the resulting manifest. Verify that the output contains your base configuration with `metadata.env: dev` applied.

Commit and push the `_kustomize/` directory to your repository.

## Step 2: Create the Service Manifest

Create a file named `service.yaml` (this file lives outside your application repository -- it is a Planton API resource, not part of your application code):

```yaml
apiVersion: service-hub.planton.ai/v1alpha1
kind: Service
metadata:
  name: my-service
  org: your-org
spec:
  description: My first backend service on Planton
  gitRepo:
    gitConnection: your-github-connection
    ownerName: your-github-org
    name: your-repo
    defaultBranch: main
  build:
    buildpacks: {}
    registry: your-registry-connection
    imageRepositoryPath: your-github-org/your-repo
    triggers:
      branches:
        - main
  deploy:
    kustomize: {}
```

Replace the placeholder values with your actual configuration:

| Field | What to set | Where to find it |
|-------|-------------|------------------|
| `metadata.name` | A name for your service (lowercase, hyphens allowed) | You choose this |
| `metadata.org` | Your Planton organization slug | `planton get organization` |
| `gitRepo.gitConnection` | The slug of your GitHub connection | `planton get github-connection` |
| `gitRepo.ownerName` | Your GitHub organization or username | Your GitHub account |
| `gitRepo.name` | Your repository name | Your GitHub repository |
| `build.registry` | The slug of your container registry connection | `planton get container-registry-connection` |
| `build.imageRepositoryPath` | The path within your registry where images are pushed | Must match `image.repo` in your kustomize base |

Here is what each section does:

**`gitRepo`**: Links this Service to your GitHub repository. `gitConnection` is the slug of the GitHub connection you created in the [GitHub connection tutorial](/tutorials/how-to-connect-your-github-account-to-planton). Planton uses it to read your code at each built commit and to report every run's verdict back to GitHub as a check on the commit.

**`build`**: How the service is built. Exactly one builder is set:
- `buildpacks: {}` means Planton's release-pinned Buildpacks track detects your language and builds a container image with Cloud Native Buildpacks. No Dockerfile, no pipeline YAML.
- `dockerfile: {}` is the alternative when you have a Dockerfile (see below); `tektonPipeline: {}` is for a pipeline you write yourself.
- `registry` and `imageRepositoryPath` say where the image is pushed: `<registry host>/<imageRepositoryPath>:<commit sha>`.
- `triggers.branches` lists the branches whose pushes build. Leaving it empty means the repository's default branch. Every trigger rule — branches, paths, pull requests, tags — lives here and nowhere else.

**`deploy.kustomize: {}`**: Declares that your repository's `_kustomize/` tree is what the service deploys. The overlays under `_kustomize/overlays/` ARE the environment set — a `dev` overlay means the service deploys to `dev`; add a `production` overlay later and it deploys there too, in your organization's promotion order. There is no separate list of environments to keep in sync.

## Step 3: Apply the Service and Watch the Initial Pipeline

Deploy the Service:

```bash
planton apply -f service.yaml
```

Your GitHub connection's App installation already delivers push events for every repository it covers, so nothing needs to be registered on the repository. Start the first run yourself, on the head of your default branch:

```bash
planton service run my-service --branch main --follow
```

`--follow` streams the run's progress to your terminal. You will see the build stage's tasks execute:

1. **git-checkout**: Clones your repository at the built commit
2. **build-image**: Buildpacks detects your language, compiles your application, and pushes a container image tagged with the commit SHA
3. **kustomize-build**: Renders each overlay against the base and stores the result for the deploy stage

Then the deploy stage walks your environments in promotion order — here, `dev` — applying the rendered manifests with the built image reference injected.

To find a run later, or to read a build's own output (the place a build failure explains itself):

```bash
planton service last-pipeline my-service        # the most recent run
planton service runs my-service                 # everything that ran, newest first
planton service logs <run-id>                   # the build's logs, every task's lines
```

## Step 4: Verify the Deployment

Once the run completes, every environment it deployed has a deployment record — an immutable receipt with the exact artifact, the applied manifests, the URLs, and the rollout verdict:

```bash
planton service deployments my-service      # every deployment, newest first, with commit and artifact
planton service urls my-service             # where the service answers, per environment
```

## Step 5: Push a Code Change and Watch the Run

Make a change to your application code, commit it, and push to a branch in `build.triggers.branches`:

```bash
git add .
git commit -m "update: my first change"
git push origin main
```

The push births a run. Find and follow it:

```bash
planton service last-pipeline my-service
planton service follow <run-id>
```

The new run builds a fresh container image tagged with the new commit SHA and deploys it to `dev`. Each push to `main` produces a new deployment with a traceable image tag, and the commit on GitHub carries a check named after the service with the run's verdict.

## Common Patterns and Tips

### Switching from Buildpacks to Dockerfile

If your application needs a custom build process, switch the builder to a Dockerfile:

```yaml
spec:
  build:
    dockerfile:
      dockerfilePath: Dockerfile    # relative to the build context; defaults to "Dockerfile"
    registry: your-registry-connection
    imageRepositoryPath: your-github-org/your-repo
```

Everything else -- the compiled pipeline, the kustomize structure, the deployment flow -- remains the same. If you later need steps neither track offers (tests, scans, a multi-stage build), `tektonPipeline: {}` lets you write the pipeline yourself; see [Self-Managed Pipelines](/docs/ci-cd/self-managed-pipelines).

### Monorepo Support

If your service lives in a subdirectory of a monorepo:

```yaml
spec:
  gitRepo:
    gitConnection: your-github-connection
    ownerName: your-github-org
    name: your-monorepo
    defaultBranch: main
    projectRoot: services/my-service
    sparseCheckoutDirectories:
      - services/my-service
  build:
    buildpacks: {}
    registry: your-registry-connection
    imageRepositoryPath: your-github-org/my-service
    triggers:
      paths:
        - "packages/shared-lib/**"
```

- `gitRepo.projectRoot` is where the build runs and where `_kustomize/` is found; a change under it always triggers
- `build.triggers.paths` adds paths OUTSIDE the project root whose changes should also trigger (a shared library)
- `gitRepo.sparseCheckoutDirectories` speeds up the clone by fetching only the directories the build needs

See [Monorepo Support](/docs/ci-cd/monorepo-support) for the full model, including a Docker build context wider than the project root.

### Enabling Ingress

To expose your service externally, add ingress configuration to your kustomize overlay:

```yaml
# _kustomize/overlays/dev/service.yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesDeployment
metadata:
  name: my-service
  env: dev
spec:
  ingress:
    enabled: true
    hostname: my-service-dev.your-domain.com
  container:
    app:
      ports:
        - name: rest
          containerPort: 8080
          networkProtocol: TCP
          appProtocol: http
          servicePort: 80
          isIngressPort: true
```

Note the `isIngressPort: true` flag on the port -- this tells the ingress controller which port to route traffic to.

### Adding a Production Environment

Create `_kustomize/overlays/production/kustomization.yaml` and `_kustomize/overlays/production/service.yaml` with production-specific settings (higher resource limits, more replicas), commit, and push. That push is a run: the platform reads the new overlay, records `production` as one of the service's environments, and every run from then on walks `dev` and then `production` in your organization's promotion order. There is no list on the Service to update — the overlays are the environment set. To pause a run before `production` for a human decision, mark the environment protected in your organization; see [How to Configure Branch Deployments and Tag Releases](/tutorials/how-to-configure-branch-deployments-and-tag-releases).

### Overriding the Kustomize Directory

By default the tree is `_kustomize/` under `projectRoot`. If it lives elsewhere:

```yaml
spec:
  deploy:
    kustomize:
      baseDirectory: deploy/kustomize    # relative to projectRoot
```

## What to Do Next

- **[How to Deploy Redis on Kubernetes](/tutorials/how-to-deploy-redis-on-kubernetes)** -- provision backing infrastructure for your service
- **[How to Configure Branch Deployments and Tag Releases](/tutorials/how-to-configure-branch-deployments-and-tag-releases)** -- protected environments, pull-request previews, tag releases, branch-to-environment bindings
- **[How to Switch to Self-Managed Tekton Pipelines](/tutorials/how-to-switch-to-self-managed-tekton-pipelines)** -- write the pipeline yourself when the platform tracks are not enough
