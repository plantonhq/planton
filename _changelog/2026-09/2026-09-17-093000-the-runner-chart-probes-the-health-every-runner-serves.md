# The runner chart probes the health every runner serves

**Date**: September 17, 2026
**Type**: Fix
**Components**: The `planton-runner` Helm chart (`helm/planton-runner/`, released as `0.5.0`), the `KubernetesPlantonRunner` kind's default chart version (`catalog/kubernetes/kubernetesplantonrunner/`)

## Summary

The first runner appliance declared from a self-hosted Planton instance joined, registered, worked both its deploy queues -- and was restarted by the kubelet every liveness window for six hours, because the chart probed `GET :8093/healthz`: the tunnel agent's health port, which a runner opens only when the instance that enrolled it operates a runner tunnel. A self-hosted instance operates none, by design, so the port never existed and every probe was refused. Chart `0.5.0` probes the standard gRPC health service on the runner's own port instead -- the one surface every execution mode serves -- exactly the way the operator's in-cluster runner has always been probed. One probe contract for every runner, whatever its tunnel.

## What Changed

### The probes

Readiness is a gRPC probe on the IaC worker's health key (`ai.planton.runner.iac.worker`): SERVING once the worker is actually polling its Temporal queue, not merely once the port is bound. An explicit `runner.executionMode: grpc` runs no worker and never registers that key, so the template probes the overall health service for it instead. Liveness probes only the overall service, so a Temporal outage reads as not-ready and never as a restart loop of a healthy process. The delays and periods are the operator's.

### The Service

Rendered only when builds are enabled, carrying the webhook port -- Tekton is its one caller. The `health: 8093` Service port is gone: it advertised the tunnel agent's health, which nothing dialed through a Service and which a tunnel-less runner never serves; and a Service with no ports is invalid, so a runner without builds now has none. The tunnel agent's container ports stay declared for what they are.

### The words

`Chart.yaml` pins `appVersion` to `v0.0.69`, the first runner release that derives a pure-worker mode for a tunnel-less instance, and drops the note that called the deployment path unexercised (it was exercised on GKE, 2026-09-16 and 17). `values.yaml` says what `auto` derives and what an explicit `grpc` changes about the probe; the README gains "How the pod is probed" and says why the Service exists only with builds.

### The kind

`default_chart_version` moves to `0.5.0` in both engines and in the kind's catalog page; the enrollment-contract floor stays `0.4.0`. An install that pins nothing gets the probed-right chart at the next catalog release; an install that needs it before then names `chartVersion: "0.5.0"` in its manifest.

## Why It Matters

Every remote runner enrolled with a self-hosted Planton is a worker without a tunnel. Until `0.5.0`, every one of them installed green and then crash-looped, with the pipeline that declared it already reporting success. The runner's own fix (deriving `temporal` mode and binding its webhook for the cluster) ships in platform `v0.0.69`; this chart is the half that lets the kubelet see the runner the way it is.
