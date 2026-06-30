package module

const (
	// OpPoolId is the exported stack output name for the node pool's
	// regional ID.
	OpPoolId = "pool_id"

	// OpPoolVersion is the exported stack output name for the actual
	// Kubernetes version running on pool nodes.
	OpPoolVersion = "pool_version"

	// OpCurrentSize is the exported stack output name for the actual
	// number of nodes in the pool (may differ from spec when autoscaling).
	OpCurrentSize = "current_size"
)
