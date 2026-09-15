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
anyone signs in) and `gateway.local_port` (two
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

`runner.service_account_annotations` (workload identity — IRSA, GKE
Workload Identity, AKS) is the right answer wherever the cluster
supports it; `runner.cloud_credentials_secret_name` (a Secret YOU own in
the platform's namespace) is the static-keys fallback. Either way the
platform stores nothing — rotation is your annotation or your Secret,
never a platform record.

## Storage: one global dial, honest failures

`storage.size` + `storage.storage_class_name` lift EVERY platform volume
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
cloud identity through `backup.service_account_annotations`.

Bringing a platform back is a declaration too. Declare it again with
`database.postgresql.recover_from`: the same store, and `server_name` set
to what the source's `status.backup.serverName` said (or the folder name
under the destination path in the bucket's own listing, when the source is
gone). The database is bootstrapped from the archive — every record and
the identity realm, so existing passwords sign in — and the restored
platform archives its own backups under a NEW server name, so it never
writes over the archive it restored from; keep `backup` declared beside
`recover_from` and the two share one bucket and one path. Recovery is
honored only when the database is first created: on a running platform
the operator leaves the database alone and the status names the
procedure. Nothing here ever destroys data to honor a declaration. The
operator that honors all of this is chart `0.16.3` or newer: on earlier
charts the backup engine could not finish installing and a restore could
come back as an empty database in the archive's place, so pin the operator
kind's `chart_version` at or above it wherever a backup is declared.

The archive carries the secrets manager too. The bundled vault (OpenBAO)
stores its data in the platform's own PostgreSQL, so every connection
credential, every managed secret, the license signing key, and the OIDC
issuer's signing key ride the same WAL stream as the records and come back
to the same instant — keyless connections keep verifying against the same
key. What the archive cannot carry is the keys that OPEN the vault, so a
backup requires them to outlive the platform, one of two ways.
`vault.auto_unseal` seals the vault with a key in your cloud (AWS KMS, GCP
Cloud KMS, Azure Key Vault, or a central OpenBao's transit engine — the
four arms the standalone `KubernetesOpenBao` kind speaks, byte for byte);
the restored vault opens itself, on a cluster that has never seen it. The
key and its grants must exist BEFORE the platform, because the seal is
checked at server start: a GCP identity needs
`roles/cloudkms.cryptoKeyEncrypterDecrypter` AND `roles/cloudkms.viewer`,
or the pod crash-loops on `Permission 'cloudkms.cryptoKeys.get' denied`
before init can open. Or `vault.init_secret_name` names a Secret you own:
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
seal opens it, and which Secret to keep.

## Destroy and the reinstall truth

Teardown is Kubernetes garbage collection: every operator-created object
is owner-referenced to the declaration, so deleting the platform
completes with or without the operator running, and database credentials
and volumes die together — no orphaned volume can hold a password a
reinstall cannot match. With a backup declared, the archive in the object
store is the one thing that survives the teardown, and `recover_from` is
how a reinstall becomes a restore instead of a fresh start. Two residues
to know: build caches and workflow volumes may survive in the namespace,
so a reinstall into the SAME namespace should be preceded by deleting it
(automatic when this resource owned the namespace via
`create_namespace`); and the platform's namespace-qualified token-review
ClusterRole/Binding lingers inert (its subject ServiceAccount died with
the platform) until an operator release adds the janitor.

## On the diagram

The platform draws an explicit `depends_on` edge TO its
KubernetesPlantonOperator — no spec field consumes an operator output
(the coupling is the cluster-global CRD contract), so composed charts
declare the edge in metadata. The `planton-on-kubernetes` infra-chart
carries namespace + operator + platform as one deployable arm; its
multi-platform variant is one operator plus N namespace+platform pairs.
