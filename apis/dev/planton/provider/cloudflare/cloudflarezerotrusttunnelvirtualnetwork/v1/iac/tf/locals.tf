locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-tunnel-virtual-network")
}
