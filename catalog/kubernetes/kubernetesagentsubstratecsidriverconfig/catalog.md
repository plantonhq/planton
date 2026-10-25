# Agent Substrate CSI Driver Config

Declares an Agent Substrate CSIDriverConfig: how Substrate reaches one CSI driver — its Controller endpoint, its node socket and how the connection is trusted — so an agent's external volume is provisioned and mounted by your storage instead of riding in the agent's snapshots. When an actor template asks for a volume of a StorageClass, Substrate finds this config by the class's provisioner and calls the driver itself. The spec is faithful to the upstream `ate.dev/v1alpha1` CSIDriverConfig (pinned to Agent Substrate v0.6.0) — this is how agents use the storage you already run.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **A CSIDriverConfig** — one cluster-scoped object, named after the driver, holding its endpoints and TLS settings.
- **Kubernetes Labels** — your own labels on the object, with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

The config registers a driver; it does not install one.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Agent Substrate installed** — it installs the CSIDriverConfig definition and runs the API server and node agents that call the driver.
- **The CSI driver running** — its Controller service reachable from the API server over TCP (usually a small bridge in front of its unix socket), and its node plugin bind-mounting `/var/lib/ate` with bidirectional propagation.
- **A StorageClass** whose provisioner is the driver's name, for actor templates to name.

## Deploy

### Console

Open the deployment store, find **Agent Substrate CSI Driver Config**, and click **Deploy**. The creation wizard walks you through the driver's name, its endpoints and TLS.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrateCsiDriverConfig
metadata:
  name: hostpath
  org: acme-corp
  env: prod
spec:
  name: hostpath.csi.k8s.io
  driverName: hostpath.csi.k8s.io
  controllerEndpoint: tcp://csi-hostpath-controller.default.svc.cluster.local:50051
  nodeSocketOverride: unix:///var/lib/kubelet/plugins/csi-hostpath/csi.sock
  tls:
    enabled: true
    usePodIdentity: true
    serverName: csi-hostpath-controller.default.svc
```

```shell
planton apply -f csi-driver-config.yaml
```

Actor templates whose external volumes name a StorageClass provisioned by `hostpath.csi.k8s.io` now get their volumes from that driver. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, declare the config beside the control plane and a Kubernetes Storage Class whose provisioner is the same driver name, so actor templates have a class to ask for:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesStorageClass
metadata:
  name: agent-volumes
spec:
  name: agent-volumes
  provisioner: hostpath.csi.k8s.io
```

The config's `driver_name` output carries the same name for anything else that needs it.

## Key Configuration

These are the most important decisions when configuring a CSI driver config. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The name is the driver's name.** Substrate finds a driver's config by its object name, so `name` must equal `driverName`.

**The Controller endpoint is TCP.** CSI drivers serve a unix socket; front the Controller with a small TCP bridge behind a Service, and point `controllerEndpoint` at it (`tcp://` or `dns://`).

**Encrypt with Substrate's pod identity.** With TLS on, the API server presents its Substrate identity and trusts Substrate's service-DNS signer; give the bridge a serving certificate from that signer and name it in `serverName`.

**Override the node socket only when needed.** The default is the kubelet convention, `unix:///var/lib/kubelet/plugins/<driverName>/csi.sock`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|-------------------|
| **Agent Substrate** | (prerequisite: the definition and the components that read it) | — |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `driver_name` | The driver's name, also the config's name | A StorageClass's `provisioner`, which actor templates' external volumes name |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**An encrypted driver connection** — the Controller endpoint over TLS with Substrate's pod identity. Start from the **TLS With Pod Identity** preset.

**An in-cluster plaintext driver** — a Controller service only the API server can reach. Start from the **In-Cluster Plaintext Driver** preset.

## Works With

- [**Agent Substrate**](/infra-catalog/kubernetes-agent-substrate) — the control plane whose API server and node agents call the driver.
- [**Agent Substrate Worker Pool**](/infra-catalog/kubernetes-agent-substrate-worker-pool) — the workers whose actors mount the volumes.
- [**Agent Substrate Sandbox Config**](/infra-catalog/kubernetes-agent-substrate-sandbox-config) — the runtime those actors boot.
- [**Kubernetes Storage Class**](/infra-catalog/kubernetes-storage-class) — the class actor templates name, with this driver as its provisioner.
