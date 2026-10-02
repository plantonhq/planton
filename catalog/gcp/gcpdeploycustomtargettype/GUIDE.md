# GcpDeployCustomTargetType Guide

The judgment this guide protects: a custom target type is a reusable deployer, shared by every target that deploys the same way, and it runs with each target's identity -- never its own.

## Tasks or custom actions

`tasks` is the direct path: Cloud Deploy runs your `deploy` container (and a `render` container, if you give one) in Cloud Build, passing the release, target, and output location through its `CLOUD_DEPLOY_*` environment variables. Your image does the deploying and writes the results Cloud Deploy expects. `customActions` names Skaffold custom actions instead -- `deployAction` (required) and optionally `renderAction` -- which suits teams whose releases are already Skaffold projects. Set one or the other, never both. With neither, the type defines nothing. Leave render unset to let Cloud Deploy render with Skaffold as usual.

## Sharing Skaffold modules

`includeSkaffoldModules` adds remote Skaffold configs to every release that deploys through the type, so the custom actions live in one place. Each module names exactly one source: `git` (a clone URL plus optional path and branch or tag), `googleCloudBuildRepo` (a repository linked through a Cloud Build connection -- reference the `GcpCloudBuildRepository`'s `name`), or `googleCloudStorage` (a `gs://` path copied recursively). `configs` picks configs by name from the source; empty takes them all.

## Identity, location, and lifecycle

The containers run in the execution environment of the target being deployed: its execution service account and its worker pool. That account needs to pull the image, read any module source, and call whatever system it deploys to -- grant those on the target's account, not on the type. Targets use a type in their own project and location. `location` and `customTargetTypeId` are fixed at creation; everything else updates in place and takes effect on the next rollout. Delete the targets that use the type before the type itself; in a chart, their references order it.
