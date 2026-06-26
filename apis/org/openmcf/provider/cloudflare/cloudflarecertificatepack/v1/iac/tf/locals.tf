locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-certificate-pack")

  # Zone (StringValueOrRef flattened to a plain string by the converter).
  zone_id = try(var.spec.zone_id, "")

  # type default ("advanced") coalesced here to match the Pulumi module and the
  # control-plane middleware.
  cert_type = try(var.spec.type, "") != "" ? var.spec.type : "advanced"
}
