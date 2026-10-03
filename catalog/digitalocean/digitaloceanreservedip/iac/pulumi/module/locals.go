package module

import (
	"strconv"

	"github.com/pkg/errors"
	digitaloceanreservedipv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanreservedip/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The reserved
// IP resources have no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanReservedIp *digitaloceanreservedipv1alpha1.DigitalOceanReservedIp

	// IsIpv6 selects the v6 resource pair; unset ip_version means ipv4.
	IsIpv6 bool

	// DropletId is the parsed numeric droplet id, nil when unassigned.
	// Droplet references resolve to the literal id before the module runs.
	DropletId *int
}

// initializeLocals copies iac-input fields into the Locals struct and
// parses the optional droplet assignment.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanreservedipv1alpha1.DigitalOceanReservedIpIacInput) (*Locals, error) {
	locals := &Locals{
		DigitalOceanReservedIp: iacInput.Target,
		IsIpv6:                 iacInput.Target.Spec.IpVersion == "ipv6",
	}

	if dropletRef := iacInput.Target.Spec.Droplet.GetValue(); dropletRef != "" {
		dropletId, err := strconv.Atoi(dropletRef)
		if err != nil {
			return nil, errors.Wrapf(err, "droplet %q is not a numeric Droplet ID", dropletRef)
		}
		locals.DropletId = &dropletId
	}

	return locals, nil
}
