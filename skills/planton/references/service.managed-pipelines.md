---
title: Planton-Managed Builds — The Platform Tracks, the Task Catalog, and the Pin
description: What a service built by Planton-managed Tekton actually runs — the platform tracks a build selects (dockerfile, buildpacks), the catalog tasks a repository pipeline references by plain name (git-clone, buildkit-daemonless, buildpacks, kustomize-build), the param contract the platform supplies to every build and the compiler verifies in both directions, the workspaces the platform binds, what the `platform-content/<version>` pin on a run record means and where those bytes live in the open-source repository, and the digest-pinned image set an air-gapped build cluster must mirror. Read when someone asks what the platform's build does or which images it pulls, is authoring a repository pipeline that should reuse a platform task, hits an `undeclared_param`, `missing_required_param`, `unbindable_workspace`, `unresolved_task_ref`, or `resolver_ref_unsupported` compile error, or asks what a run's pin means.
---

# Planton-Managed Builds — The Platform Tracks, the Task Catalog, and the Pin

Every Planton-managed build is compiled at dispatch into one self-contained Tekton run: the pipeline's task references resolved inline, the platform's parameter contract verified, hazards rewritten, and the compiled definition stamped on the run record before anything executes. How to READ that stamp on a run (`spec.resolved_pipeline`: the exact YAML, its source, its pin) and why a rerun cannot drift is `references/service.reading-a-run.md`. This reference is about the CONTENT: what the platform's own pipelines are, what a repository pipeline may reuse from them, and what the contract demands. Writing or changing a repository pipeline — the files, tasks beside it, the trigger facts, every verdict's fix — is `references/service.pipeline-authoring.md`.

## Where the content lives

The platform's pipeline catalog is the `cicd/tekton/` folder of the open-source repository (`github.com/plantonhq/planton`): `pipelines/` holds one file per build track, `tasks/` the tasks they reference, as plain Tekton YAML anyone can read. That same folder is a Go package the platform imports through its pinned release of the module — the content a platform release runs IS the content of the open-source release it pins. Nothing on the platform fetches a pipeline definition from a live git branch, ever.

## The pin

A run compiled from a platform track stamps `source: platform_release` and `pin: platform-content/<version>` — the catalog's own content version, bumped on every change to the YAML and held to it by a digest test. It is deliberately not the open-source module's release number: a release that touches no Tekton content must not give one content state a second pin. Two runs with the same pin executed byte-identical platform content; a rerun replays the stamped bytes regardless of what the catalog says today.

## The platform tracks

A service's `spec.build` selects a track by its builder:

- `build.dockerfile` → the **dockerfile** track: clone the repository at the built commit, build and push the image with BuildKit from the service's Dockerfile (respecting `dockerfilePath` and `context`), render the kustomize overlays for the deploy stage.
- `build.buildpacks` → the **buildpacks** track: clone, build and push the image with a Cloud Native Buildpacks builder (language detected from the source), render the kustomize overlays.
- `build.tektonPipeline` → no platform track: the service's own pipeline (from its repository at the built commit, or an organization-published record — `references/service.org-publishing.md`) compiles through the same compiler with the same contract.

There is no platform track for Cloudflare Worker scripts; a Worker service builds through its own repository pipeline (`references/service.pipeline-authoring.md`).

## The task catalog — reuse by plain name

A repository pipeline references a catalog task with a plain `taskRef`; the compiler inlines the definition at compile time. Names are the tasks' Tekton `metadata.name`, not their file names:

- `git-clone` — clones the repository at the built commit into the `source` workspace; takes `url`, `revision`, and the pinned `gitInitImage` the platform tracks pass.
- `buildkit-daemonless` — builds and pushes an OCI image with BuildKit (daemonless, privileged); reads the Dockerfile the pipeline exported to a ConfigMap; `cacheImage` names one stable cache reference (the tracks pass `<repository>:buildcache`).
- `buildpacks` — builds and pushes an image with a Buildpacks builder image (`BUILDER_IMAGE`, `APP_IMAGE`, `RUN_IMAGE`, `CACHE_IMAGE`).

