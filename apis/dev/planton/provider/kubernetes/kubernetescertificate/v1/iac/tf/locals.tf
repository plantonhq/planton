locals {
  namespace        = var.spec.namespace
  certificate_name = var.metadata.name
  secret_name      = var.spec.secret_name
  is_ca            = var.spec.is_ca

  labels = {
    "app.kubernetes.io/managed-by" = "planton"
    "resource.planton.dev/id"      = var.metadata.name
  }

  # Resolve the issuer ref oneof: exactly one of cluster_issuer or issuer is set.
  issuer_ref_kind = var.spec.issuer_ref.cluster_issuer != null ? "ClusterIssuer" : "Issuer"
  issuer_ref_name = var.spec.issuer_ref.cluster_issuer != null ? var.spec.issuer_ref.cluster_issuer.name : var.spec.issuer_ref.issuer.name

  # Build the cert-manager spec.privateKey block only when configured.
  # null means "omit from manifest" which lets cert-manager apply CRD defaults.
  private_key = var.spec.private_key != null ? {
    algorithm      = var.spec.private_key.algorithm
    size           = var.spec.private_key.size
    encoding       = var.spec.private_key.encoding
    rotationPolicy = var.spec.private_key.rotation_policy
  } : null

  # Duration config is optional; null means cert-manager defaults apply.
  duration     = var.spec.duration_config != null ? var.spec.duration_config.duration : null
  renew_before = var.spec.duration_config != null ? var.spec.duration_config.renew_before : null
}
