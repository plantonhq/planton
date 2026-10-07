# Cluster Runner

This preset installs the standard in-cluster runner: the official
`planton-runner` Helm chart in its own namespace, an always-on worker
that receives deploy operations from the control plane and executes them
from inside the cluster's network. Two decisions -- a namespace and a
token reference -- and everything else self-configures. The 30-second
decision for making a cluster's private targets deployable.

## When to Use

- Private services or a cluster API endpoint no hosted runner fleet can
  reach need to be deployed to and operated
- You want zero inbound exposure: the runner only ever dials out to the
  control plane
- You already run the workloads' cluster and want the runner beside them

## Key Configuration Choices

- **Token as a managed-secret reference** -- the token authorizes the
  runner to JOIN, nothing more: on first boot the runner registers
  itself with the control plane and receives its own individually
  revocable identity, and revoking the token never touches runners it
  already admitted. On Planton the platform mints the token and writes
  it at exactly this reference before the infrastructure applies; the
  module carries it in a Kubernetes Secret, never in rendered chart
  values.
- **No mode or replica knobs** -- everything beyond the join (work
  queue, tunnel, API endpoints) arrives in the join response, so the
  runner self-configures on arrival; the chart pins replicas to 1 by
  design -- more capacity means more runners, never more copies of this
  one
- **`createNamespace: true` with `planton-runner`** -- a dedicated
  namespace, created with the standard governance labels and deleted
  with the resource
- **Chart defaults for sizing** -- requests 100m/256Mi, limits 1/2Gi;
  the runner runs one IaC operation at a time in 2Gi and queues the rest,
  and a larger memory limit in `resources` runs more at once

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<runner-name>` | Name for the runner appliance | Any name you choose |

The `runner-token` secret slug is yours to choose -- on Planton the
platform writes the token there automatically; elsewhere, create a token
with `planton runner token create` and store it under that slug.

## Related Presets

- `02-build-runner` -- adds the Tekton build worker for container-image builds
- `03-self-hosted-control-plane` -- joins a self-hosted control plane instead of the hosted one
