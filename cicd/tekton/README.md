# Planton-managed Tekton pipelines

This folder is the pipeline catalog behind Planton-managed builds: the platform's build tracks under `pipelines/` and the tasks they reference under `tasks/`, as plain Tekton YAML. It is both the source of truth a person reads and a Go package the platform compiles from.

## One folder, one pin, two consumers

The YAML here is embedded into the Go package in this same directory (`content.go`). The Planton platform imports that package through its pin of this module, so the pipeline content a platform release runs is exactly the content of the open-source release it pins. There is no second copy anywhere to drift, and no build ever fetches a pipeline definition from a live git branch.

Two things on the platform read this package and nothing else does:

- The pipeline compiler. At dispatch it compiles a service's build into one self-contained run: a platform track's pipeline with every task reference resolved inline from this catalog, the platform's parameter contract verified, and the compiled definition stamped on the run record with its source and pin. Reruns replay those exact bytes.
- The build-readiness probe. It derives the set of container images this content can run (`Images()`) and checks that the registries serving them are reachable from the build cluster. That derived set is also the mirror list an air-gapped or egress-restricted cluster needs.

## The pin

`Version` in `content.go` names exactly one content state, and a ledger in `testdata/content-digest.txt` holds the two together: it records every `Version` with the digest of the content it named, append-only. Changing any YAML here without bumping `Version` fails the tests, and so does bumping `Version` with no content change -- re-recording cannot hide either, because the ledger refuses a `Version` already bound to other content and content already bound to another `Version`. The pin stamped on run records is `platform-content/<Version>`. It is deliberately not this module's release version, because a release that touches no Tekton content must not give one content state a second pin.

When you change any file under `pipelines/` or `tasks/`: bump `Version`, then record the new pair with `UPDATE_CONTENT_DIGEST=1 go test ./cicd/tekton/` (it appends one line to the ledger), and add that line to `reviewedLedger` in `content_test.go` in the same change -- the test holds every recorded line to that second copy, so an old line can never be rewritten quietly. The image-set test will also show you, as a diff, every image your change adds or removes.

## Image pinning discipline

Every container image referenced here is pinned as `tag@sha256:digest`: the tag documents intent, the digest is what the cluster pulls, so a build is reproducible and immune to upstream tag mutation. The tests refuse an unpinned image. When bumping an image, change the tag, resolve the new digest for that tag (`crane digest <image:tag>` or the registry's manifest API), and update both together — never drop the digest back to a bare tag.

## Tracks

A track is one pipeline file; its stem is the track name a service's build selects. Every track declares exactly the workspaces the platform binds (`source`) and the platform's param contract; a track nobody can select does not live here.

- `dockerfile.yaml` — clone, build and push the image with BuildKit from the service's Dockerfile, render the kustomize overlays for the deploy stage.
- `buildpacks.yaml` — clone, build and push the image with Cloud Native Buildpacks (language auto-detected), render the kustomize overlays.

## Tasks

Tasks are referenced by their Tekton `metadata.name`, which is what a pipeline's plain `taskRef` names; the file stem is only where the definition lives.

- `git-clone.yaml` — task `git-clone`: clone the repository at the built commit into the source workspace.
- `buildkit.yaml` — task `buildkit-daemonless`: build and push an OCI image with BuildKit in daemonless mode.
- `buildpacks.yaml` — task `buildpacks`: build and push an image with a Cloud Native Buildpacks builder.
- `kustomize-build.yaml` — task `kustomize-build`: render the service's kustomize overlays and store them for the deploy stage.

## Referencing a catalog task from your own pipeline

A repository-resident pipeline references a catalog task by its plain name — no resolver, no repository URL, no revision:

```yaml
tasks:
  - name: git-checkout
    taskRef:
      name: git-clone
```

The compiler inlines the task's definition from this catalog when the pipeline is compiled, so the run is self-contained and the record shows exactly what executed. A `taskRef` that uses a remote resolver is refused at compile time with an error naming the task.

## Conventions the content follows

- Outputs that the deploy stage consumes are written to ConfigMaps in the run's own namespace (`$(context.pipelineRun.namespace)`), never a hardcoded one, so compiled pipelines are target-neutral.
- Owner-identifier labels are stamped on created resources from parameters the platform passes.
- Secrets are runtime references (secret names, workspace mounts), never values in a definition.
