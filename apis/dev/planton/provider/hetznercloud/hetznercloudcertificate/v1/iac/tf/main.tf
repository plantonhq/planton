resource "hcloud_uploaded_certificate" "this" {
  count = local.is_uploaded ? 1 : 0

  name        = local.certificate_name
  certificate = var.spec.uploaded.certificate
  private_key = var.spec.uploaded.private_key
  labels      = local.standard_labels
}

resource "hcloud_managed_certificate" "this" {
  count = local.is_managed ? 1 : 0

  name         = local.certificate_name
  domain_names = var.spec.managed.domain_names
  labels       = local.standard_labels
}
