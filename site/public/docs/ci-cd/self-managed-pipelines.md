---
title: "Self-Managed Pipelines"
description: "Bring your own Tekton pipeline — and the Tasks it references — in your repository, compiled by Planton at dispatch into a self-contained run with platform deployment, status tracking, and gates."
icon: pipeline
order: 35
tags:
  - Pipelines
  - Tekton
  - CI/CD
  - Service Hub
---

# Self-Managed Pipelines

A self-managed pipeline is plain Tekton YAML in your repository. Planton compiles it at dispatch into one self-contained run — every task reference resolved inline, the platform's parameter contract verified, the compiled definition stamped on the run record — and then handles what it always handles: webhook triggering, credentials, status streaming, the deploy stage, and approval gates.

## Platform track or your own pipeline

A service's `spec.build` selects exactly one builder:

| Builder | Who writes the pipeline | Deploy stage |
|---|---|---|
| `dockerfile` | Planton's release-pinned dockerfile track | Planton |
| `buildpacks` | Planton's release-pinned buildpacks track | Planton |
| `tektonPipeline` | You, as Tekton YAML in your repository (or an organization-published pipeline) | Planton |

Choose your own pipeline when you need build steps the platform tracks do not have: test suites, linting, security scanning, a build tool like Bazel, parallel or conditional tasks.

## The pipeline file

By default the pipeline lives at `.planton/pipeline.yaml` under the service's project root — beside the service's Dockerfile and its `_kustomize` tree. Selecting it is one line:

```yaml
spec:
  build:
    tektonPipeline: {}
    registry: my-registry
    imageRepositoryPath: acme/storefront
```

To keep it elsewhere, name the path (relative to the project root):

```yaml
spec:
  build:
    tektonPipeline:
      yamlFile: ci/build.yaml
```

The file is a `tekton.dev/v1` `Pipeline` document. It is read at the exact commit being built — never from a branch.

## Tasks beside the pipeline

A pipeline references tasks by their plain Tekton name:

```yaml
tasks:
  - name: lint
    taskRef:
      name: eslint-check
```

The compiler resolves a plain reference from three places, most specific first:

1. **Your repository** — a `Task` document in a `tasks/` directory beside the pipeline file (`.planton/tasks/*.yaml` for the default location), or a second document in the pipeline file itself (`---`-separated, exactly as `kubectl apply -f` accepts).
2. **Your organization** — a published `TektonTask` record with that name.
3. **The platform catalog** — `git-clone`, `buildkit-daemonless`, `buildpacks`, `kustomize-build` (their definitions live in the open-source repository under `cicd/tekton/tasks/`).

A repository task with the same name as a catalog task wins — your file is what runs, and the run record shows it. A `taskRef` that uses a remote resolver (`resolver: git`, `hub`, `bundles`) is refused: compiled pipelines are self-contained, so nothing is fetched at run time.

A repository task is an ordinary Tekton `Task`:

```yaml
# .planton/tasks/eslint-check.yaml
apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: eslint-check
spec:
  workspaces:
    - name: source
  steps:
    - name: lint
      image: node:20-alpine@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293
      workingDir: $(workspaces.source.path)
      script: |
        npm ci && npx eslint .
```

## The parameter contract

Planton supplies one set of parameters to every build, and the compiler verifies it against your pipeline's `spec.params` in both directions before dispatch: a supplied parameter your pipeline does not declare is a compile error, and a declared parameter with no default that nothing supplies is a compile error. Your pipeline must therefore declare every one of these (defaults are fine):

| Parameter | Value |
|---|---|
| `git-url` | The repository clone URL |
| `git-revision` | The commit SHA being built |
| `git-branch` | The branch name (empty for a tag build) |
| `project-root` | The service's project root within the repository |
| `sparse-checkout-directories` | Comma-separated sparse-checkout patterns (may be empty) |
| `owner-identifier-label-key`, `owner-identifier-label-value` | Stamp these on everything the pipeline creates — log attribution selects by them |
| `kustomize-manifests-config-map-name` | Where the deploy stage reads the rendered overlays |
| `kustomize-base-directory` | The service's kustomize base directory |
| `image-name` | Present when the build names a registry: `<registryHost>/<imageRepositoryPath>:<commitSha>` |
| `dockerfile-config-map-name`, `build-context`, `dockerfile-path` | Dockerfile builds only |

