package module

const (
	OpPolicyId       = "policy_id"
	OpPolicyName     = "policy_name"
	OpPolicySelfLink = "policy_self_link"
	OpFingerprint    = "fingerprint"
	// OpRegion is the region of a regional policy; empty for a global one.
	OpRegion = "region"
	// OpNetworkEdgeSecurityServiceSelfLink is the self-link of the network
	// edge security service enrolling the region in advanced network DDoS
	// protection; empty when the spec declares none.
	OpNetworkEdgeSecurityServiceSelfLink = "network_edge_security_service_self_link"
)
