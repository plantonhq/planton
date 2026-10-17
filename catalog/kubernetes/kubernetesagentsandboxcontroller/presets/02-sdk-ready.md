# SDK-ready

Everything upstream's Python and Go SDKs expect. The SDKs create a `SandboxClaim`, which binds to a ready member of a `SandboxWarmPool` built from a `SandboxTemplate` (that's the extensions). They then reach the sandbox through the router at `sandbox-router-svc` in `agent-sandbox-system`, which the `router_endpoint` output names.

The router lets every request through. Every Service stays inside the cluster, and the router's NetworkPolicy (on by default) limits who may call it on a cluster whose network enforces policies. That suits a cluster where only your own services run. For anything reachable by untrusted callers, switch `authorization` to `tokenReview` (Kubernetes bearer tokens) or `scopedToken` (per-sandbox tokens your product signs), as the production preset does.

Change first: narrow `router.networkPolicy.clientNamespaceLabels` to the namespaces that call the router. List every port your sandboxes serve in `router.networkPolicy.sandboxPorts`; only 8888 is open by default.

See [02-sdk-ready.yaml](./02-sdk-ready.yaml) for the manifest.
