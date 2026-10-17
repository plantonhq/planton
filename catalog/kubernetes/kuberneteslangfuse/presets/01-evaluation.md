# Evaluation

The fastest way to see traces in Langfuse: one web replica and one worker against four in-cluster stores, a single ClickHouse server with cluster mode off, and every secret generated. Langfuse seeds an organization, a project and an API key pair at first start, so an SDK can send a trace as soon as the install is ready.

## When to Use

- Trying Langfuse on a development cluster before committing to a production shape
- A team sandbox where losing the data is acceptable
- Checking that your application's Langfuse instrumentation works end to end

## Key Configuration Choices

- **Stores by reference** (`postgres`, `clickhouse`, `redis`, `blob_storage`) -- each reference names only the resource (`langfuse-pg`, `langfuse-ch`, `langfuse-valkey`, `langfuse-s3`); the kind and field path come from the field's composition default.
- **One namespace for everything** (`namespace: langfuse`, no `create_namespace`) -- store passwords reach the pods through secretKeyRefs, which cannot cross namespaces, so the stores create the namespace and Langfuse joins it.
- **Cluster mode off** (`clickhouse.cluster.enabled: false`) -- a single ClickHouse server without Keeper. Switching later means migrating data, so pick the production preset if this install will grow up.
- **Generated secrets** -- the salt and encryption key are generated and kept across applies, but a delete and recreate against the same Postgres generates new ones and breaks every API key. Fine for evaluation, not for data you keep.

## Placeholders to Replace

This preset has no placeholders. Rename the store references if your stores use other names, and make sure they exist in the `langfuse` namespace: a PostgreSQL with the `langfuse` database owned by the `langfuse` user, a ClickHouse with a `langfuse` user that declares `networks`, a Valkey with `maxmemory-policy noeviction`, and a SeaweedFS that declares the `langfuse` bucket.

## Related Presets

- **02-production-in-cluster** -- the same stores with replicated workloads, durable secrets, a seeded admin and cluster mode on.
