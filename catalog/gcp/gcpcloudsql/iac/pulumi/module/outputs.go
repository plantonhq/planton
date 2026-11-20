package module

// Keys exported by the gcpcloudsql Pulumi module.
const (
	OpInstanceName             = "instance_name"               // Composition key databases/users/replicas reference
	OpConnectionName           = "connection_name"             // project:region:instance for Auth Proxy / connectors
	OpPrivateIp                = "private_ip"                  // Private IP (empty unless private_network set)
	OpPublicIp                 = "public_ip"                   // Public IPv4 (empty unless ipv4_enabled)
	OpSelfLink                 = "self_link"                   // GCP resource self link
	OpServiceAccountEmail      = "service_account_email"       // Instance's Google-managed service account
	OpDnsName                  = "dns_name"                    // DNS name (PSC-enabled instances)
	OpPscServiceAttachmentLink = "psc_service_attachment_link" // PSC service attachment for consumer endpoints
	OpSocketPath               = "socket_path"                 // /cloudsql/{connection_name} Unix socket path
)
