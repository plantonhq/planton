# The Firebase-assigned app id (mobilesdk_app_id in google-services.json).
output "app_id" {
  description = "The Firebase-assigned app id"
  value       = google_firebase_android_app.this.app_id
}

# The app's full resource name -- the handle the Firebase Management API
# addresses the app by.
output "name" {
  description = "The app's full resource name, projects/{project}/androidApps/{app_id}"
  value       = google_firebase_android_app.this.name
}

# The UID of the API key associated with the app -- the one from the spec,
# or the key Firebase associated or provisioned when none was given.
output "api_key_id" {
  description = "The UID of the API key associated with the app"
  value       = google_firebase_android_app.this.api_key_id
}

# The configuration file -- a build input that ships in the APK, not a
# secret. Both values come from the deferred config lookup.
output "config_filename" {
  description = "The configuration file's name (google-services.json)"
  value       = data.google_firebase_android_app_config.this.config_filename
}

output "config_file_contents" {
  description = "The configuration file's contents, base64-encoded"
  value       = data.google_firebase_android_app_config.this.config_file_contents
}
