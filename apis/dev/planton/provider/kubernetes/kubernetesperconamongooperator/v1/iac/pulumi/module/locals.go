package module

import (
	"strconv"

	kubernetesperconamongooperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesperconamongooperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals keeps all frequently-used, derived values in one place –
// similar to a Terraform "locals {}" block.
type Locals struct {
	Namespace                      string
	Labels                         map[string]string
	KubernetesPerconaMongoOperator *kubernetesperconamongooperatorv1.KubernetesPerconaMongoOperator

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	HelmReleaseName string
}

// initializeLocals builds the Locals struct and immediately exports the
// values required by KubernetesPerconaMongoOperatorStackOutputs.
func initializeLocals(ctx *pulumi.Context,
	stackInput *kubernetesperconamongooperatorv1.KubernetesPerconaMongoOperatorStackInput) *Locals {

	locals := &Locals{}
	locals.KubernetesPerconaMongoOperator = stackInput.Target
	target := stackInput.Target

	// ------------------------------- labels ----------------------------------
	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesPerconaMongoOperator.String(),
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

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name} for the Helm release name
	locals.HelmReleaseName = target.Metadata.Name

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	return locals
}

// vars contains Helm chart configuration constants.
var vars = struct {
	HelmChartName    string
	HelmChartRepo    string
	HelmChartVersion string
}{
	HelmChartName:    "psmdb-operator",
	HelmChartRepo:    "https://percona.github.io/percona-helm-charts/",
	HelmChartVersion: "1.20.1",
}
