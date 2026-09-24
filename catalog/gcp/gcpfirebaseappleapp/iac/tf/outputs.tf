# The Firebase-assigned app id (GOOGLE_APP_ID in GoogleService-Info.plist).
output "app_id" {
  description = "The Firebase-assigned app id"
  value       = google_firebase_apple_app.this.app_id
}

# The app's full resource name -- the handle the Firebase Management API
# addresses the app by (the API path keeps the historical iosApps segment).
output "name" {
  description = "The app's full resource name, projects/{project}/iosApps/{app_id}"
  value       = google_firebase_apple_app.this.name
}

# The UID of the API key associated with the app -- the one from the spec,
# or the key Firebase associated or provisioned when none was given.
output "api_key_id" {
  description = "The UID of the API key associated with the app"
  value       = google_firebase_apple_app.this.api_key_id
}

# The configuration file -- a build input that ships in the app bundle, not
# a secret. Both values come from the deferred config lookup.
output "config_filename" {
  description = "The configuration file's name (GoogleService-Info.plist)"
  value       = data.google_firebase_apple_app_config.this.config_filename
}

output "config_file_contents" {
  description = "The configuration file's contents, base64-encoded"
  value       = data.google_firebase_apple_app_config.this.config_file_contents
}
