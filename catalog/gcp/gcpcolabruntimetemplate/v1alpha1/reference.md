# GcpColabRuntimeTemplate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpColabRuntimeTemplateSpec defines a Colab Enterprise runtime template
(`google_colab_runtime_template`) -- the machine, disk, network, image,
and security settings every notebook runtime created from it gets. Admins
publish templates; data scientists pick one in Colab Enterprise (or a
GcpColabRuntime / GcpColabSchedule names it) and get a runtime that
already meets the team's rules: the right machine and GPU, a private
network, customer-managed encryption, idle shutdown to cap spend.

Replaced on change: location, runtime_template_id, description, labels
(Google fixes a template's labels at creation), network_tags, and every
machine, disk, and network setting. Fixed at creation, and never
changed by a deploy: idle_timeout, euc_disabled, and enable_secure_boot
(see each field) -- a different value needs a new template. Mutable in
place: display_name, kms_key_name, and software_config.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntimeTemplate
metadata:
  name: standard-runtime
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Standard runtime
  description: Everyday notebooks on a mid-size CPU machine
  machineSpec:
    machineType: e2-standard-4
  dataPersistentDiskSpec:
    diskType: pd-balanced
    diskSizeGb: 100
  networkSpec:
    enableInternetAccess: true
  # Stop idle runtimes after an hour -- the main cost control.
  idleTimeout: 3600s
  enableSecureBoot: true
  labels:
    team: data-science
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.runtimeTemplateId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.machineSpec` | `GcpColabRuntimeTemplateMachineSpec` |  |  |  |
| `spec.machineSpec.machineType` | `string` |  |  |  |
| `spec.machineSpec.acceleratorType` | `string` |  |  |  |
| `spec.machineSpec.acceleratorCount` | `int32` |  |  |  |
| `spec.dataPersistentDiskSpec` | `GcpColabRuntimeTemplateDataPersistentDiskSpec` |  |  |  |
| `spec.dataPersistentDiskSpec.diskType` | `string` |  |  |  |
| `spec.dataPersistentDiskSpec.diskSizeGb` | `int64` |  |  |  |
| `spec.networkSpec` | `GcpColabRuntimeTemplateNetworkSpec` | yes |  |  |
| `spec.networkSpec.enableInternetAccess` | `bool` |  |  |  |
| `spec.networkSpec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.networkSpec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.idleTimeout` | `string` |  |  |  |
| `spec.eucDisabled` | `bool` |  |  |  |
| `spec.enableSecureBoot` | `bool` |  |  |  |
| `spec.networkTags` | `[]string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.softwareConfig` | `GcpColabRuntimeTemplateSoftwareConfig` |  |  |  |
| `spec.softwareConfig.env` | `[]GcpColabRuntimeTemplateEnvVar` |  |  |  |
| `spec.softwareConfig.env[].name` | `string` | yes |  |  |
| `spec.softwareConfig.env[].value` | `string` |  |  |  |
| `spec.softwareConfig.postStartupScriptConfig` | `GcpColabRuntimeTemplatePostStartupScriptConfig` |  |  |  |
| `spec.softwareConfig.postStartupScriptConfig.postStartupScript` | `string` |  |  |  |
| `spec.softwareConfig.postStartupScriptConfig.postStartupScriptUrl` | `string` |  |  |  |
| `spec.softwareConfig.postStartupScriptConfig.postStartupScriptBehavior` | `string` |  |  |  |
| `spec.softwareConfig.colabImage` | `GcpColabRuntimeTemplateColabImage` |  |  |  |
| `spec.softwareConfig.colabImage.releaseName` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the template lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Colab Enterprise region, e.g. "us-central1". Runtimes created from
the template run here. Changing it replaces the template: both modules
force the replacement, since Google's provider would otherwise try to
move it in place.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.runtimeTemplateId

`string`

The template's id -- the last segment of its resource name. Lowercase
letters, digits, and hyphens. Defaults to metadata.name. Immutable.

- rule: runtime_template_id must be lowercase letters, digits, and hyphens, starting with a letter or digit, up to 128 characters

### spec.displayName

`string`

The name users see when they pick a runtime template -- up to 128
characters. Defaults to metadata.name. Mutable.

- rule: {"string":{"maxLen":"128"}}

### spec.description

`string`

What the template is for (e.g. "GPU runtime for the vision team").
Immutable.

### spec.labels

`map<string, string>`

Labels on the template. The platform attribution labels (the
organization and environment among them) are merged in and win on key
conflicts. A label change replaces the template -- moving the resource
to another environment changes its attribution labels, so it replaces
the template too.

### spec.machineSpec

`GcpColabRuntimeTemplateMachineSpec`

The machine every runtime gets. Omit for Google's default machine.

- rule: accelerator_type needs accelerator_count

### spec.machineSpec.machineType

`string`

The Compute Engine machine type, e.g. "e2-standard-4" or
"n1-standard-8" (GPUs attach to N1 and the accelerator-optimized
families). Unset keeps Google's default.

### spec.machineSpec.acceleratorType

`string`

The accelerator, e.g. "NVIDIA_TESLA_T4", "NVIDIA_L4",
"NVIDIA_TESLA_A100". Needs accelerator_count.

### spec.machineSpec.acceleratorCount

`int32`

How many accelerators each runtime gets.

- rule: {"int32":{"gte":0}}

### spec.dataPersistentDiskSpec

`GcpColabRuntimeTemplateDataPersistentDiskSpec`

The runtime's data disk (mounted as the home directory). Omit for
Google's default disk.

- rule: disk_size_gb needs disk_type

### spec.dataPersistentDiskSpec.diskType

`string`

"pd-standard", "pd-balanced", "pd-ssd", or "pd-extreme". Unset keeps
Google's default.

- rule: disk_type must be pd-standard, pd-balanced, pd-ssd, or pd-extreme

### spec.dataPersistentDiskSpec.diskSizeGb

`int64`

Disk size in GB, 10 to 65536. Needs disk_type.

- rule: disk_size_gb must be between 10 and 65536

### spec.networkSpec

`GcpColabRuntimeTemplateNetworkSpec` · required

The network the runtime joins. Required: a template that sends no
network spec has no internet access, and Google then refuses it unless
it names both a network and a subnetwork ("Both [networkSpec.network]
and [networkSpec.subnetwork] fields must be provided if Internet Access
is disabled"). Turn internet access on, or name both.

- rule: {"required":true}
- rule: a network_spec without internet access must name both a network and a subnetwork: Google refuses a private template that leaves either out

### spec.networkSpec.enableInternetAccess

`bool`

Give runtimes a public internet path. Turn off for runtimes that must
stay private; then name the network and the subnetwork, which needs
Private Google Access.

### spec.networkSpec.network

`string | valueFrom`

The VPC network: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}. Unset: the project's network
named "default", which many organizations never create -- in a project
without one, Google refuses the template, so name a network.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.networkSpec.subnetwork

`string | valueFrom`

The subnetwork in the template's region: a GcpSubnetwork reference or
a literal projects/{project}/regions/{region}/subnetworks/{name}.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.idleTimeout

`string`

Shut an idle runtime down after this long -- the main cost control for
interactive notebooks. A duration in seconds ending in "s": "0s"
disables idle shutdown; otherwise between 600s (10 minutes) and 86400s
(24 hours). Unset keeps Google's default idle shutdown. Fixed at
creation: Google's provider never sends a changed value, so a change
shows on every plan and Google keeps the first one -- create a new
template for a different timeout.

- rule: idle_timeout must be a duration in seconds such as 3600s

### spec.eucDisabled

`bool`

Block the user's own Google credentials inside the runtime, so
notebooks act only as the runtime's service identity. Sent only when
true. Fixed at creation: turning it off later sends nothing, and
Google keeps it on -- create a new template instead.

### spec.enableSecureBoot

`bool`

Boot runtimes with Secure Boot (Shielded VM), refusing unsigned boot
components. Sent only when true. Fixed at creation: turning it off
later sends nothing, and Google keeps it on -- create a new template
instead.

### spec.networkTags

`[]string`

Compute Engine network tags applied to every runtime -- the handle VPC
firewall rules target. Immutable.

- rule: {"repeated":{"unique":true}}

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the runtimes' disks: a
GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the template's region. The Compute Engine and Vertex AI service
agents need roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Mutable
(applies to runtimes created afterwards).

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.softwareConfig

`GcpColabRuntimeTemplateSoftwareConfig`

The notebook software: environment variables, a post-startup script,
and the Colab image release. Mutable.

### spec.softwareConfig.env

`[]GcpColabRuntimeTemplateEnvVar`

Environment variables set in the notebook container.

### spec.softwareConfig.env[].name

`string` · required

The variable name -- a valid C identifier.

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_]*$"}}

### spec.softwareConfig.env[].value

`string`

The value. $(VAR_NAME) expands a previously defined variable; $$(...)
escapes the expansion. Stored in the template in plain text, where
anyone who can view the template or a runtime made from it reads it:
never a secret. The template cannot hold one safely: a runtime runs
as whoever creates it (end-user credentials by default), so no grant
made with the template could open a stored secret to its runtimes.
Notebook code reads credentials at run time instead -- from Secret
Manager with the runtime user's own access, or from Colab's own
secrets panel.

### spec.softwareConfig.postStartupScriptConfig

`GcpColabRuntimeTemplatePostStartupScriptConfig`

A script run after the runtime starts (installing packages, mounting
data, configuring proxies).

### spec.softwareConfig.postStartupScriptConfig.postStartupScript

`string`

The script itself, inline.

### spec.softwareConfig.postStartupScriptConfig.postStartupScriptUrl

`string`

A URL the script is downloaded from, e.g.
gs://bucket/setup.sh or https://example.com/setup.sh.

### spec.softwareConfig.postStartupScriptConfig.postStartupScriptBehavior

`string`

When the script runs:
  "RUN_ONCE"                     -- on the first start only
  "RUN_EVERY_START"              -- on every start
  "DOWNLOAD_AND_RUN_EVERY_START" -- re-downloaded and run on every
                                    start (picks up script changes)

- rule: post_startup_script_behavior must be RUN_ONCE, RUN_EVERY_START, or DOWNLOAD_AND_RUN_EVERY_START

### spec.softwareConfig.colabImage

`GcpColabRuntimeTemplateColabImage`

Which Colab image release the runtime boots.

### spec.softwareConfig.colabImage.releaseName

`string`

The image release, e.g. "py310". Unset: the latest release.

### spec.deletionPolicy

`string`

What happens to the template when this resource is destroyed:
  "" / "DELETE" -- the template is deleted (runtimes already created
                   from it keep running)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the template leaves management and stays in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpColabRuntimeTemplate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/notebookRuntimeTemplates/{runtime_template_id} -- what GcpColabRuntime.runtime_template and GcpColabSchedule's notebook arm take. |
| `status.outputs.runtime_template_id` | `string` | The template's id (the last segment of name). |
| `status.outputs.location` | `string` | The region the template lives in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.networkSpec.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.networkSpec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpColabRuntime | `spec.runtimeTemplate` | `status.outputs.name` |
| GcpColabSchedule | `spec.notebookExecutionJob.notebookRuntimeTemplateResourceName` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
