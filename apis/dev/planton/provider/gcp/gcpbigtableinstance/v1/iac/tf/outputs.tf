output "instance_id" {
  description = "Fully qualified instance resource name"
  value       = google_bigtable_instance.this.id
}

output "instance_name" {
  description = "Short instance name"
  value       = google_bigtable_instance.this.name
}
