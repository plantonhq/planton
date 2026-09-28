---
title: Authoring a Repository Pipeline — The Files, the Tasks Beside It, the Contract, and Every Verdict's Fix
description: How to write, change, and repair a service's own Tekton pipeline exactly the way the platform compiles it — the default file and how to relocate it, Tasks beside the pipeline (a tasks/ directory or extra documents in the same file) and the three-rung name resolution, the full parameter contract every pipeline must declare plus the optional trigger facts (git-tag, pull-request-number, trigger-type) gated with Tekton's own when, the one bindable workspace, the local validate loop and its JSON, and every compile verdict mapped to the exact edit that fixes it. Read when someone asks to switch a service to its own pipeline, add a build step (lint, tests, a scan, an SBOM), run a step only on releases or pull requests, reuse or override a catalog task, share a task across services, or when a run fails with a compile verdict.
---

# Authoring a Repository Pipeline — The Files, the Tasks Beside It, the Contract, and Every Verdict's Fix

You are the developer's only teacher here: they will not read the Tekton contract, they will ask you for what they want ("add a lint step", "only build the SBOM on releases"). Edit their files the way the platform compiles them, then run the validate loop before you call anything ready. What a Planton-managed build IS (tracks, catalog tasks, the pin, the images) is `references/service.managed-pipelines.md`; reading a run's stamp is `references/service.reading-a-run.md`. This file is the authoring craft.

## The files

- **The pipeline**: `.planton/pipeline.yaml` under the service's project root (`spec.gitRepo.projectRoot`), a single `tekton.dev/v1` `Pipeline` document. Selected by `spec.build.tektonPipeline: {}`; relocated with `spec.build.tektonPipeline.yamlFile: <path relative to the project root>`. Read at the built commit, never a branch.
- **Tasks beside it**: any `*.yaml` in the `tasks/` directory beside the pipeline file — `.planton/tasks/` for the default location, `ci/tasks/` for `ci/build.yaml` — each a `tekton.dev/v1` `Task`. Or, in the pipeline file itself, additional `---`-separated `Task` documents (one Pipeline, any number of Tasks). Prefer the directory when a team will grow the set; prefer the single file for one small helper task.
- **The service's own params**: `spec.build.tektonPipeline.params: {key: value}` — free-form strings the pipeline declares and uses; they can never shadow the platform's names below.

## Name resolution — most specific wins

A plain `taskRef: {name: X}` resolves, in order: (1) a Task named `X` in the repository (beside the pipeline or in the file); (2) the organization's published `TektonTask` named `X`; (3) the platform catalog — `git-clone`, `buildkit-daemonless`, `buildpacks`, `kustomize-build`. A repository task shadows a catalog task of the same name; the run record shows which ran. `resolver:` refs (git, hub, bundles) are refused — nothing is fetched at run time; the compiler inlines every task's spec into the run.

Read a catalog task's params in the open-source repository under `cicd/tekton/tasks/<file>.yaml` before wiring it. Never guess a param name.

## The contract the pipeline MUST declare

The platform supplies these to every build and the compiler checks both directions (supplied-but-undeclared and declared-required-but-unsupplied are verdicts). Declare all of them under `spec.params` (defaults allowed):

`git-url`, `git-revision`, `git-branch`, `project-root`, `sparse-checkout-directories`, `owner-identifier-label-key`, `owner-identifier-label-value`, `kustomize-manifests-config-map-name`, `kustomize-base-directory`; `image-name` when `spec.build.registry` is set; `dockerfile-config-map-name`, `build-context`, `dockerfile-path` when the service's builder is dockerfile (a repository pipeline never is — omit them). Plus every key under `tektonPipeline.params`.

**Optional facts — supplied only when declared**: `git-tag` (tag name or `""`), `pull-request-number` (number or `""`), `trigger-type` (`webhook`, `manual`, `rerun`), `target-platform` (the deploy target's `os/arch`, comma-separated when the service targets several — e.g. `linux/amd64`; `""` when no target asks for one, meaning the build machine's own). Declaring one is how a pipeline opts in; nothing else changes. The platform's own tracks declare `target-platform` and pass it to BuildKit's `platforms`, so a laptop builds for the cloud the image will run on, emulating when the machine differs. A pipeline that builds for a target should declare it too, and give its build task a result named `build-node-architecture` (the pod's own `uname -m`, normalized to `amd64`/`arm64`) — the run record reads that result beside the requested platform to say *built for linux/amd64 on an arm64 machine — emulated* as a fact. Without the result nothing about platforms is claimed for the run. Two more facts serve the build task: `cache-image` (one stable reference for the build cache, `<repository>:buildcache` — hand it to BuildKit's `cacheImage` or Buildpacks' `CACHE_IMAGE` so a build tagged per commit still reuses layers) and `registry-sign-in-expires-at` (Unix seconds, supplied by the runner when the registry sign-in it minted expires; BuildKit's `registrySignInExpiresAt` then says so when a late push is refused).

