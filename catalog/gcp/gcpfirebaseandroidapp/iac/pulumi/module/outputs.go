package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpAppId is the Firebase-assigned app id (mobilesdk_app_id).
	OpAppId = "app_id"
	// OpName is the app's full resource name, projects/{p}/androidApps/{app_id}.
	OpName = "name"
	// OpApiKeyId is the UID of the API key associated with the app.
	OpApiKeyId = "api_key_id"
	// OpConfigFilename is the configuration file's name (google-services.json).
	OpConfigFilename = "config_filename"
	// OpConfigFileContents is the configuration file, base64-encoded -- a
	// build input that ships in the APK, exported plain.
	OpConfigFileContents = "config_file_contents"
)
