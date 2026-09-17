# Planton Runner Helm Chart

Deploy [Planton Runner](https://github.com/plantonhq/planton) to any Kubernetes cluster.

The Planton Runner is an agent that connects to the Planton control plane via a secure
mTLS reverse tunnel to execute cloud operations and IaC workflows on behalf of your
organization. The chart carries a **runner token**: on first boot the runner presents it
to the control plane's public join door, registers itself under its name, and receives
its own individually revocable identity. The token only ever authorizes joining -- it is
never the runner's identity, and one token can enroll many runners, each with its own
identity.

## Prerequisites

- Kubernetes 1.24+
- Helm 3.x
- A runner token for your organization (`planton runner token create <token-name>` --
  the secret is shown exactly once)

## Quick Start

The chart is published to GitHub Container Registry as an OCI artifact:

```bash
# Install a runner named after the release; it enrolls itself on first boot
helm install prod-a oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --create-namespace \
  --set enrollment.token=prt_...
```

The runner registers itself as `prod-a` (the release name) and appears in your
organization's Runners list the moment it joins. Set `enrollment.runnerName` to use a
name different from the release name.

## Installation

### Via `planton runner deploy` (recommended)

The Planton CLI automates target selection, values generation, and installation:

```bash
planton runner deploy <runner-name> --token prt_...
```

### Manual Helm install

```bash
helm install my-runner oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --create-namespace \
  --set enrollment.token=prt_...
```

For a self-hosted or local control plane, also set the join endpoint:

```bash
  --set enrollment.endpoint=planton.example.com:443
```

### Using an existing Kubernetes Secret

If you manage the token Secret externally (e.g., via sealed-secrets, external-secrets,
or a GitOps pipeline), reference it instead of providing the token directly:

```bash
# Create the Secret yourself
kubectl create secret generic runner-token \
  --namespace planton-runner \
  --from-literal=token=prt_...

# Install the chart referencing the existing Secret
helm install my-runner oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --set enrollment.existingSecret=runner-token
```

## Configuration

### Enrollment

| Parameter | Description | Default |
|-----------|-------------|---------|
| `enrollment.token` | The runner token (`prt_` prefixed) that authorizes joining | `""` |
| `enrollment.runnerName` | The name the runner registers itself under | Release name |
| `enrollment.endpoint` | Control-plane gRPC endpoint (host:port) to join | Runner's hosted default (`api.live.planton.ai:443`) |
| `enrollment.existingSecret` | Name of a pre-existing Secret holding the token | `""` |
| `enrollment.existingSecretKey` | Key within the existing Secret | `"token"` |

Either `enrollment.token` or `enrollment.existingSecret` must be provided. If neither
is set, the chart fails at render time with an error.

Re-enrollment is safe by design: a pod recreated without its identity volume simply
re-joins, and the same token re-admits its own runner with a fresh key. A DIFFERENT
token can never take over an existing runner's name -- that join is refused (token
lineage). Revoking a token never touches the runners it admitted; each keeps its own
identity, individually revocable.

### Runner

| Parameter | Description | Default |
|-----------|-------------|---------|
| `runner.port` | gRPC server port (CloudOps routes via tunnel to this port) | `50051` |
| `runner.logLevel` | Log level: debug, info, warn, error | `"info"` |
| `runner.executionMode` | Execution mode: auto, grpc, temporal, dual | `"auto"` |

`auto` derives the mode from the identity document the runner receives when it
enrolls and from its perimeter. A Temporal address with a tunnel endpoint means
`dual`: deploys and builds over Temporal, live cloud operations through the tunnel.
A Temporal address without one means `temporal`, a pure worker -- the shape of
every runner enrolled with a self-hosted instance, because a self-hosted instance
operates no runner tunnel (its own in-cluster runner is dialed directly; yours
pulls work from the deploy queue the front door routes). Set an explicit mode
only to override.

### How the pod is probed

