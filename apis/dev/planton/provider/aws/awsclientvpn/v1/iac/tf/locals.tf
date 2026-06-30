locals {
  safe_cidr_authorization_rules = try(var.spec.cidr_authorization_rules, [])
  safe_dns_servers              = try(var.spec.dns_servers, [])
  safe_security_groups          = try(var.spec.security_groups, [])
}