**Results the run reads**: give the build task a result named `image-digest` carrying the `sha256:…` of what it pushed, and the deploy stage pins that exact image (`repo:tag@sha256:…`), so re-pushing the tag never changes what runs. A pipeline without it deploys by its tag. A pipeline that also pushes a tag build's image under its release tag gives the build task a result named `release-tag` carrying that tag; the run records it as the artifact's `release_tag` only when it equals the `git-tag` the run was dispatched with and is a valid image tag (`[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}`), so the registry and the run both answer which image a release is.

**Workspaces**: declare exactly `source` at pipeline level and pass it to tasks (`workspaces: [{name: source, workspace: source}]`; `git-clone` calls its workspace `output`). Any other required pipeline-level workspace is refused — mark it `optional: true` or drop it. ConfigMaps the deploy stage reads are written into `$(context.pipelineRun.namespace)`, never a literal namespace.

## Worked example — the complete file pair

A pipeline that clones, lints with a repository task, builds with the catalog's BuildKit task, and renders overlays:

```yaml
# .planton/pipeline.yaml
apiVersion: tekton.dev/v1
kind: Pipeline
metadata: {name: storefront-build}
spec:
  params:
    - {name: git-url, type: string}
    - {name: git-revision, type: string}
    - {name: git-branch, type: string, default: ""}
    - {name: project-root, type: string, default: "."}
    - {name: sparse-checkout-directories, type: string, default: ""}
    - {name: image-name, type: string}
    - {name: kustomize-manifests-config-map-name, type: string}
    - {name: kustomize-base-directory, type: string, default: "_kustomize"}
    - {name: owner-identifier-label-key, type: string, default: ""}
    - {name: owner-identifier-label-value, type: string, default: ""}
    - {name: dockerfile-config-map-name, type: string, default: "storefront-dockerfile"}
  workspaces:
    - {name: source}
  tasks:
    - name: git-checkout
      taskRef: {name: git-clone}
      params:
        - {name: url, value: "$(params.git-url)"}
        - {name: revision, value: "$(params.git-revision)"}
        - {name: deleteExisting, value: "true"}
        - {name: sparseCheckoutDirectories, value: "$(params.sparse-checkout-directories)"}
        - {name: gitInitImage, value: "ghcr.io/tektoncd/github.com/tektoncd/pipeline/cmd/git-init:v0.45.0@sha256:8ab0f58d8381b0b71f5b2bae1f63522989d739e3154d8cab1bacfa0ef5317214"}
      workspaces: [{name: output, workspace: source}]
    - name: lint
      runAfter: [git-checkout]
      taskRef: {name: eslint-check}
      workspaces: [{name: source, workspace: source}]
    - name: build-image
      runAfter: [lint]
      taskRef: {name: buildkit-daemonless}
      params:
        - {name: image, value: "$(params.image-name)"}
        - {name: contextDir, value: "$(params.project-root)"}
        - {name: dockerfilePath, value: "Dockerfile"}
        - {name: cache, value: "true"}
        - {name: dockerfile-config-map-namespace, value: "$(context.pipelineRun.namespace)"}
        - {name: dockerfile-config-map-name, value: "$(params.dockerfile-config-map-name)"}
        - {name: owner-identifier-label-key, value: "$(params.owner-identifier-label-key)"}
        - {name: owner-identifier-label-value, value: "$(params.owner-identifier-label-value)"}
      workspaces: [{name: source, workspace: source}]
    - name: kustomize-build
      runAfter: [git-checkout]
      taskRef: {name: kustomize-build}
      params:
        - {name: config-map-name, value: "$(params.kustomize-manifests-config-map-name)"}
        - {name: config-map-namespace, value: "$(context.pipelineRun.namespace)"}
        - {name: project-root, value: "$(params.project-root)"}
        - {name: kustomize-base-directory, value: "$(params.kustomize-base-directory)"}
        - {name: owner-identifier-label-key, value: "$(params.owner-identifier-label-key)"}
        - {name: owner-identifier-label-value, value: "$(params.owner-identifier-label-value)"}
      workspaces: [{name: source, workspace: source}]
```

```yaml
# .planton/tasks/eslint-check.yaml
apiVersion: tekton.dev/v1
kind: Task
metadata: {name: eslint-check}
spec:
  workspaces: [{name: source}]
  steps:
    - name: lint
      image: node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293
      workingDir: $(workspaces.source.path)
      script: |
        npm ci && npx eslint .
```

Pin every image by digest (`crane digest <image:tag>`); secrets are `valueFrom.secretKeyRef` or mounts, never literals.

## Asked for → what you edit

