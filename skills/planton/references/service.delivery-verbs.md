---
title: Deploy, Promote, and Rollback — The Delivery Verbs
description: The three verbs that deliver a version without a build of their own (deploy an image built anywhere, promote a captured pair between environments, roll an environment back), plus tag releases, feature-branch deploys, and standing branch-to-environment mappings. Read when someone wants a version moved between environments, an image they built running, a tag push to be the release, or a refused delivery explained.
---

# Deploy, Promote, and Rollback — The Delivery Verbs

A service's version reaches an environment two ways: a push builds it and deploys it, or someone delivers a version that already exists. This file is about the second way — the three verbs that deploy without a build of their own — plus the two build-and-deliver shapes people ask for by name: releasing by tag, and trying a branch in one environment.

Read this when someone wants a version that runs in one environment to run in another, when something is wrong in an environment and the previous version should come back, when someone has an image built elsewhere (their CI, their laptop) and wants it running, when someone wants a tag push to BE the release or a feature branch running in dev, or when a deployment you started refuses and you need to explain why in terms the person can act on.

## Releasing by tag, and trying a branch

**Tag releases**: with `build.triggers.tags.deploy: true` on the service (optionally narrowed by `tags.patterns` like `["v*"]`), a matching tag push builds and rides the FULL promotion-ordered environment walk — gates, protection, and approvals exactly as a default-branch push. A tag is a release; without the flag, tag pushes build only, and the run says so.

**Branch deploys** (`planton service run <service> --branch <branch> --deploy-env <env>`): builds the branch's head and deploys it into exactly ONE environment — no promotion walk. A branch outside the service's trigger branches NEVER walks the promotion order: without `--deploy-env` it builds and stops, with the reason on the run ("name a deploy environment"). This is the try-my-feature-branch-in-dev shape. An explicit `--commit <sha>` is different — the deliberate release act: it keeps the full walk, and its branch label is empty unless `--branch` states the claim (the record never claims a branch it cannot prove).

**Standing branch-to-environment mappings** (`deploy.branchDeployments` on the service): for teams whose branches ARE their environments — `develop` deploys dev, `release` deploys staging — a mapping makes every PUSH to that branch deploy into exactly its one environment, no manual run needed. Gates, protection, and approvals apply exactly as anywhere else; promotion on this topology happens by MERGING between branches. A branch can be a trigger branch (the full walk) or a mapped branch (one environment), never both — the save refuses the ambiguity. For git-maintained services the mapping is also the record sync's authority: the mapped branch's pushes keep its environment's configuration entry current from its OWN tree, so an overlay that exists only on the dev branch still shows on the record, marked for dev with its branch and commit named.

## What the verbs actually do

**Deploy** (`planton service deploy <service> --env <env> --image <ref>`) takes a container image built ANYWHERE and deploys it into one environment, rendering the manifests from the service's current declared configuration (the record's `deploy.environments` — manually authored or git-synced alike) with the image injected. Optional `--commit`/`--branch` carry the image's provenance so the run's history reads honestly. This is the terminal twin of the external-CI deploy step; with `--follow` it exits non-zero unless the deployment verifiably succeeded (a failed run OR a failed rollout verdict fails the command — an honestly unverifiable one does not).

**Deploy from the working tree** (`planton service deploy <service> --env <env> --image <ref> --from-tree`, run from the repository's project root) is the same verb with the configuration taken from the folder the person is standing in instead of the record: the environment's `_kustomize/overlays/<env>` renders on their machine (the same renderer `service env` uses) and travels with the deploy; the control plane injects the image, the pull login, and the serving hostname exactly as for a synced entry, writes the render onto the record's entry for that ONE environment — stamped with who rendered it, at which commit, and whether the tree carried uncommitted changes — and then runs the deploy through the same gates, verification, URLs, and receipt. Recommend it when a git-maintained service has never been pushed through the build lane (the record's environments are empty, so a plain deploy refuses), or when someone is shaping an overlay and wants to see it run before committing. What to know when recommending it: it accepts only git-maintained services (`deploy.kustomize`; a manually-declared service deploys the record without the flag, or ejects first); it is for a person at a terminal, never a CI workflow (an external-CI credential is refused — the platform cannot verify which branch a checkout came from); the run is named "Deploy … from the working tree" and the receipt carries the tree's provenance; and git stays the stronger writer — the next push to the branch that drives the environment with the same content takes the entry back for git and redeploys nothing, so the record stops saying "working tree" the moment the tree agrees. The CLI prints one line before sending — the overlay, how many manifests, the commit and branch, and whether there are uncommitted changes under the kustomize directory — read it back to the person; "uncommitted changes" means the render is something no commit can reproduce yet.

**Promote** takes the artifact from an existing deployment and deploys it into another environment, together with the configuration that was captured alongside that artifact. It does not rebuild, and it does not read the service's current configuration. What ran in staging is what runs in production — the same image, the same settings, verified together.

