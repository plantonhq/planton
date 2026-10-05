// Package azuretagkeys names the identity tags every Azure module writes on
// the resources it creates.
//
// Unlike AWS (planton.ai/* keys) and GCP (planton-ai_* labels), Azure tags
// use plain snake_case keys. The OpenTofu modules spell these same keys as
// literals in iac/tf/locals.tf, and
// hack/guards/ensure_cross_engine_azure_tag_values.sh holds both engines to
// the same keys and values: resource_kind is the kind's CatalogKind
// enum name lowercased, and resource_id is written only when the resource
// has an id.
package azuretagkeys

const (
	Resource     = "resource"
	ResourceName = "resource_name"
	ResourceKind = "resource_kind"
	ResourceId   = "resource_id"
	Organization = "organization"
	Environment  = "environment"
)
