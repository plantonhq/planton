locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  zone_id     = var.spec.zone_id.value
  record_name = var.spec.record_name

  record_type_names = {
    1  = "A"
    2  = "AAAA"
    3  = "CNAME"
    4  = "MX"
    5  = "TXT"
    6  = "SRV"
    7  = "NS"
    8  = "PTR"
    9  = "CAA"
    10 = "SOA"
    11 = "SPF"
    12 = "SSHFP"
    13 = "NAPTR"
  }

  record_type = lookup(local.record_type_names, var.spec.type, "A")
}
