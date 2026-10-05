# Infra Charts

> **Deploy a whole environment from one template.** An Infra Chart bundles the
> dozens of Infra Components behind a real environment — network, DNS, cluster,
> load balancer, certificates, registry — into a single, parameterized blueprint
> you deploy with your own values.

> Not to be confused with the repository's [`helm/`](../helm) directory — those
> are **Helm charts** that deploy Planton itself onto your Kubernetes cluster.
> The charts here are Planton's own blueprint format for deploying
> infrastructure *through* Planton. Looking to self-host Planton? Start at
> [`helm/`](../helm).

## What you get

A curated catalog of production-ready infrastructure blueprints — cluster
platforms, data and streaming platforms, analytics and ML stacks,
observability, GitOps delivery, CI and registry, identity and secrets — for
Kubernetes and the major clouds. Instead of hand-wiring operators, databases,
DNS, certificates, and credentials one resource at a time — and getting the
dependencies right — you pick a chart, set a handful of values, and deploy the
whole thing as one coherent unit.

Every chart earns its place: the catalog is deliberately small enough that
each entry is a complete architecture a team recognizes and wants, built only
from kinds whose schemas and modules meet the catalog's full depth bar.

## Where a chart lives

The tree is provider-rooted, with one home rule:

- **A chart lives under the provider that hosts its centerpiece.**
- **Cluster charts nest under a `kubernetes/` subfolder inside their
  cloud.** A chart whose centerpiece is a managed Kubernetes cluster —
  the cluster plus everything wired onto it — lives at
  `<cloud>/kubernetes/<chart>` (for example `aws/kubernetes/` for EKS
  platforms, `gcp/kubernetes/` for GKE, `azure/kubernetes/` for AKS).
  Non-Kubernetes cloud charts live directly under `<cloud>/`.
- **[`kubernetes/`](kubernetes) holds charts that deploy onto a cluster you
  already have**, whichever cloud or datacenter it runs in. Cross-cloud
  combinations — a cluster on one provider, DNS or secrets on another — are
  parameters inside these charts, never separate catalog entries.

## The mental model

Planton's [catalog](../catalog) is made of kinds:
each one describes a single piece of infrastructure (a VPC, a database, a
cluster) with its own schema and IaC module. **An Infra Chart is a LEGO kit**
— a curated set of those blocks that fit together to build something complete.

And the runtime relationship mirrors Kubernetes and Helm:

> **An Infra Chart is to an Infra Stack what a Helm chart is to a Helm
> release.** The chart is the reusable blueprint; the Infra Stack is a deployed
> instance configured with your values.

## Using a chart

1. Pick a chart under `<provider>/<chart>`.
2. Read its `README.md` for what it provisions, and `values.yaml` for every
   tunable parameter and its default.
3. Provide your values and deploy it through Planton.

Each chart's templates render standard Planton catalog objects — the same
`apiVersion: <provider>.planton.dev/v1` manifests you would write by hand — so
nothing about a chart is a black box: it is a transparent composition of the
kinds in this repo.

## Anatomy of a chart

```
<provider>/<chart>/
├── Chart.yaml      # identity + description + catalog metadata
├── values.yaml     # parameters and their defaults (your knobs)
├── templates/      # manifests that render the Infra Components
└── README.md       # what it provisions and how to configure it
```

## Design principles

- **Every chart is a real-world architecture.** A chart earns its slot by
  being a complete, production-shaped environment for one real scenario —
  something teams recognize and want, that would take a skilled engineer days
  to compose by hand — and by removing genuinely hard wiring (dependency
  ordering, reference plumbing, posture decisions). No filler charts, no demo
  charts; a single resource is not a chart — that is what presets are for.
- **Each provider's charts stand on their own merit.** Chart anatomy is
  shared structure; which charts exist for a provider and how each is designed
  comes only from that cloud's own architectural grain — the architectures its
  services are actually shaped for — never by mirroring how another provider's
  catalog composed something similar.
- **Composability first.** Charts compose first-class, independently ownable
  resources by reference (`valueFrom`), so a chart is a starting point you can
  extend and recombine — not a monolith.
- **Defaults deploy.** Rendering any chart with its `values.yaml` defaults
  produces valid, deployable manifests. Feature toggles render valid in both
  positions.
- **Secure by default.** Where the composed components offer a hardened path
  (private networking, identity-based auth, RBAC-only data planes,
  customer-managed keys), the chart defaults to it and makes relaxation the
  explicit parameter — never the reverse.
- **Charts provision platforms; Service Hub deploys applications.** The
  catalog contains no "deploy my app" charts — a chart ends where the
  application begins. What a chart delivers is the platform an application
  team lands on: the cluster, the addons, the databases, the pipelines.
- **Kubernetes resources bind to their cluster through a connection.** A
  chart that provisions a cluster AND deploys onto it publishes the cluster's
  connection under a chart-controlled name (`planton.dev/connection-name`)
  and every Kubernetes resource consumes it (`planton.dev/connection`), with
  `runs_on` relationship edges making deploy order structural. A chart that
  deploys onto an existing cluster selects it the same way — a connection
  parameter, or the deploying environment's default.
- **Documentation is part of the artifact.** Template comments, parameter
  descriptions, and READMEs render publicly and are held to the same bar as
  the kind's schemas' field comments.
- **No hardcoded provisioner.** Chart resources must not carry a
  `planton.dev/provisioner` annotation. The IaC provisioner (OpenTofu vs
  Pulumi) is a property of the deployment target, resolved from the
  organization's mapping, not baked into the chart. Omit the annotation and
  let each resource inherit the deploying organization's choice.

## Validating a chart

Two gates, one offline and one server-side:

```bash
# Offline (this repo's CLI): renders every template with its default values —
# flipping each bool toggle once so conditional manifests are exercised in
# both branches — and validates each rendered manifest against the compiled-in
# protos, rendered the way the control plane renders them (a template's final
# newline is kept): the kind must exist, every field must exist on the spec,
# the whole document (metadata as well as spec) must pass its validation
# rules, and every valueFrom reference must resolve. No control plane needed.
# The Planton Platform CLI's `chart validate` also checks where secrets go: a
# secret field takes only a $secret/ reference, and a generated secret feeds
# only a secret field.
planton chart validate charts/<provider>/<chart>

# Exercise a specific parameter combination beyond the automatic toggle flips:
planton chart validate charts/<provider>/<chart> --set dnsEnabled=false

# Every chart in the tree (run from charts/):
make validate

# Server-side (Planton Platform CLI): renders and validates through the live
# control plane, exactly as the console does — the authoritative gate before
# publishing.
planton chart build <provider>/<chart> --no-browser
```

The full authoring standard — the catalog design method, the per-file quality
bar, and the template-language contract — lives in
[`_rules/charts/forge-planton-infra-chart.mdc`](../_rules/charts/forge-planton-infra-chart.mdc).
