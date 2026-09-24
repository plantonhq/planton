locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The bare cluster id Google's resource is keyed by. The spec's cluster
  # arrives either as the cluster's full resource path (a
  # GcpManagedKafkaCluster reference resolves to its name output,
  # projects/{p}/locations/{l}/clusters/{id}) or as the bare id; the last
  # path segment is the id in both cases -- identical to the Pulumi
  # module's ClusterId.
  cluster = element(split("/", var.spec.cluster), length(split("/", var.spec.cluster)) - 1)

  # The topic's name defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  topic_id = var.spec.topic_id != "" ? var.spec.topic_id : var.metadata.name

  # Unset optionals become null so the provider omits them and Google's
  # defaults apply.
  partition_count = var.spec.partition_count > 0 ? var.spec.partition_count : null
  configs         = length(var.spec.configs) > 0 ? var.spec.configs : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
