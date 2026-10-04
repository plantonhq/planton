# The operator's backup plugin can actually install, and a declared restore is never silently an empty database

**Date**: September 14, 2026
**Type**: Bug Fix
**Components**: planton-operator (the PostgreSQL component's backup arm, the manager ClusterRole, the Helm chart's RBAC)

## Summary

The first live run of a self-hosted platform's backup and restore found four defects in one afternoon, each of which alone would have kept an adopter's database unprotected or, worse, made a restore look like data loss. The operator's ClusterRole never granted the verbs on cert-manager `issuers`, so the Barman Cloud plugin install applied eleven objects and was refused on the twelfth, the self-signed Issuer its two Certificates reference; the Certificates never became ready and the plugin pod waited on a TLS Secret nobody would mint. The install error was swallowed into one status pass and then overwritten, so the platform said `Deploying` for as long as anyone watched. When that error happened on a fresh install with `recoverFrom` declared, the operator created the database anyway, empty, where the archive's copy was expected. And the base-backup schedule's immediate first run was requested before the archiving sidecar had attached, failed with "plugin not available", and nothing retried it until the next scheduled hour, while `status.backup` read Cluster fields CloudNativePG 1.30 leaves empty for plugin-driven backups and so reported `Failing` over an archive that was complete and recoverable.

## What Changed

- **RBAC**: the manager ClusterRole grants `get, list, watch, create, update, patch, delete` on `issuers.cert-manager.io`, declared by a kubebuilder marker beside the plugin's sub-operator definition and regenerated into `config/rbac/role.yaml` and the chart's `rbac/manager-role.yaml`.
- **The hold**: when the plugin cannot be installed and no database exists yet, the PostgreSQL component holds the Cluster and says so (`Unavailable`: "the backup plugin could not be installed: <the cluster's refusal>; holding the database so it is not created empty"). A running database keeps today's behavior: it runs without backups and the column says why.
- **The schedule waits for the sidecar**: the `ScheduledBackup` is rendered only once the Cluster's `status.pluginStatus` names the plugin, so its immediate first base backup lands against a live archiver.
- **The truth of `status.backup`**: the first recoverability point and the last successful base backup are read from the plugin's own record, `ObjectStore.status.serverRecoveryWindow[<serverName>]`, falling back to the Cluster's fields; a failed run that a later success superseded no longer reads as `Failing`.
- The operator README's backup paragraph says all four.

## Proof

Unit tests on the component package: a refused Issuer apply (an interceptor on the fake client) holds a fresh install with the sentence and spares a running database; the schedule is absent until the Cluster reports the plugin and present after; the ObjectStore's window for this server wins over empty Cluster fields and over another server's window in the same store. Found live on a GKE management instance: the plugin stuck `ContainerCreating` for 27 hours on `secret "barman-cloud-client-tls" not found`, the one status pass that carried `cannot patch resource "issuers"`, a `recoverFrom` install that came back with `bootstrap.initdb` and a new server name, and an ObjectStore whose recovery window named a complete base backup while the platform said the last one failed.
