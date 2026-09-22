# The Operator Drops a Task-Queue Variable Nothing Read

**Date**: September 22, 2026
**Type**: Fix
**Components**: Operator

## Summary

The operator no longer renders `TEMPORAL_PLATFORM_RUNNER_TASK_QUEUE_AWS` on the control plane it deploys. The control plane's per-provider queue overrides are a map with no entries, so the variable bound to nothing; only `TEMPORAL_PLATFORM_RUNNER_TASK_QUEUE_DEFAULT` -- derived from the bootstrap organization and the runner slug, the same derivation the runner resources use for the worker's queue -- carries the fleet queue. The boot-contract fixture loses the one name; no platform version floor moves, because no platform ever read it.

## What Changed

- `operator/internal/resources/control_plane.go`: the `_AWS` line removed; the comment says why there is exactly one queue variable.
- `operator/internal/resources/control_plane_test.go`: the derivation test now pins that the variable is absent.
- `operator/internal/resources/testdata/boot-contract.txt`: refreshed.

## Verification

`go test ./internal/resources/` in `operator/`.
