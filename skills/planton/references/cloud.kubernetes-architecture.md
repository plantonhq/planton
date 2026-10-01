# Kubernetes Architecture Judgment

The catalog's Kubernetes kinds compose into a platform, and some roads through
them are paved — deeper modules, presets, first-class integration — while
others are technically present but shallow. Recommending well means knowing
which is which. This reference is judgment for what runs ON the cluster;
`aws-architecture.md` covers the cloud around it, `environments.md` covers how
many clusters and environments, and `kubernetes-on-cluster.md` covers the
wiring mechanics.

## The paved road: traffic, DNS, and TLS

When an app needs to be reachable on a hostname, this is the stack — end to
end, in dependency order:

1. **`KubernetesGatewayApiCrds`** — the Gateway API resource definitions.
2. **`KubernetesIstio`** — the platform's deepest-supported gateway stack.
   The kind bundles istio/base, istiod, AND the ingress gateway in one
   resource — do not also add `KubernetesIstioBaseCrds` (that kind exists
   only for advanced split installs).
3. **`KubernetesGatewayClass`** — registers Istio as the Gateway API
   controller.
4. **`KubernetesGateway`** — the listener (typically one, with HTTPS and the
   cert).
5. **One `KubernetesHttpRoute` per hostname/app** — "route my app on
   `app.example.com`" is exactly one of these.
6. **`KubernetesCertManager`** — TLS certificates, automatically.
7. **`KubernetesExternalDns`** — watches the routes/gateways and writes the
   DNS records into the Route 53 zone automatically; its EKS config takes the
   zone by reference (`valueFrom` the chart's `AwsRoute53Zone`, or the zone id
   of an existing zone).

**Hard rule: never hand-wire load-balancer IPs or Elastic IPs into DNS
records when external-dns can be part of the architecture.** Static-IP
annotation plumbing is exactly the manual-work class this platform exists to
delete. If you catch yourself proposing "create an EIP and inject it via
annotation," stop and add external-dns instead.

`KubernetesIngressNginx` exists in the catalog but is the unpaved
alternative — offer it only when the user explicitly wants nginx.

## Monitoring rides with the cluster

A cluster that runs anything a person relies on gets its monitoring stack
and working alert delivery in the same shared-cluster chart, not as a
follow-up: `KubernetesKubePrometheusStack` with `alertmanager.notifications`
and an outside heartbeat. The craft, the questions to ask and the proof are
in `cloud.kubernetes-observability.md`.

## Educate at moments of leverage

When a platform capability erases work the user was bracing for, say so in
one sentence at the moment it lands — "external-dns will create that DNS
record automatically whenever you add a route; you never touch Route 53" is
the canonical example. These moments are where developers discover what the
platform (and Kubernetes) can do, and they are the single best use of
education. Everything else about the machinery stays in reserve per
`deployment-model.md`.

## The two-chart pattern: shared infrastructure + environment chart

When an ask mixes platform components and app workloads, propose the split —
by name, up front:

- **The shared-infrastructure chart** (deployed once): VPC/network, the EKS
  cluster, Gateway API CRDs, Istio, GatewayClass, cert-manager, external-dns,
  and the Route 53 zone. This is Scenario 1 of `kubernetes-on-cluster.md` —
  it publishes the cluster's connection for everything that follows.
- **The environment chart** (deployed once per environment — dev, prod, …):
  the environment's namespace and the `KubernetesGateway` its apps' routes
  attach to. Scenario 2 of `kubernetes-on-cluster.md` — it carries NO
  connection wiring (the shared cluster's materialized connection is the
  platform's default binding); `values.env` differentiates the deployments so
  one chart serves every environment.
- **The app itself** goes where its lifecycle goes. When the app is a Planton
  service (or will be — the usual case for code the user ships), its
  workload, the `KubernetesHttpRoute` for its hostname, and any gateway or
  certificate only it uses are the SERVICE's resources, declared per
  environment on the service — never in either chart. The platform reads a
  service's URL from its own environment's resources, and the service's
  delete cascade removes only what the service declares
  (`service.configuring-deployments.md`, "What belongs on the service, and
  what belongs on the infra project" — the one-question test, and the
  three-step move when a chart already carries them). Only when there is no
  service — the user wants the infrastructure alone — does the environment
  chart carry a **placeholder `KubernetesDeployment`** (image as a param) and
  its route.

**Placement doctrine — operators follow the cluster, never the app.**
Cluster-scoped, shared-by-design components — the Gateway API CRDs, Istio,
the GatewayClass, cert-manager, external-dns, any operator or controller —
live in the shared-infrastructure chart, exactly once. A per-environment
chart never installs one: the same chart deployed into dev and prod would
install a cluster-wide singleton twice, and the platform components'
lifecycle belongs with the cluster they serve, not with any one app. The
same rule one level down: the charts keep what outlives any one service —
the cluster, the shared ingress controller and gateway, the DNS zone, the
namespace — and a service keeps what exists because of it. When the
environment chart needs a VALUE the shared chart's resources produce (a
zone id, a cluster name), it wires a cross-chart `valueFrom` reference —
never a param the user must fill (`dependencies.md`, "References cross
chart boundaries").

**The validated-base philosophy:** the goal is the fastest *validated
base*. The first workload answering on the real hostname — the service's
first deploy, or the placeholder when there is no service — proves the
entire road: cluster, gateway, DNS, TLS. From there every push ships the
real app. Say this arc out loud when proposing the split, so the user knows
what "done" unlocks and where CI/CD picks up.

**Application-first:** ask about the APP — what it is, what port it listens
on, what hostname it should answer on. The infrastructure exists in service
of it; a composition that never asked what the user is deploying has failed
regardless of how clean the charts are.

**Sequencing:** propose both charts up front, then compose and finish the
shared-infrastructure chart before the environment chart — producers before
consumers, each chart driven green before the next begins, and the
service's first deploy last, onto the namespace and gateway the environment
chart made. In your
workspace, each chart is its own top-level subfolder (the identity check in
the skill), so both live side by side in one conversation; when the folder
you were given IS a single chart, finish it and have the user open or
create the second chart's folder — the grounding duty (`discovery.md`)
means the next conversation discovers the deployed cluster automatically.
Either way, one folder holds one chart: never mix two charts' files into
one chart folder.

## Cost shape on the cluster

The cluster itself is the big always-on charge (control plane + nodes + NAT —
see `cost-transparency.md`); everything in the paved road above is free
software running on nodes the user already pays for. One more reason the
shared-infrastructure/environment-chart split is cost-honest: environments
share the paid platform instead of multiplying it (`environments.md`).
