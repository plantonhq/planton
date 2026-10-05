# Dev preset

The smallest useful Temporal: all four server services, the Web UI,
and one Temporal namespace (`default`) — against a composed
KubernetesPostgres named `temporal-db` in the same Kubernetes
namespace. Point a worker at the exported `frontend_endpoint` and run
workflows; reach the UI over the port-forward command in the
outputs.

The database references do all the wiring: the host resolves to the
Postgres read-write Service and the credential to the
operator-maintained application Secret, so nothing password-shaped is
ever written into this manifest. What the preset expects from the
database side: the KubernetesPostgres declares `temporal` as its
bootstrap database (owner `temporal`) and creates
`temporal_visibility` via one `post_init_sql` line — both owned by the
same user the schema Jobs connect as.

Know the one irreversible default: `num_history_shards` (512) is baked
into the schema at first install and cannot be changed later. 512 is
the upstream production default — fine to keep even for dev, but never
lower it "because dev".

Change first: size the four services (`services`) before real load —
history is the heavy one; declare more `temporal_namespaces` as teams
arrive.

See [01-dev.yaml](./01-dev.yaml) for the manifest.
