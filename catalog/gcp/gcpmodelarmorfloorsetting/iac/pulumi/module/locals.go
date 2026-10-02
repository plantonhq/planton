package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpmodelarmorfloorsettingv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmodelarmorfloorsetting/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig         *gcpprovider.GcpProviderConfig
	GcpModelArmorFloorSetting *gcpmodelarmorfloorsettingv1alpha1.GcpModelArmorFloorSetting
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpmodelarmorfloorsettingv1alpha1.GcpModelArmorFloorSettingStackInput) *Locals {
	locals := &Locals{}
	locals.GcpModelArmorFloorSetting = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
