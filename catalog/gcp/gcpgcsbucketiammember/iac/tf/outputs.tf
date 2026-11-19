# The grant, resolved (after reference resolution) -- echoed so audits see
# exactly what was applied without re-resolving references.
# The provider reads the bucket back as "b/<name>" after a refresh; the
# output is the bucket's name, so it never changes between plans.
output "bucket" {
  description = "The bucket whose IAM policy received the grant"
  value       = trimprefix(google_storage_bucket_iam_member.this.bucket, "b/")
}

output "role" {
  description = "The role that was granted"
  value       = google_storage_bucket_iam_member.this.role
}

output "member" {
  description = "The member the role was granted to, in IAM member format"
  value       = google_storage_bucket_iam_member.this.member
}

output "etag" {
  description = "The etag of the bucket's IAM policy after the grant"
  value       = google_storage_bucket_iam_member.this.etag
}
