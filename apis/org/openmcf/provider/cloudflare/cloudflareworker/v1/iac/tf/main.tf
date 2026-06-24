# Fetch worker script bundle from R2
# Uses AWS S3 provider configured for R2
data "aws_s3_object" "worker_bundle" {
  provider = aws.r2
  bucket   = local.r2_bucket
  key      = local.r2_path
}

# Cloudflare Worker Script
resource "cloudflare_workers_script" "main" {
  account_id  = var.spec.account_id
  script_name = local.script_name

  # Worker content from R2 bundle
  content = data.aws_s3_object.worker_bundle.body

  # Module-format worker (the entrypoint is an ES module).
  main_module = "index.js"

  # Compatibility settings
  compatibility_date  = local.compatibility_date
  compatibility_flags = ["nodejs_compat"]

  # Flat bindings list. Each element carries the full attribute set; unused
  # attributes are null so all elements share one object type.
  bindings = concat(
    [for k, v in local.env_variables : {
      name         = k
      type         = "plain_text"
      text         = v
      namespace_id = null
    }],
    [for b in local.kv_bindings : {
      name         = b.name
      type         = "kv_namespace"
      text         = null
      namespace_id = b.field_path
    }],
    [for k, v in local.env_secrets : {
      name         = k
      type         = "secret_text"
      text         = v
      namespace_id = null
    }],
  )

  # Enable Workers Logs for observability.
  observability = {
    enabled            = true
    head_sampling_rate = 1
  }
}

# DNS Record for custom domain (if DNS is enabled)
resource "cloudflare_dns_record" "worker_dns" {
  count = local.dns_enabled ? 1 : 0

  zone_id = local.dns_zone_id
  name    = local.dns_hostname
  type    = "AAAA"
  content = "100::" # Dummy IPv6 address; the proxied Worker handles requests at the edge
  ttl     = 1
  proxied = true # Orange cloud - required for Workers
}

# Worker Route (if DNS is enabled)
resource "cloudflare_workers_route" "main" {
  count = local.dns_enabled ? 1 : 0

  zone_id = local.dns_zone_id
  pattern = local.route_pattern
  script  = cloudflare_workers_script.main.script_name

  # Ensure DNS record exists before creating route
  depends_on = [cloudflare_dns_record.worker_dns]
}

