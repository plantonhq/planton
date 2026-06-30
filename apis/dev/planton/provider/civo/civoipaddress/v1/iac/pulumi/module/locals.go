package module

import (
	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	civoipaddressv1 "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civoipaddress/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds frequently accessed values so that downstream functions
// read like Terraform locals rather than deeply nested getters.
type Locals struct {
	CivoProviderConfig *civoprovider.CivoProviderConfig
	CivoIpAddress      *civoipaddressv1.CivoIpAddress
}

// initializeLocals prepares the struct in the simplest possible way.
func initializeLocals(_ *pulumi.Context, stackInput *civoipaddressv1.CivoIpAddressStackInput) *Locals {
	locals := &Locals{}
	locals.CivoIpAddress = stackInput.Target
	locals.CivoProviderConfig = stackInput.ProviderConfig
	return locals
}
