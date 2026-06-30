package module

import (
	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewaydnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaydnszone/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module.
//
// NOTE: Unlike most other Scaleway resource modules, there is no
// ScalewayTags field here. Neither Scaleway DNS zones nor DNS records
// support tags in the API. This is a Scaleway platform limitation, not
// a design choice. The zone name and metadata.name serve as the primary
// identifiers for resource tracking.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayDnsZone        *scalewaydnszonev1.ScalewayDnsZone
	// ZoneName is the computed zone name: "{subdomain}.{domain}" or
	// just "{domain}" for root zones. Computed once here so all other
	// module files reference a single source of truth.
	ZoneName string
}

// initializeLocals copies stack-input fields into the Locals struct
// and computes the zone name from domain + subdomain.
//
// Unlike other Scaleway resource modules, this does not build a tag
// slice because Scaleway DNS zones and records do not support tags.
func initializeLocals(_ *pulumi.Context, stackInput *scalewaydnszonev1.ScalewayDnsZoneStackInput) *Locals {
	spec := stackInput.Target.Spec

	// Compute zone name following Scaleway's convention:
	// subdomain.domain (if subdomain is set) or just domain (root zone).
	zoneName := spec.Domain
	if spec.Subdomain != "" {
		zoneName = spec.Subdomain + "." + spec.Domain
	}

	return &Locals{
		ScalewayDnsZone:        stackInput.Target,
		ScalewayProviderConfig: stackInput.ProviderConfig,
		ZoneName:               zoneName,
	}
}
