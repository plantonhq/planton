# Semantic outputs mirroring GcpComputeImageStackOutputs -- names and
# shapes identical to the Pulumi module's exports.

output "name" {
  description = "Name of the image in GCP"
  value       = google_compute_image.this.name
}

output "self_link" {
  description = "Self-link URL of the image -- what disks, instances, and other images boot or copy from"
  value       = google_compute_image.this.self_link
}

output "family" {
  description = "The image's family, or empty when it has none"
  value       = google_compute_image.this.family == null ? "" : google_compute_image.this.family
}

output "disk_size_gb" {
  description = "The image's size in GB"
  value       = google_compute_image.this.disk_size_gb
}

output "image_id" {
  description = "The image's resource ID, projects/{project}/global/images/{name} -- the relative form GKE secondary boot disks consume"
  value       = google_compute_image.this.id
}
