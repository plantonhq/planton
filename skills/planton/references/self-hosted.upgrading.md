---
title: Upgrading a Self-Hosted Planton — The Operator First, Then the Platform
description: The order and the sentences of an upgrade: helm upgrade of the operator chart carrying the definitions (and the preflight that names the two adoption commands for an install from before chart 0.8.0), the version-floor refusal, moving spec.version, and the one boundary a preview-era install cannot cross. Read when a person asks to upgrade their operator or platform, when an upgrade stopped with a sentence, or when a platform created by an older operator behaves oddly under a newer one.
---

# Upgrading a Self-Hosted Planton

Read this when an adopter wants a newer operator or platform, or when an upgrade stopped and printed something. The people-facing steps are `site/public/docs/self-hosting/index.md` ("Upgrades and uninstall") and the operator chart's README ("CRD Management"); this file carries the order, the sentences you will meet, and the boundary.

## The doctrine in one sentence

The operator is upgraded first and the platform second, each by changing one declaration -- the chart release, then `spec.version` -- and every refusal along the way is a sentence that names what to change.

## Step 1 -- the operator

```bash
helm upgrade planton-operator oci://ghcr.io/plantonhq/charts/planton-operator \
  --version <chart-version> -n <operator-namespace>
kubectl -n <operator-namespace> rollout status deploy/planton-operator
```

The chart owns the two definitions (`PlantonPlatform`, `PlantonIdentityProvider`) as release resources: `helm upgrade` upgrades them, `helm rollback` rolls them back with the operator, and `helm uninstall` keeps them by default (deleting a definition deletes every platform of that kind). There is no separate "apply the CRDs" step.

One case has a preflight. An install from before chart 0.8.0 got its definition from the chart's old `crds/` directory, which Helm does not own, and Helm refuses to take over a resource it did not create. The chart says so, with the fix:

> observed: CustomResourceDefinition plantonplatforms.planton.ai exists on this cluster but belongs to no Helm release (it was installed once from an earlier chart's crds/ directory, or applied by hand)
> meaning: this chart owns its definitions as release resources so upgrades carry the schema, and Helm will not take over a resource it did not create
> next step: adopt the definition into release planton-operator, then re-run:
>   kubectl label crd plantonplatforms.planton.ai app.kubernetes.io/managed-by=Helm
>   kubectl annotate crd plantonplatforms.planton.ai meta.helm.sh/release-name=planton-operator meta.helm.sh/release-namespace=planton-operator

Run the two printed commands (with the release name and namespace the preflight prints, which are the adopter's), then the same `helm upgrade` again. The `PlantonIdentityProvider` definition appears with the upgrade; nothing else changes until the platform moves.

The operator upgrade leaves the running platform untouched: its pods do not restart, its realm is not touched, and the platform stays `Ready` throughout. (Proven on a production cluster, 0.17.0 to 0.18.0.)

## Step 2 -- the platform

After the operator is up, it judges `spec.version` against its floor. A platform below the floor is refused in words and left running exactly as it was:

> spec.version vX is older than the oldest platform release this operator runs (vY); the settings and images this operator expects belong to vY and newer. Nothing running was changed. Set spec.version to vY or newer …

Move the version the way the platform was declared -- `helm upgrade planton oci://ghcr.io/plantonhq/charts/planton -n <ns> --set platform.spec.version=<release>` when the `planton` chart declared it, `kubectl -n <ns> patch plantonplatform <name> --type merge -p '{"spec":{"version":"<release>"}}'` when a manifest did, or the catalog kind's field when infrastructure as code did. Then watch:

```bash
kubectl -n <ns> get plantonplatform <name> -w
```

Every versioned component rolls to the release's images; the platform reads `Ready` only when every rollout has finished on the new template. Ten minutes on a fresh cluster is normal (images pull once).

Both steps are mutations under the protocol in `cloud.exploration.md`: state the command and what it changes, get one clear yes for each step, report what happened.

## The boundary: installs from before chart 0.8.0

The operator moved its PostgreSQL from a StatefulSet to CloudNativePG when it moved into the open-source repository (chart 0.8.0). A platform created by chart 0.7.0 or older -- the `-selfhosted-preview` era -- keeps its data in the old StatefulSet's volume; an operator from 0.8.0 onward creates a new CloudNativePG database beside it and starts the platform on the empty one, so the realm is recreated rather than repaired and the old records are not carried across. The old volume is not deleted, but nothing reads it.

Say this plainly when you meet such an install: the version-floor sentence tells the adopter to raise the version and does not mention the data. The honest path is to treat it as a new install -- export what matters from the old database first (the identity realm and the records), stand the new platform up, and re-declare -- rather than an upgrade in place. Installs from chart 0.8.0 onward upgrade in place with their data intact, which is the path the docs describe.

## What was verified, and what to confirm on your install

Verified on running clusters: the operator upgrade on a production cluster with the platform untouched; the two-step from chart 0.7.0 to 0.18.0 on a fresh cluster (the preflight, the adoption commands, the floor refusal, the platform reaching `Ready` on the new release); a component the newer operator introduced deploying on the upgraded platform. **Not verified: a live front-door change on a running install** -- moving `spec.ingress.hostname` (or switching doors) and then signing in at the new address without touching the identity server by hand. The operator repairs the realm's redirect URIs for a moved front door, and that repair is proven in its own test suite, but no session watched a browser complete the round trip. If an adopter changes their front door and sign-in at the new address fails, read the operator's log for the `Realm repaired` line first; if the repair is missing or sign-in still fails after it, file it on the open-source repository with the before-and-after `spec.ingress` and the sign-in error (`craft.filing-platform-gaps.md`).

## What never to do

- Never upgrade the platform's version before the operator that runs it; the floor is judged by the operator, and an older operator does not know a newer release's shape.
- Never `kubectl apply` a definition (CRD) by hand on a chart-managed install; the chart owns them, and a hand-applied one is exactly the ownership conflict the preflight exists to catch.
- Never present a preview-era install's upgrade as data-preserving.
