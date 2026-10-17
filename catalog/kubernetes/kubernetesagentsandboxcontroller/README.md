# Kubernetes Agent Sandbox Controller

## When NOT to Use This

**One resource is the agent-sandbox ENGINE**: the SIG Apps controller that turns `Sandbox` declarations into isolated, stateful pods for AI agents, plus, with extensions, the controllers behind templates, warm pools and claims, and the optional router callers reach sandboxes through. One installation per cluster (an upstream contract).

Not the right kind when:

- **You want a sandbox** — that is [`KubernetesAgentSandbox`](../kubernetesagentsandbox/README.md) (one sandbox), or [`KubernetesAgentSandboxTemplate`](../kubernetesagentsandboxtemplate/README.md) with [`KubernetesAgentSandboxWarmPool`](../kubernetesagentsandboxwarmpool/README.md) (ready sandboxes your product claims, through [`KubernetesAgentSandboxClaim`](../kubernetesagentsandboxclaim/README.md) or agent-sandbox's SDKs). Installing this controller alone runs no sandboxes.
- **Your cluster is on GKE with the Agent Sandbox add-on on** — the add-on is the same controller, managed by Google, in the same namespace. Keep one of the two. This kind refuses to install over the add-on and names it.
- **You want a general sandbox for untrusted code with no agent lifecycle** — a plain Deployment on a gVisor RuntimeClass is enough; agent-sandbox earns its place through stable identity, suspend and resume, expiry, warm pools and claims.

## Kept definitions and the destroy contract

The four resource definitions are applied before the controller, as kept resources stamped with their source and release. Destroying this resource removes the controller and the router and leaves the definitions, so no `Sandbox`, pod or volume is deleted: sandboxes stop being reconciled until a controller returns, and a reinstall adopts them again. `crds.keepOnUninstall: false` deletes the definitions on destroy, and with them every sandbox, template, warm pool and claim on the cluster; keep it for throwaway clusters.

Every write is checked first: a definition another tool applied (the GKE add-on, a second installation, a `kubectl apply`) is refused with its owner named, and so is an older release than the one that wrote the cluster's definitions. The controller Deployment gets the same check: a controller this kind did not create is refused rather than taken over.

## Done means serving

The controller has no webhooks and no finalizers at v1.0.7, so nothing on the cluster waits on it to admit an object. The deploy waits until the controller (and the router, when set) has rolled out and its pods pass their health probes. Upstream's controller manifest declares none, so this kind adds them on the controller's own health endpoints. A sandbox applied right after this resource is reconciled at once.

## Distribution

Installed from upstream's own release manifest at the pinned tag (`sandbox.yaml`, or `sandbox-with-extensions.yaml` with extensions on), with the definitions split out to the kept path; the router from upstream's deployment files at the same tag, with its image pinned to the release. The typed settings are applied onto upstream's own container arguments, so a setting left unset keeps upstream's value and the installed objects stay recognisably upstream's. The namespace is the manifest's fixed `agent-sandbox-system`; this resource creates it with `createNamespace`, or a KubernetesNamespace owns it when the Secrets the router reads must be there first. The spec deliberately has no version field: the catalog's sandbox kinds mirror the pinned definitions field for field. `imageRegistry` moves both images to a mirror of registry.k8s.io.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
