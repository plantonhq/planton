package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName                          = "name"
	OpConnectionId                  = "connection_id"
	OpLocation                      = "location"
	OpCloudResourceServiceAccountId = "cloud_resource_service_account_id"
	OpSparkServiceAccountId         = "spark_service_account_id"
	OpCloudSqlServiceAccountId      = "cloud_sql_service_account_id"
	OpConnectorServiceAccount       = "connector_service_account"
	OpAwsIdentity                   = "aws_identity"
	OpAzureIdentity                 = "azure_identity"
	OpAzureApplication              = "azure_application"
	OpAzureClientId                 = "azure_client_id"
	OpAzureObjectId                 = "azure_object_id"
	OpAzureRedirectUri              = "azure_redirect_uri"
)
