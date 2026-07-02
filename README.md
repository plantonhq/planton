<p align="center">
  <img src="site/public/icon.png" alt="Planton logo" width="72">
</p>

# Planton

> **Deploy real infrastructure to your own cloud — without writing Terraform.**

Planton is a free desktop app and CLI you download and open. It finds the cloud
you're already signed into; pick a stack and fill a short form, or
`planton apply -f` a manifest from your terminal — and watch it deploy, with
clean, auditable infrastructure-as-code running underneath. No account. No
connections. No ceremony.

<p align="center">
  <b><a href="https://planton.app/download">Download Planton</a></b>
  &nbsp;·&nbsp; Free forever, including for commercial use
  &nbsp;·&nbsp; Runs on macOS · Linux · Windows
</p>

**[Website](https://planton.app)** · **[Getting Started](https://planton.app/docs/getting-started)** · **[Documentation](https://planton.app/docs)**

---

## How it works

You never start from a blank Terraform file. Planton runs **proven, pre-built
infrastructure-as-code modules** — built for secure, well-architected,
cost-efficient defaults — against your own cloud account, with real state and
history. Create like a console: pick a stack, fill a short form. Manage like
Kubernetes: `planton apply -f` deploys a single component, and
`planton chart install` stands up a whole environment — the `kubectl apply` and
`helm install` gestures, freed from Kubernetes and extended to every cloud.

Whichever way you deploy, it's real infrastructure-as-code the whole way down —
stored, versioned, every change a diff. On your own cloud. Nothing to lock you
in.

## What's in this repository

This repo holds the **open building blocks** that power Planton, all under
Apache-2.0. Audit them, fork them, or take your configuration and run it
yourself.

<!-- AI-AGENT NOTE: The component and provider counts below MUST be verified against
     the authoritative source: apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto
     (count non-test `(kind_meta)` annotations and distinct providers). The chart
     count MUST be verified against `charts/*/*/Chart.yaml`. Never re-type a stale number. -->

- **[`apis/`](apis/dev/planton/provider)** — **400+ deployment components**
  across **17 cloud providers** (AWS, GCP, Azure, Kubernetes, OCI, Alibaba
  Cloud, Hetzner Cloud, DigitalOcean, Cloudflare, Civo, Scaleway, OpenStack,
  Confluent, Snowflake, Auth0, MongoDB Atlas, and OpenFGA). Each component is a
  Protocol Buffer definition in the Kubernetes Resource Model shape
  (`apiVersion`/`kind`/`metadata`/`spec`) with field-level validations and
  auto-generated SDKs in Go, Python, TypeScript, and Java.
- **[`charts/`](charts)** — **49 ready-made infra charts**: whole environments
  (network + compute + data + DNS) composed from the components above and
  installed in one command — the Helm-chart idea, for cloud infrastructure.
- **[`cmd/planton`](cmd/planton)** — the open-source CLI and IaC engine that
  validates manifests and executes the Pulumi and OpenTofu/Terraform modules
  that ship with every component.
- **[`site/`](site)** — the [planton.app](https://planton.app) website and
  documentation.

## The CLI

The desktop app is the product; the CLI is its companion — the same deploys,
from your shell, driving the same engine.

```bash
brew install plantonhq/tap/planton
```

Write a manifest in the shape you already know:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: my-first-postgres
spec:
  namespace:
    value: my-first-postgres
  createNamespace: true
  container:
    replicas: 1
    diskSize: 1Gi
```

Then deploy it:

```bash
planton apply -f postgres.yaml
```

Validation catches mistakes in seconds — before anything touches your cloud —
and the deploy streams live output from the underlying IaC module. See the
[Getting Started guide](https://planton.app/docs/getting-started) and the
[CLI reference](https://planton.app/docs/cli/cli-reference).

## Licensing

Planton the app is **free**, including for commercial use. The building blocks
that power it — the infrastructure components, the charts, and the CLI — are
**open source under [Apache-2.0](LICENSE)**: audit them, fork them, or take
your configuration and run it yourself. No lock-in.

## Contributing

Visit [CONTRIBUTING.md](CONTRIBUTING.md) for information on building Planton
from source, and the [Contributor Guide](https://planton.app/docs/contributing)
for details about becoming a contributor.

## Acknowledgments

- **Brian Grant & the Kubernetes API team** for their foundational work on the
  Kubernetes Resource Model.
- The **[Protobuf Team](https://protobuf.dev/)** for laying the foundation for
  a powerful language-neutral contract definition language.
- The **[Buf](https://github.com/bufbuild/buf) Team** for their Protobuf
  tooling — including BSR Docs, BSR SDKs, and ProtoValidate — which
  collectively democratized protobuf adoption and made this project possible.
- The **[Pulumi](https://github.com/pulumi/pulumi)** team for providing a
  powerful infrastructure-as-code platform that enables multi-language support.
- The **[spf13/cobra](https://github.com/spf13/cobra)** team for making
  building command line tools a bliss.

---

<p align="center">
  Built by <a href="https://planton.ai">Planton</a>
</p>
