# KubernetesAgentSubstrateCsiDriverConfig Guide

The judgment this guide carries: an external volume is the one piece of an actor's state that lives outside its snapshots, on storage you choose. A CSIDriverConfig is the line from Substrate to that storage: who Substrate calls, at which address, and how it trusts the answer.

## When to use external volumes

An actor's durable directories are captured in its snapshots and travel with it into the control plane's bucket. That suits an agent's working files. An external volume is provisioned by your driver for one actor and lives exactly as long as that actor: it suits per-actor data a snapshot should not carry, such as large scratch space, data your storage encrypts with its own keys, or a performance class your storage offers. It is never shared between actors and never a volume that existed before the actor. Declare one CSIDriverConfig per driver you want actors to use, and a StorageClass that names the driver as its provisioner; templates then ask for volumes of that class.

## How the pieces connect

- **The StorageClass** is the user-facing name. Its `provisioner` is the driver's CSI name.
- **The CSIDriverConfig** is found by that name: `name` must equal `driver_name`, because the API server and the node agents look the config up by its object name. (Upstream's grammar also admits a `substrate.io/` prefix, which names the volume drivers built into Substrate; those need no config, and no object name can carry a `/`, so the spec refuses it.)
- **The Controller service** at `controller_endpoint` is dialed by the API server to create, attach and delete volumes, and must report the same driver name. CSI drivers serve a unix socket, so an in-cluster driver needs a small TCP bridge (a socat sidecar) behind a Service.
- **The Node service** is reached by each node agent over the kubelet plugin socket, `unix:///var/lib/kubelet/plugins/<driver_name>/csi.sock`, unless `node_socket_override` names another. The driver's node plugin must bind-mount `/var/lib/ate` with bidirectional propagation, so a volume it mounts there is visible to the node agent.

## TLS

The Controller endpoint carries volume operations, so encrypt it when it crosses anything you do not fully trust:

- `tls.enabled: true` with `use_pod_identity: true` (the only mode at this release): the API server presents its Substrate pod identity as a client certificate and trusts Substrate's service-DNS signer for the driver's serving certificate.
- The driver's endpoint must therefore serve a certificate from that signer: give the bridge a pod-certificate projected volume with signer `servicedns.podcert.ate.dev/identity`, and set `server_name` to the name the certificate carries (`<service>.<namespace>.svc`).
- Plaintext (`tls` unset or `enabled: false`) is for a Controller service only the API server can reach.

## Changing or removing a config

The API server and every node agent read a driver's config the first time an actor needs the driver, then keep the client they built from it until they restart. An edit (a new endpoint, TLS turned on) or a delete does not reach a driver already in use until both restart:

```shell
kubectl -n <control plane namespace> rollout restart deployment/ate-api-server daemonset/atelet
```

A driver no actor has used since the last restart picks up the change at its first use. Upstream's own NFS setup restarts the node agents after registering its driver for the same reason.

## What a config does not do

It does not install the driver, create a StorageClass, or check that the driver answers: nothing in Substrate's API reports a driver config's health. A wrong endpoint shows up when an actor with an external volume resumes and the API server cannot dial the driver.

## Parity accounting

Built for 100% parity with the upstream CSIDriverConfig (`ate.dev/v1alpha1`) at Agent Substrate v0.6.0, the install shape the control plane composes. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.name`, `metadata.labels`, `metadata.annotations` | `name`, `labels`, `annotations` (the envelope) |
| `spec.driverName` | `driverName` |
| `spec.controllerEndpoint` | `controllerEndpoint` |
| `spec.nodeSocketOverride` | `nodeSocketOverride` |
| `spec.tls.enabled`, `.usePodIdentity`, `.serverName` | `tls.enabled`, `.usePodIdentity`, `.serverName` |

### What the spec adds beyond upstream's schema

- **The name equals the driver's name**, the lookup Substrate makes, so a config that would never be found is refused before the apply.
- **The definition's TLS rules** (pod identity and a server name when TLS is on) checked before the apply, and `enabled` always written.

### Not modeled

- **Status** — a CSIDriverConfig has none.

## Pairs well with

- **[KubernetesAgentSubstrate](../kubernetesagentsubstrate/README.md)** — the control plane whose API server and node agents call the driver.
- **[KubernetesAgentSubstrateWorkerPool](../kubernetesagentsubstrateworkerpool/README.md)** — the workers whose actors mount the volumes.
- **[KubernetesAgentSubstrateSandboxConfig](../kubernetesagentsubstratesandboxconfig/README.md)** — the runtime those actors boot.
- **[KubernetesStorageClass](../kubernetesstorageclass/README.md)** — the class actor templates name, with this driver as its provisioner.
