locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The bare Connect cluster id Google's resource is keyed by. The spec's
  # connect_cluster arrives either as the Connect cluster's full resource
  # path (a GcpManagedKafkaConnectCluster reference resolves to its name
  # output) or as the bare id; the last path segment is the id in both
  # cases -- identical to the Pulumi module's ConnectClusterId.
  connect_cluster = element(split("/", var.spec.connect_cluster), length(split("/", var.spec.connect_cluster)) - 1)

  # The connector's id defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  connector_id = var.spec.connector_id != "" ? var.spec.connector_id : var.metadata.name

  # Google stores the connector id as Kafka Connect's own "name" config, so
  # the module sends it too; leaving it out would re-plan its removal on
  # every apply.
  configs         = merge(var.spec.configs, { name = local.connector_id })
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
