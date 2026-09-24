# Kubernetes OpenBao

## When NOT to Use This

**One resource is ONE OpenBao install** — the Linux
Foundation-governed secrets manager (MPL-2.0 fork of Vault): secret
storage, dynamic secrets, encryption as a service — from the official
`openbao` chart (0.28.x = server 2.6.x).

Not the right component when:

- **A managed service already covers you** — the platform's managed
  cloud KMS and secret-manager kinds exist for teams that want keys
  and secrets without operating a server.
- **You need HashiCorp Vault compatibility guarantees** — OpenBao
  forked from MPL-2.0 Vault and evolves independently; behavior past
  the fork point is not guaranteed to track Vault.

## The seal lifecycle

The fact everything else follows from: a fresh server starts
UNINITIALIZED and SEALED. `bao operator init` (which generates the
unseal key shares and the root token) and unsealing are RUNTIME API
operations no deployment tool performs — this component deliberately
does not try. Until then the pod reports NotReady BY DESIGN (the
readiness probe is `bao status`, non-zero for sealed servers); the
chart keeps sealed pods addressable through its Services, so
port-forward and the DNS names work for the init/unseal calls. In
Shamir mode every restart returns a SEALED server. Auto-unseal
(below) removes the unseal step from restarts; the one-time
initialization is always yours — with auto-unseal it produces
RECOVERY keys instead of unseal keys.

## The storage engine and the replica count

The server declares what OpenBao has: `dev`, or a storage engine and
`replicas`. Declaring nothing means one server on integrated Raft — a
legal Raft cluster of one, with transactions and snapshots.

- `raft` — integrated storage: one volume per replica
  (`raft.dataStorage`), leader election, and the engine `backup` and
  `restore` exist for (Raft snapshots). This module synthesizes the
  `retry_join` stanzas for every peer — the chart alone ships NONE,
  and without them a multi-replica install never forms a cluster.
- `postgresql` — OpenBao's production-ready external backend, declared
  BY REFERENCE to a `KubernetesPostgres` (host from its `-rw` Service,
  password from its operator-maintained Secret). No volume; HA through
  the backend's lock table; the vault is backed up by its database's
  own backup, so `backup` is refused on this engine. The connection
  reaches the server as the standard `PG*` environment and the
  password never enters a manifest or the server's configuration.
- `dev` — in-memory, auto-initialized, root token literally `root`,
  no volume, ServiceAccount annotations dropped (a chart behavior) —
  evaluation only, never real secrets. Takes no engine and one replica.

The Helm chart offers modes and a raw configuration string; this kind
declares the engine and turns the chart's own knobs from it (Raft
through the chart's Raft mode, PostgreSQL through its HA mode with Raft
off and no data volume, dev through its dev flag). Every server on an
engine runs the chart's HA mode, so the `-active` Service exists at
every replica count. Scheduling truth: the chart's REQUIRED pod
anti-affinity means replicas need as many schedulable nodes; relax it
through `helm_values` in labs only. The chart's PodDisruptionBudget is
disabled at one replica (its default there blocks every node drain
while protecting nothing) and follows the count above it.

## Auto-unseal

Four seal arms: `awsKms`, `gcpKms`, `azureKeyVault`, `transit`.
Keyless-first: on EKS/GKE/AKS annotate the server ServiceAccount for
workload identity and leave the credential fields empty. Static
credentials, when unavoidable, ride a module-owned Secret delivered
as environment variables — nothing credential-bearing lands in the
config ConfigMap. Version horizon: the cloud KMS seals are built in
but deprecated at the pinned 2.6.x — upstream moves them to external
plugins at 2.7.

## TLS is a composite

The chart's `global.tlsDisable` value alone does NOT configure the
listener — flipping it produces a plaintext server addressed as
https, an instant outage. The `tls` block is owned end to end:
listener cert/key files, the certificate Secret mount, and every
derived URL and probe switch together. A `KubernetesCertificate` is
the natural issuer for `cert_secret_name`.

## Injector, metrics, backups

The Agent Injector is OFF by default — a deliberate divergence from
the chart: it is a CLUSTER-WIDE mutating webhook on pod creation,
fail-open by default (downtime skips injection rather than blocking
pods). Metrics, when enabled, make /v1/sys/metrics UNAUTHENTICATED on
the listener — that is how Prometheus scrapes. `backup` is the Raft
disaster-recovery story (snapshots exist only for Raft — a vault stored
in PostgreSQL is backed up by its database and refuses the block; dev
has nothing to snapshot): scheduled snapshots taken through OpenBao's
own API and shipped to S3, Google Cloud Storage, Azure Blob, or
Cloudflare R2 — each in its own vocabulary, by reference to the
catalog's bucket, identity, and token kinds, keyless where the cloud
offers it — with one runtime prerequisite: the job's Kubernetes-auth
login inside OpenBao is a four-command recipe run after initialization
(the spec prints it; so does a failing job). `restore` declares a fresh
cluster's recovery from that store, offered when a seal arm is set so
the same key unseals what comes back. `helm_values` merges last for
chart surfaces deliberately not modeled; `fullnameOverride` is
re-pinned after the merge.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
