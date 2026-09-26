# Variables and Secrets — the `$var` / `$secret` Reference Grammar

Planton manifests never carry credentials, and rarely carry org-specific
config literals. Instead, a string field can hold a REFERENCE to a value
managed in the platform's config manager: a variable (plain config) or a
secret (encrypted, resolved just-in-time inside the deployment runner, never
stored in the manifest or readable back out of it). A variable reference
works in any string field; a secret reference works in any field EXCEPT the
ones a component marks as read by every viewer -- see "Where a secret
reference may go" below, because the field, not the reference, decides
whether a secret stays secret. This file is the grammar, when to use it, and
how to ground references against what actually exists.

This is a DIFFERENT instrument from `valueFrom` (`dependencies.md`):

| Value comes from | Wire it with |
|---|---|
| Another resource's deployment output (an id, ARN, endpoint the platform creates) | `valueFrom` — see `dependencies.md` |
| An operator-managed config value (a region-independent setting, a team-owned constant) | `$var/...` |
| A credential or any sensitive value (password, API key, token, private key; a workload's `env.secrets[].value`, every value in a `KubernetesSecret`, an Auth0 action's secret) | `$secret/...` — the ONLY thing a sensitive field accepts; a value written there is refused before anything is created, on every write path |
| A registry login a Kubernetes workload pulls with (`spec.pod.imageRegistries[].password`, a `KubernetesSecret`'s docker-registry password, a GHCR connection's `pullToken.token`) | `$secret/...` — and for the service's own registry, usually nothing: the deploy fills it from the registry connection (`references/service.pulling-private-images.md`) |

## The grammar

Both kinds share one shape. Scope is encoded in the reference itself:

```
$var/<slug>                    # organization-scoped variable
$var/@<env>/<slug>             # environment-scoped variable
$var/<group>/<entry>           # entry in a variable group (org scope)
$var/@<env>/<group>/<entry>    # entry in a variable group (env scope)

$secret/<slug>                 # organization-scoped secret
$secret/@<env>/<slug>          # environment-scoped secret
$secret/<name>/<key>           # key inside a key-value secret (org scope)
$secret/@<env>/<name>/<key>    # key inside a key-value secret (env scope)
```

The `@` sigil as the FIRST path segment is the scope switch, and it is
strict: an org-scoped reference never resolves an environment-scoped record
and vice versa — there is no fallback. A secret that exists only in the
`staging` environment is `$secret/@staging/db-password`; writing
`$secret/db-password` for it fails validation with "secret not found",
exactly like a misspelled slug. Scope is part of the address.

Example — a sensitive field on a database user, environment-scoped:

```yaml
spec:
  userName: app
  password: $secret/@staging/db-password
```

## Sensitive fields accept ONLY a secret reference

Every field a component's schema marks sensitive (the explain report and the
component reference page flag these) rejects plaintext before anything
deploys — the control plane validates that the field holds a well-formed
`$secret/...` reference to an EXISTING secret. So for a password/key/token
field there are exactly two failure modes to avoid:

- a plaintext value (rejected outright — never write one, not even a
  placeholder), and
- a reference whose slug or scope does not match a real secret (rejected as
  not found).

Never expose a chart param that asks the user to paste a credential — that
is the same failure as a plaintext literal, one step removed. The param, if
any, carries the REFERENCE (see "In chart templates" below).

## Where a secret reference may go — and why

The runner resolves every `$secret/...` to its plain value just before the
component's module runs, and the module writes that value wherever the
manifest put it. A reference is therefore only as secret as the field it sits
in. Components mark the two kinds of field that matter, and the explain
report (`planton explain <Kind>`) and the component reference page show both
marks:

- **`(sensitive)`** — secret material. Accepts only a `$secret/...`
  reference (rules above), and the module keeps the value out of anything a
  viewer reads.
- **`(no secrets: use <field>)`** — the value is written where anyone who
  can view the resource reads it: an environment variable in a Kubernetes
  pod template, a Cloud Run revision, an ECS task definition (and ECS keeps
  every revision forever). A secret reference anywhere inside such a field is
  refused before anything deploys, and the refusal names the sibling field
  to move it to — the component's SECRET HOME, which keeps the value in a
  secret store the workload reads by reference.

On the three targets a service deploys to, the pattern is one field for
configuration and a sibling for secrets. Read the marks on the page rather
than memorizing this table — it shows the shape, not the full list:

| Kind | Configuration (every viewer reads it) | A Planton secret goes in |
|---|---|---|
| `KubernetesDeployment` (any Kubernetes workload) | `env.variables[].value` | `env.secrets[].value` — kept in a Kubernetes Secret the workload owns |
| `GcpCloudRun`, `GcpCloudRunJob` | `env[].value` | `env[].secretValue` — a Secret Manager secret the service owns |
| `AwsEcsTaskDefinition` | `environment` | `secretEnvironment` — a Secrets Manager secret only the execution role reads |

