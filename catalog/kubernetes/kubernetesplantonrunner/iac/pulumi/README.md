# Pulumi Module to Deploy KubernetesPlantonRunner

This module installs a standing Planton runner from the official
`planton-runner` chart (oci://ghcr.io/plantonhq/charts) as a real Helm
release. The pieces, in dependency order: the installation namespace
(created only when `create_namespace` is true, with the Planton
governance labels), the `<name>-token` Secret holding the runner token,
and the release itself -- named `metadata.name`, pinned to
`chart_version` (default 0.11.0). Versions below 0.4.0 predate token
enrollment and silently ignore the enrollment values, so the module
refuses them with an explicit error before touching the cluster.

The chart values the module renders: the image block (repository +
tag), the enrollment block (the token Secret's NAME via the chart's
existingSecret form, the runner's registration name, and the
control-plane endpoint only when one is declared), the resources block
only when customized (the chart's own defaults are the baseline), and
the build block when the build worker is enabled. The spec's
`helm_values` escape hatch merges over the rendered values with Helm
`-f` semantics -- and the enrollment block is re-pinned after the
merge, so the token never rides rendered values.

The release holds exactly one runner: the chart pins replicas to 1 with
a Recreate strategy (two live pods under one runner name would revoke
each other's keys). Scaling execution capacity means more runners,
never more copies of this one.

The release deliberately does not await resource readiness (SkipAwait):
the runner's readiness contract is its work queue (operations wait
there until the worker polls), never pod liveness.

OCI wiring: Pulumi's helm.v3.Release resolves OCI registries through
the chart reference -- the joined `oci://.../planton-runner` string
with no RepositoryOpts; the Terraform module reaches the same chart
through repository + bare chart name. Same chart bytes, different
wiring; the two modules match output-for-output.

## CLI usage (Planton pulumi)

```bash
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
planton pulumi refresh \
  --manifest ../../e2e/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack <org>/<project>/<stack> \
  --module-dir .
```

## Debugging

You can debug the Pulumi program with Delve by pointing
`runtime.options.binary` in `Pulumi.yaml` at a wrapper script:

```yaml
runtime:
  name: go
  options:
    binary: ./debug.sh
```

Then run your Pulumi commands as usual. For detailed steps, see
`docs/pages/docs/guide/debug-pulumi-modules.mdx`.