Both build tasks write the `image-digest` result — the `sha256:…` of what they pushed — and fail their own step, in words, when they pushed and cannot name it. The deploy stage pins what that result names (`repo:tag@sha256:…`), so re-pushing a tag never changes what runs; a repository pipeline that declares no `image-digest` result deploys by its tag, as before it existed. Both tracks declare the optional `cache-image` fact, which the platform supplies only to a pipeline that declares it.
- `kustomize-build` — renders the service's kustomize overlays under `kustomize-base-directory` and stores them in the ConfigMap the deploy stage reads (`config-map-name`, `config-map-namespace`, the owner-identifier label pair).

A `taskRef` with a `resolver:` (git, hub, bundles) is refused at compile time as `resolver_ref_unsupported`: compiled pipelines are self-contained by definition. A name that is neither a catalog task nor an organization-published `TektonTask` is `unresolved_task_ref`, naming the pipeline task and the ref. Read the task YAML in `cicd/tekton/tasks/` for the exact params and results before wiring one — never guess a param name.

## The param contract — verified both ways

The platform supplies ONE param map to every build, and the compiler verifies it against the pipeline's `spec.params` in both directions before dispatch. Every supplied name the pipeline does not declare is an `undeclared_param` verdict; every declared param without a default that nothing supplies is `missing_required_param`. A custom pipeline must therefore DECLARE all of these (defaults are fine):

- `git-url`, `git-revision` (the commit sha), `git-branch`, `project-root`, `sparse-checkout-directories`
- `owner-identifier-label-key`, `owner-identifier-label-value` — stamp them on everything the pipeline creates; log attribution selects by them
- `kustomize-manifests-config-map-name`, `kustomize-base-directory`
- `image-name` — present when the build names a registry (`build.registry` + `build.imageRepositoryPath`): `<registryHost>/<imageRepositoryPath>:<commitSha>`
- dockerfile builds only: `dockerfile-config-map-name`, `build-context`, and `dockerfile-path` when declared

Plus the service's own `build.tektonPipeline.params` (free-form, cannot shadow the platform's names). The dogfood example of a complete declaration is the control plane's own pipeline in the platform repository; the compile verdict names every mismatch in one pass, so read all of them, not the first.

## Workspaces the platform binds

A pipeline may require exactly one pipeline-level workspace: `source` (the clone). Any other REQUIRED pipeline-level workspace is `unbindable_workspace` — mark it optional or drop it (task-level optional workspaces such as the catalog tasks' `dockerconfig` are fine). Secrets are always runtime references (secret names, workspace mounts); an env var whose value looks like a literal secret is `literal_secret_env`.

## Validate before you push

`planton service pipeline validate <pipeline.yaml> [--param <key>=<value>]...` runs the very same compiler locally against the platform catalog and reports every verdict (`-o json` for the stable shape: `valid`, `source`, `pin`, `compiler_version`, `compiled_bytes`, `tasks_resolved`, `errors`; exit 1 on any error). The platform contract above counts as supplied — pass `--param` only for the pipeline's own params. `--track <name>` validates a platform track instead of a file, and `--task <ref>=<path>` supplies a task definition the way an organization-published record would. On the platform-tools arm, `validate_service_pipeline` is the same compiler over submitted `files` (or a `track`), answering in the same JSON and resolving the organization's published tasks itself when `org` is given (`references/service.pipeline-authoring.md`). Fix every verdict, then push — a pipeline that validates clean compiles clean at dispatch.

## The images a build cluster pulls

Every container image the catalog runs is pinned `tag@sha256:digest`, and the set is derived from the content itself (never hand-listed): the git-init and git-clone images, BuildKit rootless, the Paketo jammy builder, bash, kubectl, node, and the AWS CLI. The runner's build-readiness check probes the registries serving exactly that set from the build cluster and, when one is unreachable, names it and lists every pinned image as the mirror list for an air-gapped or egress-restricted cluster. Pipeline definitions themselves need no egress at all — they are compiled into the run.
