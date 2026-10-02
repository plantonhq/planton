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
  description = "GcpGkeFleetScope specification"
  type = object({
    # The fleet host project the scope lives in: a literal project ID or a
    # GcpGkeFleet reference (the fleet the scope belongs to, which Google
    # requires first and a chart then orders before the scope). Empty means
    # the provider's default project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The scope's ID, unique in the fleet; usually the team's name. 1-63
    # lowercase letters, digits, and hyphens, starting and ending with a
    # letter or digit. Defaults to metadata.name. Immutable.
    scope_id = optional(string, "")

    # Labels on the scope resource itself. The platform attribution labels
    # are added on top and win on key conflicts.
    labels = optional(map(string), {})

    # Kubernetes labels Google applies to every namespace of this scope on
    # every bound cluster. On a key collision with a namespace's own
    # namespace_labels, this scope-level value wins.
    namespace_labels = optional(map(string), {})

    # The scope's fleet namespaces. Each is created as a Kubernetes namespace
    # on every cluster bound to the scope (onboarding a namespace that
    # already exists there).
    namespaces = optional(list(object({
      # The namespace's name, which is also the Kubernetes namespace created
      # on every bound cluster: a DNS label (1-63 lowercase letters, digits,
      # and hyphens). Google reserves the system namespaces (default,
      # kube-system, gke-connect, istio-system, config-management-system, and
      # the others the spec refuses). Immutable.
      scope_namespace_id = string

      # Labels on the fleet namespace resource. The platform attribution
      # labels are added on top and win on key conflicts.
      labels = optional(map(string), {})

      # Kubernetes labels Google applies to this namespace on every bound
      # cluster; the scope's namespace_labels win on a key collision.
      namespace_labels = optional(map(string), {})
    })), [])

    # Who gets which Kubernetes access within the scope's namespaces on the
    # bound clusters.
    rbac_role_bindings = optional(list(object({
      # The binding's ID, unique in the scope. 1-63 lowercase letters, digits,
      # and hyphens. Immutable.
      scope_rbac_role_binding_id = string

      # A user, as the clusters see it: a person's email ("alice@example.com")
      # or a service account's email -- a GcpServiceAccount reference, for the
      # CI or workload identity that deploys into the team's namespaces.
      # Exactly one of user or group.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      user = optional(string, "")

      # A Google group's email, the usual way to give a whole team access:
      # a literal ("team-a@example.com") or a GcpCloudIdentityGroup reference.
      # Exactly one of user or group.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      group = optional(string, "")

      # The access granted. Required.
      role = object({
        # One of Google's predefined roles, applied in each of the scope's
        # namespaces:
        #   "ADMIN" -- full control of the namespace, including its RBAC
        #   "EDIT"  -- read and write most objects; no RBAC changes
        #   "VIEW"  -- read-only
        predefined_role = optional(string, "")

        # The name of a Kubernetes ClusterRole on the bound clusters. Google
        # honors it only when the fleet's rbacrolebindingactuation feature
        # (GcpGkeFleetFeature with feature "rbacrolebindingactuation") lists it
        # in allowed_custom_roles.
        custom_role = optional(string, "")
      })

      # Labels on the role binding resource. The platform attribution labels
      # are added on top and win on key conflicts.
      labels = optional(map(string), {})
    })), [])

    # The clusters the team may use, each by its fleet membership.
    membership_bindings = optional(list(object({
      # The binding's ID, unique in the scope. 1-63 lowercase letters, digits,
      # and hyphens. Immutable.
      membership_binding_id = string

      # The cluster's fleet membership, by full name
      # ("projects/{p}/locations/{l}/memberships/{id}"): a GcpGkeFleetMembership
      # reference, a GcpGkeCluster reference (its fleet_membership output, for
      # a cluster that joined through fleet_project), or the literal name. The
      # membership must be in this scope's fleet. Both modules derive the
      # binding's project, location, and membership ID from it. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      membership = string

      # Labels on the binding resource. The platform attribution labels are
      # added on top and win on key conflicts.
      labels = optional(map(string), {})
    })), [])

    # What destroy does, for the scope and everything folded into it:
    #   "" / "DELETE" -- bindings, namespaces, and the scope are deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
