package module

import (
	gcpgkeclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpgkecluster/v1"
)

// Locals groups commonly accessed input data.
type Locals struct {
	GcpGkeCluster     *gcpgkeclusterv1.GcpGkeCluster
	ReleaseChannelStr string
}

// initializeLocals converts the raw stack input into a Locals struct.
func initializeLocals(stackInput *gcpgkeclusterv1.GcpGkeClusterStackInput) *Locals {
	l := &Locals{
		GcpGkeCluster: stackInput.Target,
	}

	// Map the proto enum to the literal string Google expects.
	switch l.GcpGkeCluster.Spec.GetReleaseChannel() {
	case gcpgkeclusterv1.GkeReleaseChannel_RAPID:
		l.ReleaseChannelStr = "RAPID"
	case gcpgkeclusterv1.GkeReleaseChannel_REGULAR:
		l.ReleaseChannelStr = "REGULAR"
	case gcpgkeclusterv1.GkeReleaseChannel_STABLE:
		l.ReleaseChannelStr = "STABLE"
	case gcpgkeclusterv1.GkeReleaseChannel_NONE:
		l.ReleaseChannelStr = "UNSPECIFIED"
	default:
		l.ReleaseChannelStr = "REGULAR" // sensible default
	}

	return l
}
