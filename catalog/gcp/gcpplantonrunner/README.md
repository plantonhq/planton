# GcpPlantonRunner

Runs a standing Planton runner appliance inside your GCP project: an
always-on worker that receives deploy operations from the Planton control
plane and executes them from within the project's network perimeter --
with an outbound-only network posture (the runner dials out; nothing
dials in).

## Purpose

Some infrastructure is reachable only from inside the network. The
canonical case is a **private GKE control plane** (or a private-IP
Cloud SQL instance): no hosted runner fleet can reach it, so nothing
outside the VPC can deploy into it. Placing a runner beside it makes that
target deployable and operable -- initial installs, day-2 updates,
destroys, and live resource browsing -- without opening a single inbound
port.

The appliance is standing infrastructure, not a bootstrap step. It
survives rebuilds of the clusters it deploys to, which is what makes
teardown orderly: in-cluster workloads are destroyed through the runner,
the cluster is destroyed over the GCP path, and the runner itself is
destroyed last.

The compute substrate is **Cloud Run**: a serverless service pinned to
exactly one always-on instance, no hosts to patch, restarted
automatically if the runner ever exits. The spec deliberately does not
model the substrate -- it models intent (placement, sizing, version, and
the token the runner joins with), so the API stays stable however the
implementation evolves.

## Token-first enrollment

The runner is born with a runner **token**, never an identity. On first
boot it presents the token to the control plane, registers itself, and
receives its own individually revocable identity. The token only gates
joining -- it is never the runner's identity: revoking a token never
touches runners it already admitted, and instance replacement re-joins
with the same token (the token's lineage re-admits the runner it
originally admitted; no other token can). The token reaches the container
through a module-created Secret Manager secret (`<name>-token`), resolved
at instance start -- it never appears in the service definition.

## Key Features

- **Outbound-only networking** -- the runner initiates every connection
  (control plane, its work queue, image pulls). The service accepts no
  meaningful inbound traffic: no `run.invoker` grants exist, so the
  default authenticated-only posture refuses every caller.
- **Pull-based execution** -- the runner polls its own queue for deploy
  operations; work waits in the queue while the runner boots, so ordering
  never depends on timing. There is no execution-mode knob: the runner
  derives its mode from the identity the join returns.
- **Direct VPC egress** -- optional `vpcAccess` routes the runner's
  private-range traffic into your VPC (the path to a private GKE control
  plane); the control-plane dial-out keeps its normal internet path, so a
  misconfigured VPC can never sever the runner from the control plane.
- **First-class runtime identity** -- the runner runs as a service
  account (reference your own `GcpServiceAccount`, or a permissionless
  dedicated one is created -- never the Compute Engine default), the seam
  that lets keyless cloud access run through the runner.
- **Token handled as a secret end to end** -- stored in Secret Manager
  and resolved into the container by Cloud Run at instance start; the
  runtime service account is granted read access to exactly that one
  secret and nothing else.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPlantonRunner
metadata:
  name: gke-runner
spec:
  region: us-central1
  token: $secret/gke-runner-token
```

There is no manual credential step: before the infrastructure applies,
the platform mints a runner token and writes it at exactly the
managed-secret reference the manifest declares. Pick any secret slug and
deploy with:

```shell
planton apply -f runner.yaml
```

## Field Highlights

- `region` (required) -- deploy the runner in the same region as the
  private endpoints it needs to reach.
- `token` (required, secret) -- the join authorization, as a
  `$secret/<slug>` managed-secret reference; never inline plaintext.
- `controlPlaneEndpoint` -- host:port of a self-hosted control plane;
  leave unset for Planton's hosted endpoint.
- `vpcAccess` -- VPC network, subnetwork, and network tags for Direct VPC
  egress; omit when the runner only needs public endpoints.
- `serviceAccount` -- the runner's GCP identity; unset creates a
  dedicated permissionless account.
- `cpu` / `memory` -- Cloud Run instance sizing (defaults 1 vCPU /
  512Mi); memory pressure shows up as failed operations mid-apply, so
  size memory up before cpu.

## Outputs

| Output | Description |
|--------|-------------|
| `service_name` | The fully qualified Cloud Run service name (`projects/{project}/locations/{region}/services/{name}`) |
| `service_short_name` | The service's short name (`metadata.name`) |
| `service_account_email` | The runner's runtime identity -- grant it roles for keyless cloud access |
| `token_secret_id` | The Secret Manager secret holding the runner token |
| `runner_name` | The name the runner registers itself under -- shown by `planton runner list` the moment it joins |
| `project_id` | The GCP project the runner was deployed in |
| `region` | The GCP region the runner was deployed in |

Both a Pulumi module and a Terraform/OpenTofu module implement this
kind at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
