variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpGkeWorkloadIdentityBinding specification"
  type = object({
    # The GCP project that hosts the GKE cluster — and therefore the implicit
    # workload-identity pool <project>.svc.id.goog the principal lives in.
    # This may differ from the GSA's own project in cross-project setups; the
    # GSA's project is derived from its email.
    # If omitted, the provider's default project is used — the common case
    # when the cluster lives in the credentials' project.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The email of the Google Service Account the KSA impersonates. The grant
    # is attached to THIS service account's IAM policy.
    # Example: "cert-manager@my-project.iam.gserviceaccount.com"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account_email = string

    # Kubernetes namespace of the ServiceAccount running in the cluster.
    # Validated as an RFC 1123 label — a typo here would otherwise produce a
    # syntactically valid but permanently broken principal.
    ksa_namespace = string

    # Name of the Kubernetes ServiceAccount. Validated as an RFC 1123 DNS
    # subdomain (Kubernetes object-name rules).
    ksa_name = string

    # Optional IAM Condition restricting when this grant applies. The
    # condition is part of the grant's identity: the same grant with and
    # without a condition are two independent grants that do not interfere.
    condition = optional(object({
      # Short human-readable title identifying the condition's intent,
      # e.g. "expires-2026-12-31".
      title = string

      # The CEL condition expression, e.g.
      # request.time < timestamp("2027-01-01T00:00:00Z").
      expression = string

      # Optional longer explanation of what the condition does and why it exists.
      description = optional(string, "")
    }))
  })
}
