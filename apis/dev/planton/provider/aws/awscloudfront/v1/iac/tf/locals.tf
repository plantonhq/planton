locals {
  safe_aliases = try(var.spec.aliases, [])
}