Both probes are gRPC probes against `runner.port`, the one port that serves the
standard gRPC health service in every execution mode (the CloudOps server's in
`grpc` and `dual`, a health-only server's in `temporal`) -- so the chart probes
every runner one way, whatever its tunnel, and exactly the way the operator's
in-cluster runner is probed:

- **Readiness** watches the IaC worker's health key (`ai.planton.runner.iac.worker`):
  SERVING once the worker is actually polling its Temporal queue, not merely once
  the port is bound. An explicit `runner.executionMode: grpc` runs no worker, so
  the chart probes the overall health service instead.
- **Liveness** watches only the overall health service: a Temporal outage reads as
  not-ready, never as a restart loop of a healthy process.

The tunnel agent's HTTP health and metrics ports (8093, 8094) are declared on the
container for what they are; they exist only when the identity document carries a
tunnel endpoint and are never a probe target. (Charts before 0.5.0 probed 8093 and
restarted a tunnel-less runner every liveness window.)

### Temporal (override lane)

The identity document the runner receives at enrollment already carries the
instance's runner-reachable Temporal coordinates -- these values exist only to point
the worker somewhere else. The environment variables are only injected into the pod
when `temporal.address` is set.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `temporal.address` | Temporal server address (host:port) override | `""` |
| `temporal.namespace` | Temporal namespace override | `"default"` |
| `temporal.maxConcurrency` | Maximum concurrent Temporal activities | `10` |

### Build capability

Enable this to make the cluster a **build cluster**: the runner executes build
pipelines on it (Tekton PipelineRuns), streams task pod logs to the control
plane, receives Tekton CloudEvents on its webhook, and serves the readiness
checks a registered build connection reports.

Prerequisites:

- **Tekton Pipelines** installed on the cluster.
- `runner.executionMode` left at `auto` (or set to `temporal`/`dual`); the chart
  fails at render time if builds are enabled with an explicit `grpc` mode.
- Exactly **one** build-capable runner per watched Tekton namespace (the build
  log streamer is a singleton; the chart already pins one replica).

```bash
helm install my-runner oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --create-namespace \
  --set enrollment.token=prt_... \
  --set build.enabled=true \
  --set build.tektonNamespace=build-pipelines
```

| Parameter | Description | Default |
|-----------|-------------|---------|
| `build.enabled` | Run the pipeline-build worker and the Tekton CloudEvents webhook | `false` |
| `build.tektonNamespace` | Namespace where builds land and the log streamer watches; empty uses the runner's own namespace | `""` |
| `build.webhookPort` | Container port for the Tekton CloudEvents webhook | `8086` |
| `build.rbac.create` | Create the Role/RoleBinding the build capability needs | `true` |
| `build.tekton.installNamespace` | Namespace of the Tekton Pipelines installation | `"tekton-pipelines"` |

Two steps complete the build-cluster setup after install:

1. **Point Tekton's CloudEvents sink at the runner's webhook** so live build
   status flows without waiting on the reconciliation safety net. In the
   `tekton-pipelines/config-defaults` ConfigMap set:

   ```yaml
   default-cloud-events-sink: http://<release-fullname>.<namespace>.svc.cluster.local/service-hub/tekton/cloud-event
   ```

   (The chart's Service serves the webhook on port 80, so the URL needs no
   explicit port -- but the `/service-hub/tekton/cloud-event` path is
   required: the webhook serves only that route, and a sink pointed at the
   bare Service root gets a 404 on every event, silently degrading builds to
   the reconciliation safety net. The exact URL is printed in the
   post-install notes.)

2. **Register the cluster as a build connection** (console: Connections →
   Build, or `planton apply` a `TektonConnection` naming this runner), then
   verify it. The readiness check runs on this runner over the same queue real
   builds ride, so a passing verify also proves end-to-end routing.

The RBAC the chart creates mirrors exactly what the build capability performs:
create PipelineRuns and their per-build supporting resources, reconcile by
list, label-scoped cleanup, and follow task pod logs — plus a read grant on
Tekton's `config-defaults` ConfigMap so the events-sink readiness check can
report a conclusive verdict.

### Image

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Container image repository | `ghcr.io/plantonhq/planton/runner` |
| `image.tag` | Container image tag | Chart `appVersion` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |

### Resources and Scheduling

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `256Mi` |
| `resources.limits.cpu` | CPU limit | `1` |
| `resources.limits.memory` | Memory limit | `1Gi` |
| `serviceAccount.create` | Create a Kubernetes ServiceAccount | `true` |
| `serviceAccount.annotations` | ServiceAccount annotations (e.g., IRSA) | `{}` |
| `nodeSelector` | Node selector for pod scheduling | `{}` |
| `tolerations` | Tolerations for pod scheduling | `[]` |
| `affinity` | Affinity rules for pod scheduling | `{}` |

## Architecture

The runner makes only **outbound** connections. No Ingress or public Service is required.

```
Runner Pod (your cluster)
  ├─ join + gRPC + TLS ──► control plane API  (enrollment, secrets, variables)
  └─ mTLS tunnel       ──► tunnel server      (cloud operations requests)
```

On first boot the runner joins the control plane with its token and receives its
identity document -- the runner's identity, mTLS certificates, API key, and endpoint
configuration, minted server-side and delivered only to the runner. It persists the
document on the pod's writable identity volume (`/var/lib/planton-runner`), so
container restarts reuse it; a recreated pod re-joins. The runner then maintains a
persistent reverse tunnel through which the Planton control plane sends cloud
operations requests, and makes authenticated gRPC calls to the control plane API to
resolve secrets and variables at runtime.

With builds enabled there is one additional, **cluster-internal** listener: Tekton posts
pipeline CloudEvents to the runner's webhook through the chart's ClusterIP Service. No
Ingress and no public exposure — the traffic never leaves the cluster. That Service
exists for exactly this caller and is rendered only when builds are enabled; a runner
without builds has no Service, because nothing in the cluster needs to dial it (health
is probed on the pod).

### Automatic Rollouts

When the chart manages the token Secret (i.e., `enrollment.existingSecret` is not
set), the Deployment includes a `checksum/token` annotation that forces a pod rollout
whenever the token changes. When using an existing Secret, you are responsible for
triggering rollouts (e.g., via `kubectl rollout restart`). The deployment strategy is
`Recreate`: two live pods under one runner name would revoke each other's keys, so
the old pod terminates before the new one starts and re-joins.

## Verification

After installation, verify the runner is running:

```bash
kubectl -n planton-runner get pods
kubectl -n planton-runner logs -l app.kubernetes.io/name=planton-runner
```

The runner also appears in your organization's Runners list (`planton runner list`)
the moment it joins.

## Upgrading

Image or configuration upgrades do not need the token again -- Helm keeps the release's
values:

```bash
helm upgrade my-runner oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --reuse-values \
  --set image.tag=v1.2.3
```

To rotate to a NEW token (e.g. after revoking the old one), upgrade with the new value:

```bash
helm upgrade my-runner oci://ghcr.io/plantonhq/charts/planton-runner \
  --namespace planton-runner \
  --reuse-values \
  --set enrollment.token=prt_...
```

Note: a runner can only be re-admitted by the token that admitted it. If that token is
revoked, reset the runner's enrollment first (`planton runner reset-enrollment <name>`)
so the next join re-admits it under the new token.

## Uninstalling

```bash
helm uninstall my-runner --namespace planton-runner
```

Uninstalling removes the pod and its identity volume. The runner's registration and
identity remain in your organization (delete the runner from the console or CLI to
revoke them).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
