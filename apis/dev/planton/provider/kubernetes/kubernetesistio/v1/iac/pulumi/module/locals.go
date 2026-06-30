package module

import (
	"fmt"

	kubernetesistiov1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesistio/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values for the Istio module
type Locals struct {
	// KubernetesIstio holds the target resource
	KubernetesIstio *kubernetesistiov1.KubernetesIstio

	// Computed Helm release names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{chart}
	BaseReleaseName    string
	IstiodReleaseName  string
	GatewayReleaseName string

	// Namespace configuration
	SystemNamespace  string
	GatewayNamespace string
}

// initializeLocals creates and initializes the Locals struct with computed values
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetesistiov1.KubernetesIstioStackInput) *Locals {
	locals := &Locals{}
	target := stackInput.Target
	spec := target.Spec
	resourceName := target.Metadata.Name

	locals.KubernetesIstio = target

	// Computed Helm release names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{chart}
	// Users can prefix metadata.name with component type if needed (e.g., "istio-prod")
	locals.BaseReleaseName = fmt.Sprintf("%s-base", resourceName)
	locals.IstiodReleaseName = fmt.Sprintf("%s-istiod", resourceName)
	locals.GatewayReleaseName = fmt.Sprintf("%s-gateway", resourceName)

	// Namespace configuration
	// Use spec.namespace if provided, otherwise fall back to standard istio-system
	namespace := spec.Namespace.GetValue()
	if namespace == "" {
		locals.SystemNamespace = vars.SystemNamespace
	} else {
		locals.SystemNamespace = namespace
	}
	locals.GatewayNamespace = vars.GatewayNamespace

	return locals
}
