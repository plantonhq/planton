package module

import (
	alicloudcontainerregistryv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudcontainerregistry/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AliCloudContainerRegistry *alicloudcontainerregistryv1.AliCloudContainerRegistry
}

func initializeLocals(_ *pulumi.Context, stackInput *alicloudcontainerregistryv1.AliCloudContainerRegistryStackInput) *Locals {
	return &Locals{
		AliCloudContainerRegistry: stackInput.Target,
	}
}

func paymentType(spec *alicloudcontainerregistryv1.AliCloudContainerRegistrySpec) string {
	if spec.PaymentType != nil {
		return *spec.PaymentType
	}
	return "Subscription"
}

func namespaceAutoCreate(ns *alicloudcontainerregistryv1.AliCloudContainerRegistryNamespace) bool {
	if ns.AutoCreate != nil {
		return *ns.AutoCreate
	}
	return false
}

func namespaceDefaultVisibility(ns *alicloudcontainerregistryv1.AliCloudContainerRegistryNamespace) string {
	if ns.DefaultVisibility != nil {
		return *ns.DefaultVisibility
	}
	return "PRIVATE"
}

func optionalString(s string) pulumi.StringPtrInput {
	if s == "" {
		return nil
	}
	return pulumi.String(s)
}
