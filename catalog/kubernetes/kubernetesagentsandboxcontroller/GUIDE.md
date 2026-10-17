# KubernetesAgentSandboxController Guide

The judgment this guide carries: this kind is the one agent-sandbox engine on a cluster. It owns the resource definitions every sandbox is built on, so it installs them as kept resources and never shares them, and it decides who may reach a sandbox, so the router's authorization is always a choice you make.

## One per cluster, in a fixed namespace

Everything installs into `agent-sandbox-system`, with upstream's own resource names. The namespace is not a field because upstream bakes it in twice: in the release manifest, and in the default network policy every SandboxTemplate gets, which admits traffic only from pods labelled `app: sandbox-router` in `agent-sandbox-system`. That second fact is why the router is part of this kind and not a kind of its own: a router anywhere else could not reach a single policed sandbox.

## The namespace's owner is your choice

The name is fixed; the owner is not. Two shapes:

- **`createNamespace: true`** — this resource creates `agent-sandbox-system` and deletes it on destroy, with everything in it. Right when nothing else ever lives there.
- **A KubernetesNamespace named `agent-sandbox-system`**, referenced from `namespace`. Right the moment anything this kind reads must exist before its pods start: the router's TLS Secret, scoped-token verification keys, a client CA bundle, OTLP headers, image pull secrets. The namespace comes first, the Secrets go into it, then this controller; destroying the controller leaves them all in place, and the namespace's own quotas and policies apply.

The deploy checks the choice before it writes anything. With `createNamespace` off and no namespace yet, it stops and says to deploy the namespace first. With `createNamespace` on and a namespace someone else created, it stops rather than take it over, because a later destroy would delete what others put there. And turning `createNamespace` off on a live installation is refused, because the engines would delete the namespace this resource created, running controller included. To move the namespace to a KubernetesNamespace, destroy this resource first (the resource definitions and every sandbox outside `agent-sandbox-system` stay), deploy the namespace resource, then deploy this one again with `createNamespace` off.

## This kind or the GKE add-on, never both

GKE's managed Agent Sandbox add-on (the `GcpGkeCluster` addon `agent_sandbox_enabled`) installs the same controller and definitions into the same namespace, and Google keeps them current. Choose one per cluster:

- **The add-on** when you want Google to own upgrades and you need none of the settings below.
- **This kind** when you want the release pinned with the rest of your catalog, any controller or router setting, the router at all (the add-on installs none), or the same installation on every cloud.

On a cluster that already carries the definitions, from the add-on or anyone else, the deploy stops before it changes anything:

- **Observed:** CRD sandboxes.agents.x-k8s.io already exists on the cluster and was not applied by this module (and who did apply it, as the cluster records it).
- **Meaning:** two owners would write one schema, and the module cannot tell whether the existing definition is newer or older than agent-sandbox release v1.0.7, so it stops before changing it.
- **Next step:** remove the other installation (turn the add-on off), or keep it and do not install this kind. `crds.install: false` is only for a GitOps repository that applies upstream's definitions while this kind runs the one controller.

A controller Deployment someone else created gets the same refusal, so the kind never takes over another installation's workloads and never deletes them on destroy.

## When you need the router

You need the router when anything outside the cluster's pod network, or any SDK, talks to a sandbox:

- Upstream's Python and Go SDKs reach sandboxes through `sandbox-router-svc`.
- Sandboxes running under gVisor or Kata cannot be port-forwarded; the router is the only way in.
- The default SandboxTemplate network policy admits only the router, so a policed sandbox is reachable through nothing else.

A product that runs its own proxy inside the cluster, and writes its own network policies, can leave `router` out.

## Choosing router authorization

Upstream's router lets every request through unless told otherwise, so `router.authorization` is required and has no default:

- **`allowAll`** — no check. The router's NetworkPolicy (on by default) and the cluster network are the whole fence. Right for development and for a router reachable only from your own product's namespace (`networkPolicy.clientNamespaceLabels`).
- **`tokenReview`** — the caller presents a Kubernetes ServiceAccount token and the router asks the API server whether it is valid (`audiences` narrows which tokens count; `requireToken` refuses a request with none instead of passing it through unauthenticated). Right when your callers run on the cluster. The module adds the TokenReview RBAC only in this mode.
- **`scopedToken`** — the caller presents a token signed for one sandbox. Version 2 tokens are verified with public keys from `verificationKeys` and bound to the sandbox's UID, which needs the router's pod-IP cache (`cache.enabled`, on by default). `v1Secret` and `v1AcceptUntil` keep version 1 tokens valid while you move off them.
- **`cookie`** — added to any mode for browsers: the router turns a token in a query parameter into a session cookie. It needs `proxy.pathRoutingPrefix`, because a browser reaches a sandbox only by path.

## Kept definitions and the destroy contract

