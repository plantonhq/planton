# Kubernetes OpenFGA

## When NOT to Use This

**One resource is ONE OpenFGA server** — the CNCF authorization
engine implementing Google-Zanzibar-style relationship-based access
control, from the official `openfga` chart (0.3.x = OpenFGA 1.18+),
on a datastore you pick.

Not the right kind when:

- **You want roles and permissions inside your identity provider** —
  that is `KubernetesKeycloak` territory. OpenFGA answers "is user U
  related to object O in the required way" from stored relationship
  tuples, not "what roles does this token carry".
- **You want a managed authorization service** — point your
  applications at it and deploy nothing here.
- **You want to declare the authorization data** — stores, models
  and tuples are API-managed, never deployment config. This kind
  installs the ENGINE; the platform's `OpenFgaStore`,
  `OpenFgaAuthorizationModel` and `OpenFgaRelationshipTuple` kinds
  compose against the exported `api_http_endpoint`.
- **The memory datastore, beyond a demo** — the zero-dependency
  evaluation arm: data lost on every restart, replicas forced to 1.

## The datastore contract

Exactly one engine: `postgres` (recommended — a `KubernetesPostgres`
composes by reference: the host resolves to its read-write Service
and the password rides the operator-maintained `<cluster>-app`
Secret), `mysql`, or `memory`. Credentials go through Secret
indirection everywhere: the connection URI renders WITHOUT
credentials and username/password arrive as environment variables
read from the Secret — nothing credential-bearing lands in rendered
values. One Kubernetes constraint: a secretKeyRef reads only its own
namespace's Secrets — co-locate OpenFGA with its database or
replicate the Secret.

## Migrations that cannot deadlock

`openfga migrate` runs as an idempotent init container in every
server pod. The chart's default hook-Job mode is deliberately not
used: a post-install hook Job deadlocks engines that wait on rollout
readiness, and its post-delete hook dials the database during
uninstall. `migration_timeout` (default 3m) covers the composed
case where the database provisions concurrently with OpenFGA.

## Authentication

Unset means an OPEN API — anyone who can reach the Service reads and
writes every store; lab clusters only. Pre-shared keys reach the
server through a Secret in BOTH arms: declare them and the module
materializes `<name>-authn-keys`, or point at a Secret you maintain
(comma-separated under the data key `keys` — the chart's contract).
OIDC requires BOTH issuer and audience — the server hard-fails at
startup without them (audience validation became mandatory at v1.18).

## Playground, always off

The chart ships its demo playground enabled; this module always
disables it. Upstream turned it off by default for security at
v1.18, the server refuses to start when it is combined with any
authentication method, and at this version it binds pod-local
anyway. Evaluate models with the `fga` CLI or the VS Code extension
against the API instead.

## Scaling and the closed schema

OpenFGA is stateless — all state lives in the datastore, so replicas
scale checks linearly, and the `hpa` arm hands the count to an
autoscaler. The chart's values schema is CLOSED
(`additionalProperties: false`): `helm_values` can only override
keys the chart defines (`extraEnvVars` carries the ~50 server flags
without values paths), never invent new ones; `fullnameOverride` is
re-pinned after the merge.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
