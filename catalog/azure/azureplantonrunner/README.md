# AzurePlantonRunner

Runs a standing Planton runner appliance inside your Azure subscription:
an always-on worker that receives deploy operations from the Planton
control plane and executes them from within your network perimeter --
with an outbound-only network posture (the runner dials out; nothing
dials in).

## Purpose

Some infrastructure is reachable only from inside the network. The
canonical case is a **private AKS API server** (or a private-endpoint
database): no hosted runner fleet can reach it, so nothing outside the
VNet can deploy into it. Placing a runner beside it makes that target
deployable and operable -- initial installs, day-2 updates, destroys,
and live resource browsing -- without opening a single inbound port.

The appliance is standing infrastructure, not a bootstrap step. It
survives rebuilds of the clusters it deploys to, which is what makes
teardown orderly: in-cluster workloads are destroyed through the runner,
the cluster is destroyed over the Azure path, and the runner itself is
destroyed last.

The compute substrate is **Azure Container Apps**: a single-revision app
pinned to exactly one replica, no hosts to patch, restarted
automatically if the runner ever exits. The runner runs inside a
Container App Environment you reference -- the environment decides the
network boundary: a VNet-integrated environment gives the runner reach
into that network's private endpoints. The spec deliberately does not
model the substrate -- it models intent (placement, sizing, version, and
the token the runner joins with), so the API stays stable however the
implementation evolves.

## Token-first enrollment

The runner is born with a runner **token**, never an identity. On first
boot it presents the token to the control plane, registers itself, and
receives its own individually revocable identity. The token only gates
joining -- it is never the runner's identity: revoking a token never
touches runners it already admitted, and replica replacement re-joins
with the same token (the token's lineage re-admits the runner it
originally admitted; no other token can). The token lives in the
Container App's own secret store (`runner-token`); the container reads
it as a secret-backed environment variable, never as plaintext.

## Key Features

- **Outbound-only networking** -- the runner initiates every connection
  (control plane, its work queue, image pulls). The app exposes no
  ingress at all: there is nothing to dial in to.
- **Pull-based execution** -- the runner polls its own queue for deploy
  operations; work waits in the queue while the runner boots, so ordering
  never depends on timing. There is no execution-mode knob: the runner
  derives its mode from the identity the join returns.
- **The environment decides the reach** -- reference a VNet-integrated
  Container App Environment and the runner can deploy to that network's
  private endpoints; the module references the environment, never
  creates or mutates it.
- **Token handled as a secret end to end** -- stored in the app's own
  secret store and referenced by name from the container's environment;
  reading the app definition reveals nothing.
- **Single-revision by design** -- new revisions replace the old one,
  the only sane model for a workload whose identity contract forbids two
  live copies.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePlantonRunner
metadata:
  name: aks-runner
spec:
  resourceGroup:
    value: my-resource-group
  containerAppEnvironmentId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-resource-group/providers/Microsoft.App/managedEnvironments/my-environment
  token: $secret/aks-runner-token
```

There is no manual credential step: before the infrastructure applies,
the platform mints a runner token and writes it at exactly the
managed-secret reference the manifest declares. Pick any secret slug and
deploy with:

```shell
planton apply -f runner.yaml
```

## Field Highlights

- `resourceGroup` (required) -- the resource group the app is created
  in; a literal name or an `AzureResourceGroup` reference.
- `containerAppEnvironmentId` (required) -- the environment the runner
  runs in; pick (or compose) a VNet-integrated one when the runner must
  reach private endpoints.
- `token` (required, secret) -- the join authorization, as a
  `$secret/<slug>` managed-secret reference; never inline plaintext.
- `controlPlaneEndpoint` -- host:port of a self-hosted control plane;
  leave unset for Planton's hosted endpoint.
- `cpu` / `memory` -- Consumption-plan sizing (defaults 0.5 vCPU / 1Gi);
  the pairing is fixed at memory = cpu x 2 and validated up front.

## Outputs

| Output | Description |
|--------|-------------|
| `container_app_id` | The Azure resource ID of the Container App keeping the runner running |
| `container_app_name` | The Container App's name (`metadata.name`) |
| `token_secret_name` | The Container App secret holding the runner token |
| `runner_name` | The name the runner registers itself under -- shown by `planton runner list` the moment it joins |
| `resource_group_name` | The resource group the runner was deployed in |

Both a Pulumi module and a Terraform/OpenTofu module implement this
kind at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
