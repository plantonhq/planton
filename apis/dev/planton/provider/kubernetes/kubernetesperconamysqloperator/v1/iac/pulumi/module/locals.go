package module

import (
	"fmt"
	"strconv"

	kubernetesperconamysqloperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesperconamysqloperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used, derived values in one place –
// similar to a Terraform "locals {}" block.
type Locals struct {
	Namespace                      string
	Labels                         map[string]string
	KubernetesPerconaMysqlOperator *kubernetesperconamysqloperatorv1.KubernetesPerconaMysqlOperator
	// Computed resource names to avoid conflicts when multiple instances share a namespace
	HelmReleaseName string
}

// initializeLocals builds the Locals struct and immediately exports the
// values required by KubernetesPerconaMysqlOperatorStackOutputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetesperconamysqloperatorv1.KubernetesPerconaMysqlOperatorStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesPerconaMysqlOperator = stackInput.Target
	target := stackInput.Target

	// ------------------------------- labels ----------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPerconaMysqlOperator.String(),
	}
	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	// Users can prefix metadata.name with component type if needed (e.g., "pxc-operator-prod")
	locals.HelmReleaseName = fmt.Sprintf("%s-pxc-operator", target.Metadata.Name)

	return locals
}
