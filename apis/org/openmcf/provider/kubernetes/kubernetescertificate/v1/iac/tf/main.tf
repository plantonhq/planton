# Create the cert-manager Certificate custom resource
resource "kubernetes_manifest" "certificate" {
  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "Certificate"

    metadata = {
      name      = local.certificate_name
      namespace = local.namespace
      labels    = local.labels
    }

    spec = merge(
      {
        dnsNames   = var.spec.dns_names
        secretName = local.secret_name
        isCA       = local.is_ca

        issuerRef = {
          kind = local.issuer_ref_kind
          name = local.issuer_ref_name
        }
      },
      local.duration != null ? { duration = local.duration } : {},
      local.renew_before != null ? { renewBefore = local.renew_before } : {},
      local.private_key != null ? { privateKey = local.private_key } : {},
    )
  }
}
