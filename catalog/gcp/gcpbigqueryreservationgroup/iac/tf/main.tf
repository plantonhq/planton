# Enable the BigQuery Reservation API first so a fresh project works on the
# first deploy. disable_on_destroy is false: tearing down one group must
# never disable the API its admin project's reservations run on.
resource "google_project_service" "bigqueryreservation_api" {
  project = local.project_id
  service = "bigqueryreservation.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The group. Every argument is immutable; reservations join it from their
# own reservation_group field.
resource "google_bigquery_reservation_group" "this" {
  project  = local.project_id
  location = local.location
  name     = local.reservation_group_name

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.bigqueryreservation_api]
}
