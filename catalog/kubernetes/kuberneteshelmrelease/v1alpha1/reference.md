# KubernetesHelmRelease

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesHelmReleaseSpec** installs an upstream Helm chart as a real
Helm release — the catalog's sole intentional passthrough. Both engines
perform an actual `helm install`: hooks run, the release secret is
written, and `helm list` shows the release exactly as if installed by the
Helm CLI.

WHEN NOT TO USE THIS: a first-class catalog kind always wins. Typed
kinds validate their configuration before deploy, export composable
outputs, and teach their trade-offs field by field — a generic chart
install does none of that. Reach for KubernetesHelmRelease only when the
catalog has no kind for the chart you need.

VALUES MODEL (Helm's own): `values_yaml` is the values file; `set`,
`set_string`, and `set_sensitive` are the `--set`-style overrides applied
on top. Both engines merge these layers with identical precedence
(values_yaml, then set, then set_string, then set_sensitive) and hand Helm
one final values map — the same manifest installs byte-identical releases
on either engine.

CRD LIFECYCLE: Helm installs the CRDs in a chart's `crds/` directory once
and never upgrades or removes them. The modules OWN that surface instead:
the CRDs are derived from the pinned chart at deploy time (rendered with
the release's own values), applied keyed by CRD name ahead of the
release, kept when the resource is destroyed (`crds.keep_on_uninstall`),
re-adopted on reinstall, and moved with every `version` bump; a `version`
lower than the CRDs already on the cluster is refused before anything
changes. CRDs a chart templates as ordinary release resources stay
Helm's: a chart that marks them `helm.sh/resource-policy: keep` protects
them itself and installs as is; a chart that templates them without that
mark would delete them with the release, so it is refused with the CRD
names and the remedies unless `crds.allow_helm_managed` accepts it.

## Example

```yaml
# Full-surface test manifest for the offline plan proofs. Exercises the
# chart identity, all three values layers, release-name override, and the
# lifecycle knobs both engines express (take_ownership stays out: the pulumi
# module rejects it loudly at the pinned SDK, and this manifest feeds BOTH
# engines' offline proofs).
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHelmRelease
metadata:
  name: test-helm-release
spec:
  namespace:
    value: test-helm-release-ns
  create_namespace: true
  repo: https://stefanprodan.github.io/podinfo
  chart: podinfo
  version: 6.9.2
  release_name: podinfo-proof
  values_yaml: |
    replicaCount: 2
    resources:
      requests:
        cpu: 100m
        memory: 64Mi
  set:
    replicaCount: "1"
    ui.color: "#34577c"
  set_string:
    image.tag: "6.9.2"
  set_sensitive:
    secret.apiKey: proof-api-key
  atomic: true
  cleanup_on_fail: true
  wait_for_jobs: true
  timeout_seconds: 600
  crds:
    install: true
    keep_on_uninstall: true
    allow_helm_managed: false
  max_history: 5
  description: offline plan proof
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.repo` | `string` | yes |  |  |
| `spec.chart` | `string` | yes |  |  |
| `spec.version` | `string` | yes |  |  |
| `spec.releaseName` | `string` |  |  |  |
| `spec.valuesYaml` | `string` |  |  |  |
| `spec.set` | `map<string, string>` |  |  |  |
| `spec.setString` | `map<string, string>` |  |  |  |
| `spec.setSensitive` | `map<string, string>` (sensitive) |  |  |  |
| `spec.repositoryUsername` | `string` |  |  |  |
| `spec.repositoryPassword` | `string` (sensitive) |  |  |  |
| `spec.atomic` | `bool` |  |  |  |
| `spec.cleanupOnFail` | `bool` |  |  |  |
| `spec.skipAwait` | `bool` |  |  |  |
| `spec.waitForJobs` | `bool` |  |  |  |
| `spec.timeoutSeconds` | `int32` |  | `300` |  |
| `spec.dependencyUpdate` | `bool` |  |  |  |
| `spec.maxHistory` | `int32` |  | `10` |  |
| `spec.replace` | `bool` |  |  |  |
| `spec.forceUpdate` | `bool` |  |  |  |
| `spec.reuseValues` | `bool` |  |  |  |
| `spec.resetValues` | `bool` |  |  |  |
| `spec.disableWebhooks` | `bool` |  |  |  |
| `spec.disableOpenapiValidation` | `bool` |  |  |  |
| `spec.takeOwnership` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.crds` | `KubernetesHelmReleaseCrds` |  |  |  |
| `spec.crds.install` | `bool` |  | `true` |  |
| `spec.crds.keepOnUninstall` | `bool` |  | `true` |  |
| `spec.crds.allowHelmManaged` | `bool` |  | `false` |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

The namespace to install the release into. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the release is installed, and deleted with the
resource. When false, the namespace must already exist.

### spec.repo

`string` · required

The chart repository: an HTTP(S) Helm repository URL (e.g.
"https://stefanprodan.github.io/podinfo") or an OCI registry reference
(e.g. "oci://ghcr.io/stefanprodan/charts"). For OCI, the chart is pulled
as `<repo>/<chart>:<version>`.

- rule: Chart repository must be an http(s):// Helm repository URL or an oci:// registry reference
- rule: {"required":true}

### spec.chart

`string` · required

The chart name within the repository (e.g. "podinfo", "cert-manager").

- rule: {"required":true}

### spec.version

`string` · required

The exact chart version to install (e.g. "6.9.2"). Required — an
unpinned "latest" install is not reproducible, and reproducibility is
the point of declaring the release here.

- rule: {"required":true}

### spec.releaseName

`string`

Overrides the Helm release name. When omitted, the release is named
after the resource (`metadata.name`). Helm limits release names to 53
characters (DNS label rules with dots allowed by Helm, kept strict here:
lowercase alphanumerics and hyphens).

- rule: Release name must be lowercase alphanumeric with hyphens, at most 53 characters (Helm's release-name limit)

### spec.valuesYaml

`string`

Chart values as a YAML document — the equivalent of a Helm values file
passed with `-f`. Full YAML expressiveness: nested maps, lists, numbers,
booleans. Applied first; `set*` overrides layer on top. Do not put
secrets here — use `set_sensitive`.

Example:
```yaml
replicaCount: 2
resources:
  requests:
    cpu: 100m
```

### spec.set

`map<string, string>`

Targeted value overrides in Helm `--set` form: the key is a dotted path
(e.g. "image.tag", "ingress.hosts[0].host"), the value is parsed with
Helm's own coercion — "true"/"false" become booleans, digits become
numbers, "null" removes the key. For values that must stay literal
strings (version-like tags such as "1.30"), use `set_string` instead.
Applied after `values_yaml`.

### spec.setString

`map<string, string>`

Targeted value overrides in Helm `--set-string` form: same dotted-path
keys as `set`, but the value is always kept as a literal string — no
type coercion. Applied after `set`.

### spec.setSensitive

`map<string, string>` · sensitive

Secret value overrides in Helm `--set-string` form (always literal
strings): same dotted-path keys, for credentials and other sensitive
chart values. Kept out of rendered plans and state where each engine
supports it. Applied last — highest precedence.

### spec.repositoryUsername

`string`

Username for a private chart repository (HTTP basic auth or OCI registry
login). Set together with `repository_password`.

### spec.repositoryPassword

`string` · sensitive

Password or token for a private chart repository. Set together with
`repository_username`.

### spec.atomic

`bool`

When true, a failed install or upgrade rolls everything back and purges
new resources — the release is never left half-deployed. Atomic implies
waiting for readiness, so it cannot be combined with `skip_await`.

### spec.cleanupOnFail

`bool`

When true, a failed upgrade deletes the resources that upgrade newly
created (a lighter cleanup than `atomic`, which also rolls back).

### spec.skipAwait

`bool`

When true, the deploy returns as soon as Helm records the release —
without waiting for the chart's resources to become ready. When false
(the default), both engines wait for readiness. Cannot be combined with
`atomic` (atomic must wait to know whether to roll back).

### spec.waitForJobs

`bool`

When true (and awaiting readiness), also wait until all Jobs the chart
creates have completed. Ignored when `skip_await` is true.

### spec.timeoutSeconds

`int32` · optional (explicit presence)

Maximum seconds to wait for each Kubernetes operation (and for readiness
when awaiting). Helm's default is 300.

- default: `300`
- rule: {"int32":{"gt":0}}

### spec.dependencyUpdate

`bool`

When true, run `helm dependency update` before installing — fetches the
chart's declared subchart dependencies. Charts published with bundled
dependencies (the common case) do not need this.

### spec.maxHistory

`int32` · optional (explicit presence)

Maximum number of release revisions Helm keeps for rollback history.
0 means no limit. Helm's default is 10.

- default: `10`
- rule: {"int32":{"gte":0}}

### spec.replace

`bool`

When true, re-use the release name even if it is already taken by a
failed/uninstalled release that left history behind. Helm marks this
unsafe in production — prefer cleaning up the old release.

### spec.forceUpdate

`bool`

When true, force resource updates through delete-and-recreate when a
field cannot be patched in place. Disruptive — pods are replaced, not
rolled.

### spec.reuseValues

`bool`

On upgrade, when true, reuse the last release's values and merge these
on top (Helm `--reuse-values`). Mutually exclusive with `reset_values`.

### spec.resetValues

`bool`

On upgrade, when true, reset the values to the chart's built-in defaults
before applying these (Helm `--reset-values`). Mutually exclusive with
`reuse_values`.

### spec.disableWebhooks

`bool`

When true, chart hooks (pre/post install, upgrade, delete) do not run.

### spec.disableOpenapiValidation

`bool`

When true, skip validating the rendered manifests against the cluster's
OpenAPI schema before applying. Only needed for charts that render
fields newer than the cluster's schema.

### spec.takeOwnership

`bool`

When true, install/upgrade skips existing-resource conflict checks and
adopts matching resources into the release (Helm `--take-ownership`).
The migration knob: point a release at resources an earlier tool
created. Currently honored by the terraform provisioner only — the
pulumi engine's pinned SDK predates the flag and rejects it loudly
rather than silently ignoring it.

### spec.description

`string`

Free-form note stored with the release (visible in `helm status`).

### spec.crds

`KubernetesHelmReleaseCrds`

CRD lifecycle for the chart's `crds/` directory, and the posture toward
CRDs the chart templates itself. Unset means the module owns the
directory's CRDs and keeps them, and refuses Helm-managed ones.

### spec.crds.install

`bool` · optional (explicit presence)

Derive the chart's `crds/` directory at the pinned version and apply
each CRD ahead of the release, keyed by its name. Default TRUE. Set
false ONLY when those CRDs are owned elsewhere (another release, a
GitOps-managed bundle): the release still installs with the directory
skipped, and nothing is derived or checked.

- default: `true`

### spec.crds.keepOnUninstall

`bool` · optional (explicit presence)

Keep the module-owned CRDs (and therefore every custom resource built
on them, cluster-wide) when the resource is destroyed. Default TRUE:
deleting a CRD deletes every object of that type, a destructive act
that must be an explicit false. A later reinstall re-adopts kept CRDs.

- default: `true`

### spec.crds.allowHelmManaged

`bool` · optional (explicit presence)

Accept CRDs the chart templates as release resources WITHOUT
`helm.sh/resource-policy: keep`. Helm installs and upgrades them with
the release and deletes them when the release is uninstalled, along
with every custom resource built on them; nothing the module applies
can protect a resource Helm owns. Default FALSE: such a chart is
refused at plan time with the CRD names and the remedies (the typed
catalog kind for the chart if one exists; the chart's own keep switch
in its values; or this dial). CRDs the chart already marks keep are
never refused: the chart protects them itself.

- default: `false`

## Validation Rules

- `spec.atomic_vs_skip_await`: atomic and skip_await cannot both be set: atomic must wait for readiness to know whether to roll back. Drop one of the two.
- `spec.reuse_vs_reset_values`: reuse_values and reset_values are mutually exclusive: one keeps the previous release's values, the other discards them. Set at most one.
- `spec.repo_auth_pair`: repository_username and repository_password must be set together (both or neither) for private chart repositories.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesHelmRelease, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | The namespace the release is installed in. |
| `status.outputs.release_name` | `string` | The Helm release name (as shown by `helm list`) — `spec.release_name` when set, otherwise the resource's metadata.name. |
| `status.outputs.version` | `string` | The installed chart version (e.g. "6.9.2"). |
| `status.outputs.app_version` | `string` | The chart's appVersion — the upstream application version the chart packages (e.g. "6.9.2" for podinfo, "v1.18.2" for cert-manager). |
| `status.outputs.status` | `string` | The release status as Helm records it (e.g. "deployed"). |
| `status.outputs.revision` | `int32` | The release revision number (1 on first install, incremented by each upgrade or rollback). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
