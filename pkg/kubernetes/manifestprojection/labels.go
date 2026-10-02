package manifestprojection

// Planton's identity labels on a projection kind's object: the planton.ai/*
// family both engines stamp, in the order the generated Terraform module lists
// them. Values come from the manifest's metadata; the last three are stamped
// only when set.
const (
	LabelResource     = "planton.ai/resource"
	LabelResourceName = "planton.ai/resource-name"
	LabelResourceKind = "planton.ai/resource-kind"
	LabelResourceID   = "planton.ai/resource-id"
	LabelOrganization = "planton.ai/organization"
	LabelEnvironment  = "planton.ai/environment"
)

// Identity is the manifest metadata the identity labels are read from.
type Identity struct {
	Kind         string
	Name         string
	ID           string
	Organization string
	Environment  string
}

// IdentityLabels returns the identity labels for one object.
func IdentityLabels(id Identity) map[string]string {
	labels := map[string]string{
		LabelResource:     "true",
		LabelResourceName: id.Name,
		LabelResourceKind: id.Kind,
	}
	if id.ID != "" {
		labels[LabelResourceID] = id.ID
	}
	if id.Organization != "" {
		labels[LabelOrganization] = id.Organization
	}
	if id.Environment != "" {
		labels[LabelEnvironment] = id.Environment
	}
	return labels
}

// MergeLabels lays the identity labels over a manifest's own labels: a key in
// both keeps the identity value.
func MergeLabels(own, identity map[string]string) map[string]string {
	merged := make(map[string]string, len(own)+len(identity))
	for k, v := range own {
		merged[k] = v
	}
	for k, v := range identity {
		merged[k] = v
	}
	return merged
}
