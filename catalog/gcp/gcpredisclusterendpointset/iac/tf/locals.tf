locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The bare cluster name Google's resource is keyed by. The spec's cluster
  # arrives either as the cluster's full resource path (a GcpRedisCluster
  # reference resolves to its name output,
  # projects/{p}/locations/{r}/clusters/{name}) or as the bare name; the
  # last path segment is the name in both cases -- identical to the Pulumi
  # module's ClusterName.
  cluster_name = element(split("/", var.spec.cluster), length(split("/", var.spec.cluster)) - 1)
  region       = var.spec.region

  connection_count = sum([for e in var.spec.endpoints : length(e.connections)])
}