**Rollback** puts an environment's previous deployment back by re-applying that deployment's own artifact and configuration exactly as they ran. With no target named it restores the immediately preceding deployment, so rolling back twice returns to where you started. Naming a specific past deployment restores exactly that one.

All three answer with a **run**, not a deployment record. A deployment takes time, can pause for an approval, and can fail — so what comes back is something to follow. The deployment record appears when the environment succeeds.

Promote and rollback are also clickable: the console's service page carries them on its Deployments tab (each environment row offers Promote and Roll Back; a past deployment in the history offers Roll Back to This), confirming into the same run page. When a person would rather click than type — or wants to SEE what a rollback restores before committing to it — pointing them there is a complete answer.

Deploy differs from promote in what configuration ships: deploy renders the service's CURRENT declaration (today's configuration, the named image); promote re-applies the CAPTURED configuration (the tested pair). When a version already runs in another environment, promote is the right verb; deploy is for versions Planton has never deployed.

## Why they are exact, and why that matters when explaining them

Every run captures the resolved manifests for each environment it planned, and every deployment record additionally carries the manifests it applied. Those captures are what make these verbs trustworthy: the platform is re-applying bytes that already ran rather than re-deriving them. Practical consequences worth stating to a person:

- A rollback during an incident restores a state that genuinely existed. It cannot produce old code paired with configuration someone changed last week.
- Neither verb needs the repository to be reachable or a build system to be healthy.
- A promotion moves a tested pair. If the configuration for the target environment changed after the artifact was built, promotion still deploys the tested configuration — that is the promise, not a limitation to apologize for.

If someone wants the same version with *current* configuration, that is a different intent. Say so plainly: the honest path today is a new build (for services whose configuration lives in git, a configuration change IS a commit, so pushing is the natural route).

## Protection is never bypassed

Promoting into a protected environment still stops at that environment's approval gate, and the person who promoted cannot be the person who approves. Report the gate and who can resolve it. Never attempt an approval on someone's behalf — approval is a human decision, and the assistant holds no approval rights anywhere.

## Naming the source of a promotion

Promotion needs to know which deployment moves. Name it one of two ways, never both: the deployment record itself, or the service plus the environment whose current deployment should move. Ambiguity is refused rather than guessed at, because guessing here deploys the wrong version.

Before promoting, read the delivery history for the source environment and confirm which version is about to move. Say the version and the commit out loud in your answer — the person is about to change what production runs.

## The refusals, and what each one means

Each refusal names a real next step. Relay it rather than retrying:

- **The target environment was never prepared for this artifact.** The run that produced this deployment did not resolve configuration for that environment, so there is nothing tested to promote. The fix is to add the environment to the service's deploy environments and push once; that build prepares it.
- **A deployment to that environment is already waiting for approval.** Someone needs to decide on the run that is already there. Starting a second deployment would leave the first waiting forever, so the answer is to surface the pending gate.
- **A deployment to that environment is already in progress.** Wait for it, and follow it.
- **Deployments are turned off for this service.** The switch is deliberate — either the environment is still being prepared, or something outside Planton owns this service's deployment. Do not work around it.
- **The deployment names no artifact**, or **did not record the configuration it applied.** There is nothing to redeploy exactly. Deploy the version you want from a build instead.
- **Promoting into the environment it already runs in.** That is a rollback in place, not a promotion.
- **No push has rendered the kustomize tree onto the record yet** (deploy verb only). A git-maintained service's configuration lands on the record when a branch that drives an environment pushes; until the first such push there is nothing to render from — push to a driving branch (the promotion branch or a mapped one), or promote an existing deployment.
- **The environment is not among the service's declared deploy environments** (deploy verb only). The refusal names the declared set; add the environment to the service's deploy declaration and apply it first.
- **The target is a preview environment** (all three verbs). A preview (`{service}-pr-{n}`) deploys only from its own pull request's pushes — push to the pull request to update its preview, or close it to tear the preview down; deliveries target durable environments. The mirror case refuses too: **promoting FROM a preview's deployment** answers "merge the pull request" (the merge commit builds and deploys through the normal lane), or deploy its image directly with the deploy verb. See `preview-environments.md`.
- **The deployment was observed from a GitOps system** (trigger provenance `gitops`). Planton did not perform it — the record is a report of another system's deploy, not a restorable fact — so promote and rollback both refuse, including when "roll back one" lands on an observed record. The real path is the git repository that system watches: the change belongs there, and that system will deliver it. Never advise working around the fence.

## How a delivery run looks while it runs

It is an ordinary run with no build stage: one environment, and the deploy work inside it. Follow it exactly as you would a push-triggered run. If it is parked at an approval gate, the run reports itself as awaiting approval — that is the truth to relay, along with the fact that nothing is executing until a person decides.
