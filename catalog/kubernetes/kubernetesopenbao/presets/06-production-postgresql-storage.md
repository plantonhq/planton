# Production on PostgreSQL storage preset

An OpenBao server that keeps its data in a PostgreSQL database you
already run and back up, instead of on its own volume. The database is
a `KubernetesPostgres` declared BY REFERENCE: `host` resolves to its
`-rw` Service and `passwordSecret.secretName` to its operator-maintained
credential Secret, so the password never appears in a manifest and never
enters the server's configuration — it reaches the server as the
standard `PGPASSWORD` environment variable from that Secret, beside
`PGHOST`, `PGPORT`, `PGDATABASE`, `PGUSER`, and `PGSSLMODE`. OpenBao
calls this backend production-ready; it is transactional and HA-capable.

WHAT YOU GET AND WHAT YOU GIVE UP: no data volume (the audit volume is
still yours to declare), high availability through the backend's lock
table at any replica count (raise `server.replicas` for availability —
there is no quorum arithmetic here, one live server is availability),
and disaster recovery that is your database's: the vault is backed up
whenever the database is, and restored whenever the database is. In
exchange there are no Raft snapshots — a `backup` block is refused on
this engine, with the reason.

THE DATABASE SIDE IS YOURS: the referenced cluster must bootstrap the
`openbao` database owned by the `openbao` role (`bootstrap.initdb` on
the `KubernetesPostgres`); OpenBao creates its two tables on first
start, and ownership covers them. The referenced Secret must live in
the vault's namespace — a Secret is namespace-local. `maxParallel`
bounds the connections EACH server opens: OpenBao's default is 128 per
server, a default PostgreSQL allows 100 in total, and the replica count
multiplies it — bound it on a shared database.

THE BOOTSTRAP IS YOURS, by design, exactly as on Raft: fresh pods run
but report NotReady (the readiness probe is `bao status`, which fails
for sealed servers). Initialize once through pod 0 (`bao operator
init`), then unseal EVERY pod. After every pod restart the affected
server is SEALED again until unsealed — that is Shamir-mode reality,
and an `autoUnseal` arm exists to remove exactly that step. TLS to the
database defaults to `require` (encrypted, no CA check); choose
`verify-full` and mount the cluster's CA through `helmValues` when the
network is not yours.

Change first: `server.postgresql.host` and `passwordSecret.secretName`
to your database's name, `database`/`username` to the pair its
`initdb` bootstraps, `maxParallel` to your database's headroom, and
`server.replicas` for availability.

See [06-production-postgresql-storage.yaml](./06-production-postgresql-storage.yaml) for the manifest.
