---
title: "How to Switch to Self-Managed Tekton Pipelines"
date: "2026-04-02"
author:
  - name: "Planton Team"
    title: "Platform Engineering"
    bio: "Helping teams deploy infrastructure and services without the DevOps bottleneck"
tags:
  - "service-hub"
  - "tekton"
  - "self-managed"
  - "pipeline"
  - "custom-build"
category: "service-hub"
excerpt: "Take control of your CI/CD pipeline by writing Tekton YAML in your repository -- add SonarQube analysis, unit tests, or any custom build step while Planton handles triggering, credentials, and deployment."
---

# How to Switch to Self-Managed Tekton Pipelines

If you followed [How to Deploy Your First Service with Zero-Config CI/CD](/tutorials/how-to-deploy-your-first-service-with-zero-config-cicd), you have a Service building on one of Planton's platform tracks — dockerfile or buildpacks. The platform builds your image, renders your kustomize overlays, and deploys — with no pipeline file in your repository.

That works until you need a step the platform track does not have: SonarQube analysis before every production deploy, a unit-test stage, a build tool like Bazel. A self-managed pipeline is the answer: plain [Tekton](https://tekton.dev) YAML in your repository, which Planton compiles at dispatch into a self-contained run and executes on every commit — while it keeps handling webhook triggering, credentials, status streaming, and deployment.

## What You Will Learn

- Write a Tekton pipeline that clones your repository, runs a SonarQube scan, builds the image, and renders your kustomize overlays
- Reuse Planton's catalog tasks by name and add a Task of your own beside the pipeline
- Validate the pipeline locally with the same compiler that runs at dispatch
- Switch your Service to the pipeline with one field
- Gate a step on the trigger — for example, run it only on tag releases

## Prerequisites

- [ ] A working Service deployed through Planton on a platform track (see the tutorial above)
- [ ] A Dockerfile in your repository (the catalog's `buildkit-daemonless` task builds from it)
- [ ] A [SonarCloud](https://sonarcloud.io) account with a project for your repository, and its token stored as a Planton secret: `planton secret set sonar-token --string` (the build reads it from there — nothing is placed in a cluster by hand)
- [ ] The `planton` CLI installed and authenticated (`planton auth login`)
- [ ] Basic familiarity with Tekton's Pipeline and Task concepts

## Step 1: Write the pipeline

Create `.planton/pipeline.yaml` under your service's project root. It declares the platform's parameter contract, the `source` workspace (the cloned repository), the `secrets` workspace (the build secrets your Service names, one file each), and four tasks. Three come from Planton's catalog by plain name; one — the SonarQube scan — lives beside the pipeline in your repository.

```yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: build-with-sonar
spec:
  params:
    - name: git-url
      type: string
    - name: git-revision
      type: string
    - name: git-branch
      type: string
      default: ""
    - name: project-root
      type: string
      default: "."
    - name: sparse-checkout-directories
      type: string
      default: ""
    - name: image-name
      type: string
    - name: dockerfile-config-map-name
      type: string
    - name: build-context
      type: string
      default: "."
    - name: dockerfile-path
      type: string
      default: "Dockerfile"
    - name: kustomize-manifests-config-map-name
      type: string
    - name: kustomize-base-directory
      type: string
      default: "_kustomize"
    - name: owner-identifier-label-key
      type: string
      default: ""
    - name: owner-identifier-label-value
      type: string
      default: ""
    - name: sonar-project-key
      type: string
  workspaces:
    - name: source
    - name: secrets
  tasks:
    - name: git-checkout
      taskRef:
        name: git-clone
      params:
        - name: url
          value: "$(params.git-url)"
        - name: revision
          value: "$(params.git-revision)"
        - name: deleteExisting
          value: "true"
        - name: sparseCheckoutDirectories
          value: "$(params.sparse-checkout-directories)"
        - name: gitInitImage
          value: "ghcr.io/tektoncd/github.com/tektoncd/pipeline/cmd/git-init:v0.45.0@sha256:8ab0f58d8381b0b71f5b2bae1f63522989d739e3154d8cab1bacfa0ef5317214"
      workspaces:
        - name: output
          workspace: source
    - name: sonar-analysis
      runAfter: [git-checkout]
      taskRef:
        name: sonar-scan
      params:
        - name: project-key
          value: "$(params.sonar-project-key)"
        - name: project-root
          value: "$(params.project-root)"
      workspaces:
        - name: source
          workspace: source
        - name: secrets
          workspace: secrets
    - name: build-image
      runAfter: [sonar-analysis]
      taskRef:
        name: buildkit-daemonless
      params:
        - name: image
          value: "$(params.image-name)"
        - name: contextDir
          value: "$(params.build-context)"
        - name: dockerfilePath
          value: "$(params.dockerfile-path)"
        - name: cache
          value: "true"
        - name: dockerfile-config-map-namespace
          value: "$(context.pipelineRun.namespace)"
        - name: dockerfile-config-map-name
          value: "$(params.dockerfile-config-map-name)"
        - name: owner-identifier-label-key
          value: "$(params.owner-identifier-label-key)"
        - name: owner-identifier-label-value
          value: "$(params.owner-identifier-label-value)"
      workspaces:
        - name: source
          workspace: source
    - name: kustomize-build
      runAfter: [git-checkout]
      taskRef:
        name: kustomize-build
      params:
        - name: config-map-name
          value: "$(params.kustomize-manifests-config-map-name)"
        - name: config-map-namespace
          value: "$(context.pipelineRun.namespace)"
        - name: project-root
          value: "$(params.project-root)"
        - name: kustomize-base-directory
          value: "$(params.kustomize-base-directory)"
        - name: owner-identifier-label-key
          value: "$(params.owner-identifier-label-key)"
        - name: owner-identifier-label-value
          value: "$(params.owner-identifier-label-value)"
      workspaces:
        - name: source
          workspace: source
```

Read each catalog task's own parameters before wiring it — their definitions are plain YAML in the open-source repository under `cicd/tekton/tasks/`. Never guess a parameter name; the compiler will tell you, but the file tells you first.

## Step 2: Add your own Task beside the pipeline

`sonar-scan` is not in the catalog, so define it yourself in `.planton/tasks/sonar-scan.yaml` — an ordinary Tekton Task. The compiler discovers every Task in the `tasks/` directory beside your pipeline file and resolves `taskRef: {name: sonar-scan}` to it.

```yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: sonar-scan
spec:
  params:
    - name: project-key
      type: string
    - name: project-root
      type: string
      default: "."
  workspaces:
    - name: source
    - name: secrets
  steps:
    - name: scan
      image: sonarsource/sonar-scanner-cli:11
      workingDir: $(workspaces.source.path)/$(params.project-root)
      script: |
        export SONAR_TOKEN="$(cat $(workspaces.secrets.path)/sonar-token)"
        sonar-scanner -Dsonar.projectKey=$(params.project-key) -Dsonar.host.url=https://sonarcloud.io
```

Pin the image by digest before you rely on it (`crane digest sonarsource/sonar-scanner-cli:11`, then `image: sonarsource/sonar-scanner-cli:11@sha256:...`) so a moved tag can never change your build. The token is a file the platform mounts at build time from your Planton secret — never a value in the definition, which the compiler refuses — and the platform masks it in the build's logs.

## Step 3: Validate before you push

```bash
planton service pipeline validate .planton/pipeline.yaml \
  --param sonar-project-key=acme_storefront --secret sonar-token -o json
```

The same compiler that runs at dispatch runs here: it discovers `sonar-scan` beside the file, inlines `git-clone`, `buildkit-daemonless`, and `kustomize-build` from the catalog, and checks the parameter contract in both directions. Every problem comes back in one pass as a named verdict — `undeclared_param`, `unresolved_task_ref`, `unbindable_workspace` — each naming the field to fix. Exit code 0 means it compiles clean.

## Step 4: Switch the Service to your pipeline

Replace the platform builder with `tektonPipeline`, supply your own parameter, and name the build secret the pipeline reads:

```yaml
spec:
  build:
    tektonPipeline:
      params:
        sonar-project-key: acme_storefront
      secrets:
        - name: sonar-token          # the file: $(workspaces.secrets.path)/sonar-token
          ref: $secret/sonar-token   # an organization secret reaches every build
    registry: your-registry-connection
    imageRepositoryPath: acme/storefront
```

The default file location is `.planton/pipeline.yaml`; set `tektonPipeline.yamlFile` to keep it elsewhere. Every secret reference is checked to exist when you apply. To give each environment its own token, name the file once per environment — `ref: $secret/@staging/sonar-token`, `ref: $secret/@prod/sonar-token` — and a build reads the one for the environment it deploys to. Apply the manifest:

```bash
planton apply -f service.yaml
```

## Step 5: Push and watch

```bash
git add .planton/pipeline.yaml .planton/tasks/sonar-scan.yaml service.yaml
git commit -m "Build with SonarQube analysis"
git push origin main
```

The push starts a run through the same webhook as before. Watch it:

```bash
planton service runs storefront
planton service follow <run-id>
planton service logs <run-id>
```

You will see `git-checkout`, `sonar-analysis`, `build-image`, and `kustomize-build`. If the scan fails its quality gate, `build-image` never runs and the run fails with the scanner's own output. The run record carries the exact compiled pipeline that executed (`spec.resolved_pipeline`), so a rerun replays it byte-for-byte.

## Common Patterns

**Run a step only on releases.** Declare the optional `git-tag` parameter — the platform supplies it only because you declared it — and gate the task with Tekton's `when`:

```yaml
  params:
    - name: git-tag
      type: string
      default: ""
  tasks:
    - name: sbom
      when:
        - input: "$(params.git-tag)"
          operator: notin
          values: [""]
      taskRef:
        name: generate-sbom
```

`pull-request-number` and `trigger-type` (`webhook`, `manual`, `rerun`) work the same way.

**Keep everything in one file.** A pipeline file may carry its Tasks as additional `---`-separated documents — one Pipeline, any number of Tasks — exactly as `kubectl apply -f` accepts.

**Override a catalog task.** A Task in your `tasks/` directory with the same name as a catalog task wins; the run record shows yours.

**Share a task across services.** Publish it once as a `TektonTask` record (`planton apply -f task.yaml`); every service in the organization resolves it by name.

## What to Do Next

- [Self-Managed Pipelines](/docs/ci-cd/self-managed-pipelines) — the full contract, every compile verdict, and its fix
- [Pipelines](/docs/ci-cd/pipelines) — the run model, stages, and gates
- [How to Configure Branch Deployments and Tag Releases](/tutorials/how-to-configure-branch-deployments-and-tag-releases)
