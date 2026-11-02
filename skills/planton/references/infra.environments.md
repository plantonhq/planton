# Environments — Planton's Model and How Many Clusters

Developers speak environment language — "dev", "staging", "prod" — and
translating it into the right infrastructure shape is your job, not theirs.
Two models must not be conflated: what an environment IS in Planton, and how
many clusters environments need.

## The Planton model

- An **organization** owns everything; **environments** partition it
  (dev, prod, …).
- **Infra Components are environment-scoped** — the same resource name can
  exist in dev and prod without collision (which is why chart resource names
  carry `{{ values.env }}`).
- **Connections and charts are organization-scoped** — one AWS connection,
  one chart, usable across environments.
- **A chart deploys into an environment**: the same chart deployed into
  `dev` and `prod` produces two independent sets of resources, differentiated
  by `values.env`. This is how "multiple environments" works — one chart,
  many deployments — never one chart per environment.

## Promotion Paths: the order work moves through environments

An environment carries no order of its own. The order lives in one record
per organization, its **Promotion Paths** (kind `PromotionPathSet`, slug
`promotion-paths`, born with the organization; nobody creates or deletes it):

- **A path** is one product line's way through the environments: a name (a
  slug, like `homes` or `mortgage`) and its stages, first to last. Paths run
  in the order the record lists them.
- **A stage** is a set of environments at one point on the path. Every
  environment of a stage comes after every environment of the stages before
  it; environments in one stage are **siblings**, with no order among them.
  "After dev, demo and prod are both next" is `dev → {demo, prod}`, and each
  sibling is promoted on its own.
- **Promotion Only.** Each path names the first stage that changes only by
  promotion (`reachedByPromotionFromStage`, counted from 1; 0 when pushes
  deploy every stage). A push walks the path's stages and stops there: that
  stage and every one after it are reached by promoting a deployment that
  already ran, or by the deploy verb.
- **An environment sits on at most one stage of one path, or on none.** One
  on no path is never guessed into an order: a trigger-branch push or tag
  release skips it and says why (a branch mapped to it still deploys it),
  and no promotion leads into or out of it. Shared infrastructure
  environments that serve every path through references (a `preprod` or
  `shared` holding shared databases, say) usually sit on no path on purpose.
- **Services and stacks never name a path.** A service follows the paths its
  declared environments sit on; a stack follows its environment's, so the
  same stack name on another path is another stack.

**Creating an environment is order-free; placing it is a second call**, made
as the person:

```
planton env create staging                         # at a terminal it asks where it goes:
                                                   # a path's new last stage, or no path
planton env create prod-eu --path homes --stage 3  # joins stage 3 of homes, beside its environments
planton env create qa --path homes                 # a new last stage of homes
planton env create sandbox --no-path               # on no path: pushes don't deploy it
planton env get prod-eu                            # its place: "homes · stage 3 of 3, beside prod-us"
```

