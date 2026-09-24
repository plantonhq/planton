# KubernetesPlantonPlatform Guide

The judgment this guide carries: the platform is zero-config on purpose —
`version` is the ONE decision, and every other field is a refinement of a
platform that already works. Resist the urge to configure upfront; the
two settings that genuinely reward deciding early are the ones sign-in
bakes at first boot.

## Two settings are sticky — decide them before the first sign-in

The identity server bakes the platform's URL into its realm at first
boot. That makes exactly two fields effectively first-boot-sticky:
`ingress.hostname` (when you know the platform will live at a real URL —
through an Ingress controller or a Gateway API Gateway — set it before
anyone signs in) and `gateway.localPort` (two
port-forward platforms on one laptop need distinct ports, chosen before
the first visit). Everything else — storage, replicas, the runner's
cloud identity, the opt-in components — changes cleanly on a running
platform.

## Tell the platform whether the internet reaches its door

`ingress.reachability` is the one fact about the front door the operator
cannot observe from inside the cluster. Everything the platform offers
that needs an inbound path from the internet — keyless cloud connections,
where the cloud fetches the platform's identity documents from the door,
and GitHub webhook delivery — is offered only where the door is public.
`auto` (the default) reads a hostname served over HTTPS as public and
anything else as private, which is right for most installs. Two shapes
need a word from you: an HTTPS address only your network reaches (split
DNS, a corporate CA, an internal load balancer) is `private`, so those
doors stay honestly closed instead of failing at the cloud's first fetch;
and a door whose TLS terminates outside the cluster (an internet-facing
ALB with an ACM certificate, hence no in-cluster `tls` block) is `public`,
or `auto` will read it as private. Changing the declaration is safe on a
running platform; it changes which doors the console offers, never the
platform's address.

## Declare email once, and both senders use it

A platform a team runs day to day needs to reach people: invitations
that land in inboxes, alerts someone reads, a "Forgot password?" that
works on the sign-in page. `email` is that one declaration. The control
plane and the identity server both send through it, from the address you
name, so there is no second place where a relay is configured and no way
for the two to disagree. Without it the platform is fully usable —
invitations are shared as links and the sign-in page offers no password
reset — and the console's Email settings show the exact `email` fragment
and the one `kubectl create secret` command for this install.

