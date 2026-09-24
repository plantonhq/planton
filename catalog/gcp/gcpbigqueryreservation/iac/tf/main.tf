# Enable the BigQuery Reservation API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one reservation
# must never disable the API for every other reservation and commitment in
# the admin project.
resource "google_project_service" "bigqueryreservation_api" {
  project = local.project_id
  service = "bigqueryreservation.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The reservation. project, location, name, and edition are immutable;
# capacity, autoscaling, concurrency, idle-slot sharing, the group, the
# secondary location, and labels update in place.
resource "google_bigquery_reservation" "this" {
  project            = local.project_id
  location           = local.location
  name               = local.reservation_name
  slot_capacity      = var.spec.slot_capacity
  edition            = local.edition
  ignore_idle_slots  = local.ignore_idle_slots
  concurrency        = local.concurrency
  reservation_group  = local.reservation_group
  secondary_location = local.secondary_location
  labels             = local.final_labels

  # The spec lifts the block's one input; 0 means no autoscaling, so the
  # block is sent only when a ceiling is declared.
  dynamic "autoscale" {
    for_each = var.spec.autoscale_max_slots > 0 ? [var.spec.autoscale_max_slots] : []
    content {
      max_slots = autoscale.value
    }
  }

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.bigqueryreservation_api]
}

# The folded assignments. Every argument is immutable, so any change to an
# assignment replaces it (its key changes with it). They share the
# reservation's project, location, and destroy stance.
resource "google_bigquery_reservation_assignment" "this" {
  for_each = local.assignments

  project     = local.project_id
  location    = google_bigquery_reservation.this.location
  reservation = google_bigquery_reservation.this.name
  assignee    = each.value.assignee
  job_type    = each.value.job_type
  principal   = each.value.principal != "" ? each.value.principal : null

  deletion_policy = local.deletion_policy
}