In a script or a coding agent (no terminal), `env create` without `--path`
or `--no-path` lands the environment on no path and says how to place it
("staging is on no Promotion Path, so pushes don't deploy it. Place it:
planton promotion-path list"). A `--path` or `--stage` that does not exist
is refused before anything is created. If placing is refused, the
environment still exists, on no path, and the command says why and exits 1.
In the console, New Environment asks where it goes. Deleting an environment
takes it off its path; a stage or a path left empty disappears.

**Reading and changing the paths** (the console's Environments → Promotion,
titled Promotion Paths & Protection, edits the same record):

```
planton promotion-path list                  # one line per path, in walk order: its stages, siblings
                                             # in braces, where Promotion Only starts; then the
                                             # environments on no path
planton promotion-path list -o yaml          # the record; metadata.version.id is what a change is made from
planton promotion-path set homes --stage dev --stage stage --stage prod-us,prod-eu --promotion-only-from 3
planton promotion-path delete mortgage       # its environments go on no path; nothing deployed is touched
planton apply -f promotion-paths.yaml        # the whole PromotionPathSet document
```

- **`set <name>`** writes one path whole: each `--stage` is a stage, first to
  last, and commas make siblings (`--stage demo,prod`);
  `--promotion-only-from N` names the first Promotion Only stage (left out,
  pushes deploy every stage); `--description` says what the path is for. A
  new name is added after the organization's last path; the other paths are
  untouched. It asks first only when pushes would start or stop deploying an
  environment, showing the path now and after.
- **`delete <name>`** always asks, naming what pushes stop deploying.
- Both take `--org` and `--yes` (`-y`), which answers the question ahead:
  pass it only once the person agreed to exactly this change, or told you to
  carry on without asking.

**Agents (MCP)** use the four twins, each made with the caller's own key:

- `get_promotion_path_set` (`org`): the paths as the caller sees them, with
  `metadata.version.id`. Read it before any change.
- `set_promotion_path` (`org`; `path` with `name`, `stages` as lists of
  environment slugs — `[["dev"], ["demo", "prod"]]` — optional
  `reached_by_promotion_from_stage` and `description`; `version_id`, required):
  replaces the path of that name whole, or adds it after the last path.
- `delete_promotion_path` (`org`, `name`, `version_id`, and `confirm: true`):
  its environments go on no path.
- `apply_promotion_path_set` (`org`, `paths` in walk order, `version_id`
  optional but passed unless the person declares the paths from a file they
  own): replaces every path at once; a path left out is deleted.

`create_environment` leaves a new environment on no path, so pushes never
deploy it; place it with `set_promotion_path`. Every write decides what
pushes deploy: confirm it with the person first.

**Writes are guarded.** Each lands only over the version it was made from; a
stale one is refused ("Promotion Paths changed since you opened them; reload
them and make your change again") — read again, make the change over what
comes back, and never resend the old version. A change needs Update on the
organization's Promotion Paths (Infrastructure Admin and the organization
owner hold it) and Update on every environment whose place it moves, so
moving prod beside demo, or moving where Promotion Only starts past prod,
needs prod's grant; a protected environment needs its explicit one. Adding a
new last stage, or adding to the last stage, moves no one else. Reading
needs only membership of the organization.

**Stage numbers.** A Delivery Path and a run number stages as the path itself
does, so a service with no environment in a path's second stage shows stages
1 and 3. The Promotion Paths reads (`planton env get`, the console's
Promotion page, `get_promotion_path_set`) leave out the environments the
reader cannot see, and a stage left with none, and number the stages the
reader sees. The two agree for anyone who can see every environment.

## Preview environments are environments too

An environment whose name looks like `{service}-pr-{number}` (e.g.
`storefront-pr-123`) is a **pull-request preview**: a real, short-lived
Environment record the platform mints when a PR opens on a service that opted
in via `build.triggers.pullRequests.deploy`, and destroys when the PR closes
or its TTL expires. The `spec.preview` block on the record is the tell — base
environment, service, pull request number, expiry — and it is server-managed
(create, update, and apply refuse it). Two laws to relay:

- **Never try to delete one by hand** — `planton env delete` REFUSES a
  preview, because a record-only delete would leave its Infra Components
  running with no environment to answer for them. If
  a user asks to remove a preview, close its pull request — that IS the
  delete button (or let the expiry pass).
- **Previews never sit on a Promotion Path** — a path refuses one; they
  exist beside the durable environments without changing any walk.

Everything else about them (authoring the previews tree, the verified URL,
the one-call read) lives in `references/service.preview-environments.md`.

## Deleting a durable environment

Deleting an environment removes Planton's records, secrets, variables, grants
and the Planton-managed state of its resources -- it never touches the cloud.
So `planton env delete <env>` is REFUSED while anything in the environment is
still deployed; the refusal names each resource and the command that destroys
it (`planton infra stack undeploy <stack> --env <env> --yes` for an Infra
Stack's components, `planton purge <Kind> <slug> --yes` for anything else --
`--yes` only after the user agreed, or told you to carry on without asking,
since nothing in your session can answer the CLI's question; then say exactly
what was destroyed). Relay the refusal; destroy
first when the user wants the infrastructure gone. When they want it kept
running outside Planton, `--retain-infra-components` deletes the environment
and leaves the infrastructure alone -- refused while a running resource keeps
its state in Planton-managed storage, until
`planton infra state-backend move-off-planton` has moved that state to a backend
they own. Never pass the flag on your own judgment: it is the user's choice to
stop Planton managing live infrastructure.

The delete itself asks, naming what goes with the environment, and `--yes`
answers ahead; without a terminal it refuses and prints itself with `--yes`
added. Pass it only once the user agreed to lose the environment.

## One cluster serves multiple environments (the default recommendation)

A developer asking for "dev and prod environments" on Kubernetes does NOT
need two clusters. The default: **one cluster, with environments separated by
namespace** (each environment chart creates its own namespace — see the
two-chart pattern in `cloud.kubernetes-architecture.md`). The reasoning, worth
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
