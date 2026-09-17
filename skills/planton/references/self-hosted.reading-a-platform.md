---
title: Reading a Self-Hosted Platform — What the Operator Tells You Before You Touch Anything
description: How to read a self-hosted Planton (a PlantonPlatform reconciled by the Planton operator) from its own status, what each column and condition means, the three shapes a component reports, and the one law of changing it. Read when a person's Planton is self-hosted (deployment kind self_hosted, a PlantonPlatform in their cluster), when they ask why their platform is not Ready, what a status sentence means, or before proposing any change to the install.
---

# Reading a Self-Hosted Platform

Read this when the Planton you are helping with runs on the person's own cluster. You will know because `planton instance show` reads `deployment_kind: self_hosted`, or because they say "our Planton", "the operator", or "the platform resource". Installing it is documented for people at `site/public/docs/self-hosting/index.md` in the open-source repository (two Helm charts: `planton-operator` first, then `planton` declaring the platform; or the catalog kinds `KubernetesPlantonOperator` and `KubernetesPlantonPlatform` through the infrastructure craft this skill already has). This file is the agent's half: how to read what is there.

## The doctrine in one sentence

A self-hosted Planton is one declaration -- a `PlantonPlatform` resource -- that the operator converges everything from, so every fact about the install is on that resource, and every change to the install is an edit to that declaration, never to the objects the operator renders from it.

## Read the platform first, every time

```bash
kubectl get plantonplatform -A
kubectl -n <namespace> get plantonplatform <name> -o yaml
```

The columns are the platform's own words:

- **`PHASE`** -- `Pending`, `Deploying`, `Ready`, or `Error`. `Ready` means every component's rollout has finished on the current template; a version change never reads `Ready` while the previous release still serves.
- **`MESSAGE`** -- the `Ready` condition's message. It speaks for the worst-off component in that component's own words, prefixed by the component's key (`controlPlane: Waiting for dependency: openfga`). Relay it; it already names the object and the cause.
- **`VERSION`**, **`URL`**, **`REACHABILITY`** -- the release the platform runs, the public origin it serves at, and whether the operator concluded the public internet reaches that front door (`public` | `private`; `auto` while it decides). The URL is the identity issuer's origin too, which is why a front door is never changed casually (`self-hosted.front-doors-and-the-cli.md`).
- **`LICENSE`**, **`EMAIL`**, **`GITHUB`**, **`BACKUP`** -- configuration echoes, never delivery verdicts. `LICENSE` says how a key is delivered (`Community`, `InlineKey`, `SecretRef`), not whether it verified -- the live license state is the control plane's own entitlements, read by the console's License page. `EMAIL` says which provider arm is declared (`NotConfigured` | `SMTP` | `Resend`), not whether mail arrives. `GITHUB` lists the declared hosts, with `(App)` where an install-wide App is registered. `BACKUP` reads the archive's own state (`NotConfigured`, `Deploying`, `Unavailable`, `Healthy`, `Failing`).
- **`VersionSupported`** (a condition) -- whether `spec.version` names a release this operator runs. The operator runs releases from a floor upward; its first log line (`Platform version floor`) names the floor. A version below it is refused before anything is created, in words: *"spec.version vX is older than the oldest platform release this operator runs (vY); the settings and images this operator expects belong to vY and newer. Nothing running was changed. Set spec.version to vY or newer …"*. A platform already running is left exactly as it was.

`spec.version` is a contract: it names the release whose settings and images the operator expects. To run a custom build, keep `spec.version` at a release and set `image.tag` on the component -- the version names the contract, the tag names the bytes.

## The three shapes a component reports

Every component under `status.components` carries a phase, a one-word `reason`, the `object` the reason is about, a sentence, and when the condition began. Read them as three shapes:

1. **Ready** -- the rollout finished; the sentence is what the component offers (the identity component's says where the first admin's credentials are, or that a restored realm keeps its existing users).
2. **Not ready, explained** -- `Pending` or `Deploying` with a reason such as `WaitingForDependency`, a pull failure, a missing Secret key, a crash loop, an unprovisionable volume claim, or the calm "running, not yet answering its health check". The operator read the pods, the claims, the workload conditions, and the namespace Events and classified them in order of certainty; you do not need to repeat that read to explain it.
3. **Refused** -- the declaration cannot be honored as written (a seal whose Secret lacks its key, a contradictory `spec.email` block, a missing referenced Secret). Nothing deploys until the named field changes. This is never an outage; it is a sentence to relay with the field it names.

Failure reasons become a Warning Event on the platform when a component enters them, once, and a Normal Event when it recovers -- `kubectl -n <ns> get events --field-selector involvedObject.kind=PlantonPlatform` is the timeline when the person asks "when did this start".

## What the console already shows

Settings pages on the self-hosted console render the same facts as sentences: License (the entitlements and seats), Email (the declaration or the two setup hints when nothing is declared), Directory (the identity manifest's verdicts and the live checks, `self-hosted.identity-connecting.md`). When the person is looking at the console, read the page they are on before reading the cluster; the words match.

## Changing the install

Every change is an edit to the declaration -- `spec.version`, `spec.ingress`, `spec.email`, `spec.database.postgresql.backup`, `spec.vault`, `spec.github`, the `PlantonIdentityProvider` beside it -- applied with `kubectl apply` (or through the `planton` Helm chart's values, or the catalog kind, whichever installed it). The operator reconciles within a pass (about thirty seconds) and reports each component again. An apply is a mutation under the protocol in `cloud.exploration.md`: the exact command and its blast radius in one sentence, one clear yes, then the report of what happened.

Before a change, the reading above is the ground: never propose an edit to a platform you have not read, and never guess a reason the status already names.

## What never to do

- **Never scale, edit, or delete an operator-rendered object** -- a Deployment, Service, ConfigMap, or Secret the operator owns. The operator restores it within a pass (a console scaled to zero came back in thirty seconds, observed live); the change you meant belongs on the declaration.
- **Never change realm state in the identity server's admin console** for anything the operator owns (clients, mappers, the federation component, the directory group mirror, the browser flow's redirector). It converges them back; state an admin created outside that owned set is left alone, which is why the distinction matters.
- **Never touch a Secret the operator wrote** (the bootstrap admin, the vault's init Secret, the setup code) except to read it when a journey needs the value.
- **Never trust a green pod over a red status.** A running pod that the platform reports not ready is not ready; the status knows about rollouts, health checks, and dependencies the pod list does not show.
