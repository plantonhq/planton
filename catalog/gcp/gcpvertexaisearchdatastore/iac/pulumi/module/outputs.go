package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	OpName            = "name"
	OpDataStoreId     = "data_store_id"
	OpLocation        = "location"
	OpDefaultSchemaId = "default_schema_id"
	OpSchemaName      = "schema_name"
	OpTargetSiteNames = "target_site_names"
	OpSitemapNames    = "sitemap_names"
)
