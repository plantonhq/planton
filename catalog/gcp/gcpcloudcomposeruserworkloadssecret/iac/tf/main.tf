# The user workloads Secret — a Kubernetes Secret Composer manages in
# the environment's GKE cluster. Airflow DAGs (KubernetesPodOperator
# tasks, connections) consume it by name; the material never has to be
# baked into DAG code.
#
# The data updates in place; name, environment, region, and project are
# immutable. The environment must already exist — it is a first-class
# resource this one composes against by reference.
#
# No API enablement here: the Composer API is enabled by the environment
# this Secret is delivered into (a Secret cannot exist without one).
resource "google_composer_user_workloads_secret" "secret" {
  name        = local.secret_name
  environment = local.environment
  region      = local.region
  project     = local.project_id

  # Values are base64-encoded secret material (the Kubernetes Secret
  # contract). The provider marks the attribute sensitive — plans redact
  # it — and it is never surfaced in outputs.
  data = var.spec.data

  # Client-side destroy behavior: DELETE (default), PREVENT (destroy
  # fails — protects credentials live pipelines depend on), or ABANDON
  # (drop from state, keep the Secret in the cluster).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
