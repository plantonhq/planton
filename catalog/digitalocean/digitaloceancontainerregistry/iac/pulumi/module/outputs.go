package module

const (
	// OpRegistryName is the exported output with the registry's name
	// (also the registry's resource identifier in DigitalOcean).
	OpRegistryName = "registry_name"

	// OpServerUrl is the exported output with the registry host,
	// always "registry.digitalocean.com".
	OpServerUrl = "server_url"

	// OpEndpoint is the exported output with the full docker push/pull
	// endpoint, i.e. "registry.digitalocean.com/<registry_name>".
	OpEndpoint = "endpoint"

	// OpRegion is the exported output with the region slug.
	OpRegion = "region"

	// OpDockerCredentials is the exported output with the base64-encoded
	// Docker config.json -- a SECRET, exported via pulumi.ToSecret. Empty when
	// the spec's docker_credentials block is unset.
	OpDockerCredentials = "docker_credentials"

	// OpCredentialExpirationTime is the exported output with the RFC 3339
	// expiry of the minted credentials. Empty when credentials are unset.
	OpCredentialExpirationTime = "credential_expiration_time"
)
