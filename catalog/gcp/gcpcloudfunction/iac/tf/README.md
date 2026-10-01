# Terraform module for GcpCloudFunction

Provisions a `google_cloudfunctions2_function` (plus the public-invoker `google_cloud_run_service_iam_member` when `allowUnauthenticated` is set) from the validated protobuf spec. Enables the Cloud Functions, Cloud Build, Cloud Run, Artifact Registry, and Eventarc APIs automatically.

A `secretEnvironmentVariables` entry with a `value` gets its own `google_secret_manager_secret` (user-managed replication in the function's region), a pinned `google_secret_manager_secret_version` holding the value, and a `google_secret_manager_secret_iam_member` granting `secretmanager.secretAccessor` on that secret alone to the runtime identity (the default compute account when `serviceAccountEmail` is empty). The function reads the stored version natively and depends on the grant. The resources live in `secrets.tf`.

See the component [README](../../README.md) for the full configuration reference.