Pick one arm. `smtp` reaches every workplace mail system and every
transactional vendor's SMTP endpoint; `resend` uses Resend's API. On a
relay, `security` is a promise the platform keeps: `starttls` requires
the upgrade and fails a relay that will not offer it, `tls` opens TLS from
the first byte, and `none` is plaintext for a credential-free internal
smart host — the platform refuses credentials over it. Sign in one way: a
username and password in a `kubernetes.io/basic-auth` Secret, an OAuth2
app registration (Exchange Online after Microsoft's password retirement),
or no credential. Every credential is a Secret name or a Secret key
reference, never a value; the operator preflights each one and, when a
Secret is missing, says so in words while the platform runs as if no
email were declared. Credentials reach the control plane as mounted
files, so rotating a password is a Secret edit that is live on the next
send. The relay must permit sending as `from.address` — SPF and DKIM for
that domain are the domain owner's job — and the one failure only a real
send can reveal (a From the account may not send as) is what the
console's "Send Me a Test Email" is for. The platform never probes the
relay on a timer; the checks run when someone asks.

## Version is the upgrade lever, and it is never automated

`version` is required with no default, deliberately: a module-owned
default would turn a catalog update into a silent whole-platform upgrade
on the next apply. Day-2 for this kind IS editing `version` — the
operator rolls the platform to the new line. Pin it like you would pin a
database engine version.

## One operator, many platforms — and the two shared facts

Platforms are namespace-isolated (own URL, own identity server, own
databases) and one operator serves them all. Two cluster-level facts are
honest limits, not bugs: Tekton allows ONE cluster-wide build-events
sink, so keep `build.enabled` on for at most one platform per cluster;
and the cluster has one PlantonPlatform CRD schema (the operator's
version), while each platform still pins its own `spec.version`.

## The runner is where cloud credentials DON'T live

`runner.serviceAccountAnnotations` (workload identity — IRSA, GKE
Workload Identity, AKS) is the right answer wherever the cluster
supports it; `runner.cloudCredentialsSecretName` (a Secret YOU own in
the platform's namespace) is the static-keys fallback. Either way the
platform stores nothing — rotation is your annotation or your Secret,
never a platform record.

## Storage: one global dial, honest failures

`storage.size` + `storage.storageClassName` lift EVERY platform volume
at once — built for backends with minimum-size floors (some NAS backends
refuse volumes under hundreds of Gi; one `size: 800Gi` satisfies them
all). The operator preflights that the chosen class can actually
provision and, when a volume sticks, the CR's per-component status names
the exact problem and fix — read `kubectl get plantonplatforms` status
before reading pod logs.

## Back up the platform's own database, and bring it back

Without `database.postgresql.backup`, everything a platform knows —
organizations, environments, connections, projects, pipeline history,
members, the identity realm with its users — lives on one volume in one
namespace, and `kubectl get plantonplatform` says so: `BACKUP` reads
`NotConfigured`. Declaring the backup turns on continuous WAL archiving
into an object store you own plus a base backup on a schedule (the first
one the moment backups are declared — WAL alone restores nothing), with a
retention the store enforces. The column then reads `Deploying`,
`Healthy`, or `Failing` in the plugin's own words, and `status.backup`
carries the archive's server name, the first recoverability point, and the
last successful base backup. A failing backup never takes a working
platform out of Ready.

The declaration speaks the same vocabulary as the catalog's
`KubernetesPostgres` kind, and on Cloudflare R2 it is composed entirely
from other resources, so nothing is typed: a `CloudflareR2Bucket` (the
archive; `jurisdiction` is fixed at creation and decides which host serves
it), a `CloudflareAccountApiToken` scoped to that bucket with `Workers R2
Storage Bucket Item Write` (the credential — R2 has no keyless posture from
any cluster; the token kind exports itself as the S3 key pair
`r2_access_key_id` / `r2_secret_access_key`), and the platform's `r2` arm
referencing the bucket's `account_id` and `jurisdiction` outputs and the
token's key pair. The module materializes the credential as a Secret
(`<platform>-postgres-backup-creds`) BEFORE the platform resource, in the
same apply, and names it to the operator — so the database is born
archiving, and rotating the token is a new key pair the Secret follows on
the next apply. The other arms mirror the operator's postures: S3 keyless
(IRSA or an instance profile) or access keys, GCS keyless (Workload
Identity — the identity needs `roles/storage.objectAdmin` AND
`roles/storage.legacyBucketReader`) or a service-account key, Azure Blob
keyless or a connection string. Keyless postures bind the database pods'
cloud identity through `backup.serviceAccountAnnotations`.

Bringing a platform back is a declaration too. Declare it again with
`database.postgresql.recoverFrom`: the same store, and `serverName` set
to what the source's `status.backup.serverName` said (or the folder name
under the destination path in the bucket's own listing, when the source is
gone). The database is bootstrapped from the archive — every record and
the identity realm, so existing passwords sign in — and the restored
platform archives its own backups under a NEW server name, so it never
writes over the archive it restored from; keep `backup` declared beside
`recoverFrom` and the two share one bucket and one path. Recovery is
honored only when the database is first created: on a running platform
the operator leaves the database alone and the status names the
procedure. Nothing here ever destroys data to honor a declaration. The
operator that honors all of this — the database's backup and restore, and
the vault's seal, keys Secret, and place in the archive below — is chart
`0.17.0` or newer: earlier charts backed up the database alone and left
the vault on its own volume, and before `0.16.3` the backup engine could
not finish installing, so a restore could come back as an empty database
in the archive's place. Pin the operator kind's `chartVersion` at or above
`0.17.0` wherever a backup is declared.

The archive carries the secrets manager too. The bundled vault (OpenBAO)
stores its data in the platform's own PostgreSQL, so every connection
credential, every managed secret, the license signing key, and the OIDC
issuer's signing key ride the same WAL stream as the records and come back
to the same instant — keyless connections keep verifying against the same
key. What the archive cannot carry is the keys that OPEN the vault, so a
backup requires them to outlive the platform, one of two ways.
`vault.autoUnseal` seals the vault with a key in your cloud (AWS KMS, GCP
Cloud KMS, Azure Key Vault, or a central OpenBao's transit engine — the
four arms the standalone `KubernetesOpenBao` kind speaks, byte for byte);
the restored vault opens itself, on a cluster that has never seen it. The
key and its grants must exist BEFORE the platform, because the seal is
checked at server start: a GCP identity needs
`roles/cloudkms.cryptoKeyEncrypterDecrypter` AND `roles/cloudkms.viewer`,
or the pod crash-loops on `Permission 'cloudkms.cryptoKeys.get' denied`
before init can open. Or `vault.initSecretName` names a Secret you own:
the operator writes the vault's unseal keys and root token into it at
first boot without an owner reference and never deletes it, so deleting the
`PlantonPlatform` leaves it standing -- but a namespace this resource owns
(`create_namespace: true`) is deleted with the resource and takes every
Secret in it, so keep a copy outside the cluster (or place the platform in
a namespace you own), and a restore unseals with it once you have recreated
it in the new cluster from that copy. The spec refuses a backup with neither.
Under either seal, that Secret is the vault's break-glass (the root token;
the recovery quorum) — the one object a lost cluster takes with it that no
archive brings back, so keep a copy outside the cluster.
`status.backup.vault` says whether the archive covers the vault, which
seal opens it, and which Secret to keep. The resource sets that make this
whole on each cloud, and the runbook for the bad day, are the three
sections that follow.