Plus anything you put under `build.tektonPipeline.params` (free-form key/value pairs; they cannot shadow the platform's names).

### Trigger facts — supplied only if you declare them

Three more parameters describe the run, and the platform supplies them **only when your pipeline declares them**, so an existing pipeline never has to change:

| Parameter | Value |
|---|---|
| `git-tag` | The tag name for a tag build; empty otherwise |
| `pull-request-number` | The pull request number for a pull-request build; empty otherwise |
| `trigger-type` | `webhook`, `manual`, or `rerun` |

Gate a task with Tekton's own `when` — no platform grammar to learn:

```yaml
spec:
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

## Workspaces

The platform binds `source`, the cloned repository, for every build — declare it and pass it to your tasks — and `secrets`, the build secrets your Service names, one file each. A pipeline that requires any other workspace is refused at compile time — declare it `optional: true` or drop it.

## Secrets and Variables

A pipeline reads its configuration from Planton, never from values typed into the Service:

- **Build secrets.** Name each secret under `build.tektonPipeline.secrets` — a file name and a `$secret/` reference — and read it at `$(workspaces.secrets.path)/<name>`. The platform reads it at build time with the build's own short-lived identity, deletes it when the run ends, and masks its value in the build's logs.
- **Per environment.** `$secret/@<env>/<slug>` reaches only builds for that environment (a branch mapped to it, a run scoped to it, a pull request previewing it), so name a file once per environment and each branch reads its own value. A build that serves several environments — the promotion walk, a tag release — receives your organization's secrets only.
- **Variables.** A param's value may be `$var/<slug>`; it is filled in when the build starts. A `$secret/` param is refused — params are visible on the run object, so secrets are files.
- Every reference is checked to exist when you apply the Service, and each run records which build secrets it could read (names and references, never values).

```yaml
build:
  tektonPipeline:
    params:
      sonar-host-url: $var/sonarqube-host-url
    secrets:
      - name: sonar-token
        ref: $secret/@staging/sonar-token
      - name: sonar-token
        ref: $secret/@prod/sonar-token
```

## Validate before you push

```bash
planton service pipeline validate .planton/pipeline.yaml -o json
```

The same compiler that runs at dispatch runs locally: it discovers the tasks beside the file, resolves catalog references, checks the parameter contract, and reports every problem in one pass. The platform's contract stands in for the dispatch — the parameters every build receives count as supplied, and a pipeline that forgets to declare one is refused here exactly as the dispatch would refuse it — so pass `--param` only for the pipeline's own parameters (`--param git-tag=v1.4.0` exercises a `when`); `--task <name>=<path>` supplies a task the way an organization-published record would. The JSON carries `valid`, `source`, `pin`, `compiler_version`, `compiled_bytes`, `tasks_resolved` (each with the rung that answered it), and `errors`; the exit code is 1 on any error. A pipeline that validates clean compiles clean. An agent working through Planton's MCP server has the same loop as the `validate_service_pipeline` tool: pointed at your service it reads the pipeline and the task files beside it straight from your repository — at HEAD, or at the exact commit a run built — answers with the same report, and resolves the tasks your organization has published without you naming them; it also accepts the files themselves for a change that exists only in the conversation.

## Compile verdicts

| Code | Meaning | Fix |
|---|---|---|
| `invalid_yaml` | A document is not valid YAML | Fix the YAML; the message names the file |
| `invalid_pipeline_schema` | The document is not a `tekton.dev/v1` Pipeline, or fails Tekton's own validation | Correct `apiVersion`/`kind`, or the field Tekton names |
| `multiple_pipeline_documents` | The source contains more than one Pipeline | Keep one Pipeline per source; extra documents may only be Tasks |
| `unexpected_document_kind` | A document beside the pipeline is neither a Pipeline nor a Task | Remove it, or move it out of the pipeline's `tasks/` directory |
| `duplicate_task_definition` | Two repository documents define a Task with the same name | Rename one |
| `unresolved_task_ref` | A `taskRef` names nothing in your repository, your organization, or the catalog | Add the Task beside the pipeline, publish it, or fix the name |
| `resolver_ref_unsupported` | A `taskRef` uses a remote resolver | Reference the task by name; inline it or add it beside the pipeline |
| `undeclared_param` | The platform supplies a parameter the pipeline does not declare (or a `build.tektonPipeline.params` key is not declared) | Declare it under `spec.params` — every pipeline declares the nine always-supplied parameters |
| `missing_required_param` | A declared parameter has no default and nothing supplies it | Give it a default, or set it under `build.tektonPipeline.params` |
| `unbindable_workspace` | The pipeline requires a workspace the platform does not bind | Use `source` (or `secrets`), or mark the workspace `optional: true` |
| `secrets_workspace_without_secrets` | The pipeline requires `secrets`, but the Service names no build secrets | Name them under `build.tektonPipeline.secrets` (validate with `--secret <name>`), or mark the workspace optional |
| `build_secrets_never_mounted` | Build secrets reach this build, but the pipeline declares no `secrets` workspace | Declare it and read `$(workspaces.secrets.path)/<name>` |
| `dangling_result_reference` | `$(tasks.X.results.Y)` names a task or result that does not exist | Fix the task or result name |
| `literal_secret_env` | An env var looks like a literal secret value | Name it as a build secret and read it from the `secrets` workspace |
| `repo_file_unreadable` | The pipeline file could not be read at the built commit | Check `yamlFile` and that the file exists at that commit |

## What a run records

Every run stamps the compiled definition on its record: the exact YAML that executed, `source: repo`, and the commit SHA as its pin. A rerun replays those bytes byte-identically. The run view and `planton service pipeline get` show them; see [Pipelines](/docs/ci-cd/pipelines) for the run model.

## Organization-published pipelines and tasks

An organization writes a pipeline once and every service consumes it by name. Two record kinds carry the Tekton content verbatim:

- **`TektonTask`** — exactly one `tekton.dev/v1` Task document in `spec.yamlContent`. Any pipeline in the organization (a repository pipeline or a published one) references it with a plain `taskRef: {name: <name>}`; it resolves after the pipeline's own tasks and before the platform catalog.
- **`TektonPipeline`** — one `tekton.dev/v1` Pipeline plus any Task documents it bundles (`---`-separated) in `spec.yamlContent`. A service consumes it with one line: `spec.build.tektonPipeline: {pipeline: <name>}`.

Both are organization-owned records with a `description` (one line, indexed for search) and an `overviewMarkdown` where the publishing team documents what a consumer must supply — the parameters a service passes under `build.tektonPipeline.params`, and whether the service must name a container registry. The record's `metadata.name` must equal the Tekton document's `metadata.name` and must be a Tekton name (lowercase letters, digits, hyphens), so the record's name is also its slug: `planton get TektonPipeline <name>`, the service's `pipeline: <name>`, and a `taskRef` all use the same one name.

```yaml
apiVersion: service-hub.planton.ai/v1alpha1
kind: TektonPipeline
metadata: {name: node-service-ci, org: acme}
spec:
  description: Clone, lint, build with BuildKit, render kustomize
  overviewMarkdown: |
    Set `spec.build.registry` (this pipeline pushes an image) and
    `build.tektonPipeline.params.node-version`.
  yamlContent: |
    apiVersion: tekton.dev/v1
    kind: Pipeline
    metadata: {name: node-service-ci}
    spec: ...   # declares the platform contract, image-name, node-version
```

**Publishing is one `planton apply -f` per record, tasks first.** The publish check runs before the record is written — at the CLI, through the assistant's tools, and at the API: the same compiler that runs at dispatch verifies the document shape, the name law, every task reference (against the bundled tasks, the organization's published tasks, and the catalog), the parameter contract, workspaces, result references, and Tekton's own admission rules. A record that would not compile is refused with the verdicts above in one sentence, and nothing is written. A pipeline that references a task the organization has not published yet is refused with `unresolved_task_ref`: publish the task first, or bundle it as a second document.

**Consuming is verified when the service is applied**: a `build.tektonPipeline.pipeline` naming no published record is refused there, with the fix in the sentence — never at the first build. Each run of a consuming service stamps `source: org_published` and the pin `<record id>@<record spec updated-at>`; updating a record changes what the next fresh run compiles, while reruns replay the stamp they carry.

**Deleting is refused while anything depends on the record**: a `TektonPipeline` any service still builds from (the refusal names the services), a `TektonTask` any published pipeline still references (the refusal names the pipelines). Repository pipelines that reference a deleted task learn at their next compile through `unresolved_task_ref`.

There is no catalog page. `planton search --kind TektonPipeline` (or `TektonTask`) lists what the organization has published; `planton get TektonPipeline <name>` shows a record with its documentation; the assistant reads the same through its `get_tekton_pipeline` / `get_tekton_task` tools and search.

## Related Documentation

- [Pipelines](/docs/ci-cd/pipelines) — the run model, stages, and gates
- [Build Methods](/docs/ci-cd/build-methods) — the platform's dockerfile and buildpacks tracks
- [Monorepo Support](/docs/ci-cd/monorepo-support) — project roots and trigger paths
