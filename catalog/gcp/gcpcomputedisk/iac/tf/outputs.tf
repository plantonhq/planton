# Semantic outputs mirroring GcpComputeDiskOutputs — names and
# shapes byte-identical to the Pulumi module's exports.

output "name" {
  description = "Name of the disk in GCP"
  value       = google_compute_disk.this.name
}

output "disk_id" {
  description = "Server-assigned unique numeric identifier of the disk"
  value       = google_compute_disk.this.disk_id
}

output "self_link" {
  description = "Self-link URL of the disk — the attachment composition key"
  value       = google_compute_disk.this.self_link
}

output "zone" {
  description = "Zone the disk lives in (plain zone name)"
  value       = var.spec.zone
}

output "size_gb" {
  description = "Provisioned size of the disk in GB"
  value       = google_compute_disk.this.size
}

# Normalized to the plain type name (last path segment) on BOTH engines:
# provider lines differ on whether the attribute is a bare name or a full
# self-link path, and an un-normalized export would silently break output
# parity between engines.
output "type" {
  description = "Disk type (plain type name, e.g. pd-balanced)"
  value       = element(reverse(split("/", google_compute_disk.this.type)), 0)
}
