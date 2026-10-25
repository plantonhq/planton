# KubernetesAgentSubstrateSandboxConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSubstrateSandboxConfigSpec defines an Agent Substrate
SandboxConfig: the exact sandbox runtime an actor boots with, for one
sandbox class. A config lists versions, each an exact set of files the node
agent downloads (the gVisor release tarball, or a microVM's hypervisor,
kernel and root filesystem), checked by SHA-256. Every boot and every
restore of an actor uses the default version of the config its worker pool
resolves, as the config stands at that moment.

100% fidelity with the upstream SandboxConfig custom resource
(ate.dev/v1alpha1, cluster-scoped), pinned to the definition of the
Agent Substrate release KubernetesAgentSubstrate composes (v0.6.0) and held
to it by the mirror test. The definition's two rules and the per-class
asset rules of upstream's admission policy (sandboxconfig-assets) are
written below, so a config the cluster would refuse fails validation
before the apply.

Who reads it:
  - a worker pool resolves one config of each class it runs
    (KubernetesAgentSubstrateWorkerPool `sandbox_classes[].config_ref`, or
    the class default) and keeps it from deletion while it uses it; an
    actor runs in the class of the pool its template selects;
  - the node agent on every node downloads each config's default version
    ahead of time (a microVM config only where /dev/kvm exists), so an
    actor's boot does not wait for the download.

The control plane ships one: `gvisor-default`, gVisor's class default.
Declare another to pin a different gVisor release, to roll a new one out
beside the old (a second version, then `default_version` moved to it, the
old one kept listed so a rollback is moving it back), or to run microVM
sandboxes, whose files you stage in your own bucket.

## Example

```yaml
# The full surface, for offline proofs (tofu plan, pulumi preview, a server
# dry run): every field of the config, the same object the full-surface lane
# deploys.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateSandboxConfig
metadata:
  name: e2e-asubsc-full
spec:
  name: e2e-gvisor-rollout
  labels:
    team: platform
    runtime: gvisor
  annotations:
    planton.dev/e2e-note: full-surface
  sandboxClass: gvisor
  defaultVersion: gvisor-2026-10-17
  versions:
    - name: gvisor-2026-10-17
      state: Enabled
      assets:
        amd64:
          entries:
            gvisor:
              url: gs://gvisor/releases/nightly/2026-10-17/x86_64/gvisor.tar.zstd
              sha256: fa72f0d45756c69a30beec21222c78241b0d330c36f4f2185648617cc23415c8
        arm64:
          entries:
            gvisor:
              url: gs://gvisor/releases/nightly/2026-10-17/aarch64/gvisor.tar.zstd
              sha256: 238634978d99f0dd3e631c4f88eae7b19fb36937065428467f0b6d6b10697689
    # Retired, so no sandbox ever boots it and no node downloads its file:
    # the gVisor release before the one Agent Substrate's workers need.
    - name: gvisor-2026-10-10
      state: Disabled
      assets:
        amd64:
          entries:
            gvisor:
              url: gs://gvisor/releases/nightly/2026-10-10/x86_64/gvisor.tar.zstd
              sha256: 4ae7fab988c3647d2a835d6c3c70c6a99021c562db97827592562b6c51c3e88c
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.sandboxClass` | `string` | yes |  |  |
| `spec.defaultVersion` | `string` | yes |  |  |
| `spec.versions` | `[]KubernetesAgentSubstrateSandboxConfigVersion` | yes |  |  |
| `spec.versions[].name` | `string` | yes |  |  |
| `spec.versions[].state` | `string` |  |  |  |
| `spec.versions[].assets` | `map<string, KubernetesAgentSubstrateSandboxConfigAssetFiles>` |  |  |  |
| `spec.versions[].assets.*.entries` | `map<string, KubernetesAgentSubstrateSandboxConfigAssetFile>` |  |  |  |
| `spec.versions[].assets.*.entries.*.url` | `string` | yes |  |  |
| `spec.versions[].assets.*.entries.*.sha256` | `string` | yes |  |  |

## Field Details

### spec.name

`string` · required

The SandboxConfig's name in the cluster (its metadata.name): the value a
worker pool's `config_ref.name` names. A DNS subdomain name. Renaming
creates a new config and deletes the old one, which stays (Terminating)
while a worker pool still runs it.

- rule: name must be a valid DNS subdomain (lowercase alphanumeric, hyphens and dots, starting and ending with an alphanumeric)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Labels on the SandboxConfig object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the SandboxConfig object itself (its
metadata.annotations).

One annotation is Substrate's own switch:
`sandboxconfig.ate.dev/is-class-default: "true"` makes this config the
default for its class, the one a worker pool resolves when its
`config_ref` is unset. More than one config of a class may carry it;
the one created last wins (on a tie, the name that sorts first), and a
config being deleted is passed over. The control plane's `gvisor-default`
carries it for gVisor, so a gVisor config created with it takes over the
moment it is created. Two consequences: adding the annotation to a config
created earlier than `gvisor-default` never makes it win, and a control
plane reinstall recreates `gvisor-default`, which then wins again. The
value must be exactly "true".

