locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The bare cluster id Google's resource is keyed by. The spec's cluster
  # arrives either as the cluster's full resource path (a
  # GcpManagedKafkaCluster reference resolves to its name output) or as the
  # bare id; the last path segment is the id in both cases -- identical to
  # the Pulumi module's ClusterId.
  cluster = element(split("/", var.spec.cluster), length(split("/", var.spec.cluster)) - 1)

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
