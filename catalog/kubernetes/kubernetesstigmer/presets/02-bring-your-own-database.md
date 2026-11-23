# Bring Your Own Database

Stigmer on a Postgres and a Temporal you already run: a KubernetesPostgres and a KubernetesTemporal in the same environment, wired by reference, so the InfraPipeline deploys them first. The operator runs neither, so CloudNativePG is never installed for Stigmer.

## When to Use

- Your team already runs Postgres and Temporal and wants Stigmer's records there
- You back up and upgrade databases with your own tooling

## Key Configuration Choices

- **`database.external` with `temporal.external`** — the two go together: Stigmer's operator creates Temporal's databases only on a Postgres it runs
- **`password.value`** — a Planton secret reference, written into a Secret in the Stigmer's namespace. When the database runs in the same namespace, `password.existing_secret` can name a KubernetesPostgres's own application Secret instead
- **No front door** — your own Postgres cannot back the bundled sign-in (it keeps its accounts on the bundled database). To expose this install, sign people in through your own issuer (`identity.oidc`)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `stigmer-db` | Your KubernetesPostgres | The Postgres resource in this environment |
| `stigmer-temporal` | Your KubernetesTemporal | The Temporal resource in this environment |
| `user` | The role Stigmer signs in as | It must own the `stigmer` database |
| `$secret/stigmer-db-password` | That role's password | A Planton secret |

## Related Presets

- **01-minimal** — everything bundled
