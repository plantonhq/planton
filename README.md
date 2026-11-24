<p align="center">
  <img src=".github/icon.png" alt="Planton logo" width="72">
</p>

# Planton

> **Planton deploys every push to GitHub into your own cloud, and shows the cost and permissions of every infrastructure change before it runs.**

Start by pasting one prompt into your coding agent. On planton.ai, a new organization gets a Setup Prompt: paste it into Claude Code, Codex or Cursor from your app's repository, and the agent installs the CLI, connects the cloud you're already signed into, and ships your first deployment, asking before anything changes your cloud or costs money. Underneath is clean, auditable infrastructure-as-code from the open-source modules in this repository.

<p align="center">
  <b><a href="https://planton.ai/signup">Start free on planton.ai</a></b>
  &nbsp;·&nbsp; Free for up to 3 seats, no card
  &nbsp;·&nbsp; Or <a href="https://planton.ai/desktop">Planton Desktop</a>, free on macOS · Linux · Windows
</p>

**[Website](https://planton.ai)** · **[Getting Started](https://planton.ai/docs/getting-started)** · **[Documentation](https://planton.ai/docs)**

---

## How it works

You never start from a blank Terraform file. Planton runs **proven, pre-built
infrastructure-as-code modules** — built for secure, well-architected,
cost-efficient defaults — against your own cloud account, with real state and
history. Create like a console: pick a stack, fill a short form. Manage like
Kubernetes: `planton apply -f` deploys one manifest (one Infra Component) or a whole directory of
them in dependency order, and `planton chart install` stands up a templated
environment — the `kubectl apply` and `helm install` gestures, freed from
Kubernetes and extended to every cloud.

Whichever way you deploy, it's real infrastructure-as-code the whole way down —
stored, versioned, every change a diff. On your own cloud. If you leave, you
keep the modules and your manifests, and the open-source CLI keeps deploying
them.

## What's in this repository

This repo holds the **open building blocks** that power Planton, all under
Apache-2.0. Audit them, fork them, or take your configuration and run it
yourself.

<!-- AI-AGENT NOTE: The kind and provider counts below MUST be verified against
     the authoritative source: shared/catalogkind/catalog_kind.proto
     (count non-test `(kind_meta)` annotations and distinct providers). Never re-type a stale number. -->

- **[`catalog/`](catalog)** — **850+ catalog kinds**
  across **9 cloud providers** (AWS, GCP, Azure, Kubernetes, DigitalOcean,
  Cloudflare, Auth0, Stripe, and OpenFGA). Each kind is a
  Protocol Buffer definition in the Kubernetes Resource Model shape
  (`apiVersion`/`kind`/`metadata`/`spec`) with field-level validations and
  auto-generated SDKs in Go, Python, TypeScript, and Java — plus verified,
  machine-checked fact-sheets: cost data with generated per-preset estimates
  (priced from the pinned price books in `catalog/_pricing/`), technical
  control posture with framework crosswalks (`catalog/_compliance/`), and
  least-privilege provisioning-permission manifests validated against the
  providers' own published inventories.
- **[`cmd/planton`](cmd/planton)** — the open-source CLI and IaC engine that
  validates manifests and executes the Pulumi and OpenTofu/Terraform modules
  that ship with every kind, and validates your own Infra Charts offline with
  `planton chart validate`.
- **[`site/`](site)** — the source of [planton.ai](https://planton.ai):
  website, docs, and blog (Next.js static export, deployed to GitHub Pages on
  every `site/` change). Run it locally with `make run-site`, or build and
  preview the static export with `make preview-site`.

## The CLI

The desktop app is the product; the CLI is its companion — the same deploys,
from your shell, driving the same engine.

```bash
curl -fsSL https://planton.ai/install.sh -o /tmp/planton-install.sh && sh /tmp/planton-install.sh
```

The installer puts the CLI at `~/.planton/bin/planton` on macOS or Linux and prints how to add it to your PATH. Homebrew, Windows and direct downloads are in [the CLI docs](https://planton.ai/docs/cli).

Write a manifest in the shape you already know:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
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
[Getting Started guide](https://planton.ai/docs/getting-started) and the
[CLI reference](https://planton.ai/docs/cli).

## From your coding agent

In the Planton console (planton.ai, or your own self-hosted Planton), your organization's Getting Started hands you a **Setup Prompt**: paste it into your coding agent from your app's repository, and the agent installs the CLI and the skills, signs in as you with a one-time code, and takes the repository to its first deployment, asking before anything changes your cloud or costs money.

Planton ships its knowledge as agent skills in the open Agent Skills format,
so Cursor, Claude Code, Codex, and any other coding agent can compose,
validate, price, and deploy from the same craft. Install them once:

```bash
npx skills add plantonhq/skills
```

Two skills ship together: `planton` (the working craft, and the boundaries the
agent never crosses — no mutation without your consent, never outside your
repository) and `multi-cloud-catalog` (one reference page per kind, with
its cost, control posture, and runner permissions). Agents that want the
platform's own operations reach them over MCP at `https://mcp.planton.ai/`.
See [Coding Agents](https://planton.ai/docs/coding-agents).

## Where it runs

Planton runs where you decide: hosted at [planton.ai](https://planton.ai),
self-hosted on your own Kubernetes cluster with a license that verifies
offline, or free on your laptop as Planton Desktop. In every shape it is your
cloud account, your keys, your state, and your bill. The same manifests and the
same model run on all three, so nothing is redone when a person becomes a team.

Self-hosting is two Helm installs — the operator, then the platform at the
release you name — with the control plane, the console, the identity server,
the secrets manager, the databases, and an in-cluster runner all inside your
cluster. See [Self-Hosting](https://planton.ai/docs/self-hosting).

## Licensing

Planton the app is **free**, including for commercial use. The building blocks
that power it — the infrastructure kinds and the CLI — are
**open source under [Apache-2.0](LICENSE)**: audit them, fork them, or take
your configuration and run it yourself. No lock-in.
The Planton name and logo are trademarks of Planton Cloud, Inc. — the code is
yours to fork, the identity is not; see [TRADEMARKS.md](TRADEMARKS.md).

## Contributing

Visit [CONTRIBUTING.md](CONTRIBUTING.md) for information on building Planton
from source and for details about becoming a contributor.

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
