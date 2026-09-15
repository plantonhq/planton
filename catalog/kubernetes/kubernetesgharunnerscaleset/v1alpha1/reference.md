# KubernetesGhaRunnerScaleSet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesGhaRunnerScaleSetSpec** declares an autoscaling fleet of
self-hosted GitHub Actions runners for one GitHub repository,
organization or enterprise — from the official
`gha-runner-scale-set` chart (OCI,
ghcr.io/actions/actions-runner-controller-charts).

PREREQUISITE: the runner scale set controller must be installed
first (declare a KubernetesGhaRunnerScaleSetController resource).
This kind renders an AutoscalingRunnerSet; the controller runs a
listener that long-polls GitHub for queued jobs and creates one
EPHEMERAL runner pod per job — each runner executes exactly one job
and is replaced.

HOW WORKFLOWS TARGET THIS FLEET: by NAME. The scale set registers in
GitHub under `runner_scale_set_name` (default: this resource's
metadata.name), and workflows select it with
`runs-on: <that name>` — runner labels beyond the name are not how
routing works in scale sets. The name may be at most 45 characters.

AUTHENTICATION IS SECRET-NATIVE: the GitHub credential (a PAT or a
GitHub App) always lives in a Kubernetes Secret. Either reference an
existing Secret (`auth.existing_secret_name` — the recommended
posture) or declare the credential inline (`auth.pat` /
`auth.github_app`) and the module materializes the Secret — inline
values are marked sensitive and never rendered into chart values
(the chart reads the Secret by name).

DOCKER BUILDS NEED A CONTAINER MODE: the default runner runs plain
jobs only. Set `container_mode.mode: dind` for `docker build`-style
jobs (runs a privileged Docker-in-Docker sidecar — the cluster must
allow privileged pods) or `kubernetes` for container jobs via the
Kubernetes container hook (no privilege, but requires a work volume
with a dynamic-provisioning StorageClass).

## Example

