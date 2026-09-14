---
kinds:
  - CloudflareR2Bucket
  - CloudflareAccountApiToken
  - KubernetesPostgres
  - KubernetesMongodb
  - KubernetesOpenBao
---

# Stateful Kind Disaster Recovery: the Backup Store, the Identity, and the Restore

Three stateful kinds in this catalog — the PostgreSQL database, the MongoDB
replica set, and the OpenBao secrets vault — back themselves up to an object
store the adopter owns and restore a fresh instance from it by declaration.
They share one composition shape and one set of failure modes, and a
proposal that gets the shape wrong fails late: a backup job that authenticates
anonymously, two live instances corrupting one archive, or a restore that
"succeeds" into an empty target. This pattern states the shared truth once;
each kind's guide carries what is that kind's alone.

## The problem

"Can I get it back?" is the first question a team asks of any stateful
component, and the answer is a composition, never one manifest: the store,
the credential the instance presents to it, the instance's backup
declaration, and — on the bad day — a second instance declared to restore
from the first one's store. Each hop is a catalog kind wired by reference,
and each hop has a way to go wrong that no single kind's schema can refuse:

- **Anonymous uploads.** A keyless posture (`keyless: true` on an arm)
  authenticates as the pods' cloud identity. Declare it without the identity
  and every upload fails at runtime as an anonymous request; the manifest is
  valid.
- **Two writers, one archive.** Every kind prunes or archives under a path or
  prefix it assumes it owns alone. Two live instances sharing one corrupt each
  other's history — the second writer's pruning deletes the first's objects.
- **A restore into the wrong key.** OpenBao's snapshot is protected by the
  seal key that took it; PostgreSQL's and MongoDB's restores carry the
  source's roles and passwords. A target that does not hold the same key (or
  reference the same credential Secret) restores data nobody can read.
- **A bucket that cannot be retired.** Cloudflare R2 refuses to delete a
  non-empty bucket and the provider has no force-destroy; a store that was
  never emptied blocks its own teardown.

## The composition

Four nodes on the store side and one or two on the instance side, every edge
a `valueFrom` reference so the diagram shows the whole recovery path:

1. **The store** — a bucket kind in the cloud the adopter chose:
   `GcpGcsBucket`, or `CloudflareR2Bucket` for a store outside the cloud that
   runs the cluster. The kind's outputs (`bucket_name`, and for R2
   `account_id` and `jurisdiction`) are what the instance references, so the
   store follows the bucket and never drifts.
2. **The identity or the credential** — keyless where the cloud allows
   (`GcpServiceAccount` + `GcpGkeWorkloadIdentityBinding` on GKE, an IAM role
   on EKS, a managed identity on AKS), declared keys where it does not. R2 has
   no keyless posture from any cluster: its credential is a
   `CloudflareAccountApiToken` whose id and SHA-256 value ARE the S3 key pair,
   exported as `r2_access_key_id` / `r2_secret_access_key`.
3. **The instance's backup block** — the store in its own vocabulary (never
   S3-with-an-endpoint for R2), a path or prefix the instance owns alone, a
   schedule, and a retention.
4. **The restore target, on the bad day** — a second instance of the same kind
   declaring the SAME store and path, the restore member the kind offers, and
   the credential-continuity reference the kind needs (OpenBao: the same
   `autoUnseal` key and a Secret for the fresh cluster's init root token;
   PostgreSQL: the source's `<name>-app` Secret; MongoDB: the source's
   `<name>-secrets` Secret).

The R2 store pair, complete — one private bucket and one token scoped to it
with the least privilege an S3 client that writes backups needs (`Workers R2
Storage Bucket Item Write`: read, write, list objects in THIS bucket, nothing
at the account level). The policy's resource string spells the bucket by
account, jurisdiction, and name because a map key cannot carry a reference:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareR2Bucket
metadata:
  name: prod-archive
spec:
  bucketName: prod-archive
  accountId: 0123456789abcdef0123456789abcdef
  jurisdiction: default
---
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAccountApiToken
metadata:
  name: prod-archive-writer
spec:
  account_id: 0123456789abcdef0123456789abcdef
  name: prod-archive-writer
  policies:
    - effect: allow
      permission_group_ids:
        - 2efd5506f9c8494dacb1fa10a3e7d5b6
      resources:
        "com.cloudflare.edge.r2.bucket.0123456789abcdef0123456789abcdef_default_prod-archive":
          permission: "*"
```

The vault on that store — the OpenBao kind's own `05-production-ha-r2-backups`
preset, where the store is declared entirely by reference and the module
composes the S3 dialect underneath:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenBao
metadata:
  name: openbao
spec:
  namespace:
    value: openbao
  createNamespace: true
  server:
    ha:
      replicas: 3
    dataStorage:
      size: 10Gi
  backup:
    schedule: "0 * * * *"
    retentionDays: 14
    objectStore:
      prefix: openbao/openbao
      r2:
        bucket:
          valueFrom:
            kind: CloudflareR2Bucket
            name: prod-archive
            fieldPath: status.outputs.bucket_name
        accountId:
          valueFrom:
            kind: CloudflareR2Bucket
            name: prod-archive
            fieldPath: status.outputs.account_id
        jurisdiction:
          valueFrom:
            kind: CloudflareR2Bucket
            name: prod-archive
            fieldPath: status.outputs.jurisdiction
        credentials:
          accessKeyId:
            valueFrom:
              kind: CloudflareAccountApiToken
              name: prod-archive-writer
              fieldPath: status.outputs.r2_access_key_id
          secretAccessKey:
            valueFrom:
              kind: CloudflareAccountApiToken
              name: prod-archive-writer
              fieldPath: status.outputs.r2_secret_access_key
```

