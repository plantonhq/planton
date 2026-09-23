# The Operator Hands the Control Plane No Runner Queue, and the Platform Floor Moves to v0.0.75

**Date**: September 24, 2026
**Type**: Fix
**Components**: Operator, Helm (planton-runner), Catalog (KubernetesPlantonPlatform), Docs

## Summary

From platform `v0.0.75`, a deploy whose connection names no runner follows the organization's default runner -- on a self-hosted install, the runner the operator seeds -- and the control plane derives that runner's queue itself. With no default runner and no Planton-hosted runner, it refuses the deploy with a sentence that names both ways out. So the operator stops rendering `TEMPORAL_PLATFORM_RUNNER_TASK_QUEUE_DEFAULT`, a typed copy of a queue the control plane now derives.

The platform floor moves from `v0.0.65` to `v0.0.75`. An older control plane sends every deploy whose connection names no runner to that variable's queue, so under this operator those deploys would have nowhere to go. The same release serves the runner door this operator already routes a remote runner's work calls to.

The runner chart pins the `v0.0.75` runner line: the first that carries its jobs through the control plane's door and pre-creates a build's ConfigMaps under the verbs the chart's build Role grants.

The public docs now say the one order deploys and live cloud operations follow: the connection's runner, else the organization's default runner, else a Planton-hosted runner.

## What Changed

- `operator/internal/resources/control_plane.go`: the variable and its comment removed.
- `operator/internal/resources/runner.go`: `RunnerTaskQueue` removed; that line was its only caller.
- `operator/internal/resources/control_plane_test.go`: pins that neither runner queue variable is set.
- `operator/internal/resources/runner_test.go`: pins the channel identifier, the one fact both sides derive the queue from.
- `operator/internal/platformversion/platformversion.go`: `MinimumSupported` is `v0.0.75`, with its reason.
- `operator/internal/resources/testdata/boot-contract.txt`: refreshed (the floor and the one name).
- `helm/planton-runner/Chart.yaml`: `appVersion` `v0.0.75`, with its reason.
- `catalog/kubernetes/kubernetesplantonplatform/e2e/scenarios/with-vault-keys.yaml` and `GUIDE.md`: the examples that run at the oldest release the operator runs move to `v0.0.75`.
- `site/public/docs/runner/planton-hosted-runners.md`, `runner/deployment.md`, `runner/index.md`, `connections/state-backends.md`: the one order; a default runner is refused while the organization's state lives in Planton-managed storage, and `planton state-backend list-planton-managed` and `move-off-planton` show and move it.

## Release order

Release this operator and chart only after platform `v0.0.75` has published its images: this operator refuses any older platform.

## Verification

- `go test ./internal/resources ./internal/platformversion`, `go vet` on both, and `go build ./cmd` in `operator/`.
- `helm lint helm/planton-runner` with the release lane's flags; `helm template` renders `ghcr.io/plantonhq/planton/runner:v0.0.75`.
- The site's internal link gate: every internal link resolves.
