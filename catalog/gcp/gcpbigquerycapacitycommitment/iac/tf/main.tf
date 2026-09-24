# Enable the BigQuery Reservation API first so a fresh project works on the
# first deploy. disable_on_destroy is false: a commitment's teardown must
# never disable the API its admin project's reservations run on.
resource "google_project_service" "bigqueryreservation_api" {
  project = local.project_id
  service = "bigqueryreservation.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The commitment -- a purchase. Creating it starts a billed term; Google
# refuses the delete before the term ends. project, location, id, slot
# count, edition, and the single-admin-project guard are immutable; the
# plan (to a longer term) and the renewal plan update in place.
resource "google_bigquery_capacity_commitment" "this" {
  project                              = local.project_id
  location                             = local.location
  capacity_commitment_id               = local.capacity_commitment_id
  slot_count                           = var.spec.slot_count
  plan                                 = var.spec.plan
  renewal_plan                         = local.renewal_plan
  edition                              = local.edition
  enforce_single_admin_project_per_org = local.enforce_single_admin_project_per_org

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.bigqueryreservation_api]
}
