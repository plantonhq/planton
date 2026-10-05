package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpsccnotificationconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccnotificationconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig        *gcpprovider.GcpProviderConfig
	GcpSccNotificationConfig *gcpsccnotificationconfigv1alpha1.GcpSccNotificationConfig
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpsccnotificationconfigv1alpha1.GcpSccNotificationConfigIacInput) *Locals {
	locals := &Locals{}
	locals.GcpSccNotificationConfig = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
