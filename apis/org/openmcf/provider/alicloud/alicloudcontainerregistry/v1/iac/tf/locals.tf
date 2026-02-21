locals {
  namespaces_map = {
    for ns in var.spec.namespaces : ns.name => ns
  }

  internet_endpoint = try(
    [for ep in alicloud_cr_ee_instance.main.instance_endpoints : ep if ep.endpoint_type == "internet"][0],
    null
  )

  vpc_endpoint = try(
    [for ep in alicloud_cr_ee_instance.main.instance_endpoints : ep if ep.endpoint_type == "vpc"][0],
    null
  )

  public_endpoint = try(local.internet_endpoint.domains[0].domain, "")
  vpc_endpoint_domain = try(local.vpc_endpoint.domains[0].domain, "")
}