## Disaster recovery on GKE: the resource set

"Every record and every secret back by declaration, with no human holding
a key" is, on GKE, the platform plus the seal set from the standalone
vault's guide plus the store set from the disaster-recovery pattern —
every one a kind in this catalog, wired by reference. Two identities on
purpose: the seal key and the archive bucket are different blast radii.

| # | Resource | What it is for | Wiring |
|---|---|---|---|
| 1 | `GcpServiceAccount` (e.g. `planton-vault-unseal`) | The VAULT's identity: wraps and unwraps the master key — KEYLESS | — |
| 2 | `GcpServiceAccount` (e.g. `planton-backup`) | The DATABASE's identity: archives WAL and base backups — KEYLESS | — |
| 3 | `GcpKmsKeyRing` | Holds the unseal key. Permanent by GCP design — never deletable, its name occupied forever | `location` = the region the platform runs in |
| 4 | `GcpKmsKey` | The unseal key. A restored platform opens with the SAME key | `keyRingId` by reference to #3; `deletionPolicy: PREVENT` — destroying the key makes every archive sealed by it unreadable, backups included |
| 5 | `GcpKmsKeyIamMember` (two per key) | Lets the vault use the key AND read it | `cryptoKeyId` by reference to #4, `member` by reference to #1; one `roles/cloudkms.cryptoKeyEncrypterDecrypter` (wrap on init, unwrap on every start) and one `roles/cloudkms.viewer` (the server reads the key's metadata when it configures its seal at START; without it the pod crash-loops on "Error configuring seal") |
| 6 | `GcpGcsBucket` | The archive | `iamMembers`: **two** roles for #2 — `roles/storage.objectAdmin` AND `roles/storage.legacyBucketReader` (Barman reads the bucket's attributes before it writes) |
| 7 | `GcpGkeWorkloadIdentityBinding` (two per platform) | Lets the two Kubernetes ServiceAccounts act as the identities | for the vault: `ksaName` = `<platform>-openbao` (the operator names the vault's ServiceAccount after its release) bound to #1; for the database: `ksaName` = `<platform>-postgres` (CloudNativePG names the instance ServiceAccount after the Cluster) bound to #2; `ksaNamespace` = the platform's namespace. A platform declared into a new cluster needs both bindings there |
| 8 | `KubernetesPlantonPlatform` | The platform, born archiving and sealed by the key | `database.postgresql.backup.objectStore.gcs.keyless: true` with a `gs://` destination and #2's email in `backup.serviceAccountAnnotations`; `vault.autoUnseal.gcpKms` with `keyRing` and `cryptoKey` by reference to #3/#4 and `workloadIdentityServiceAccount` by reference to #1 (the module merges that identity onto the vault's ServiceAccount, so `vault.serviceAccountAnnotations` stays empty); `vault.initSecretName` for the break-glass Secret |

Where each piece lives, validated: rows 2, 6, and the database's half of
row 7 are the
[GKE keyless store set](../../_patterns/stateful-kind-disaster-recovery.md#the-gke-keyless-store-set-complete)
in the disaster-recovery pattern — the same three nodes every keyless GKE
backup stands on; change its `ksaName` to `<platform>-postgres` and its
`ksaNamespace` to the platform's namespace. Rows 1, 3, 4, 5, the vault's
half of row 7, and row 8 are here. One honesty about row 8: the vault's
seal identity travels by reference (the arm's `workloadIdentityServiceAccount`
field), while the database's backup identity is a literal email in
`backup.serviceAccountAnnotations` — a map cannot carry a reference — so
that one value is copied from the identity's email, not wired.

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpServiceAccount
metadata:
  name: planton-vault-unseal
spec:
  serviceAccountId: planton-vault-unseal
  projectId:
    value: my-gcp-project
  displayName: Planton bundled vault auto-unseal via Workload Identity
  description: Keyless identity the platform's vault assumes through GKE Workload Identity to wrap and unwrap its master key with the Cloud KMS key
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: planton-vault-unseal-wi
spec:
  projectId:
    value: my-gcp-project
  serviceAccountEmail:
    valueFrom:
      kind: GcpServiceAccount
      name: planton-vault-unseal
      fieldPath: status.outputs.email
  ksaNamespace: planton
  # The operator names the vault's ServiceAccount after its release,
  # "<platform>-openbao". A platform declared into a new cluster needs this
  # binding there too, before the platform.
  ksaName: planton-openbao
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyRing
metadata:
  name: planton-vault-unseal
spec:
  projectId:
    value: my-gcp-project
  # Permanent by GCP design: the ring can never be deleted and its name is
  # occupied in this project and location forever. Choose it once.
  keyRingName: planton-vault-unseal
  location: asia-south1
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKey
metadata:
  name: planton-vault-unseal
spec:
  keyRingId:
    valueFrom:
      kind: GcpKmsKeyRing
      name: planton-vault-unseal
      fieldPath: status.outputs.key_ring_id
  keyName: planton-vault-unseal
  # Destroying the key makes every archive this vault ever wrote unreadable;
  # PREVENT is the production posture.
  deletionPolicy: PREVENT
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyIamMember
metadata:
  name: planton-vault-unseal-encrypter-decrypter
spec:
  cryptoKeyId:
    valueFrom:
      kind: GcpKmsKey
      name: planton-vault-unseal
      fieldPath: status.outputs.key_id
  # Wrap on init, unwrap on every start — the role the seal USES the key with.
  role:
    value: roles/cloudkms.cryptoKeyEncrypterDecrypter
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: planton-vault-unseal
      fieldPath: status.outputs.member
---
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpKmsKeyIamMember
metadata:
  name: planton-vault-unseal-viewer
spec:
  cryptoKeyId:
    valueFrom:
      kind: GcpKmsKey
      name: planton-vault-unseal
      fieldPath: status.outputs.key_id
  # The second grant: the server READS the key's metadata when it configures
  # its seal at START, and the encrypter-decrypter role does not carry
  # cloudkms.cryptoKeys.get. Without this one the vault pod crash-loops on
  # "Error configuring seal" before it can be initialized, and the
  # platform's status names that check.
  role:
    value: roles/cloudkms.viewer
  member:
    valueFrom:
      kind: GcpServiceAccount
      name: planton-vault-unseal
      fieldPath: status.outputs.member
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  createNamespace: true
  version: v0.0.75
  database:
    postgresql:
      backup:
        objectStore:
          # The bucket from the pattern's store set; the path after it is
          # this platform's. Its archive is filed beneath it under the
          # platform's own server name.
          destinationPath: gs://planton-archive-my-gcp-project/platform
          gcs:
            keyless: true
        # The database pods' identity: the backup GcpServiceAccount's email,
        # copied (a map cannot carry a reference). Its binding names
        # ksaName "<platform>-postgres".
        serviceAccountAnnotations:
          iam.gke.io/gcp-service-account: planton-backup@my-gcp-project.iam.gserviceaccount.com
        retentionPolicy: 30d
        schedule: "0 0 2 * * *"
  vault:
    autoUnseal:
      gcpKms:
        project:
          value: my-gcp-project
        region: asia-south1
        keyRing:
          valueFrom:
            kind: GcpKmsKeyRing
            name: planton-vault-unseal
            fieldPath: status.outputs.key_ring_name
        cryptoKey:
          valueFrom:
            kind: GcpKmsKey
            name: planton-vault-unseal
            fieldPath: status.outputs.key_name
        # The seal identity by reference: the module puts its email on the
        # vault's ServiceAccount as the Workload Identity annotation.
        workloadIdentityServiceAccount:
          valueFrom:
            kind: GcpServiceAccount
            name: planton-vault-unseal
            fieldPath: status.outputs.email
    # Under a cloud seal this Secret holds the recovery keys and the root
    # token — the break-glass, never needed by anything running. Keep a copy
    # outside the cluster; the operator never deletes it.
    initSecretName: planton-vault-keys
```

Apply the set in dependency order — the identities, the ring, the key, the
grants, the bucket, the two bindings — and the platform last: the vault
checks its seal when its server starts, and a key or grant that lands after
the platform is a crash-loop the status explains instead of a working
vault. On a cluster that has never run this platform, the two bindings are
the only cluster-side thing the seal and the archive need; the ring, the
key, the grants, the identities, and the bucket live in the project and
survive any cluster.

## Disaster recovery with Cloudflare R2: the resource set

The cloud-neutral shape: the archive in a Cloudflare R2 bucket, declared
entirely by reference (a `CloudflareR2Bucket` and a `CloudflareAccountApiToken`
scoped to it — R2 has no keyless posture from any cluster, so the token's
S3 key pair is the credential, referenced, never typed), and the vault on
the built-in seal with `vault.initSecretName` naming the Secret you own for
its keys. This is the `06-backups-to-r2` preset, whole (named by slug —
presets ship in the release's `presets.zip` and in the catalog repository,
not in the skill's pack); the R2 store pair is embedded in the
[disaster-recovery pattern](../../_patterns/stateful-kind-disaster-recovery.md#the-composition).
Nothing in this shape is in a cloud the platform runs on, which is the
point of it: the archive and the key that opens it are both yours.

What the Secret holds, and what to do with it: the operator writes
`unseal-keys` (five shares, threshold three) and `root-token` into the
Secret at first boot, annotates it with what each key does and which seal
it was initialized under, and never writes over a Secret that already holds
a vault's keys. Copy it out of the cluster the day the platform is born
(`kubectl get secret planton-vault-keys -n planton -o yaml`, stored where
your other break-glass material lives), because a lost cluster — or a
destroy of a platform that owns its namespace — takes the in-cluster copy
with it, and nothing in the archive brings it back. A cloud key
(`vault.autoUnseal`) removes this step for the vault's opening; the Secret
then holds recovery keys and the root token and is break-glass only.

## Restore on the bad day

The platform, its database, and its vault come back as one declaration:
the platform declared again with `database.postgresql.recoverFrom` naming
the same store and the source's `serverName`. The vault's data is in the
archive; what differs by seal is how it opens. Every step below is what the
operator's own Kind lane runs, and every sentence in quotation marks is
what `kubectl get plantonplatform` shows at that moment.

Two things the restored declaration keeps from the source, whichever seal:
the same `version` (the archive holds that release's schema — restore
first, then upgrade by editing `version`, as its own step), and the same
`ingress.hostname` when one was set (the identity realm that comes back
was baked with it at the source's first boot, and a realm that names a
different address than the door serves is the one restore that signs
nobody in). And the cluster needs cert-manager before the platform: the
backup engine cannot install without it, and the `BACKUP` column says so
while the platform runs.

**Under a cloud seal** (`vault.autoUnseal`):

1. Make sure the key, its two grants, and the vault's Workload Identity
   binding exist for the new cluster — the ring and the key never died,
   but the binding is per cluster.
2. Declare the platform with `recoverFrom` (the same store, the source's
   `serverName`), the same `autoUnseal` arm, and `backup` beside it so the
   restored platform archives itself under a new server name.
3. The database bootstraps from the archive; the vault starts against it,
   finds itself initialized, and opens itself from your key — the operator
   never unseals it. `status.components.openBao` reads Ready; the control
   plane gets a fresh token from the operator; every connection credential,
   managed secret, and signing key is back, and keyless connections keep
   verifying because the signing key is the same one.
4. Recreate the keys Secret from your copy. Until you do, the platform
   works and the vault's Ready sentence says so plainly: the vault came
   back from the archive and your key opened it, but the Secret does not
   exist, so "the vault's break-glass (its root token and recovery keys) is
   gone until the Secret is back". Nothing running needs it; you do, the
   day you need to generate a new root token.

The bad-day declaration for the GKE set above, complete — the day-one
platform plus `recoverFrom`, the same seal, `backup` kept beside it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonPlatform
metadata:
  name: planton
spec:
  namespace:
    value: planton
  createNamespace: true
  # The release the source ran; upgrade afterwards, as its own step.
  version: v0.0.75
  database:
    postgresql:
      # READS the source's archive; honored only when the database is first
      # created. serverName is what the source's status.backup.serverName
      # said, or the folder under the destination path in the bucket's own
      # listing when the source is gone.
      recoverFrom:
        objectStore:
          destinationPath: gs://planton-archive-my-gcp-project/platform
          gcs:
            keyless: true
        serverName: planton-postgres-3f9a1c2e
      # Kept beside recoverFrom: the restored platform archives itself under
      # a NEW server name in the same bucket and path, never over its source.
      backup:
        objectStore:
          destinationPath: gs://planton-archive-my-gcp-project/platform
          gcs:
            keyless: true
        serviceAccountAnnotations:
          iam.gke.io/gcp-service-account: planton-backup@my-gcp-project.iam.gserviceaccount.com
        retentionPolicy: 30d
        schedule: "0 0 2 * * *"
  vault:
    # The SAME seal the source had: the seal is decided when the vault is
    # created, and a different one is refused before anything renders.
    autoUnseal:
      gcpKms:
        project:
          value: my-gcp-project
        region: asia-south1
        keyRing:
          valueFrom:
            kind: GcpKmsKeyRing
            name: planton-vault-unseal
            fieldPath: status.outputs.key_ring_name
        cryptoKey:
          valueFrom:
            kind: GcpKmsKey
            name: planton-vault-unseal
            fieldPath: status.outputs.key_name
        workloadIdentityServiceAccount:
          valueFrom:
            kind: GcpServiceAccount
            name: planton-vault-unseal
            fieldPath: status.outputs.email
    # The same name, recreated from your copy after the restore (step 4).
    initSecretName: planton-vault-keys
```

Under the built-in seal the declaration is the same with the `autoUnseal`
block absent, and the Secret recreated BEFORE it is applied.

**Under the built-in seal** (`vault.initSecretName`, no `autoUnseal`):

1. Recreate the keys Secret in the new cluster's namespace from the copy
   you kept, under the same name, BEFORE you declare the platform. (If the
   declaration creates the namespace, apply the Secret the moment the
   namespace exists — the operator restores the database before it needs
   the Secret, so a Secret that arrives a minute after the declaration is
   in time.)
2. Declare the platform with `recoverFrom`, the same `initSecretName`, and
   `backup` beside it.
3. The database bootstraps from the archive; the vault starts against it,
   finds itself initialized and sealed, and the operator unseals it with
   the shares from your Secret. Everything is back, as above.

If you forget step 1, nothing is lost and nothing is guessed: the vault
component refuses, with `ConfigurationRefused` and the Secret as its
object, in one sentence — the vault "came back from archive
`<serverName>` initialized and sealed, but Secret `<name>` (keys
`unseal-keys`, `root-token`) does not exist in namespace `<ns>`, so the
operator holds no keys to open it. Recreate that Secret from the copy you
kept; or, on a platform declared with `spec.vault.autoUnseal`, the restored
vault opens itself from your cloud key." Recreate it and the next pass
unseals.

**What a restore cannot bring back.** The keys that open the vault are
never in the archive, by design: under the built-in seal that is the
Secret you kept; under a cloud seal it is your key, which you never lost.
The root token and (under a cloud seal) the recovery quorum live only in
the keys Secret, so a restore without that Secret is a working platform
whose break-glass is gone until you recreate it — the operator says so on
every pass rather than refusing a platform that works. And the seal is
decided when the vault is created: a declaration that names a different
seal than the archive was written under is refused before anything renders
("the seal is decided when the vault is created and cannot be changed on a
running platform"), because the server would refuse to start against its
own storage — restore with the seal the source had.

**How to read the state.** `kubectl get plantonplatform` carries the
`BACKUP` column; `status.backup.vault` carries four facts — `covered`
(the archive carries the vault), `seal` (`shamir`, `gcpKms`, `awsKms`,
`azureKeyVault`, or `transit`), `initSecretName` (the Secret to keep), and
a sentence for the posture; `status.backup.restoredFrom` names the archive
a restored platform came from for as long as it lives. Read those before
pod logs: every way the seal, the Secret, or the archive can fail is one
sentence there with the next step in it.

## Destroy and the reinstall truth

Teardown is Kubernetes garbage collection: every operator-created object
is owner-referenced to the declaration, so deleting the platform
completes with or without the operator running, and database credentials
and volumes die together — no orphaned volume can hold a password a
reinstall cannot match. With a backup declared, the archive in the object
store is the one thing that survives the teardown, and `recoverFrom` is
how a reinstall becomes a restore instead of a fresh start. Two residues
to know: build caches and workflow volumes may survive in the namespace,
so a reinstall into the SAME namespace should be preceded by deleting it
(automatic when this resource owned the namespace via
`createNamespace`); and the platform's namespace-qualified token-review
ClusterRole/Binding lingers inert (its subject ServiceAccount died with
the platform) until an operator release adds the janitor.

## On the diagram

The platform draws an explicit `depends_on` edge TO its
KubernetesPlantonOperator — no spec field consumes an operator output
(the coupling is the cluster-global CRD contract), so composed charts
declare the edge in metadata. The `planton-on-kubernetes` infra-chart
carries namespace + operator + platform as one deployable arm; its
multi-platform variant is one operator plus N namespace+platform pairs.
