output "name" {
  description = "Full resource name of the floor setting ({parent}/locations/{location}/floorSetting)"
  value       = google_model_armor_floorsetting.this.id
}

output "parent" {
  description = "The project, folder, or organization the floor governs"
  value       = google_model_armor_floorsetting.this.parent
}

output "location" {
  description = "The location the floor is managed at"
  value       = google_model_armor_floorsetting.this.location
}