```yaml
# Full-surface hack manifest for the offline plan/preview proofs: the
# GitHub App declared arm (the module materializes the Secret), dind
# mode, proxy, GHES private CA, an explicit controller reference, and
# the escape hatch — every typed block at once.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGhaRunnerScaleSet
metadata:
  name: org-runners-full
spec:
  namespace:
    value: ci-runners
  createNamespace: true
  chartVersion: 0.14.2
  githubConfigUrl: https://github.example.com/my-org
  auth:
    githubApp:
      appId: "123456"
      installationId: "654321"
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        hack-manifest-placeholder-never-a-real-key
        -----END RSA PRIVATE KEY-----
  runnerScaleSetName: org-runners
  runnerGroup: platform
  minRunners: 2
  maxRunners: 30
  containerMode:
    mode: dind
  runner:
    image: ghcr.io/actions/actions-runner:2.321.0
    resources:
      requests:
        cpu: "1"
        memory: 2Gi
      limits:
        cpu: "4"
        memory: 8Gi
  proxy:
    https:
      url: http://proxy.example.com:8080
      credentialSecretName: proxy-auth
    noProxy:
      - svc.cluster.local
  githubServerTls:
    configMapName:
      value: ghes-ca
    key: ca.crt
    runnerMountPath: /usr/local/share/ca-certificates/
  controllerServiceAccount:
    namespace: arc-system
    name: arc
  helmValues: |
    listenerTemplate:
      spec:
        containers:
          - name: listener
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.14.2` |  |
| `spec.githubConfigUrl` | `string` | yes |  |  |
| `spec.auth` | `KubernetesGhaRunnerScaleSetAuth` | yes |  |  |
| `spec.auth.existingSecretName` | `string` |  |  |  |
| `spec.auth.pat` | `KubernetesGhaRunnerScaleSetAuthPat` |  |  |  |
| `spec.auth.pat.token` | `string` (sensitive) | yes |  |  |
| `spec.auth.githubApp` | `KubernetesGhaRunnerScaleSetAuthGithubApp` |  |  |  |
| `spec.auth.githubApp.appId` | `string` | yes |  |  |
| `spec.auth.githubApp.installationId` | `string` | yes |  |  |
| `spec.auth.githubApp.privateKey` | `string` (sensitive) | yes |  |  |
| `spec.runnerScaleSetName` | `string` |  |  |  |
| `spec.runnerGroup` | `string` |  |  |  |
| `spec.minRunners` | `int32` |  |  |  |
| `spec.maxRunners` | `int32` |  |  |  |
| `spec.containerMode` | `KubernetesGhaRunnerScaleSetContainerMode` |  |  |  |
| `spec.containerMode.mode` | `string` | yes |  |  |
| `spec.containerMode.kubernetesWorkVolume` | `KubernetesGhaRunnerScaleSetWorkVolume` |  |  |  |
| `spec.containerMode.kubernetesWorkVolume.storageClass` | `string \| valueFrom` | yes |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.containerMode.kubernetesWorkVolume.size` | `string` | yes |  |  |
| `spec.runner` | `KubernetesGhaRunnerScaleSetRunner` |  |  |  |
| `spec.runner.image` | `string` |  |  |  |
| `spec.runner.resources` | `ContainerResources` |  |  |  |
| `spec.runner.resources.limits` | `CpuMemory` |  |  |  |
| `spec.runner.resources.limits.cpu` | `string` |  |  |  |
| `spec.runner.resources.limits.memory` | `string` |  |  |  |
| `spec.runner.resources.requests` | `CpuMemory` |  |  |  |
| `spec.runner.resources.requests.cpu` | `string` |  |  |  |
| `spec.runner.resources.requests.memory` | `string` |  |  |  |
| `spec.proxy` | `KubernetesGhaRunnerScaleSetProxy` |  |  |  |
| `spec.proxy.http` | `KubernetesGhaRunnerScaleSetProxyServer` |  |  |  |
| `spec.proxy.http.url` | `string` | yes |  |  |
| `spec.proxy.http.credentialSecretName` | `string` |  |  |  |
| `spec.proxy.https` | `KubernetesGhaRunnerScaleSetProxyServer` |  |  |  |
| `spec.proxy.https.url` | `string` | yes |  |  |
| `spec.proxy.https.credentialSecretName` | `string` |  |  |  |
| `spec.proxy.noProxy` | `[]string` |  |  |  |
| `spec.githubServerTls` | `KubernetesGhaRunnerScaleSetGithubServerTls` |  |  |  |
| `spec.githubServerTls.configMapName` | `string \| valueFrom` | yes |  | KubernetesConfigMap (`metadata.name`) |
| `spec.githubServerTls.key` | `string` |  | `ca.crt` |  |
| `spec.githubServerTls.runnerMountPath` | `string` |  |  |  |
| `spec.controllerServiceAccount` | `KubernetesGhaRunnerScaleSetControllerRef` |  |  |  |
| `spec.controllerServiceAccount.namespace` | `string` | yes |  |  |
| `spec.controllerServiceAccount.name` | `string` | yes |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "0.14.2"). Keep it EQUAL to
the controller's chart_version — GitHub supports controller and
scale set charts only at matching versions. Versions must exist
as SERVED charts in the OCI registry.

- default: `0.14.2`

### spec.githubConfigUrl

`string` · required

GitHub URL the runners register against — a repository
("https://github.com/my-org/my-repo"), an organization
("https://github.com/my-org") or an enterprise
("https://github.com/enterprises/my-enterprise"). GitHub
Enterprise Server URLs work the same way.

- rule: github_config_url must be an http(s) URL of a repository, organization or enterprise
- rule: {"required":true}

### spec.auth

`KubernetesGhaRunnerScaleSetAuth` · required

GitHub credential the listener authenticates with. A GitHub App
is the production posture (fine-grained, expiring tokens); a PAT
is the quick start.

- rule: {"required":true}
- rule: declare exactly one auth method — existing_secret_name, pat, or github_app

### spec.auth.existingSecretName

`string`

Existing Secret in the install namespace holding the
credential — key `github_token` for a PAT, or keys
`github_app_id`, `github_app_installation_id`,
`github_app_private_key` for a GitHub App. The recommended
posture: the credential never rides a manifest.

### spec.auth.pat

`KubernetesGhaRunnerScaleSetAuthPat`

Declared PAT — the module materializes it as a Secret; the
value never lands in rendered chart values.

### spec.auth.pat.token

`string` · required · sensitive

The token (e.g. "ghp_..."). Sensitive — materialized into a
Secret, never rendered into chart values.

- rule: {"required":true}

### spec.auth.githubApp

`KubernetesGhaRunnerScaleSetAuthGithubApp`

Declared GitHub App credential — the module materializes it as
a Secret; the values never land in rendered chart values.

### spec.auth.githubApp.appId

`string` · required

The App ID (or client ID), as a string.

- rule: {"required":true}

### spec.auth.githubApp.installationId

`string` · required

The installation ID of the App on the target org/repo, as a
string.

- rule: {"required":true}

### spec.auth.githubApp.privateKey

`string` · required · sensitive

The App's PEM private key. Sensitive — materialized into a
Secret, never rendered into chart values.

- rule: {"required":true}

### spec.runnerScaleSetName

`string`

Name this fleet registers under in GitHub — the exact value
workflows put in `runs-on`. Empty = this resource's
metadata.name. At most 45 characters (a GitHub limit the chart
enforces).

- rule: {"string":{"maxLen":"45"}}

### spec.runnerGroup

`string`

Runner group this fleet joins. Runner groups exist on
organizations/enterprises to control which repositories may use
the fleet; the group must already exist in GitHub. Empty =
"default".

### spec.minRunners

`int32` · optional (explicit presence)

Minimum number of IDLE runners kept warm. The fleet holds
min_runners + assigned jobs. Empty = 0 (fully scale-to-zero —
cold-start latency of one pod schedule per job).

- rule: {"int32":{"gte":0}}

### spec.maxRunners

`int32` · optional (explicit presence)

Maximum number of runners the fleet scales to; additional queued
jobs wait in GitHub. Empty = unbounded.

- rule: {"int32":{"gte":0}}

### spec.containerMode

`KubernetesGhaRunnerScaleSetContainerMode`

How runner pods run jobs. Empty = the plain runner (no Docker
daemon, no container jobs — shell/tool jobs only).

- rule: kubernetes container mode requires kubernetes_work_volume (dind and kubernetes-novolume must not declare it)

### spec.containerMode.mode

`string` · required

`dind` = a privileged Docker-in-Docker sidecar per runner —
`docker build`/`docker run` steps work; the cluster must allow
privileged pods. `kubernetes` = container jobs run as separate
pods via the Kubernetes container hook — no privilege, requires
`kubernetes_work_volume`. `kubernetes-novolume` = the same hook
without a shared work volume (jobs that never share files
between containers).

- rule: {"required":true,"string":{"in":["dind","kubernetes","kubernetes-novolume"]}}

### spec.containerMode.kubernetesWorkVolume

`KubernetesGhaRunnerScaleSetWorkVolume`

The per-runner work volume for `kubernetes` mode (ignored
otherwise). Each runner pod gets one ephemeral PersistentVolumeClaim
of this shape; the StorageClass must provision dynamically.

### spec.containerMode.kubernetesWorkVolume.storageClass

`string | valueFrom` · required

StorageClass to provision from. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Must support
dynamic provisioning.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.containerMode.kubernetesWorkVolume.size

`string` · required

Requested size per runner (e.g. "1Gi").

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]+)?(Ei|Pi|Ti|Gi|Mi|Ki|E|P|T|G|M|k|m)?$"}}

### spec.runner

`KubernetesGhaRunnerScaleSetRunner`

The runner container.

### spec.runner.image

`string`

Runner image (full reference). Empty =
"ghcr.io/actions/actions-runner:latest" (the chart default —
pin a tag on production fleets; "latest" changes under you).

### spec.runner.resources

`ContainerResources`

CPU and memory for the runner container — sized for the JOBS the
fleet runs, not for the runner agent (builds inherit these
limits). Empty = no requests/limits.

### spec.runner.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.runner.resources.limits.cpu

`string`

### spec.runner.resources.limits.memory

`string`

### spec.runner.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.runner.resources.requests.cpu

`string`

### spec.runner.resources.requests.memory

`string`

### spec.proxy

`KubernetesGhaRunnerScaleSetProxy`

Outbound proxy for the listener and runner pods (clusters that
reach GitHub through a proxy).

### spec.proxy.http

`KubernetesGhaRunnerScaleSetProxyServer`

Proxy for HTTP traffic.

### spec.proxy.http.url

`string` · required

Proxy URL (e.g. "http://proxy.example.com:8080").

- rule: {"required":true}

### spec.proxy.http.credentialSecretName

`string`

Existing Secret with keys `username` and `password` when the
proxy requires authentication.

### spec.proxy.https

`KubernetesGhaRunnerScaleSetProxyServer`

Proxy for HTTPS traffic.

### spec.proxy.https.url

`string` · required

Proxy URL (e.g. "http://proxy.example.com:8080").

- rule: {"required":true}

### spec.proxy.https.credentialSecretName

`string`

Existing Secret with keys `username` and `password` when the
proxy requires authentication.

### spec.proxy.noProxy

`[]string`

Hosts that bypass the proxy (e.g. in-cluster service names).

### spec.githubServerTls

`KubernetesGhaRunnerScaleSetGithubServerTls`

Trust a private CA when talking to a GitHub Enterprise Server
with a self-signed certificate.

### spec.githubServerTls.configMapName

`string | valueFrom` · required

ConfigMap holding the CA certificate. Accepts a literal name or a
reference to a KubernetesConfigMap resource.

- references: KubernetesConfigMap (`metadata.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesConfigMap, name: <that resource's name>, fieldPath: metadata.name}} -- a bare string does not parse