The definitions are applied first, as kept resources stamped with their source (`agent-sandbox`) and release. Destroying this resource removes the controller and the router and leaves the definitions, so no sandbox, pod or volume is deleted. Sandboxes then stop being reconciled (no expiry, no warm pool refill) until a controller returns, and a reinstall adopts them again. `crds.keepOnUninstall: false` deletes the definitions, and with them every sandbox, template, warm pool and claim on the cluster.

Turning `extensions.enabled` off removes the extension controllers. Their definitions follow the same rule: kept while `crds.keepOnUninstall` is true, so existing templates, pools and claims survive and resume when you turn the extensions back on; deleted with every object of their kind when it is false.

## Done means serving

At v1.0.7 the controller registers no webhooks and puts no finalizers on anything, so no object waits on it to be admitted or deleted. The deploy waits for the controller and the router to roll out with every pod ready. The router's deployment file probes its health endpoints already; the controller's release manifest declares no probes, so the module adds them on the controller's own health endpoints (`/healthz` for liveness, `/readyz` for readiness), with controller-runtime's usual timings. Without them, a finished rollout would only mean the process started.

## Pulling from a mirror

`imageRegistry` replaces the registry part of both images this kind runs (the controller and the router, both on registry.k8s.io), keeping each path and tag. An image override on `controller.image` or `router.image` wins for that one image. Sandbox images are your own and never move. `imagePullSecrets` names Secrets in `agent-sandbox-system` for a private registry.

## Parity accounting

Built for 100% parity with every setting upstream's controller and router offer at v1.0.7. Every flag, file and manifest setting is accounted for below; a setting left unset keeps upstream's value, because the module writes a flag only when it differs from upstream's default.

### The controller

| Upstream setting | Spec field |
|------------------|------------|
| `--extensions` | `extensions.enabled` (selects `sandbox-with-extensions.yaml`, which passes it) |
| `--cluster-domain` | `clusterDomain` |
| `--leader-elect`, `--leader-election-namespace` | `controller.leaderElection.enabled`, `.namespace` |
| `--metrics-bind-address`, `--enable-rest-client-metrics` | `controller.metrics.port`, `.restClientMetricsEnabled` |
| `--metrics-secure-serving`, `--metrics-cert-dir`, `--tls-min-version`, `--tls-cipher-suites` | `controller.metrics.tls` (`enabled`, `certificateSecret` mounted as the cert directory, `minVersion`, `cipherSuites`) |
| `--health-probe-bind-address` | `controller.healthProbePort` |
| `--sandbox-concurrent-workers`, `--disable-sandbox-events`, `--sandbox-write-behind-window` | `controller.sandbox.concurrentWorkers`, `.eventsDisabled`, `.writeBehindWindow` |
| `--cache-label-selectors` | `controller.cacheLabelSelectors` |
| `--kube-api-qps`, `--kube-api-burst`, `--api-connections`, `--separate-watch-connection` | `controller.kubeApi.qps`, `.burst`, `.connections`, `.separateWatchConnection` |
| `--enable-tracing` and the OTLP environment | `controller.tracing` (`otlpEndpoint`, `insecure`, `headers` from a Secret) |
| `--enable-pprof`, `--enable-pprof-debug`, `--pprof-block-profile-rate`, `--pprof-mutex-profile-fraction` | `controller.profiling.cpuEnabled`, `.debugEnabled`, `.blockProfileRate`, `.mutexProfileFraction` |
| `--zap-devel`, `--zap-encoder`, `--zap-log-level`, `--zap-stacktrace-level`, `--zap-time-encoding` | `controller.logging.development`, `.encoder`, `.level`, `.stacktraceLevel`, `.timeEncoding` |
| `--sandbox-claim-concurrent-workers`, `--sandbox-claim-warm-candidate-grace-period`, `--disable-claim-events`, `--disable-claim-observability-annotations` | `extensions.claim.concurrentWorkers`, `.warmCandidateGracePeriod`, `.eventsDisabled`, `.observabilityAnnotationsDisabled` |
| `--sandbox-warm-pool-concurrent-workers`, `-max-batch-size`, `-replenish-delay`, `-max-refill-rate`, `-readiness-grace-period`, `-unschedulable-recheck-interval`, `--enable-warm-pool-eviction` | `extensions.warmPool.concurrentWorkers`, `.maxBatchSize`, `.replenishDelay`, `.maxRefillRate`, `.readinessGracePeriod`, `.unschedulableRecheckInterval`, `.evictionEnabled` |
| `--sandbox-template-concurrent-workers` | `extensions.template.concurrentWorkers` |
| The `agent-sandbox-config` ConfigMap (allowed label domains) | `extensions.allowedLabelDomains`; the pod template carries the ConfigMap's checksum, so a change restarts the controller, which reads it at startup |
| Image, replicas, resources, node selector, tolerations | `controller.image`, `.replicas`, `.resources`, `.nodeSelector`, `.tolerations`; `imageRegistry`, `imagePullSecrets` |
| The four definitions | `crds.install`, `crds.keepOnUninstall` |

