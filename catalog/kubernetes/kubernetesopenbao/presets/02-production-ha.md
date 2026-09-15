# Production HA (integrated Raft) preset

Three OpenBao servers with integrated Raft storage: each replica
persists to its own 10Gi PVC, the module synthesizes the `retry_join`
stanzas for every peer (the chart alone ships none — without them a
multi-replica install never forms a cluster), and the cluster elects a
leader. Losing any single node loses neither data nor availability. A
dedicated audit volume is mounted at `/openbao/audit`; enable auditing
after initialization with
`bao audit enable file file_path=/openbao/audit/audit.log`.

THE BOOTSTRAP IS YOURS, by design: fresh pods run but report NotReady
(the readiness probe is `bao status`, which fails for sealed servers).
Initialize once through pod 0 (`bao operator init` — custody of the
unseal key shares and root token is the whole point of a secrets
manager), then unseal EVERY pod; peers join automatically through
retry_join. After every pod restart the affected server is SEALED
again until unsealed — that is Shamir-mode reality, and the
auto-unseal preset exists to remove exactly that step.

Scheduling truth: the chart ships a REQUIRED pod anti-affinity on
hostname, so three replicas need three schedulable nodes.

Change first: `server.ha.replicas` (odd counts only — 5 tolerates two
losses), storage sizes to your churn, and declare a `backup` block —
scheduled Raft snapshots to S3, GCS, Azure Blob, or Cloudflare R2 are the
disaster-recovery story, and a fresh cluster with the same seal key
restores from them declaratively.

See [02-production-ha.yaml](./02-production-ha.yaml) for the manifest.
