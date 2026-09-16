# The self-hosted backup documents name one operator chart floor

**Date**: September 16, 2026
**Type**: Documentation
**Components**: The platform kind's guide (`catalog/kubernetes/kubernetesplantonplatform/GUIDE.md`), the public self-hosting backup page (`site/public/docs/self-hosting/backup-and-recovery.md`)

## Summary

The two documents that tell a self-hosted team which operator chart to pin for their backup now name one number: `0.17.0` or newer. Until now each carried a `0.16.3` floor for the database's backup and a placeholder for the vault's part of the archive -- "named here the day it ships" -- because the chart that carries the bundled vault in the archive had not been released when the words were written. It has: `0.17.0` is the first operator chart on which the vault stores in the platform's PostgreSQL, is sealed by the adopter's cloud key or a Secret they own, and comes back with the records, and it is the chart on which the whole restore has since been proven end to end on a live platform. The placeholder is gone and the two floors are one.

## What Changed

### The guide

`GUIDE.md`, the backup-and-restore section: the sentence that named `0.16.3` and deferred the vault's chart now says the operator that honors all of it -- the database's backup and restore, and the vault's seal, keys Secret, and place in the archive -- is chart `0.17.0` or newer, with what earlier charts fell short on in one clause (the database alone before `0.17.0`; a backup engine that could not finish installing before `0.16.3`), and tells the reader to pin the operator kind's `chartVersion` at or above `0.17.0` wherever a backup is declared.

### The public page

`backup-and-recovery.md`, Requirements: the two bullets (a chart to be named later for the vault; `0.16.3` for the database) become one bullet naming `0.17.0`, the first chart that carries the secrets manager in the archive and the first on which the whole restore -- records and secrets together -- has been proven end to end on a live platform, with the earlier charts' shortfall in one clause.

## Why

An adopter declaring a backup needs one number to pin, not two floors and a history of which chart fixed what. `0.17.0` contains everything `0.16.3` fixed, so the older floor has no reader left; keeping it beside the new one would only make a person wonder which applies to them. The skill coding agents install ships these documents inside its reference pack, so a catalog release carries the sentence to every agent that composes a self-hosted backup from the pack.