The database and the replica set declare the same store pair in their own
shapes — PostgreSQL's archive is a `destinationPath` with the bucket inside
it (`s3://prod-archive/prod-db`), MongoDB's is a named storage with a
`bucket` and a `prefix`. Their complete manifests are the kinds' own R2
presets (`kubernetespostgres/presets/05-gke-ha-r2-backups.yaml`,
`kubernetesmongodb/presets/04-gke-replica-set-r2-backups.yaml`); the wiring
onto `CloudflareR2Bucket` and `CloudflareAccountApiToken` is byte-for-byte
the vault's.

## Choices and consequences

| Choice | What it buys | What it costs | On the diagram |
|---|---|---|---|
| **Keyless posture** (GKE Workload Identity for GCS, EKS IRSA for S3, AKS Workload Identity for Azure Blob) | No stored credential anywhere; rotation is the cloud's | Only where the cluster and the store share a cloud; the identity and its binding are two more nodes, and the binding names the exact ServiceAccount the kind renders (a restore target is another instance and needs its own binding) | The identity is a visible node with an edge to the bucket and to the instance |
| **Declared keys** (an access-key pair, a service-account key, an Azure key or connection string) | Works from any cluster to any store | A live credential the module materializes as a Kubernetes Secret; reference it (`key_base64`, an organization secret), never paste it | The credential source is a node only when it is a catalog kind; a pasted key is invisible |
| **Cloudflare R2 by reference** | The store outside the cloud that runs the cluster; one token per bucket, least privilege; the store follows the bucket's jurisdiction | No keyless posture exists; emptying the bucket before teardown is yours | Bucket and token are two nodes; the instance draws edges to both |
| **One path or prefix per live instance** | Retention and archiving stay correct | A restore target must declare the SOURCE's path to read it — the one deliberate sharing — and then (PostgreSQL) archive its own to a NEW path, or (OpenBao) suspend its schedule while the restore is declared | None; the path is a string, which is why the rule is taught here |
| **The GCS bucket's two roles** | rclone, Barman, and the storage clients read the bucket's attributes before writing | `roles/storage.objectAdmin` alone fails with `storage.buckets.get`; add `roles/storage.legacyBucketReader` on the bucket's `iamMembers` | None |

## The restore is a declaration, and one step stays yours

Every kind here restores by declaring a fresh instance against the source's
store; none can perform the runtime handshake that makes the restored state
readable, and each guide names the one step that stays with the operator:

- **OpenBao** — the snapshot is bound to the seal key that took it, so a
  declared restore requires the same `autoUnseal` key on both sides. The
  fresh cluster is initialized by hand (`bao operator init`), and the
  returned root token is handed to the restore Job through a Secret the spec
  names; the token stops existing when the source's state lands. While the
  restore is declared the target's backup schedule is suspended. A Shamir
  vault has no key to point a restore at; it is restored by hand with the
  source's shares present (the guide's runbook).
- **PostgreSQL** — the recovered data carries the source's roles and
  passwords; the target references the source's `<name>-app` Secret, so that
  Secret must outlive the source (back it up with the archive). The target
  archives to a new path.
- **MongoDB** — the restored database carries the source's users; the target
  references the source's `<name>-secrets` Secret, and the restore waits for
  the replica set to form before the Restore object exists.

Two consequences follow. First, a declared restore that is waiting is in a
designed state, not a failed one — the vault's restore pod sits in
`CreateContainerConfigError` until the token Secret exists; MongoDB's Restore
object is not even created until the new replica set reports ready — and the
kind's field doc names the state, so read it before calling the deploy stuck.
Second, "restore again" means different things per kind, and deleting the
wrong object is how a restore runs over live data. OpenBao's restore Job and
MongoDB's Restore object are each named by a hash of the declaration: a
changed declaration is a new restore, and a deleted object is recreated on
the next apply and restores AGAIN over whatever the instance has written
since — change the declaration to restore deliberately; never delete the
object to "clean up". PostgreSQL's recovery is the cluster's bootstrap: it
runs once, when the cluster is first created, and a second recovery is a
second cluster.

## When not to use this

- A store inside the same cluster (an in-cluster `KubernetesSeaweedFs`
  reached by its S3 endpoint) protects against pod loss, not cluster loss.
  It is the right lab shape and the wrong disaster-recovery shape; say so in
  the proposal.
- Dev-mode or file-storage instances (OpenBao's `dev` and `standalone`
  modes) have nothing to snapshot; the kinds refuse a backup block on them.
- A restore is never an edit to a running instance: it is a new instance
  declared with the restore member set. Editing a live instance's restore
  block changes nothing on PostgreSQL and MongoDB, and on OpenBao it only
  suspends the schedule.

## See also

- [KubernetesOpenBao guide](../kubernetes/kubernetesopenbao/GUIDE.md) — the
  resource sets for GKE and R2, the login recipe, the bad-day runbook, and
  the Shamir runbook.
- [KubernetesPostgres guide](../kubernetes/kubernetespostgres/GUIDE.md) and
  [KubernetesMongodb guide](../kubernetes/kubernetesmongodb/GUIDE.md) — the
  same two resource sets for the databases, with each kind's credential
  continuity rule.
- [CloudflareR2Bucket](../cloudflare/cloudflarer2bucket/v1alpha1/reference.md)
  and
  [CloudflareAccountApiToken](../cloudflare/cloudflareaccountapitoken/v1alpha1/reference.md)
  — the outputs every R2 arm references.
- [namespace-ownership](namespace-ownership.md) — a restore target is a new
  instance and usually a new namespace; the sole-tenant case applies.
