---
title: Detection-First Registration — the Platform Reads the Repo and Proposes
description: Registering a service by letting the platform read the repository — the detect_service_from_repo tool answers with facts plus a proposed Service assembled server-side, the human confirms, apply_service submits. Read when someone asks to register a service from a repo conversationally, when you need to propose a registration rather than interrogate, or when the console's guided creation flow needs explaining.
---

# Detection-First Registration — the Platform Reads the Repo and Proposes

The registration door for "I have code, get it on the platform": the platform reads the repository server-side and proposes the whole registration, the human confirms (editing anything), and the confirmed proposal submits through the same apply every other door uses. In the console this is the guided creation flow behind the services directory's Add New; for an agent it is the same beats conversationally: detect, present the proposal with its evidence, confirm once, apply.

## The beats

1. **Detect** — call `detect_service_from_repo` with the org, the git connection slug, the repository owner and name, and (for monorepos) the service's directory as `project_root`. The read rides the connection's own credential — it reaches exactly the repositories the connection reaches (what the App installation was granted, or what the machine's signed-in account can see), and read access on the connection gates the call. Do not ask the person for these coordinates: `search_api_resources_by_kind` on kind `github_connection` gives the organization's connection slugs, and `list_github_repositories` with `org` + `slug` lists every repository that connection reaches with its `owner_name` and `name` (`references/service.reading-a-repository.md`, "Before a service exists"). Detection reports paths, not contents — when a proposal needs a file's text to be honest (which Dockerfile, what the pipeline declares), `read_repo_files` by the same coordinates opens it.
2. **Present the proposal WITH its evidence.** The response carries `facts` (languages, Dockerfile paths, the `_kustomize` tree's environments and preview overlays, a `.planton/pipeline.yaml`, CI workflow presence) and `proposed_service` — assembled from the facts in ONE server-side home. Relay the proposal, never re-derive it; and name the evidence beside each value ("Dockerfile at the project root, so a Docker build"). That is how the console renders it, and it is how trust is earned.
3. **Ask exactly the open questions.** Two classes are deliberately left open:
   - A question the repository cannot answer arrives as an ABSENT block. Multiple Dockerfiles leave `spec.build` unset — the facts list every path; ask which one builds this service.
   - Organization choices are never proposed: the container registry (`spec.build.registry`) and image path are the user's — discover the org's container-registry connections and confirm. When the org has exactly one registry, offering it as the default is right; picking among several silently is not. When the org has none, do not stop: `planton connect registry detect` lists the registries the org's trusted connections already reach (GHCR through the GitHub sign-in on the machine; ECR, Artifact Registry, ACR through its cloud connections) and writes one with nothing stored — offer that before asking anyone for a key.
4. **Confirm once, then `apply_service`** with the (possibly edited) proposal. Apply has no side effects — no webhooks wired, no build started; the record exists and the next default-branch push runs the first build.

## The proposal's own judgments (teach them, don't re-litigate them)

- A pipeline definition at `.planton/pipeline.yaml` outranks a Dockerfile: the team authored a custom pipeline, so the tekton builder is proposed with its source unset (unset IS the default-path spelling).
- Exactly one Dockerfile proposes the Docker build; the root `Dockerfile` is spelled as an empty path (the spec's own default).
- No Dockerfile proposes Buildpacks — the facts' languages tell the user what the buildpack will meet.
- A conventional `_kustomize` tree proposes the git-maintained deploy posture (`spec.deploy.kustomize` with an empty base directory — the conventional spelling). No tree proposes NO deploy block: "add deployment configuration later" is a valid, honest registration.
- An authored `previews/<env>` overlay proposes pull-request previews ON (`build.triggers.pullRequests.deploy`) — the strongest possible statement of intent, still the user's toggle.

## Honesty rules

- `facts.tree_truncated: true` means the repository was too large to list completely — absence proves nothing. Present the proposal as partial and invite corrections.
- Detection failing is not a dead end: the error names the real cause (a revoked installation, a wrong repository name) — relay it verbatim, and fall back to composing the Service manifest by hand from what the user tells you (the apply grammar is unchanged).
- Monorepos register one service per directory: detection scoped to `project_root` ignores sibling services' Dockerfiles, and the proposed name is the directory's.

## Registering WITH deployment configuration (the inline posture)

A registration may carry real per-environment deployment configuration from birth: `spec.deploy.environments` entries (each a list of full catalog object manifests) with NO `spec.deploy.kustomize` block — that absence IS the declaration that the configuration is manually authored (console, agent, apply) rather than git-maintained. The authoring contract when composing these manifests:

- The workload's container **image stays absent** — the blank field is the injection slot the delivery engine fills with the built artifact at deploy time. Never invent a placeholder.
- `metadata.env` names the entry's environment explicitly, and `metadata.name` follows `{service}-{env}`.
- Which kinds can RECEIVE the built artifact is the `list_service_deployment_targets` catalog's `service_deployable` fact; supporting resources (an ALB, an ingress, a domain mapping) ride the same environment list beside the receiver — an environment is a resource SET.
- Deploys resolve provider credentials from the environment's default provider connection — manifests carry none.
- The workload's **pull login stays absent too** for the service's own registry: on a Kubernetes target the deploy fills `spec.pod.imageRegistries` from the registry connection when it holds a login a cluster can keep, and says so on the run. Author a login only for an image in a DIFFERENT private registry (a sidecar), and only as a `$secret/` reference. A GHCR registry connected through a GitHub sign-in needs a read-only pull token before a Kubernetes deploy can pull with it; a Cloud Run target pulls private images only from Artifact Registry — say both at registration time, not after the first pod fails (`references/service.pulling-private-images.md`).
- Pull-request previews are NOT available for inline-authored configuration (preview deploys ride the repository's kustomize previews tree); say so instead of wiring the toggle.

The console's guided flow offers this as its "Configure Here" posture: pick the target platform, pick environments, size each with dials (untouched dials keep platform defaults — the manifests carry only what the user set), add the boot-level environment variables. An agent doing the same conversationally asks the same few questions and composes the same manifests.

## Where the human-visible twin lives

The console's guided flow (services directory → Add New → the guided door, or `/orgs/{org}/services/new`) walks the same beats visually: source, a one-screen proposal with each finding named, an Environments screen when the inline posture is chosen (target platform, environment chips, sizing dials, variables), review with the exact YAML. Point humans there when they prefer clicking; everything the flow proposes is this same detection read.
