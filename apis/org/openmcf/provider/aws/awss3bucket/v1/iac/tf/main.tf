# Create S3 bucket
resource "aws_s3_bucket" "this" {
  bucket        = local.bucket_name
  force_destroy = try(var.spec.force_destroy, false)
  tags          = local.merged_tags
}

# Configure versioning
resource "aws_s3_bucket_versioning" "this" {
  count  = try(var.spec.versioning_enabled, false) ? 1 : 0
  bucket = aws_s3_bucket.this.id

  versioning_configuration {
    status = "Enabled"
  }
}

# Configure encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = try(var.spec.encryption_type, "SSE_S3") == "SSE_KMS" ? "aws:kms" : "AES256"
      kms_master_key_id = try(var.spec.encryption_type, "SSE_S3") == "SSE_KMS" ? try(var.spec.kms_key_id.value, null) : null
    }
    bucket_key_enabled = true
  }
}

# Configure public access block (security best practice)
resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = !local.is_public
  block_public_policy     = !local.is_public
  ignore_public_acls      = !local.is_public
  restrict_public_buckets = !local.is_public
}

# Grant public read access via bucket policy when is_public is true.
# With BucketOwnerEnforced ownership (ACLs disabled), a bucket policy is
# the only mechanism for granting public access to objects.
resource "aws_s3_bucket_policy" "public_read" {
  count  = local.is_public ? 1 : 0
  bucket = aws_s3_bucket.this.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicReadGetObject"
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.this.arn}/*"
    }]
  })

  depends_on = [aws_s3_bucket_public_access_block.this]
}

# Configure ownership controls (disable ACLs - bucket owner enforced)
resource "aws_s3_bucket_ownership_controls" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }

  depends_on = [aws_s3_bucket_public_access_block.this]
}

# Configure lifecycle rules
resource "aws_s3_bucket_lifecycle_configuration" "this" {
  count  = length(try(var.spec.lifecycle_rules, [])) > 0 ? 1 : 0
  bucket = aws_s3_bucket.this.id

  dynamic "rule" {
    for_each = try(var.spec.lifecycle_rules, [])
    content {
      id     = rule.value.id
      status = try(rule.value.enabled, false) ? "Enabled" : "Disabled"

      # Filter by prefix if specified
      dynamic "filter" {
        for_each = try(rule.value.prefix, "") != "" ? [1] : []
        content {
          prefix = rule.value.prefix
        }
      }

      # Add transition if specified
      dynamic "transition" {
        for_each = try(rule.value.transition_days, 0) > 0 && try(rule.value.transition_storage_class, "") != "" ? [1] : []
        content {
          days          = rule.value.transition_days
          storage_class = lookup(local.storage_class_map, rule.value.transition_storage_class, "STANDARD")
        }
      }

      # Add expiration if specified
      dynamic "expiration" {
        for_each = try(rule.value.expiration_days, 0) > 0 ? [1] : []
        content {
          days = rule.value.expiration_days
        }
      }

      # Add noncurrent version expiration if specified
      dynamic "noncurrent_version_expiration" {
        for_each = try(rule.value.noncurrent_version_expiration_days, 0) > 0 ? [1] : []
        content {
          noncurrent_days = rule.value.noncurrent_version_expiration_days
        }
      }

      # Add abort incomplete multipart upload if specified
      dynamic "abort_incomplete_multipart_upload" {
        for_each = try(rule.value.abort_incomplete_multipart_upload_days, 0) > 0 ? [1] : []
        content {
          days_after_initiation = rule.value.abort_incomplete_multipart_upload_days
        }
      }
    }
  }

  depends_on = [aws_s3_bucket_versioning.this]
}

# Configure replication
resource "aws_s3_bucket_replication_configuration" "this" {
  count  = try(var.spec.replication.enabled, false) ? 1 : 0
  bucket = aws_s3_bucket.this.id
  role   = var.spec.replication.role_arn.value

  rule {
    id       = "replication-rule"
    status   = "Enabled"
    priority = try(var.spec.replication.priority, 0)

    # Filter by prefix if specified
    dynamic "filter" {
      for_each = try(var.spec.replication.prefix, "") != "" ? [1] : []
      content {
        prefix = var.spec.replication.prefix
      }
    }

    destination {
      bucket        = var.spec.replication.destination.bucket_arn
      storage_class = try(var.spec.replication.destination.storage_class, "") != "" ? lookup(local.storage_class_map, var.spec.replication.destination.storage_class, "STANDARD") : null

      # Cross-account replication
      dynamic "access_control_translation" {
        for_each = try(var.spec.replication.destination.account_id, "") != "" ? [1] : []
        content {
          owner = "Destination"
        }
      }

      account = try(var.spec.replication.destination.account_id, "")
    }
  }

  depends_on = [aws_s3_bucket_versioning.this]
}

# Configure logging
resource "aws_s3_bucket_logging" "this" {
  count  = try(var.spec.logging.enabled, false) ? 1 : 0
  bucket = aws_s3_bucket.this.id

  target_bucket = var.spec.logging.target_bucket
  target_prefix = try(var.spec.logging.target_prefix, "")
}

# Configure CORS
resource "aws_s3_bucket_cors_configuration" "this" {
  count  = try(var.spec.cors, null) != null && length(try(var.spec.cors.cors_rules, [])) > 0 ? 1 : 0
  bucket = aws_s3_bucket.this.id

  dynamic "cors_rule" {
    for_each = try(var.spec.cors.cors_rules, [])
    content {
      allowed_methods = cors_rule.value.allowed_methods
      allowed_origins = cors_rule.value.allowed_origins
      allowed_headers = try(cors_rule.value.allowed_headers, [])
      expose_headers  = try(cors_rule.value.expose_headers, [])
      max_age_seconds = try(cors_rule.value.max_age_seconds, 0) > 0 ? cors_rule.value.max_age_seconds : null
    }
  }
}
