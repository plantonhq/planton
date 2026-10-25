# KubernetesAgentSubstrateCsiDriverConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

KubernetesAgentSubstrateCsiDriverConfigSpec defines an Agent Substrate
CSIDriverConfig: how Substrate reaches one CSI driver, so an actor's
external volume is provisioned and mounted by your storage rather than kept
in the actor's snapshots. When an actor template asks for an external
volume of a StorageClass, Substrate reads the class's provisioner (the
driver's name), finds this config by that name, and calls the driver
itself: the API server reaches the driver's Controller service at
`controller_endpoint`, and each node agent reaches its node's Node service
over the kubelet plugin socket. The driver must already run on the cluster;
this config registers it with Substrate, it does not install it. External
volumes are a preview: the API server serves them only with the
ExternalVolumes gate in the control plane's `preview_gates`.

The API server and every node agent read a driver's config the first time
an actor needs the driver, and keep the client they build from it until
they restart. A change to this config, or its deletion, therefore reaches
a driver already in use only after both restart:
`kubectl -n <control plane namespace> rollout restart
deployment/ate-api-server daemonset/atelet`.

100% fidelity with the upstream CSIDriverConfig custom resource
(ate.dev/v1alpha1, cluster-scoped), pinned to the definition of the Agent
Substrate release KubernetesAgentSubstrate composes (v0.6.0) and held to it
by the mirror test. The definition's rules are written below,
with one more: the name must equal `driver_name`, because both the API
server and the node agent look a driver's config up by its object name, so
a config named anything else is never found.

## Example

```yaml
# The full surface, for offline proofs (tofu plan, pulumi preview, a server
# dry run): every field of the config, the same object the full-surface lane
# deploys.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateCsiDriverConfig
metadata:
  name: e2e-asubcsi-full
spec:
  name: hostpath.csi.k8s.io
  labels:
    team: storage
  annotations:
    planton.dev/e2e-note: full-surface
  driverName: hostpath.csi.k8s.io
  controllerEndpoint: tcp://csi-hostpath-controller.default.svc.cluster.local:50051
  nodeSocketOverride: unix:///var/lib/kubelet/plugins/csi-hostpath/csi.sock
  tls:
    enabled: true
    usePodIdentity: true
    serverName: csi-hostpath-controller.default.svc
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.driverName` | `string` | yes |  |  |
| `spec.controllerEndpoint` | `string` | yes |  |  |
| `spec.nodeSocketOverride` | `string` |  |  |  |
| `spec.tls` | `KubernetesAgentSubstrateCsiDriverConfigTls` |  |  |  |
| `spec.tls.enabled` | `bool` | yes |  |  |
| `spec.tls.usePodIdentity` | `bool` |  |  |  |
| `spec.tls.serverName` | `string` |  |  |  |

## Field Details

### spec.name

`string` · required

The CSIDriverConfig's name in the cluster (its metadata.name), which must
equal `driver_name`.

- rule: name must be a valid DNS subdomain (lowercase alphanumeric, hyphens and dots, starting and ending with an alphanumeric)
- rule: {"string":{"minLen":"1","maxLen":"253"}}

### spec.labels

`map<string, string>`

Labels on the CSIDriverConfig object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the CSIDriverConfig object itself (its
metadata.annotations).

### spec.driverName

`string` · required

The driver's standard CSI name, such as "hostpath.csi.k8s.io": what its
StorageClasses name as `provisioner` and what the driver reports from
GetPluginInfo (Substrate checks the two agree). At most 63 characters.

Upstream's pattern also admits a "substrate.io/" prefix, the namespace of
the volume drivers built into Substrate, which need no CSIDriverConfig.
No object name can carry a "/", so the name rule above refuses it here.

- rule: {"string":{"minLen":"1","maxLen":"63","pattern":"^(substrate\\.io/)?([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*)$"}}

### spec.controllerEndpoint

`string` · required

The gRPC address of the driver's CSI Controller service, which the API
server dials to create and attach volumes: `tcp://host:port` or
`dns:///service.namespace.svc:port`. CSI drivers serve a unix socket, so
an in-cluster driver is usually fronted by a small TCP bridge (a socat
sidecar) behind a Service.

- rule: {"required":true,"string":{"pattern":"^(tcp|dns)://.+$"}}

### spec.nodeSocketOverride

`string`

The driver's Node service socket on every worker node, when it is not
the kubelet convention `unix:///var/lib/kubelet/plugins/<driver_name>/csi.sock`
(a driver that registers under another directory name, for example).
Must start with `unix://`.

- rule: node_socket_override must be a unix:// socket address

### spec.tls

`KubernetesAgentSubstrateCsiDriverConfigTls`

TLS for the connection to `controller_endpoint`. Unset, the API server
dials the endpoint in plaintext.

- rule: use_pod_identity must be true when TLS is enabled; certificates from Secrets are not supported at this release
- rule: server_name is required when TLS is enabled

### spec.tls.enabled

`bool` · required · optional (explicit presence)

Whether the connection uses TLS. Always written, so false reaches the
object as false.

- rule: {"required":true}

### spec.tls.usePodIdentity

`bool`

Present Substrate's pod identity (its SPIFFE certificate, issued by the
control plane's pod-certificate controller) as the client certificate,
and trust Substrate's service-DNS signer for the driver's serving
certificate. The only mode at this release, so required with TLS: the
driver's endpoint must serve a certificate from that signer (a
pod-certificate projected volume with signer
`servicedns.podcert.ate.dev/identity`).

### spec.tls.serverName

`string`

The name the driver's serving certificate must carry, such as
`csi-hostpath-controller.default.svc`. Required with TLS.

## Validation Rules

- `spec.name.is_driver_name`: name must equal driver_name: Substrate finds a driver's config by its object name, so a config named anything else is never found

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesAgentSubstrateCsiDriverConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.driver_name` | `string` | The CSI driver's name, which is also the config's name in the cluster. A StorageClass names it as its `provisioner`, and an actor template's external volume names that StorageClass; Substrate then finds this config by the driver name and dials the driver. |

## See Also

- [Overview](../README.md)
