---
title: Offline Deploys — Services with No Planton Backend
description: Deploying a service entirely without a Planton backend — authoring an offline-clean kustomize tree, the offline deploy verb's preflight-report-then-deploy contract and exit codes, node-addressed --set overrides, state backend truths, the GitHub Action's offline mode, and the complete gh-driven journey from "set up CI/CD on my repo" to a watched live deploy. Read when a user wants their repository deploying from GitHub Actions with no backend anywhere, asks what "offline" means for services, needs an offline refusal explained, or asks you to set up CI/CD end to end.
---

# Offline Deploys — Services with No Planton Backend

`planton service deploy --env <env>` runs in one of two lanes, decided by whether a backend is configured. Connected, the control plane runs pipelines, gates, and rollout verification. With NO backend configured — a bare CI runner, or any machine with `PLANTON_BACKEND=none` set for the invocation — the same verb deploys the working tree's own kustomize declaration through the open-source engine: the overlay renders locally, a preflight report verifies everything verifiable BEFORE anything is handed to an IaC engine, and the resources deploy sequentially in the dependency order their own `valueFrom` references declare, each resource's captured outputs resolving the next ones' references.

**One asymmetry governs everything you do here: YOU are never offline — only the artifact is.** On a desktop instance you have a backend (the local control plane), your platform tools, and this skill. The offline constraints bind the TREE you author and the WORKFLOW you wire, because they will run where no backend exists. Author with your full instruments; produce an artifact that needs none.

## Authoring an offline-clean tree

The service's `_kustomize/` tree (base + one overlay per environment — `references/service.kustomize-authoring.md` has the conventions) is offline-clean when every rendered manifest stands alone:

- **No `$var/` or `$secret/` references, anywhere.** These resolve only through a Planton backend; offline preflight refuses each one naming its field. Values you would have made a variable are resolved by YOU at authoring time and written as literals — you have the estate knowledge; the runner must not need it.
- **Runtime secrets ride PROVIDER-NATIVE references** — a Cloud Run secret reference, an ECS `valueFrom` ARN, a Key Vault reference, ESO on Kubernetes. These are references into the customer's own vault, safe to commit, resolved by the provider at runtime with no Planton anywhere. The rendered idioms come from the snippets surface — `planton secret snippet` on the CLI arm, the `list_secret_snippets` tool on the platform arm — which is the ONE authored source of every per-target shape: compose what it returns, never write the idiom from memory.
- **A private image's pull login is the same story.** `spec.pod.imageRegistries` takes only a `$secret/` password, so it is a connected-lane arm and the preflight refuses it offline naming the field. Offline, a Kubernetes workload pulls with the cluster's own identity (its cloud's registry) or through a `KubernetesExternalSecret` (the **Docker Registry Pull Secret (Template)** preset, fed from the customer's secrets manager) declared beside it and named in `spec.pod.imagePullSecrets` with `kind: KubernetesExternalSecret` and `fieldPath: status.outputs.secret_name` (`references/service.pulling-private-images.md`).
- **Cross-resource wiring stays `valueFrom`** — offline resolution feeds each reference from the producer's captured outputs, exactly like the connected lane. A reference whose target is NOT in the tree refuses at preflight (there is no backend to discover it): add the producer's manifest to the tree, or deploy connected.

## The offline verb's contract

