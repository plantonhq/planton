---
title: "Version History"
sidebar_title: "Versions"
description: "A live, provider-backed timeline — versions added in your cloud console appear immediately, deployments always read the store's latest, and deletion is honest per provider"
icon: history
order: 35
tags:
  - Secrets
  - Versions
  - Audit
---

# Version History

A secret's version history in Planton is **live**: the storage backend is the source of truth, and the timeline you see is the provider's own version list joined with Planton's authorship records. There is no mirror to drift and no sync to schedule.

## Two Kinds of Entries

**Versions written through Planton** carry an authorship record: who wrote them, when, and through which surface. The timeline shows "via Planton by *name*".

**Versions written outside Planton** — a teammate rotating a value in the GCP console, a script calling the AWS API — appear in the timeline the moment anyone looks, marked as added outside Planton, carrying the backend's own creation time. They are not second-class: you can reveal them, roll back to them, and delete them, addressed by the backend's own version identifier.

## Latest Means the Store's Latest

When a deployment, a service pipeline, or `planton secret get --reveal` resolves a secret, "latest" is the **backend's** latest version — including one added out-of-band five seconds ago. Rotating a credential directly in your provider console is a fully supported workflow: the next deployment picks it up with no Planton step in between.

## Version Pinning

Every version is addressed by the backend's own version identifier (a GCP version number, an AWS version id, an Azure version hex, a Vault KV version). Reveals and rollbacks address these coordinates, and the [read-audit trail](/docs/secrets/managing-secrets#the-read-story) pins which version every read actually served — so "latest" reads stay accountable even as the history grows.

## Honest Deletion

Providers genuinely differ on what deleting a single version means, and Planton tells each provider's truth instead of pretending:

| Provider | Deleting a version |
|----------|--------------------|
| GCP Secret Manager | Destroys it permanently; it leaves the timeline |
| Vault / OpenBAO | Destroys it permanently; it leaves the timeline |
| Azure Key Vault | Disables it; it leaves the live timeline |
| AWS Secrets Manager | Not supported by AWS — versions age out naturally as new ones are written; the timeline keeps showing what the store holds |

Deleting the **secret itself** through Planton destroys the remote container in every provider.

## Rollback

`planton secret rollback <slug> --to-version <backend-version-id>` (or the console timeline's rollback action) writes the chosen version's value as a new latest version — history stays append-only, and the rollback itself is an authored, audited write.

## Related Documentation

- [Where Secrets Live](/docs/secrets/where-secrets-live) — Naming, labels, and remote identity
- [Managing Secrets](/docs/secrets/managing-secrets) — Creating, referencing, and auditing secrets
