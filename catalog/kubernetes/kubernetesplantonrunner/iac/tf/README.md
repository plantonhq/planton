# Terraform Module to Deploy KubernetesPlantonRunner

This module installs a standing Planton runner from the official
`planton-runner` chart (oci://ghcr.io/plantonhq/charts) as a real Helm
release. The resources, in dependency order: the installation namespace
(created only when `create_namespace` is true, with the Planton
governance labels), the `<name>-token` Secret holding the runner token,
and the `helm_release` -- named `metadata.name`, pinned to
`chart_version` (default 0.10.0). A lifecycle precondition refuses
versions below 0.4.0 with an explicit error: those charts predate token
enrollment and silently ignore the enrollment values.

The release's values are three documents merged in order by the
provider (helm `-f` semantics): the typed rendering first -- the image
block (repository + tag), the enrollment block (the token Secret's NAME
via the chart's existingSecret form, the runner's registration name,
and the control-plane endpoint only when one is declared), the
resources block only when customized (the chart's own defaults are the
baseline), and the build block when the build worker is enabled -- then
the spec's `helm_values` escape hatch, then the enrollment block
re-pinned last, so the token never rides rendered values.

The release holds exactly one runner: the chart pins replicas to 1 with
a Recreate strategy (two live pods under one runner name would revoke
each other's keys). Scaling execution capacity means more runners,
never more copies of this one.

The release deliberately sets `wait = false`: the runner's readiness
contract is its work queue (operations wait there until the worker
polls), never pod liveness.

`variables.tf` reflects the proto schema for `KubernetesPlantonRunner`.
Both the kubernetes and helm providers are configured by the calling
workspace/environment (the standard kubeconfig contract); the provider
blocks in `provider.tf` are deliberately empty.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: Cluster credentials are provided via IaC input (CLI), not in the manifest `spec`.

See [`e2e/manifest.yaml`](../../e2e/manifest.yaml) for a minimal test manifest.
