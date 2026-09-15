# Production HA + GCP Cloud KMS auto-unseal preset

The production-ha shape with the restart toil removed: the master key
is wrapped by a Cloud KMS crypto key, so every server unseals ITSELF
at startup — pod restarts, node replacements and scale events need no
human with key shares. Credentials follow the keyless-first doctrine:
the server ServiceAccount is annotated for GKE Workload Identity and
the GCP service account needs two roles on the crypto key —
`roles/cloudkms.cryptoKeyEncrypterDecrypter` to wrap and unwrap, and
`roles/cloudkms.viewer` because the server reads the key's metadata when
it configures the seal at start (with only the first role the pod
crash-loops on "Error configuring seal" before init can open); no
static credential exists anywhere. (On EKS or AKS, switch the seal arm
to `awsKms` / `azureKeyVault` and the annotation to the matching
workload-identity seam.)

Initialization is STILL a one-time manual step — with auto-unseal it
produces RECOVERY keys instead of unseal keys (custody them the same
way; they authorize recovery operations, not unsealing). After that,
the seal lifecycle is invisible in operation.

Version horizon, known and accepted: at the pinned OpenBao 2.6.x the
cloud KMS seals are built in but deprecated upstream — v2.7 moves them
to external KMS plugins, and the module's rendering will follow when
the chart pin crosses that line.

Change first: the GCP project/key-ring/crypto-key references and the
service-account email (a `GcpKmsKey` resource reference composes
naturally), then everything the production-ha preset says about
replicas, storage and backups.

See [03-production-ha-gcp-auto-unseal.yaml](./03-production-ha-gcp-auto-unseal.yaml)
for the manifest.
