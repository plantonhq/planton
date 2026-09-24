# The gcp-gke batch bootstrap adopts the permanent KMS key ring a previous batch left behind

**Date**: September 16, 2026
**Type**: Fix (real-cluster E2E batch)
**Components**: `catalog/kubernetes/aa_e2e/realcluster/gcp-gke/bootstrap.sh`, `catalog/kubernetes/aa_e2e/realcluster/README.md`

## Summary

The second bootstrap of the `gcp-gke` batch in the same project failed on its fourteenth node: `Error creating KeyRing ... already exists` (409). The batch's own record says why the ring is there -- a Cloud KMS key ring is permanent by GCP design, so `teardown.sh`'s destroy of its node only abandons it: the ring stays in the project and the node's state forgets it. The consequence was written nowhere: every bootstrap after the first in a project meets the ring already present and the node's create refuses. The first batch (2026-09-14) never saw it; the second (today) did.

`bootstrap.sh` now adopts the ring before the set applies. When `gcloud kms keyrings describe` finds the ring and the ring node's workspace state does not hold a `google_kms_key_ring`, it runs the same `tofu import` an operator would, in the set lane's own workspace on the var file the set lane wrote, so the node converges as a no-op and the key, its grants, and the rest follow. The node's workspace exists only once the set lane has run on that machine, so on a fresh machine the first apply fails on the ring, the adoption runs against the workspace that apply created, and the apply is re-run -- completed nodes re-apply as no-ops, the set lane's own contract. The README's `bootstrap.sh` row says so.

## Verification

`bash -n`; the failed second bootstrap re-run after the adoption converged all 17 nodes (`GcpKmsKeyRing ... deployed` as a no-op, the batch's key `planton-e2e-gke-openbao-unseal-20260916101522` and both IAM grants created) and wrote `env.sh`; the two GKE lanes then ran on both engines against the batch (`gke-gcs-backup-restore`, `gke-r2-backup-restore`) with the Terraform GCS lane's blind import round-trip proposing no real change.
