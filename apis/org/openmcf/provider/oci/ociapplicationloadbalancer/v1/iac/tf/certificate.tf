resource "oci_load_balancer_certificate" "this" {
  for_each = local.certificates_map

  load_balancer_id  = oci_load_balancer_load_balancer.this.id
  certificate_name  = each.value.certificate_name

  ca_certificate     = each.value.ca_certificate != "" ? each.value.ca_certificate : null
  public_certificate = each.value.public_certificate != "" ? each.value.public_certificate : null
  private_key        = each.value.private_key != "" ? each.value.private_key : null
  passphrase         = each.value.passphrase != "" ? each.value.passphrase : null
}
