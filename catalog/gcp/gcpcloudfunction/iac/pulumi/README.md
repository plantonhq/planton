# Pulumi module for GcpCloudFunction

Provisions a `cloudfunctionsv2.Function` (plus the public-invoker `cloudrunv2.ServiceIamMember` when `allowUnauthenticated` is set) from the validated protobuf spec. Enables the Cloud Functions, Cloud Build, Cloud Run, Artifact Registry, and Eventarc APIs automatically.

A `secretEnvironmentVariables` entry with a `value` gets its own Secret Manager secret (user-managed replication in the function's region), a pinned version holding the value, and a `secretmanager.secretAccessor` grant on that secret alone for the runtime identity (the default compute account when `serviceAccountEmail` is empty). The function reads the stored version natively and depends on the grant. The Secret Manager API is enabled the same way as the others.

See the kind's [README](../../README.md) for the full configuration reference.
