# KubernetesPlantonRunner

Runs a standing Planton runner appliance inside your Kubernetes cluster:
an always-on worker that receives deploy operations from the Planton
control plane and executes them from within the cluster's network -- with
an outbound-only posture (the runner dials out; nothing dials in).

## Purpose

Some targets are reachable only from inside the network -- private
services, internal endpoints, the cluster itself. Running the runner in
the cluster makes those targets deployable and operable with zero
inbound exposure: the runner only ever dials OUT to the control plane.

The module installs the official `planton-runner` Helm chart (OCI,
ghcr.io/plantonhq/charts) as a real Helm release, so the deployed runner
is byte-identical to a hand-installed one. The chart carries the
load-bearing mechanics the module deliberately does not re-model:
replicas pinned to 1 with a Recreate strategy (two live pods under one
runner name would revoke each other's keys), the ephemeral identity
volume the runner persists its minted identity into (container restarts
reuse it; pod recreation re-joins with the token), and the health
endpoints.

The appliance is standing infrastructure, not a bootstrap step. It
survives rebuilds of the workloads it deploys, which is what makes
teardown orderly: in-cluster workloads are destroyed through the
runner, and the runner itself is destroyed last.

## Key Features

- **Outbound-only networking** -- the runner initiates every connection
  it uses (the control plane, its work queue, image pulls); no inbound
  path exists or is needed.
- **Token-first enrollment** -- the runner is born with a runner TOKEN,
  never an identity. On first boot it presents the token, registers
  ITSELF, and receives its own individually revocable identity; revoking
  the token never touches runners it already admitted, and pod
  recreation re-joins with the same token (the token's lineage re-admits
  the runner it originally admitted).
- **The token never rides rendered values** -- the module materializes
  it as the `<name>-token` Kubernetes Secret the chart reads by name
  (its existingSecret form); rendered chart values land in Helm's
  release Secret, where an inline token would be readable by anyone
  with release-history read access.
- **A loud chart-version floor** -- charts below 0.4.0 predate token
  enrollment and silently ignore the enrollment values (the runner would
  deploy with no way to join); the module refuses them with an error
  that names the cause.
- **Optional build worker** -- enable `build` and the runner also
  executes container-image build pipelines through Tekton on this
  cluster (Tekton Pipelines must be installed).

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonRunner
metadata:
  name: cluster-runner
spec:
  namespace:
    value: planton-runner
  createNamespace: true
  token: $secret/cluster-runner-token
```

There is no manual credential step: before the infrastructure applies,
the platform mints a runner token and writes it at exactly the
managed-secret reference the manifest declares. Pick any secret slug and
deploy with:

```shell
planton apply -f runner.yaml
```

The runner registers itself under `<env>-<metadata.name>`
(`metadata.name` outside an environment) the moment it joins.

Both a Pulumi module and a Terraform/OpenTofu module implement this
kind at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
