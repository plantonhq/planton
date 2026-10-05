# The GO Feature Flag relay proxy from the official chart, as a real Helm
# release (Pulumi twin: helm_release.go). The typed spec renders into chart
# values (locals.typed_helm_values); the helm_values escape hatch is a
# SECOND values document, merged by the provider with Helm -f semantics;
# fullnameOverride is re-pinned LAST - the Service, ServiceAccount and every
# output derive from the fullname.

resource "helm_release" "relay_proxy" {
  name       = local.release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.chart_version
  namespace  = local.namespace

  # The module owns namespace creation (create_namespace flag).
  create_namespace = false

  wait            = true
  atomic          = true
  cleanup_on_fail = true
  timeout         = 600

  values = concat(
    [yamlencode(local.typed_helm_values)],
    try(var.spec.helm_values, "") != "" ? [var.spec.helm_values] : [],
    [yamlencode({ fullnameOverride = local.release_name })]
  )

  depends_on = [
    kubernetes_namespace_v1.relay,
    kubernetes_secret_v1.env,
    kubernetes_role_binding_v1.flag_reader,
  ]

  lifecycle {
    # NAME BUDGET (Pulumi twin: the length check in main.go Resources): the
    # relay Service is named after the resource, and a Service name is a
    # 63-character DNS label (the chart truncates the fullname there).
    precondition {
      condition     = length(var.metadata.name) <= 63
      error_message = "metadata.name exceeds the relay's 63-character name budget (the relay Service is named after the resource, and a Service name is a 63-character DNS label)."
    }
    # The relay reads key lists from comma-separated environment variables.
    precondition {
      condition     = alltrue([for k in local.all_keys : !strcontains(k, ",")])
      error_message = "An API key contains a comma; the relay reads key lists comma-separated, so a key may not contain one."
    }
    # Every API key selects exactly one flag set.
    precondition {
      condition     = length(local.flag_set_keys) == length(distinct(local.flag_set_keys))
      error_message = "Two flag sets share an API key; every key must select exactly one flag set."
    }
    # Sensitive headers ride environment variables whose names the relay
    # splits on "_".
    precondition {
      condition     = alltrue([for h in local.sensitive_header_names : !strcontains(h, "_")])
      error_message = "A sensitive header name contains \"_\"; the relay reads it from an environment variable whose name it splits on \"_\"."
    }
    precondition {
      condition     = length(local.header_conflicts) == 0
      error_message = "A header is declared both plain and sensitive; declare it once."
    }
    precondition {
      condition     = length(setintersection(toset(keys(local.extra_env_values)), toset(keys(local.secret_env)))) == 0
      error_message = "An extra environment variable collides with a variable the module generates for a secret value."
    }
  }
}
