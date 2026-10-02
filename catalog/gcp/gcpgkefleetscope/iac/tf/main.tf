# The Fleet API (GKE Hub). disable_on_destroy is false: tearing down one
# team's scope must never disable the API for the rest of the fleet.
resource "google_project_service" "gkehub_api" {
  project = local.project_id
  service = "gkehub.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The team scope. The scope ID is immutable; labels and the namespace
# labels Google applies to every namespace of the scope update in place.
resource "google_gke_hub_scope" "this" {
  project          = local.project_id
  scope_id         = local.scope_id
  labels           = local.final_labels
  namespace_labels = length(var.spec.namespace_labels) > 0 ? var.spec.namespace_labels : null
  deletion_policy  = local.deletion_policy

  depends_on = [google_project_service.gkehub_api]
}

# The scope's fleet namespaces, one per spec.namespaces[] entry. Google
# takes the scope twice: its short ID in the URL and its full name in the
# body.
resource "google_gke_hub_namespace" "this" {
  for_each = local.namespaces

  project            = local.project_id
  scope_id           = google_gke_hub_scope.this.scope_id
  scope              = google_gke_hub_scope.this.name
  scope_namespace_id = each.key
  labels             = merge(each.value.labels, local.attribution_labels)
  namespace_labels   = length(each.value.namespace_labels) > 0 ? each.value.namespace_labels : null
  deletion_policy    = local.deletion_policy
}

# Who gets which access in the scope's namespaces, one per
# spec.rbac_role_bindings[] entry: exactly one of user or group, and
# exactly one of a predefined or a custom role.
resource "google_gke_hub_scope_rbac_role_binding" "this" {
  for_each = local.rbac_role_bindings

  project                    = local.project_id
  scope_id                   = google_gke_hub_scope.this.scope_id
  scope_rbac_role_binding_id = each.key
  user                       = each.value.user != "" ? each.value.user : null
  group                      = each.value.group != "" ? each.value.group : null
  labels                     = merge(each.value.labels, local.attribution_labels)
  deletion_policy            = local.deletion_policy

  role {
    predefined_role = each.value.role.predefined_role != "" ? each.value.role.predefined_role : null
    custom_role     = each.value.role.custom_role != "" ? each.value.role.custom_role : null
  }
}

# The clusters the team may use, one binding per
# spec.membership_bindings[] entry.
resource "google_gke_hub_membership_binding" "this" {
  for_each = local.membership_bindings

  project               = local.project_id
  location              = each.value.location
  membership_id         = each.value.membership_id
  membership_binding_id = each.key
  scope                 = google_gke_hub_scope.this.name
  labels                = merge(each.value.labels, local.attribution_labels)
  deletion_policy       = local.deletion_policy
}