### spec.sandboxClass

`string` · required

The sandbox runtime family this config is for. "gvisor" runs the actor
under gVisor inside its worker pod (no node changes needed). "microvm"
runs it in a cloud-hypervisor microVM with Kata's guest kernel and root
filesystem, and needs nodes that expose /dev/kvm. A worker pool only
resolves configs of its class.

Required here although the cluster would write "gvisor": the class decides
which assets every version needs, and a consumer reads it back as an
output to pair the config with a worker pool's `sandbox_classes` entry.

- rule: {"required":true,"string":{"in":["gvisor","microvm"]}}

### spec.defaultVersion

`string` · required

The version every actor boots and restores with: the name of an entry in
`versions` that is not Disabled. Moving it is how a new runtime rolls
out: an actor running now keeps its sandbox, and the next time it is
created or resumed, from a snapshot taken before the move too, it runs
the new default.

- rule: {"string":{"minLen":"1","maxLen":"63"}}

### spec.versions

`[]KubernetesAgentSubstrateSandboxConfigVersion` · required

The runtime versions this config lists, 1 to 16, each with a unique
name. At this release an actor only ever boots or restores the default
version, and only the default is downloaded ahead of time; the others
change nothing an actor runs until one is named the default, which is
what makes a rollback a one-field change.

- rule: {"repeated":{"minItems":"1","maxItems":"16"}}

### spec.versions[].name

`string` · required

The version's name within its config: a DNS label (lower-case
alphanumerics and '-', at most 63 characters), such as "v3" or
"gvisor-2026-10-10".

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.versions[].state

`string` · optional (explicit presence)

Whether the version may be used. The version `default_version` names
cannot be Disabled, so "Disabled" marks a listed version that must not
be made the default again; the node agent also leaves Disabled versions
out of the runtimes it records for each worker. Unset, the cluster
writes "Enabled".

- rule: {"string":{"in":["Enabled","Disabled"]}}

### spec.versions[].assets

`map<string, KubernetesAgentSubstrateSandboxConfigAssetFiles>`

The files the node agent downloads for this version, keyed by the node's
architecture (Go's names: "amd64", "arm64") and then by asset name. A
node uses the block for its own architecture; a version with no block for
a node's architecture is skipped there, except the default, which every
worker must be able to boot.

The asset names are the runtime's: gVisor needs "gvisor" (the release's
gvisor.tar.zstd, extracted so runsc's helpers sit beside it); microVM
needs "cloud-hypervisor", "virtiofsd", "kata-kernel" and "kata-image".
The names stay a free map because upstream keeps them generic, so a
future runtime adds assets without a schema change.

A manifest writes each architecture's files under `entries`
(`amd64: {entries: {gvisor: {url: ..., sha256: ...}}}`); the object
carries the bare map upstream defines.

### spec.versions[].assets.*.entries

`map<string, KubernetesAgentSubstrateSandboxConfigAssetFile>`

The architecture's files, keyed by asset name.

### spec.versions[].assets.*.entries.*.url

`string` · required

Where the file is downloaded from, at most 2048 characters. The node
agent fetches every URL through its object-store plugin with the
plugin's credentials; that plugin serves only its own backend's scheme
(gs:// on GCS, s3:// on S3), and any gs:// URL it cannot serve is read
anonymously, which is how public buckets such as gs://gvisor work from
any cluster with internet egress. A URL under the control plane's
snapshot location is refused: stage runtime files outside it.

- rule: {"string":{"minLen":"1","maxLen":"2048"}}

### spec.versions[].assets.*.entries.*.sha256

`string` · required

The file's lower-case hex SHA-256. It names the cached copy on each node
and verifies every download, so a changed file at the same URL is refused
rather than booted.

- rule: {"required":true,"string":{"pattern":"^[a-f0-9]{64}$"}}

## Validation Rules

- `spec.default_version.listed`: default_version must name an entry in versions
- `spec.default_version.enabled`: the version named by default_version must not be Disabled
- `spec.assets.gvisor`: a gvisor config must define a 'gvisor' asset (the release tarball) for every architecture of every version
- `spec.assets.microvm`: a microvm config must define cloud-hypervisor, virtiofsd, kata-kernel and kata-image assets for every architecture of every version
- `spec.versions.unique_names`: each version needs a unique name

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSubstrateSandboxConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.sandbox_config_name` | `string` | Name of the created SandboxConfig (the spec's `name`, its metadata.name in the cluster). |
| `status.outputs.sandbox_class` | `string` | The sandbox class the config is for: "gvisor" or "microvm". |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAgentSubstrateWorkerPool | `spec.sandboxClasses[].configRef.name` | `status.outputs.sandbox_config_name` |

## See Also

- [Overview](../README.md)
