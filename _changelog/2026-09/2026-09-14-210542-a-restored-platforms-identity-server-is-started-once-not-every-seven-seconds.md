# A restored platform's identity server is started once, not every seven seconds

**Date**: September 14, 2026
**Type**: Bug Fix
**Components**: planton-operator (the identity component's admin recovery)

## Summary

On a platform restored from its archive, the operator re-establishes the identity server's master admin in two halves: a one-shot recovery Job before the server starts, and the admin reset once the server answers. The first half stopped any running server before checking the Job, on every reconcile until the second half wrote its marker. The pass after the Job succeeded started the server; the pass after that stopped it again; the second half, which needs the server answering, never ran. Live, the identity pods were replaced every seven seconds for ten minutes and the platform never left `Deploying`.

## What Changed

The first half reads the Job first. A succeeded Job hands over to the server unconditionally, whether or not the server is already up; only a Job that is absent or still running stops a live server so the recovery command can run beside no node. One test pins it: a succeeded Job beside a running Deployment returns nothing to wait for and leaves the Deployment standing.