- **"Add a lint/test/scan step"**: write `.planton/tasks/<name>.yaml` (a Task with a `source` workspace), add a pipeline task with `taskRef: {name: <name>}`, `runAfter: [git-checkout]`, and make `build-image` `runAfter` it if the step must gate the build.
- **"Only on releases" / "only on pull requests"**: declare `git-tag` (or `pull-request-number`) with `default: ""`; on the task add `when: [{input: "$(params.git-tag)", operator: notin, values: [""]}]` (use `in` with `values: [""]` for "only when NOT a tag").
- **"Use our own clone/build step"**: put a Task with the catalog task's name in `tasks/` — it shadows the catalog's.
- **"Share this task / this pipeline across services"**: publish it as a `TektonTask` / `TektonPipeline` record and consume it by name -- the records, the publish check, and the switch are `references/service.org-publishing.md`.
- **"Switch this service to its own pipeline"**: `spec.build.dockerfile`/`buildpacks` → `spec.build.tektonPipeline: {}`; keep `registry` and `imageRepositoryPath`; write the pipeline (start from the worked example); validate; `planton apply -f service.yaml`.
- **"Move the pipeline file"**: set `tektonPipeline.yamlFile`; move the `tasks/` directory beside the new location.

## Validate, always, before "ready"

`planton service pipeline validate .planton/pipeline.yaml [--param <key>=<value>]... -o json` runs the dispatch compiler locally: discovers the tasks beside the file, resolves catalog refs, checks the contract, and returns `{valid, source, pin, compiler_version, compiled_bytes, tasks_resolved: [{name, source: repo|org|platform}], errors: [{code, subject, message}]}`; exit 1 on any error. The platform's contract stands in for the dispatch: the always-supplied params count as supplied (and a pipeline that forgets to declare one is refused with `undeclared_param`, exactly as the dispatch would), declared optional facts count as supplied, and `image-name` / the dockerfile params are treated as supplied when declared. Pass `--param` only for the pipeline's own params (every `tektonPipeline.params` key; `--param git-tag=v1.4.0` exercises a `when`); `--task <name>=<path>` stands in for an organization-published task. Relay errors in the developer's words and fix them yourself — every one below has a mechanical fix.

On the platform-tools arm the twin is `validate_service_pipeline`, with three sources and exactly one per call: the **repository itself** (`org` + `service`, or `org` + `git_connection` + `owner_name` + `repo_name`, plus an optional `ref` — the platform reads the pipeline and the tasks beside it at HEAD or at the commit a run built, nothing pasted; the report's `pin` is the commit read), submitted `files` (a path-to-content map of the pipeline file — by default `.planton/pipeline.yaml` — plus every Task file beside it under its `tasks/` directory) with the pipeline's own `params`, or `track` to validate a platform track. Prefer the repository source for "does my pipeline compile?" and for re-checking exactly what a failed run compiled (`ref` = its `spec.git_commit.sha`); use `files` for a change that exists only in the conversation. It runs the same compiler with the same stand-in rule and answers with the same JSON; verdicts arrive in the report (`valid: false`, `errors`), never as a tool error. Pass `org` and task references the files do not answer are resolved from the organization's published `TektonTask` records automatically — the twin needs no `--task` stand-ins. A missing `pipeline_path` entry, more or fewer than one source, a repository source without `org`, or an oversized payload is a malformed call, not a verdict: fix the call. Reading the files themselves — to quote a task, or to see what the developer actually committed — is `references/service.reading-a-repository.md`.

## Every verdict → the fix

- `invalid_yaml` (subject: the file) — fix the YAML the message quotes.
- `invalid_pipeline_schema` — `apiVersion: tekton.dev/v1`, `kind: Pipeline`, or the field Tekton's own validation names.
- `multiple_pipeline_documents` — the source holds more than one Pipeline; keep one, turn the rest into Tasks or separate services.
- `unexpected_document_kind` (subject: file or document index) — a document beside the pipeline is neither Pipeline nor Task; remove it or move it out of `tasks/`.
- `duplicate_task_definition` (subject: the name) — two repository documents define the same Task name; rename one.
- `unresolved_task_ref` (subject: the ref) — the name is in no repository file, no published record, no catalog; add the Task beside the pipeline, publish it (`references/service.org-publishing.md`), or fix the spelling.
- `resolver_ref_unsupported` (subject: the pipeline task) — replace `taskRef.resolver` with a plain `name` and add the Task beside the pipeline.
- `undeclared_param` (subject: the param) — add it to `spec.params` (with a default if the pipeline ignores it); for a platform contract param the message says so -- every pipeline must declare all nine.
- `missing_required_param` (subject: the param) — give it a `default`, or add it under `spec.build.tektonPipeline.params`.
- `unbindable_workspace` (subject: the workspace) — only `source` is bound; mark the workspace `optional: true` or remove it.
- `dangling_result_reference` (subject: the reference) — fix `$(tasks.<task>.results.<result>)` to a task and result that exist.
- `literal_secret_env` (subject: the task) — replace the value with `valueFrom.secretKeyRef` or a mount.
- `repo_file_unreadable` (subject: the path) — the file is not at that path at the built commit; fix `yamlFile` or commit the file.

Verdicts land on the run record as an explained failure. A rerun of a compile-failed run recompiles from the repository at the SAME commit, so the path is: fix the file, validate, push a new commit — the push starts the run.
