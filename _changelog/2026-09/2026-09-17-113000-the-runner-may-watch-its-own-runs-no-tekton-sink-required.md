# The runner may watch its own runs; no Tekton sink is required

**Date**: September 17, 2026
**Type**: Feature
**Components**: The `planton-runner` Helm chart (`helm/planton-runner/`, released as `0.6.0`), the operator's in-cluster runner role (`operator/internal/resources/runner.go`)

## Summary

Planton's build path used Tekton's one cluster-wide CloudEvents sink to learn that a run changed: Tekton's events controller POSTed every namespace's events to one URL, and one runner received them all. A cluster with two Planton control planes -- a self-hosted install beside a second one, or the hosted product's environments -- had no honest answer: whichever runner the sink named heard about everyone's builds. From platform `v0.0.70` the runner watches the PipelineRuns and TaskRuns of its own build namespace and signals each change to the owning build itself, one hop earlier than the sink ever did. This release grants what that watch needs and stops telling operators to configure a sink.

## What Changed

### The RBAC

The build Role gains `watch` on `pipelineruns` and `taskruns` in the build namespace -- an informer lists, then watches -- in both places it is rendered: the chart's `build-rbac.yaml` and the operator's `RunnerBuildRole`. The runner's readiness probe asserts the same verbs (its `run-watch` check), so the three stay in lockstep or the probe says which is behind.

The chart's second Role -- a read grant on Tekton's `config-defaults` ConfigMap in the Tekton install namespace, which existed only so the probe could report whether the sink was configured -- is gone, and `build.tekton.installNamespace` with it. Nothing the runner does now reaches outside its build namespace.

### The words

The README's build section, the post-install notes, and `values.yaml` no longer instruct anyone to point `default-cloud-events-sink` at the runner. One step remains after install: register the cluster as a build connection and verify it. The webhook and its Service still render for one release, for control planes that predate the watch; a sink pointed at them keeps working and is redundant.

## Why It Matters

Tekton allows one sink per cluster and one Tekton per cluster (its operator admits exactly one `TektonConfig`). Routing events to many control planes would have needed a router every control plane on the cluster shares and something to keep it configured. A watch on the runner's own namespace needs neither: the namespace is the boundary, and a second Planton on the cluster changes nothing for the first. An adopter with a hand-installed Tekton also loses a cluster-admin step -- live build status no longer needs a write to Tekton's namespace.
