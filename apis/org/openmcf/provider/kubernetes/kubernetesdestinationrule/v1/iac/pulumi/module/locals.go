package module

import (
	kubernetesdestinationrulev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesdestinationrule/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesDestinationRule *kubernetesdestinationrulev1.KubernetesDestinationRule
	// DestinationRuleName is the Kubernetes resource name of the DestinationRule. It equals
	// the OpenMCF resource metadata.name.
	DestinationRuleName string
	// Namespace is the resolved namespace the DestinationRule is created in.
	Namespace string
	Labels    map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *kubernetesdestinationrulev1.KubernetesDestinationRuleStackInput) *Locals {
	target := stackInput.Target
	metadata := target.Metadata
	spec := target.Spec

	destinationRuleName := metadata.Name

	// namespace is a StringValueOrRef foreign key. The platform middleware resolves
	// valueFrom references to literal strings before the IaC module runs, so GetValue()
	// returns the resolved value.
	namespace := spec.GetNamespace().GetValue()

	labels := map[string]string{
		"app.kubernetes.io/name":       "destination-rule",
		"app.kubernetes.io/instance":   metadata.Name,
		"app.kubernetes.io/managed-by": "openmcf",
		"app.kubernetes.io/component":  "destination-rule",
	}

	return &Locals{
		KubernetesDestinationRule: target,
		DestinationRuleName:       destinationRuleName,
		Namespace:                 namespace,
		Labels:                    labels,
	}
}
