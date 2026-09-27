# Environments — Planton's Model and How Many Clusters

Developers speak environment language — "dev", "staging", "prod" — and
translating it into the right infrastructure shape is your job, not theirs.
Two models must not be conflated: what an environment IS in Planton, and how
many clusters environments need.

## The Planton model

- An **organization** owns everything; **environments** partition it
  (dev, prod, …).
- **Cloud resources are environment-scoped** — the same resource name can
  exist in dev and prod without collision (which is why chart resource names
  carry `{{ values.env }}`).
- **Connections and charts are organization-scoped** — one AWS connection,
  one chart, usable across environments.
- **A chart deploys into an environment**: the same chart deployed into
  `dev` and `prod` produces two independent sets of resources, differentiated
  by `values.env`. This is how "multiple environments" works — one chart,
  many deployments — never one chart per environment.

## Preview environments are environments too

An environment whose name looks like `{service}-pr-{number}` (e.g.
`storefront-pr-123`) is a **pull-request preview**: a real, short-lived
Environment record the platform mints when a PR opens on a service that opted
in via `build.triggers.pullRequests.deploy`, and destroys when the PR closes
or its TTL expires. The `spec.preview` block on the record is the tell — base
environment, service, pull request number, expiry — and it is server-managed
(create, update, and apply refuse it). Two laws to relay:

- **Never try to delete one by hand** — `planton env delete` REFUSES a
  preview, because a record-only delete would orphan its cloud resources. If
  a user asks to remove a preview, close its pull request — that IS the
  delete button (or let the expiry pass).
- **Previews never join promotion order** — they exist beside the durable
  environments without changing any walk.

Everything else about them (authoring the previews tree, the verified URL,
the one-call read) lives in `references/service.preview-environments.md`.

## Deleting a durable environment

Deleting an environment removes Planton's records, secrets, variables, grants
and the Planton-managed state of its resources -- it never touches the cloud.
So `planton env delete <env>` is REFUSED while anything in the environment is
still deployed; the refusal names each resource and the command that destroys
it (`planton infra project undeploy <project>` for a project's resources,
`planton purge <Kind> <slug>` for anything else). Relay the refusal; destroy
first when the user wants the infrastructure gone. When they want it kept
running outside Planton, `--retain-cloud-resources` deletes the environment
and leaves the infrastructure alone -- refused while a running resource keeps
its state in Planton-managed storage, until
`planton state-backend move-off-planton` has moved that state to a backend
they own. Never pass the flag on your own judgment: it is the user's choice to
stop Planton managing live infrastructure.

## One cluster serves multiple environments (the default recommendation)

A developer asking for "dev and prod environments" on Kubernetes does NOT
need two clusters. The default: **one cluster, with environments separated by
namespace** (each environment chart creates its own namespace — see the
two-chart pattern in `kubernetes-architecture.md`). The reasoning, worth
saying out loud:

- AWS manages the EKS control plane for high availability in exchange for its
  flat monthly control-plane fee (the EKS cost fact-sheet has the figure) —
  the resilience a second cluster would buy is largely already paid for.
- A second cluster roughly doubles the always-on bill (control plane, nodes,
  NAT) to buy isolation most solo and small-team setups do not need.
- Namespaces + the gateway's per-hostname routes keep dev and prod cleanly
  separated for everyday purposes.

**Ask the availability/isolation requirement before multiplying clusters.**
Recommend cluster-per-environment only when the user states a hard need:
compliance isolation, blast-radius separation for a real customer-facing SLA,
or wildly different cluster configurations. "It feels safer" gets the honest
cost-benefit answer, then their call.

## The cross-environment authorization nuance

When a cluster deploys, the platform authorizes its auto-created Kubernetes
connection **for the environment the cluster was deployed into**. What that
means per instance type:

- **Local desktop instances**: authorization is not enforced — the one
  cluster/many environments pattern just works. Nothing to do.
- **Hosted instances** (deny-by-default): an environment chart deploying into
  a DIFFERENT environment than the cluster's needs the connection authorized
  for it. The fix is one org-scoped record the user (or you, with consent —
  it is a mutation) applies — widen to the specific environments or to the
  whole organization:

  ```yaml
  apiVersion: connect.planton.ai/v1
  kind: ProviderConnectionAuthorization
  metadata:
    name: <connection-slug>-authorization
  spec:
    provider: kubernetes
    connection: <connection-slug>
    scope: organization        # or scope: environment + authorized_environments list
  ```

Teach this proactively when composing the environment chart for a hosted
user — a deploy denied for authorization after everything else was wired
right is a demoralizing failure you can prevent in one sentence.
