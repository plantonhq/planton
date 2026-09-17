package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpAppId is the Firebase-assigned app id (firebaseConfig.appId).
	OpAppId = "app_id"
	// OpName is the app's full resource name, projects/{p}/webApps/{app_id}.
	OpName = "name"
	// OpApiKeyId is the UID of the API key associated with the app.
	OpApiKeyId = "api_key_id"
	// OpAppUrls are the URLs Firebase records the app as hosted at.
	OpAppUrls = "app_urls"
	// OpApiKey is firebaseConfig.apiKey -- the key STRING the page presents,
	// a client identifier exported plain.
	OpApiKey = "api_key"
	// OpAuthDomain is firebaseConfig.authDomain.
	OpAuthDomain = "auth_domain"
	// OpDatabaseUrl is firebaseConfig.databaseURL (empty without an RTDB).
	OpDatabaseUrl = "database_url"
	// OpStorageBucket is firebaseConfig.storageBucket (empty without a
	// default bucket).
	OpStorageBucket = "storage_bucket"
	// OpLocationId is firebaseConfig.locationId (empty until finalized).
	OpLocationId = "location_id"
	// OpMessagingSenderId is firebaseConfig.messagingSenderId -- the project
	// number a browser client registers with for push.
	OpMessagingSenderId = "messaging_sender_id"
	// OpMeasurementId is firebaseConfig.measurementId (empty without a
	// linked Google Analytics property).
	OpMeasurementId = "measurement_id"
)
