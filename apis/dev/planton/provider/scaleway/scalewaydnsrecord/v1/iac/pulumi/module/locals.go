package module

import (
	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewaydnsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaydnsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module.
//
// NOTE: Unlike most other Scaleway resource modules, there is no
// ScalewayTags field here. Scaleway DNS records do not support tags
// in the API. This is a Scaleway platform limitation, not a design
// choice. The record FQDN and metadata.name serve as the primary
// identifiers for resource tracking.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayDnsRecord      *scalewaydnsrecordv1.ScalewayDnsRecord
}

// initializeLocals copies stack-input fields into the Locals struct.
//
// Unlike other Scaleway resource modules, this does not build a tag
// slice because Scaleway DNS records do not support tags.
func initializeLocals(_ *pulumi.Context, stackInput *scalewaydnsrecordv1.ScalewayDnsRecordStackInput) *Locals {
	return &Locals{
		ScalewayDnsRecord:      stackInput.Target,
		ScalewayProviderConfig: stackInput.ProviderConfig,
	}
}
