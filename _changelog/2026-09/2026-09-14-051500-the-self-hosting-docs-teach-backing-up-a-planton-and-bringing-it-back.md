# The self-hosting docs teach backing up a Planton and bringing it back

**Date**: September 14, 2026
**Type**: Documentation
**Components**: planton.ai docs (Self-Hosting section)

## Summary

A new page, [Backup and Recovery](https://planton.ai/docs/self-hosting/backup-and-recovery), sits beside the Self-Hosting index. It is the page an adopter follows to back up their self-hosted Planton's database to an object store they own and to bring the whole platform back from it, on the same cluster or a new one. It teaches what the backup holds and what it does not (the secrets manager), shows the four-resource declaration on Cloudflare R2 with nothing typed (bucket, bucket-scoped token, the platform's `backup` block referencing both), gives one sentence each for S3, GCS, and Azure Blob, explains the `BACKUP` column and how to read and keep the archive's server name, and walks the restore: declare again with `recover_from`, what comes back, that people sign in with their existing passwords and the operator re-establishes the admin, what is re-entered, and the two rules the operator holds (first creation only; nothing is ever destroyed to honor a declaration).

## Problem Statement / Motivation

The catalog kind gained the declaration (release `v0.5.56`) and the operator gained the whole recovery path (chart `0.16.0`), but the one place an adopter looks first, planton.ai/docs/self-hosting, had a single page and said nothing about backups. A feature a person cannot find is not shipped. The catalog kind's GUIDE and preset carry the operator's reasoning for the record; the docs page is written for the person doing it.

## What Changed

- `site/public/docs/self-hosting/backup-and-recovery.md` (order 25, the section's first content page). Frontmatter per the docs style rule; YAML the adopter authors; absolute links; no protobuf names in prose. Links the catalog kind's GUIDE and its backups-to-R2 preset on GitHub. Screenshots of the console's Backups step are deferred to the platform release that carries it, and the page says so.
- The site deploys on push to `main`, so the page is public with this commit.

## How to check

Open https://planton.ai/docs/self-hosting/backup-and-recovery after the site lane finishes; the sidebar shows it under Self-Hosting. Every command and field name on the page was read from the operator source, the catalog kind's spec, and the published chart (`helm show chart oci://ghcr.io/plantonhq/charts/planton-operator` reports `0.16.0`).

## Not in this change

- Console screenshots (the platform release that ships the Backups step has not been cut).
- The docs style rule's section-order table gives Self-Hosting and Security the same `order: 70`; flagged, not fixed, in the project's records.
