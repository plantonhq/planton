# Terraform module for GcpFirestoreIndex

Provisions a `google_firestore_index` from the validated protobuf spec. Enables `firestore.googleapis.com` automatically.

Covers the full index surface: field roles (order / array-contains / vector / Enterprise search config), `multikey` and `unique` (MongoDB-compatible scope), `skip_wait` for fire-and-forget creation, and the `deletion_policy` destroy guard (PREVENT fails destroys; ABANDON unmanages without deleting).

See the kind's [README](../../README.md) for the full configuration reference.
