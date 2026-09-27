---
title: Push-to-Register — the service.yaml in the Repository IS the Registration
description: A connected repository registers and maintains its own Service record by carrying a service.yaml on its default branch — the file is the opt-in, content decides ownership, the record follows the source of truth. Read when someone asks how to register from a repository, why a pushed manifest did or didn't land, what manifest-sync status means, or whether pushing a manifest redeploys anything (it never does).
---

# Push-to-Register — the `service.yaml` in the Repository IS the Registration

A repository connected to a Planton org registers and maintains its own Service record by carrying a `service.yaml`: push the file to the default branch, the service appears in the catalog; change the file, the record follows; the same push can start the service's first build. Read this when someone asks how to register a service from a repository, why their pushed `service.yaml` did or didn't land, what the manifest-sync status on a service means, or whether pushing a manifest will redeploy anything (it won't).

## What to teach, and in what order

1. **The file is the opt-in.** No setting enables this lane — the connected GitHub connection is the trust anchor (connecting it was an org-admin act: installing a GitHub App, or recording the sign-in the machine running Planton holds), and committing a Planton `service.yaml` is the declaration of intent. With an App, GitHub delivers the push; with a machine's sign-in, Planton's repository watch notices the commit and hands it to the same door. The manifest grammar IS the Service resource's YAML: the same document the console shows, the API accepts, and `planton service register` submits.
2. **Default branch only.** Manifests apply from the repository's default branch. A `service.yaml` on a feature branch changes nothing until it merges — the record follows the source of truth, not proposals.
3. **The file needn't repeat what its location proves.** `metadata.org` and the repository identity under `spec.gitRepo` (connection, owner, name, default branch) may be omitted — the platform fills them from the push. If they ARE present and contradict the push, the manifest refuses with both sides named: a pushed manifest may only declare the repository it lives in, in its connection's own org.
4. **Content decides ownership.** Knative and others also name files `service.yaml`. Only a file carrying `apiVersion: service-hub.planton.ai/v1alpha1` and `kind: Service` is Planton's; anything else is ignored quietly. Never tell a user their Knative file "failed" — it was never read as ours.
5. **The record only, never deployments.** A pushed manifest changes what the service DECLARES. Builds still start from the trigger rules; deploys still arrive through the pipeline and the delivery verbs. Changing `spec.deploy` by push redeploys nothing.
6. **Deleting the file never deletes the service.** The record outlives the file; deleting a service stays an explicit act (`planton delete`).

## Reading the manifest-sync status

Every push outcome ALSO lands on the commit itself: each candidate `service.yaml` gets its own GitHub check named by the file path — green when the record matches the file, a red X carrying the exact error otherwise, with the error pinned as an annotation on its line in the diff when the parser names one. The commit check is the ONLY surface for a file so broken it names no service (not even parseable YAML — there is no record to stamp), so when someone says "my push did nothing", the commit page answers even when `manifestSync` cannot. A commit page with no check on a pushed manifest usually means the GitHub App installation lacks the 'Checks: Read and write' permission.

Every push outcome lands on the service record's `status.manifestSync` (visible via `planton get service <slug>`):

- `lastApplied` — the last push that landed: commit, file, time. "Landed" includes CONVERGED: a manifest matching the record exactly writes nothing (redeliveries and no-op pushes never churn history) but still stamps success.
- `lastFailed` — the most recent push whose manifest failed: commit, file, time, and the exact error. A populated `lastFailed` always means the repository's manifest is currently diverged from the record — it is CLEARED by the next success.

When walking a failure, dispatch on the error text:

- **A named unknown field** (e.g. `spec.buidl`) — a typo against the schema; fix the field. The pre-push guard is `planton validate -f service.yaml` — the same validation, before the push.
- **"declares repository X but was pushed to Y"** — the manifest names another repository; a pushed manifest may only declare its own. To register a service for another repository, commit the manifest there.
- **"declares organization X but this repository's connection belongs to Y"** — the org field contradicts the connection; omit it or fix it.
- **"belongs to repository X but this manifest was pushed to Y"** — the manifest's name collides with an EXISTING service owned by another repository; rename this manifest's service. (A name colliding with an existing service that declares NO repository is not an error — the push adopts it, which is the console-stub-then-manifest onboarding path.)

## What has no record to show

A malformed manifest for a service that does not exist yet has no record to stamp — the failure is in the platform's logs only today. If a user says "I pushed a service.yaml and nothing happened", check in order: is the push on the default branch; is the apiVersion/kind Planton's; does `planton get service <slug>` find the service (it may have landed under a slug derived from the name); and only then suspect a malformed first manifest — have them run `planton validate -f service.yaml` locally, which reproduces the same refusal with the field named.

## The register-and-first-build moment

A push that adds a valid `service.yaml` AND matches the new service's build triggers (by default: any push to the default branch once a build block exists) births the first pipeline run from that same push. When someone registers by push and asks "did it build?", check `planton service pipelines` for the service — the run's trigger reads as the git push itself.
