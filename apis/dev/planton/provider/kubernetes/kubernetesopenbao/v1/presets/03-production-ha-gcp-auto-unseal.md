# Production HA with GCP KMS Auto-Unseal

This preset deploys OpenBao in high-availability mode with GCP Cloud KMS auto-unseal and GKE Workload Identity. Pods unseal automatically on every restart without human intervention, eliminating the availability gap caused by manual unsealing.

## When to Use

- Production GKE deployments where zero-downtime recovery is required
- Environments that cannot tolerate manual unseal intervention after pod restarts, rolling updates, or node failures
- Teams already managing GCP KMS keyrings for encryption at rest

## Key Configuration Choices

- **HA mode** with 3 replicas -- uses Raft consensus for leader election and data replication; tolerates 1 node failure
- **GCP Cloud KMS auto-unseal** -- delegates master key protection to a symmetric encrypt/decrypt key in Cloud KMS; OpenBao encrypts its root key with the KMS key instead of Shamir splitting
- **GKE Workload Identity** -- the Kubernetes service account is annotated with `iam.gke.io/gcp-service-account` so the pod authenticates to Cloud KMS without credential files
- **TLS enabled** -- encrypts client-to-server and server-to-server communication
- **UI enabled** with ingress -- web interface for managing secrets accessible at the specified hostname

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project containing the KMS keyring | GCP Console > Project selector |
| `<gcp-region>` | Region of the KMS keyring (e.g., `asia-south1`) | GCP Console > KMS > Key rings |
| `<kms-key-ring-name>` | Name of the KMS keyring | GCP Console > KMS > Key rings |
| `<kms-crypto-key-name>` | Name of the crypto key within the keyring | GCP Console > KMS > Key ring > Keys |
| `<gcp-service-account-email>` | GCP service account with `roles/cloudkms.cryptoKeyEncrypterDecrypter` and a Workload Identity binding | GCP Console > IAM > Service Accounts |
| `<your-openbao.example.com>` | Hostname for the OpenBao UI and API | Your DNS provider |

## Related Presets

- **01-dev-mode** -- Simple single-server deployment for development
- **02-production-ha** -- HA deployment without auto-unseal (manual Shamir unseal)
