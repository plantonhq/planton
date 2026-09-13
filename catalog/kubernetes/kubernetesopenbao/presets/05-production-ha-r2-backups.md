# Production HA with Cloudflare R2 Backups

The production-ha shape with its snapshots outside the cloud that runs it:
three Raft servers, and hourly Raft snapshots landing in a Cloudflare R2
bucket declared in R2's own terms — the bucket, its account and
jurisdiction, and a Cloudflare API token, all by reference to the catalog's
Cloudflare kinds. The module does the S3 translation R2 needs (the
jurisdiction's endpoint host, region `auto`, path-style addressing, the
token as an S3 key pair); nothing S3-shaped is typed. R2 is reachable from
any cluster, so this preset has no cloud: it runs on GKE, EKS, AKS, or a
cluster of your own the same way.

The token is the credential: a `CloudflareAccountApiToken` whose one policy
grants `Workers R2 Storage Bucket Item Write` on this bucket (resource
`com.cloudflare.edge.r2.bucket.<account>_<jurisdiction>_<bucket>`) — objects
in this bucket only, nothing at the account level. Rotating the token mints
a new key pair and the vault follows the references on its next apply. A
token without an R2 permission group authenticates and then fails every
upload with AccessDenied.

The seal is yours to choose. On Shamir (this manifest as written) the
backups run exactly the same and the restore is the manual runbook in the
component guide — the humans holding the shares are the key. Add an
`autoUnseal` arm (Cloud KMS on the cloud you run in, or another OpenBao's
transit engine) and the restore becomes a declaration: a fresh vault on the
SAME key with this `backup` block and `restore: {latest: true}`.

Initialization is still yours, once; then the four-command login recipe the
guide prints, so the backup job can authenticate — until it runs, every
backup run fails and its log prints the recipe with the real names.

Change first: the referenced bucket and token resource names, the `prefix`
(one per live vault, forever), then the schedule and retention to your
recovery objectives — and the seal.

See [05-production-ha-r2-backups.yaml](./05-production-ha-r2-backups.yaml)
for the manifest.
