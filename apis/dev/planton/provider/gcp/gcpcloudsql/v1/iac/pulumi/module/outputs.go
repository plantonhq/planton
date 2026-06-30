package module

// Keys exported by the gcp_cloud_sql Pulumi module.
const (
	OpInstanceName   = "instance_name"   // Name of the Cloud SQL instance
	OpConnectionName = "connection_name" // Full connection name (project:region:instance)
	OpPrivateIp      = "private_ip"      // Private IP address (if enabled)
	OpPublicIp       = "public_ip"       // Public IP address
	OpSelfLink       = "self_link"       // GCP resource self link
)
