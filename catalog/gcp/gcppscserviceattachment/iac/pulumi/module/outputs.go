package module

// Output keys must match the field names in outputs.proto -- the outputs
// transformer maps raw engine outputs onto the proto by name.
const (
	// OpSelfLink is the attachment's self link: a consumer forwarding
	// rule's target.
	OpSelfLink = "self_link"
	// OpAttachmentName is the resolved cloud-side name.
	OpAttachmentName = "attachment_name"
	// OpRegion is the attachment's region.
	OpRegion = "region"
	// OpFingerprint is the server-computed fingerprint.
	OpFingerprint = "fingerprint"
	// OpConnectedEndpointsCount is the number of consumer endpoints connected
	// at provisioning time.
	OpConnectedEndpointsCount = "connected_endpoints_count"
)
