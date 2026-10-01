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
  description = "GcpGkeFleet specification"
  type = object({
    # The fleet host project: a literal project ID or a GcpProject
    # reference. Empty means the provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The fleet's name as the console and gcloud show it. 4-30 characters of
    # letters, digits, hyphens, single or double quotes, spaces, and
    # exclamation points. Empty lets Google derive one from the project's
    # name.
    display_name = optional(string, "")

    # Defaults applied to every cluster in the fleet, existing and future.
    default_cluster_config = optional(object({
      # Binary Authorization for every cluster in the fleet: which GKE
      # platform policies their workloads are evaluated against.
      binary_authorization_config = optional(object({
        # How workloads are evaluated:
        #   "DISABLED"        -- no Binary Authorization evaluation
        #   "POLICY_BINDINGS" -- workloads are audited against policy_bindings
        # Empty leaves Google's current setting.
        evaluation_mode = optional(string, "")

        # The GKE platform policies to audit against, each the policy's
        # relative name: "projects/{project_number}/platforms/gke/policies/{policy_id}".
        # Platform policies are created through the Binary Authorization API or
        # console (no catalog block creates one), so these are names, not
        # references. Distinct from the project-level GcpBinaryAuthorizationPolicy.
        policy_bindings = optional(list(string), [])
      }))

      # GKE security posture for every cluster in the fleet: workload
      # configuration auditing and vulnerability scanning.
      security_posture_config = optional(object({
        # Workload configuration auditing:
        #   "DISABLED"   -- off
        #   "BASIC"      -- Google's standard configuration checks
        #   "ENTERPRISE" -- advanced posture capabilities; check Google's
        #                   security posture pricing before choosing it
        # Empty leaves Google's current setting.
        mode = optional(string, "")

        # Vulnerability scanning of running workloads:
        #   "VULNERABILITY_DISABLED", "VULNERABILITY_BASIC" (OS vulnerabilities),
        #   "VULNERABILITY_ENTERPRISE" (adds language-package vulnerabilities).
        # Empty leaves Google's current setting.
        vulnerability_mode = optional(string, "")
      }))
    }))

    # What destroy does:
    #   "" / "DELETE" -- the fleet is deleted (Google refuses while
    #                    memberships or scopes remain)
    #   "PREVENT"     -- destroy fails; a guard for a fleet teams depend on
    #   "ABANDON"     -- the fleet leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