### The router

| Upstream setting | Spec field |
|------------------|------------|
| Installing it at all | `router` present, `router.enabled` |
| `--http-bind-address`, `--metrics-bind-address`, `--health-probe-bind-address` | `router.ports.http` (or `router.tls.httpEnabled: false` to close it), `.metrics`, `.healthProbe` |
| `--https-bind-address`, `--tls-cert-file`, `--tls-key-file` | `router.tls.port`, `router.tls.certificateSecret` (mounted, Service and policy ports opened) |
| `--mtls-mode`, `--tls-client-ca-file`, `--tls-min-version`, `--tls-cipher-suites` | `router.tls.mtlsMode`, `.clientCa`, `.minVersion`, `.cipherSuites` |
| `--cluster-domain` | `clusterDomain` |
| `--proxy-timeout`, `--response-header-timeout`, `--shutdown-timeout`, `--max-request-body-bytes`, `--allow-loopback-pod-ip`, `--path-routing-prefix` | `router.proxy.timeout`, `.responseHeaderTimeout`, `.shutdownTimeout` (also sets the pod's termination grace), `.maxRequestBodyBytes`, `.allowLoopbackPodIp`, `.pathRoutingPrefix` |
| `--upstream-max-retries`, `--upstream-retry-initial-delay`, `--upstream-retry-max-delay` | `router.proxy.upstreamMaxRetries`, `.upstreamRetryInitialDelay`, `.upstreamRetryMaxDelay` |
| `--enable-tracing`, `--enable-otel-metrics` and the OTLP environment | `router.telemetry` (`otlpEndpoint`, `insecure`, `tracesEnabled`, `metricsEnabled`, `headers` from a Secret) |
| `--access-log` | `router.accessLogEnabled` |
| `--authz-mode` and every `--authz-tokenreview-*`, `--authz-scoped-token-*` and `--authz-cookie-*` flag, `--authz-trust-forwarded-proto` | `router.authorization` (`allowAll`, `tokenReview`, `scopedToken`, `cookie`); the TokenReview RBAC follows the mode |
| `--cache-enabled`, `--cache-namespace` | `router.cache.enabled`, `.namespace`; the cache's RBAC follows the switch |
| The NetworkPolicy's placeholders (client namespaces, metrics namespace, sandbox port) | `router.networkPolicy.enabled`, `.clientNamespaceLabels`, `.metricsNamespace`, `.sandboxPorts` |
| Image, replicas, resources, node selector, tolerations, PodDisruptionBudget | `router.image`, `.replicas`, `.resources`, `.nodeSelector`, `.tolerations`, `.minAvailable` |

### What the module adds

- Liveness and readiness probes on the controller, on its own health endpoints (the router's deployment file has them already).
- The `planton.ai/kind` label on the workloads, which is how a later deploy knows the installation is this kind's.
- The `planton.ai/config-sha256` pod annotation on the controller, so a change to the allowed label domains restarts it.

### Not modeled, and why

- **`--version`** on both binaries prints a version and exits; it deploys nothing.
- **`--kubeconfig`** on both binaries points at a cluster from outside it. Inside the cluster both use their ServiceAccount, which is the only configuration that makes sense for an installed controller.
- **The router's `--config` file** carries the same settings as its flags; every one of them is typed above, so a second way to set them would only let the two disagree.
- **Other `OTEL_*` variables** (sampler, resource attributes, protocol) belong to the OpenTelemetry SDK, not to agent-sandbox. Upstream reads the endpoint, the insecure switch and the headers, and this kind types exactly those; the collector is where sampling and resource attributes are decided.

## Pairs well with

- **[KubernetesAgentSandbox](../kubernetesagentsandbox/README.md)**, **[KubernetesAgentSandboxTemplate](../kubernetesagentsandboxtemplate/README.md)**, **[KubernetesAgentSandboxWarmPool](../kubernetesagentsandboxwarmpool/README.md)**, **[KubernetesAgentSandboxClaim](../kubernetesagentsandboxclaim/README.md)** — the resources this controller runs; each declares this kind as its prerequisite.
- **[KubernetesNamespace](../kubernetesnamespace/README.md)** — owns `agent-sandbox-system` when other components put Secrets there.
- **[KubernetesCertificate](../kubernetescertificate/README.md)** — the router's TLS certificate, issued into `agent-sandbox-system`.
- **[KubernetesSecret](../kubernetessecret/README.md)** — scoped-token verification keys, a client CA bundle, OTLP headers.
- **[KubernetesServiceMonitor](../kubernetesservicemonitor/README.md)** — scrapes the controller's metrics Service (`status.outputs.metrics_service_name`).