```yaml
# GcpCloudRun container
env:
  - name: LOG_LEVEL
    value: info
  - name: STRIPE_KEY
    secretValue: $secret/@production/stripe-key

# AwsEcsTaskDefinition container
environment:
  LOG_LEVEL: info
secretEnvironment:
  STRIPE_KEY: $secret/@production/stripe-key
```

What a developer should hear, in their terms, when you choose the home:

- **The value is copied into their cloud's own secret store** — exactly as
  a Kubernetes workload's secrets become a Kubernetes Secret. The copy
  belongs to the workload: deleting the service deletes it. Only the
  workload's runtime identity can read it (on Cloud Run, give the service a
  dedicated service account; without one the grant goes to the project's
  shared Compute Engine default account).
- **Rotation is a deploy.** The workload is pinned to the exact version
  stored, so changing the secret in Planton redeploys nothing by itself;
  the next deploy of that environment — a push, a promotion, a rollback, a
  CLI deploy: every one resolves the reference again — carries the new
  value as a new revision, never as instances silently disagreeing.
- **The value also sits in the deployment's IaC state**, as it does for
  every secret a module writes; that is why state lives in a backend the
  organization controls.
- **When the secret has another owner** (another team rotates it, several
  services share it), point at the store directly instead: Cloud Run's
  `valueFromSecret`, ECS's `secrets` (an ARN), a Kubernetes
  `env.secrets[].secretRef`. Then the grant and the rotation are the
  owner's.

The refusal, when it happens, names the field and the home: "…
`spec.containers[0].environment.STRIPE_KEY` holds a secret reference, but
its value is stored where anyone who can view the resource reads it -- move
the reference to `secretEnvironment` …". Relay it and make the one move it
names — never "fix" it by pasting the value as a literal
(`references/infra.issue-catalog.md`).

## Ground before you write — never invent a slug

References are validated against the org's real records, so look up what
exists before writing one, exactly as you ground field names with
`planton explain`:

```
planton secret list -o json      # every secret; each record's "env" field
planton variable list -o json    #   distinguishes org- from env-scoped
```

Use `-o json`: the JSON records carry each entry's `env` (empty = org
scope), which the human-readable table does not show — and the scope decides
which reference form you write. On the platform-tools arm (no CLI), check
your roster for a config-manager search/list tool; when none exists, you
cannot verify existence — treat every secret you reference as
possibly-missing and follow the missing-secret protocol below. Never fake a
lookup or claim a secret exists unverified.

## Referencing a secret that does not exist yet

Normal, not an error: the reference is the design; the secret's value is
deployment-time material only its owner should supply. When the lookup finds
no matching secret (or no lookup instrument exists), still write the
reference — never plaintext, never a placeholder — and then say so plainly
in the explain-after:

1. name every reference you wrote whose secret does not exist yet,
2. say the secrets must be created before deploy, and
3. hand the user the exact ready-to-run commands, offering to run them
   yourself once they supply the values (a mutation — one confirmation), or
   point at the console's Secrets page as the click path:

```
planton secret set db-password value=<the-value> --env staging   # env-scoped
planton secret set api-key value=<the-value>                     # org-scoped
```

A composed chart with declared-but-uncreated secrets is honest and
deployable after one step; a chart with an invented value is a silent
failure.

## Creating variables and secrets (mutations — confirm first)

```
planton secret set <slug> value=<value> [--env <env>]     # key=value pairs
planton secret set <slug> --from-file value=./key.json    # file-backed value
planton variable set <slug> <value> [--env <env>]         # value is POSITIONAL
```

Note the asymmetry: `secret set` takes `key=value` pairs (a secret can be a
key-value map); `variable set` takes the value as a positional argument.
`--env` makes the record environment-scoped; without it the record is
org-scoped. Re-running `secret set` on an existing secret writes a new
version (rotation), never a duplicate.

## In chart templates

Slugs are DNS-compatible (`[a-z0-9-]`) and can never contain `@`, so the env
sigil composes cleanly with Jinja — the canonical pattern for a chart that
deploys per-environment:

```yaml
spec:
  password: $secret/@{{ values.env }}/db-password
```

Prefer carrying the WHOLE reference through a param over hardcoding a slug,
so the chart stays portable across orgs whose secrets are named differently
(param defaults are plain values — template expressions live in templates/,
never in values.yaml):

```yaml
# values.yaml
params:
  - name: db_password_secret_ref
    description: Secret reference for the database password (e.g. $secret/@staging/db-password)
    value: $secret/db-password

# templates/database.yaml
spec:
  password: "{{ values.db_password_secret_ref }}"
```

Either way the manifest field ends up holding a `$secret/...` string — the
param is a naming convenience, never a channel for the raw value.
