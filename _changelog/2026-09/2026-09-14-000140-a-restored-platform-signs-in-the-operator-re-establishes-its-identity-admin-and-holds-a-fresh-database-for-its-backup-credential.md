# A restored platform signs in: the operator re-establishes its identity admin, and holds a fresh database for its backup credential

**Date**: September 14, 2026
**Type**: Feature
**Components**: Operator (identity, PostgreSQL backup components, Keycloak admin client), PlantonPlatform CRD status

## Summary

Two gaps in the operator's recovery path, both found by reading the code against the restore a platform actually goes through. A platform declared with `spec.database.postgresql.recoverFrom` came back with its database but nobody could sign in and the operator could never finish: the restored identity realm carries the master admin password its source platform had, the new install generated a different one, and every reconcile ended in "realm could not be reconciled: 401". The operator now re-establishes its admin on a restored realm through Keycloak's own recovery command, exactly once, with each step named in status. And a fresh install with a backup declared whose credentials Secret had not landed yet was created without archiving and marked `Failing`, so attaching the archive later cost a restart; it is now held the way it is already held for the backup plugin and for a recovery source, so the database is born archiving.

## Problem Statement / Motivation

Keycloak reads `KC_BOOTSTRAP_ADMIN_*` into an empty master realm only. The operator's bootstrap admin Secret lives in the platform's namespace and dies with it; the realm lives in the database and comes back from the archive. After a restore the two disagree forever. Keycloak's documented answer, `kc.sh bootstrap-admin user`, creates a new temporary admin and requires every server node to be stopped while it runs, so it cannot be a retry inside a running reconcile; it has to happen before the identity server starts, and it cannot recreate `admin`, only a second admin that then resets `admin` and is removed.

### Pain Points

- A restored platform's status said `401` every thirty seconds with no way out but a person running the recovery command by hand, against a server they first had to stop.
- The three holds in the backup arm were asymmetric: the plugin and a recovery source held a fresh database, a missing backup credential did not.
- The recovery-source hold had no test.

## What Changed

### The database tells the truth about where it came from

- `status.backup.restoredFrom` (new, `operator/api/v1/postgresql_backup_types.go`): the source archive's server name when the database was bootstrapped from an archive, read every pass from the live Cluster's `externalClusters` entry (`clusterRecoverySource` in `internal/component/postgresql_backup.go`, before the plugin gate, so a recovery-only database carries it too). Empty for a database created empty. The identity component keys on this fact, not on the spec: `recoverFrom` added later to a running platform is ignored by design and must not trigger anything.
- CRD regenerated (`operator/config/crd/bases`, the chart's `templates/crds` copy). The ClusterRole is unchanged: the janitor's marker already grants `delete` on Secrets.

### The identity component re-establishes its admin on a restored realm (`internal/component/identity_recovery.go`)

- **Pending** while `status.backup.restoredFrom` is set and the marker ConfigMap `<platform>-identity-recovery` is absent.
- **Before the identity server** (the Deployment does not exist): a recovery credential Secret `<platform>-identity-recovery-admin` is generated and a one-shot Job of the same name runs `kc.sh bootstrap-admin user --username planton-recovery --password:env KC_BOOTSTRAP_ADMIN_PASSWORD` on the identity image with the Deployment's own database env (`resources.IdentityRecoveryAdminJob`; `identityDatabaseEnv` and `identityImage` extracted so the two can never drift; no `ensure-database` step, the restored database exists; `backoffLimit` 2, `activeDeadlineSeconds` 600). The status reads "Re-establishing the identity server's admin credential on the restored realm (job ...) before the server starts"; a failed Job names its condition and the `kubectl logs` command; only a succeeded Job lets the Deployment be created. A server already running (an operator upgraded onto a restored platform whose admin it could never reach) is stopped first, because the command refuses to run beside a live node.
- **After the identity server answers**: the real admin's own sign-in is the probe. If it works, completion is recorded. If it is refused (401, typed as `keycloak.TokenRefusedError`), `keycloak.ReestablishAdmin` signs in as `planton-recovery`, resets `admin`'s password to this install's, and deletes the recovery admin, in that order. On success the marker ConfigMap is written and the Job and recovery Secret removed. A realm with no `admin` is refused in words that name the recovery credential and the admin console; a refused recovery credential names the Job's log; a server not answering is left to the realm reconciler's own sentence.
- Three new `AdminClient` methods (`FindUserByUsername`, `ResetUserPassword`, `DeleteUser`) beside the existing user reads; `MasterRealm`, `IsCredentialRefused`, `IsAdminNotFound` for the callers.

### The backup arm holds a fresh database for its credential (`internal/component/postgresql_backup.go`)

- When the backup's credentials Secret preflight returns a sentence and no Cluster exists yet: `holdCluster`, `Deploying`, "Waiting for the backup credentials before creating the database, so it is born archiving: <the preflight's sentence>". A running database keeps today's `Failing`.

### Documentation

- `operator/README.md`: the `spec.database.postgresql.backup` / `recoverFrom` paragraph the "Declaring a Platform" section never had, in the section's voice; the package map names `recovery.go`.
- `helm/planton-operator/README.md`: the restore sentence and the born-archiving hold.

## How to check

- `go test ./internal/component/` (from `operator/`): the fresh-install credential hold and its two shapes; the running-database `Failing` contract; the recovery-source hold (previously untested); `restoredFrom` read from a restored Cluster and empty for a plain one; the identity recovery in eleven cases (pending only for a restored database without the marker; the credential and Job created and waited on; a failed Job explained with its log; a succeeded Job letting the server start; a running server stopped first; the admin re-established and the artifacts cleaned; an admin that already works only recorded; a refused admin with no recovery in flight restarting the server; no admin in the realm refused in words; a refused recovery credential naming the Job's log; a server not answering left to the realm reconciler).
- `go test ./internal/resources/`: the Job's shape (identity image, database env byte-identical to the Deployment's, the command, the password from the recovery Secret, one-shot policy).
- `go test ./internal/keycloak/`: the reset sequence against an `httptest` master realm (token as recovery, find admin, reset permanent, find self, delete self), the typed refusals.
- `go test -tags=requires_docker ./internal/keycloak/ -run TestRecovery_ReestablishAdminOnARestoredRealm`: the whole seam on real Keycloak 26.3 in Docker -- master realm bootstrapped with admin/A on a named volume, server stopped, the recovery command run on the same data creating `planton-recovery`, server started, `ReestablishAdmin` giving admin this install's password B and removing the recovery admin; admin/B signs in, admin/A is refused, the recovery credential is dead. Passed in 29 seconds.
- `make manifests generate` in `operator/`; `go vet` on the three packages.

## Not in this change

- No live run on a cluster; the first platform restored with this operator is the proof of the Job against a real CloudNativePG-restored database. The one precondition it rests on -- CloudNativePG reconciling the superuser password to the new `-superuser` Secret on a restored cluster -- is the same one the control plane, OpenFGA, and Temporal already rest on.
