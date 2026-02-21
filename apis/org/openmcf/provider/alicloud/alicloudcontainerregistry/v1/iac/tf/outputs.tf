output "instance_id" {
  description = "The ACR Enterprise Edition instance ID"
  value       = alicloud_cr_ee_instance.main.id
}

output "instance_name" {
  description = "The registry instance name"
  value       = alicloud_cr_ee_instance.main.instance_name
}

output "public_endpoint" {
  description = "The internet-facing registry endpoint domain for docker login"
  value       = local.public_endpoint
}

output "vpc_endpoint" {
  description = "The VPC-internal registry endpoint domain for pulling images within VPC"
  value       = local.vpc_endpoint_domain
}

output "namespace_ids" {
  description = "Map of namespace names to their IDs"
  value = {
    for name, ns in alicloud_cr_ee_namespace.namespaces : name => ns.id
  }
}
