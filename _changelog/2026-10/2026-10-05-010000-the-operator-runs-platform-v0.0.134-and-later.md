# The Operator Runs Platform v0.0.134 and Later

**Date**: October 5, 2026
**Type**: Fix
**Components**: Planton operator

## Summary

**An operator-run install can start platform v0.0.134 and every release after it.** That release made the control plane require two settings, and the operator set neither. So the new control plane refused to boot (`Could not resolve placeholder 'LOG_FORMAT'`), and an upgraded install kept serving on its previous pod, because the rollout held the new one back. Planton's own management instance met it while adopting v0.0.139.

**The operator now declares both:**
- `LOG_FORMAT=json`: the operator only ever runs the control plane in a cluster, where a collector reads each log line (and its trace id) into a record.
- `METRICS_PORT=-1`: the metrics surface stays off, like the rest of the operator's observability block. The operator declares no metrics port or scrape target, so a listener would serve nobody. Opening that surface to a self-hosted Prometheus is a separate feature.

Images before v0.0.134 ignore both, so the oldest supported platform release stays v0.0.113.

## What Changed

- `operator/internal/resources/control_plane.go` renders the two variables, beside `LOG_LEVEL`.
- `testdata/boot-contract.txt` lists them under `[control-plane/control-plane]`.

## Why It Slipped, and What Stops the Next One

The boot contract pins what the operator emits, but nothing compared it with what each platform release requires. The platform's release preflight now does, against the operator release its own estate pins: a platform release that requires a variable that operator doesn't emit is refused when it is tagged, not when an instance boots.
