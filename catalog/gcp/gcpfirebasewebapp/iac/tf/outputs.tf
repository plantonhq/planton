# The Firebase-assigned app id (firebaseConfig.appId).
output "app_id" {
  description = "The Firebase-assigned app id"
  value       = google_firebase_web_app.this.app_id
}

# The app's full resource name -- the handle the Firebase Management API
# addresses the app by.
output "name" {
  description = "The app's full resource name, projects/{project}/webApps/{app_id}"
  value       = google_firebase_web_app.this.name
}

# The UID of the API key associated with the app -- the one from the spec,
# or the key Firebase associated or provisioned when none was given.
output "api_key_id" {
  description = "The UID of the API key associated with the app"
  value       = google_firebase_web_app.this.api_key_id
}

# The URLs Firebase records the app as hosted at (empty when none).
output "app_urls" {
  description = "The URLs where the web app is hosted, as Firebase records them"
  value       = google_firebase_web_app.this.app_urls
}

# firebaseConfig -- every value a client identifier that ships in the page,
# from the deferred config lookup; the conditionally present ones degrade
# to "".
output "api_key" {
  description = "firebaseConfig.apiKey -- the API key string the page presents"
  value       = data.google_firebase_web_app_config.this.api_key
}

output "auth_domain" {
  description = "firebaseConfig.authDomain"
  value       = data.google_firebase_web_app_config.this.auth_domain
}

output "database_url" {
  description = "firebaseConfig.databaseURL, when the project has a Realtime Database"
  value       = try(data.google_firebase_web_app_config.this.database_url, "")
}

output "storage_bucket" {
  description = "firebaseConfig.storageBucket, when the project has a default bucket"
  value       = try(data.google_firebase_web_app_config.this.storage_bucket, "")
}

output "location_id" {
  description = "firebaseConfig.locationId, once the project's default location is finalized"
  value       = try(data.google_firebase_web_app_config.this.location_id, "")
}

output "messaging_sender_id" {
  description = "firebaseConfig.messagingSenderId -- the project number a browser client registers with for push"
  value       = data.google_firebase_web_app_config.this.messaging_sender_id
}

output "measurement_id" {
  description = "firebaseConfig.measurementId, when the app is linked to a Google Analytics property"
  value       = try(data.google_firebase_web_app_config.this.measurement_id, "")
}
