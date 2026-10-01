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

func initializeLocals(_ *pulumi.Context, stackInput *gcpsccnotificationconfigv1alpha1.GcpSccNotificationConfigStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSccNotificationConfig = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