### spec.githubServerTls.key

`string` · optional (explicit presence)

Key in the ConfigMap holding the PEM certificate. Empty =
"ca.crt".

- default: `ca.crt`

### spec.githubServerTls.runnerMountPath

`string`

Mount path in runner pods (e.g.
"/usr/local/share/ca-certificates/"). When set, runners also get
NODE_EXTRA_CA_CERTS pointed at the mounted certificate.

### spec.controllerServiceAccount

`KubernetesGhaRunnerScaleSetControllerRef`

The controller install serving this fleet. LEAVE EMPTY on
clusters with one cluster-wide controller (the chart discovers it
at install time). REQUIRED when the controller was fenced with
`flags.watch_single_namespace` — auto-discovery cannot see it.

### spec.controllerServiceAccount.namespace

`string` · required

Namespace the controller runs in.

- rule: {"required":true}

### spec.controllerServiceAccount.name

`string` · required

The controller's ServiceAccount name — a
KubernetesGhaRunnerScaleSetController exports it as
`service_account_name`.

- rule: {"required":true}

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields — the full runner pod `template` PodSpec,
`listenerTemplate` sidecars, `listenerMetrics` histogram buckets,
per-resource metadata — never the substitute for them. Do not put
secrets here; the GitHub credential belongs in `auth`.

## Validation Rules

- `spec.runner_bounds`: max_runners must be greater than or equal to min_runners

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGhaRunnerScaleSet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the scale set (listener + runner pods) runs in. |
| `status.outputs.release_name` | `string` | Helm release name (equals metadata.name). |
| `status.outputs.runner_scale_set_name` | `string` | Name the fleet registered under in GitHub — the exact value workflows put in `runs-on:` to target this fleet. |
| `status.outputs.github_config_url` | `string` | The GitHub URL the fleet serves (repository, organization or enterprise). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.containerMode.kubernetesWorkVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.githubServerTls.configMapName` | KubernetesConfigMap | `metadata.name` |

## See Also

- [Overview](../README.md)
