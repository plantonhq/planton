##############################
# Route53 Hosted Zone
##############################
resource "aws_route53_zone" "r53_zone" {
  # var.metadata.name is expected to be something like "example.com"
  name = var.metadata.name

  # Use metadata.labels for tags, fallback to {} if null
  tags = merge(
      var.metadata.labels != null ? var.metadata.labels : {},
    {
      "Name" = var.metadata.name
    }
  )
}

##############################
# Route53 Records
##############################
resource "aws_route53_record" "records" {
  for_each = local.normalized_records

  zone_id = aws_route53_zone.r53_zone.zone_id
  name    = each.value.name
  type    = each.value.record_type
  ttl     = each.value.ttl_seconds == 0 ? 300 : each.value.ttl_seconds
  records = each.value.values
}
