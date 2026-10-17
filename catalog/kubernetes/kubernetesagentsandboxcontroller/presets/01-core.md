# Core

The core controller exactly as upstream publishes it: one Sandbox definition, one controller, nothing else. A `Sandbox` is a single pod with a stable name, optional persistent volumes and an optional headless Service; the controller creates it, suspends it (pod deleted, volumes kept) and expires it.

Pick this when your product creates and manages Sandboxes itself through the Kubernetes API. It needs no warm pools, no claims and no router. This resource creates the `agent-sandbox-system` namespace and deletes it on destroy; reference a **Namespace** resource instead when other components will put Secrets there.

Change first: nothing, usually. Turn on `extensions` when you want templates, warm pools and claims, which is what upstream's SDKs use. Add a `router` when callers reach sandboxes over HTTP by name, which is the only way in under gVisor or Kata.

See [01-core.yaml](./01-core.yaml) for the manifest.
