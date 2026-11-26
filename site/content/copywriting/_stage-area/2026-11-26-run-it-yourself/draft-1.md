# Run Planton yourself (draft 1)

This is the stage area for planton.ai's one page, which is built one or two sections per session. This session adds Home's section before the close: in the page's final order it is section 8 of 9 ("run it yourself"), and later sections go in above it. It was written 2026-11-26 under `../../_rules/update-planton-ai-copy-writing.mdc`, and every sentence sits beside the chapter and ledger row it rests on.

The story changed first. The positioning story is now v3.2 (2026-11-26): chapter 8's proof carries the two-command install and the free Community edition. The fact reference replaced "less than 1 day" with the two commands. The claims ledger gained the install row and the Community row, and its self-hosted license row was reworded.

## The media every candidate shares: a real install on a fresh cluster

The docs' two Helm commands ran from ghcr.io on a fresh kind cluster, with both versions pinned. The platform reached Ready and the first visitor set up the administrator. The record (commands, timings, the watch and the first sign-in) is kept in the platform repository's website research.

**Install** (verbatim, as run):

```text
$ helm install planton-operator oci://ghcr.io/plantonhq/charts/planton-operator \
    --version 0.29.0 --namespace planton --create-namespace
$ helm install planton oci://ghcr.io/plantonhq/charts/planton \
    --version 0.4.2 --namespace planton --set platform.spec.version=v0.0.164
```

**Status** (verbatim, the same cluster once Ready):

```text
$ kubectl get plantonplatform -n planton -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,VERSION:.status.version,LICENSE:.status.license
NAME      PHASE   VERSION    LICENSE
planton   Ready   v0.0.164   Community
```

| Rests on | |
|---|---|
| The commands and the status | the walk's record (2026-11-26); the ledger's install row (2026-11-26) |
| Community, free up to the seat ceiling | `pricing.ts` `COMMUNITY_SEAT_LIMIT` (held by `check-displayed-vs-enforced.mjs`); the ledger's Community row; the status's LICENSE column |
| No license key, admin account, database or values file; the first visitor is the administrator | the docs' Install and First sign-in sections; the walk's setup page; the ledger's install row |
| The same product as planton.ai | chapter 8's proof ("the same manifests and the same model run on all three shapes") |

## Candidate A: the same Planton, on your cluster (the lead's recommendation going in)

- **Heading:** Run the same Planton on your own Kubernetes cluster.
- **Line 1:** Two Helm commands install all of it: the console, sign-in, secrets, the databases, CI/CD and infrastructure management. *(ch 8 proof; ledger install row)*
- **Line 2:** There is no license key, admin account, database or values file to set. The first person to open the console becomes its administrator, with a setup code read from your cluster. *(ch 8 proof; ledger install row)*
- **Seat line:** On your own cluster, the Community edition is free for up to {COMMUNITY_SEAT_LIMIT} seats. *(ch 8 proof, ch 12; ledger Community row; `pricing.ts`)*
- **Artifact (right):** the Install panel, then the Status panel.
- **Small print:** It runs on amd64 nodes, with a default StorageClass and about 8 GiB to spare. *(ledger install row, its condition)*
- **Link:** Read the Self-Hosting Docs (`/docs/self-hosting`, `DOORS.selfHostedDocs`).

## Candidate B: A, with the product's first screen

This is candidate A plus the "Welcome to Planton" setup page from the walk, a flat still labelled "The first screen of a fresh install". It is a dark product still on the light page, which the design law allows only in the hero, so it needs the founder's named yes.

## Candidate C: the person first

- **Heading:** On your own cluster, the first person in is the admin.
- **Line 1:** Two Helm commands install the same Planton on your own Kubernetes cluster: the console, sign-in, secrets, the databases, CI/CD and infrastructure management.
- **Line 2:** There is no license key, admin account, database or values file to set. Whoever opens the console first reads a setup code from the cluster and becomes the administrator.
- **Seat line, artifacts, small print and link:** as in A.

## Never said, checked

- No time figure.
- No "air-gapped" and no "zero lock-in".
- The platform isn't called open source.
- No internal component is named.
- Nothing unconfigured is shown: the status uses the custom-columns output, never the default columns, which show Email and Backup.
- No price literal, and no seat count typed.
- No new demo door.
