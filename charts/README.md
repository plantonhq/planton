# Infra Charts

> **Deploy a whole environment from one template.** An Infra Chart bundles the
> dozens of cloud resources behind a real environment — network, DNS, cluster,
> load balancer, certificates, registry — into a single, parameterized blueprint
> you deploy with your own values.

## What you get

A curated catalog of production-ready infrastructure blueprints across AWS, GCP,
Azure, OCI, DigitalOcean, Hetzner Cloud, Alibaba Cloud, Civo, Scaleway, and
OpenStack. Instead of hand-wiring a VPC, subnets, gateways, a Kubernetes
cluster, DNS, and certificates one resource at a time — and getting the
dependencies right — you pick a chart, set a handful of values, and deploy the
whole thing as one coherent unit.

## The mental model

Planton's [deployment components](../apis/dev/planton/provider) are LEGO blocks:
each one is a single cloud resource (a VPC, a database, a cluster) with its own
schema and IaC module. **An Infra Chart is a LEGO kit** — a curated set of those
blocks that fit together to build something complete.

And the runtime relationship mirrors Kubernetes and Helm:

> **An Infra Chart is to an Infra Project what a Helm chart is to a Helm
> release.** The chart is the reusable blueprint; the project is a deployed
> instance configured with your values.

## Using a chart

1. Pick a chart under `<provider>/<chart>` (for example `aws/ecs-environment`).
2. Read its `README.md` for what it provisions, and `values.yaml` for every
   tunable parameter and its default.
3. Provide your values and deploy it through Planton.

Each chart's templates render standard Planton Cloud Resources — the same
`apiVersion: <provider>.planton.dev/v1` manifests you would write by hand — so
nothing about a chart is a black box: it is a transparent composition of the
components in this repo.

## Anatomy of a chart

```
<provider>/<chart>/
├── Chart.yaml      # identity + description + catalog metadata
├── values.yaml     # parameters and their defaults (your knobs)
├── templates/      # manifests that render the Cloud Resources
└── README.md       # what it provisions and how to configure it
```

## Design principles

- **Composability first.** Charts compose first-class, independently ownable
  resources by reference (`valueFrom`), so a chart is a starting point you can
  extend and recombine — not a monolith.
- **No hardcoded provisioner.** Chart resources must not carry a
  `planton.dev/provisioner` label. The IaC provisioner (OpenTofu vs Pulumi) is a
  property of the deployment target, resolved from the organization's mapping,
  not baked into the chart. Omit the label and let each resource inherit the
  deploying organization's choice.

Authoring guidance for these charts lives in [`_rules/charts/`](../_rules/charts).