- `--env <env>` names the overlay to render — required.
- The **preflight report** comes first, always: schema validation, reference resolvability, `$var`/`$secret` absence, dependency order (cycles named as chains), engine binaries, module availability, state backend completeness and REACHABILITY (probed with the credentials in hand), and live cloud-credential checks. Every failure is a field-naming sentence; ALL problems report at once.
- **One approval for the whole set**: `--auto-approve` in CI, one interactive question otherwise. A non-interactive run without approval refuses rather than hanging a runner.
- **Exit codes are the CI contract**: `2` refused at preflight OR not approved — nothing was handed to an IaC engine; `1` a resource failed to deploy — completed resources are safe in state, and re-running the same command re-applies them as no-ops and continues; `0` every resource is live, URLs printed.
- **`--image <reference>` injects the built image into the tree's annotated image slots** — the same slots the connected lane's injectors write, declared in the catalog's own protos. Blank container images receive the reference; explicitly authored images (sidecars) are untouched (leaving the image blank IS the authoring contract); the Kubernetes shape receives it split into repo+tag, with `--branch` stamping the version slot when given. A tree whose kinds declare no slot refuses honestly — the manifests' values are the truth. `--set <Kind>/<name>:<fieldPath>=<value>` remains the field-exact escape hatch (a bare field path would be ambiguous across a tree's documents).
- **State backends**: on a laptop, zero configuration works — each resource keeps its state in an identity-keyed workspace under `~/.planton/setdeploy/`, so re-runs are honest. On CI runners, which are ephemeral, state must outlive the run: `PLANTON_BACKEND_TYPE` + `PLANTON_BACKEND_BUCKET` (+ region/endpoint as the backend needs) as job env, or the equivalent annotations in the manifests. The preflight report states where state lives. Encrypting that state is the engines' own setting, passed as job env like the backend: `TF_ENCRYPTION` for OpenTofu, a real `PULUMI_CONFIG_PASSPHRASE` (or another secrets provider) for Pulumi -- an empty passphrase leaves Pulumi's secrets openable by anyone who can read the bucket.

## Verifying the artifact from a connected machine

On a desktop the CLI is CONNECTED, so verify the offline artifact by forcing the lane for one invocation, preflight only:

```
PLANTON_BACKEND=none planton service deploy --env prod --preflight-only
```

The full preflight report prints and nothing deploys; the exit code is first-class — 0 means the wall is clean, 2 means it refused (with every problem named in the report). Pass the same `--image`/`--set` the workflow will use, so the wall verifies exactly what CI would deploy. And say honestly what this run does NOT prove: it verified the tree and THIS machine's credentials; the RUNNER's cloud credentials and its view of the state backend are proven by the first workflow run.

## The GitHub Actions journey

When a user says "set up CI/CD on GitHub for my repo", this is the whole choreography. Verify at every step — the finished setup must be PROVEN before you call it ready, never assumed.

1. **Author the tree offline-clean** (above) in the repository, and verify it with the forced-offline preflight.
2. **Ask for one `gh auth login`** if the `gh` CLI is not yet authenticated (`gh auth status`). This is the single manual step; everything after it is your hands.
3. **Author the workflow** using the published Action's offline mode — `plantonhq/planton/actions/deploy` with NO `org`/`audience` inputs (their absence selects offline; half-states refuse naming the exact line). The job needs the provider's own official OIDC auth action before it (`aws-actions/configure-aws-credentials`, `google-github-actions/auth`, `azure/login`) — keyless cloud auth is the providers' business, not ours — plus `environment:` and `image:` (the built reference; the CLI injects it into the tree's annotated slots). The Action's README carries the input table and the mode-switch story.
4. **Configure repository settings with `gh`** — state-backend configuration as variables or secrets (`gh variable set`, `gh secret set`) for the `PLANTON_BACKEND_*` values the workflow env needs. Deploy-time cloud auth needs NO stored secrets (OIDC above); store only what the state backend genuinely requires.
5. **Verify EVERYTHING before saying "ready"**: the repository is reachable (`gh repo view`), the workflow file is committed and well-formed, the variables/secrets exist (`gh variable list`, `gh secret list`), and the tree passed the forced-offline preflight. State plainly the one thing that remains unproven — the runner's own first contact — and then:
6. **Offer the finale**: push a commit (or `gh workflow run` if the trigger allows) and watch it live — `gh run watch` — narrating the preflight report and the deploys as they land. The job's log shows the report inside a collapsible group with the verdict outside it; the exit-code sentences match the contract above.

**The upgrade story, when the user later connects to a Planton backend**: add `org`, `audience`, `service`, and `image` inputs to the same Action step and drop the `set:` line and state env — the Action's README documents both directions. Nothing else about the workflow changes.

## Walking an offline refusal

Every refusal names its field, its reason, and its fix — relay them in the user's words, never retry around them. The classes: a schema violation (fix the manifest, the report names the field); an external `valueFrom` target (add the producer's manifest to the tree, or deploy connected); a `$var`/`$secret` reference (resolve the value at authoring time, or move the secret to a provider-native reference — the snippets surface has the idiom); a cycle (the chain is printed — break it); an unreachable state backend or invalid credentials (fix the runner's env/OIDC wiring, not the tree); a missing module for a kind (the catalog does not ship that kind's module yet — file the gap).
