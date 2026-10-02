# The grant, resolved (after reference resolution) -- echoed so audits see
# exactly what was applied without re-resolving references.
output "bucket" {
  description = "The bucket whose IAM policy received the grant"
  value       = google_storage_bucket_iam_member.this.bucket
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
