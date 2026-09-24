# The bundled vault's restore lanes run green on a real cluster

**Date**: September 16, 2026
**Type**: Testing
**Components**: The Planton operator's Kind suite (`operator/test/e2e`), the operator's Makefile and README, the operator's CI workflows

## Summary

The two vault restore lanes ran on a real Kind cluster for the first time and passed, seventeen specs in twenty-two minutes: a team on the built-in seal with a keys Secret they own, and a team on a transit seal against an in-cluster key holder, each installed, archived, destroyed, and declared again from the archive, with the secrets written before the disaster read back through the control plane's own token and a signature verified on the restored OIDC signing key. The operator's source did not change: every red spec on the way was the lane expecting the wrong thing about a true behavior, and each is corrected with the behavior recorded beside it. The Kind suites now run on a kubeconfig of their own, and the lanes have a post-merge workflow of their own.

## What Changed

### The lanes, corrected by the run

`operator/test/e2e/vault_restore_test.go`: a token minted through a token role carries no `period` of its own -- the server keeps the period on the role and reads it there at every renewal -- so the lane asserts what `token lookup` truly reports: the token is orphan, renewable, names its role, and its creation TTL equals the role's seven-day period. The control plane's Deployment is waited for rather than read the moment the vault is Ready, because the operator renders it only once its own dependencies (the identity server among them) are Ready. The transit lane's declaration is built in its `BeforeAll`, never at tree construction: `runID` does not exist until the suite's `BeforeAll` runs, and a prefix built earlier read `/transit`, so the source archived under a path the restore never looked at.

`operator/test/e2e/vault_restore_helpers_test.go`: the one read that lists a Secret's data keys goes through a Go template; kubectl's jsonpath can walk a map's values but not its keys.

### The Kind suites never read your kubeconfig

`operator/Makefile`: `E2E_KUBECONFIG` and `E2E_CHART_KUBECONFIG` -- each Kind cluster's credentials in a file of their own under `bin/`, written by `kind create cluster`, passed to the test process as `KUBECONFIG` (which every `kubectl`, `make install`, `make deploy`, and Helm SDK call the suites make inherits), and removed by `kind delete cluster`. A suite creates and deletes namespaces, deletes PlantonPlatforms, and tears its cluster down; on a machine whose current context is a live cluster those commands must have nowhere else to go. `README.md` says so under Development.

### The lanes' CI home

`.github/workflows/e2e-operator-vault-restore.yaml` (new): the lanes run post-merge -- on push to `main` touching the operator or its two charts, on a weekly schedule, and on dispatch -- never on a pull request. Its own workflow because `e2e-operator.yaml` cancels its in-progress run on every push (right for a PR gate, wrong for an hour of proof; this one queues) and skips the cert-manager install the lanes' backup plugin needs. Ninety-minute timeout against a twenty-two-minute run on a laptop that emulates the control plane's image. `e2e-operator.yaml`'s header points at it; the Makefile's target comment says where CI runs it.

## Why

A proof surface that has never run proves nothing; this one now has. What the run settled that no offline test could: the cluster's real token review accepts the operator's sign-in with no repair; the real control plane starts against the minted token and creates the platform's signing key; the archive carries the vault's tables through a base backup and the WAL after it; a restored built-in-seal vault refuses in the operator's words when its Secret is missing and unseals from the recreated one; a restored transit-sealed vault opens itself; the same signing key verifies a signature made before the disaster; and a crash-looping seal is named, with the log line to read, in under two minutes.
