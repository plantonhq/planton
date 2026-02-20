package module

import (
	alicloudcontainerregistryv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudcontainerregistry/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AlicloudContainerRegistry *alicloudcontainerregistryv1.AlicloudContainerRegistry
}

func initializeLocals(_ *pulumi.Context, stackInput *alicloudcontainerregistryv1.AlicloudContainerRegistryStackInput) *Locals {
	return &Locals{
		AlicloudContainerRegistry: stackInput.Target,
	}
}

func paymentType(spec *alicloudcontainerregistryv1.AlicloudContainerRegistrySpec) string {
	if spec.PaymentType != nil {
		return *spec.PaymentType
	}
	return "Subscription"
}

func namespaceAutoCreate(ns *alicloudcontainerregistryv1.AlicloudContainerRegistryNamespace) bool {
	if ns.AutoCreate != nil {
		return *ns.AutoCreate
	}
	return false
}

func namespaceDefaultVisibility(ns *alicloudcontainerregistryv1.AlicloudContainerRegistryNamespace) string {
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
