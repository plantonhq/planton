package module

import (
	"strconv"

	kubernetesaltinityoperatorv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesaltinityoperator/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// locals contains computed values and data transformations for the Altinity operator deployment
type locals struct {
	// Namespace is the resolved namespace where the operator will be installed
	Namespace string

	// Labels are the Kubernetes labels applied to all resources
	Labels map[string]string

	// HelmValues is the prepared configuration map for the Helm chart
	HelmValues pulumi.Map

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	HelmReleaseName string
}

// newLocals creates and initializes local values from the stack input
func newLocals(stackInput *kubernetesaltinityoperatorv1.KubernetesAltinityOperatorStackInput) *locals {
	l := &locals{}
	target := stackInput.Target

	// Determine namespace - use from spec or default
	l.Namespace = target.Spec.Namespace.GetValue()
	if l.Namespace == "" {
		l.Namespace = vars.DefaultNamespace
	}

	// Computed resource names to avoid conflicts when multiple instances share a namespace
	// Format: {metadata.name}-{purpose}
	l.HelmReleaseName = target.Metadata.Name

	// Build standard labels for all resources
	l.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesAltinityOperator.String(),
	}
	if target.Metadata.Id != "" {
		l.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		l.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		l.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// Prepare helm values with CRD installation enabled and resource limits from spec
	l.HelmValues = pulumi.Map{
		"operator": pulumi.Map{
			"createCRD": pulumi.Bool(true),
			"resources": pulumi.Map{
				"limits": pulumi.Map{
					"cpu":    pulumi.String(stackInput.Target.Spec.Container.Resources.Limits.Cpu),
					"memory": pulumi.String(stackInput.Target.Spec.Container.Resources.Limits.Memory),
				},
				"requests": pulumi.Map{
					"cpu":    pulumi.String(stackInput.Target.Spec.Container.Resources.Requests.Cpu),
					"memory": pulumi.String(stackInput.Target.Spec.Container.Resources.Requests.Memory),
				},
			},
		},
		// Configure operator to watch all namespaces cluster-wide
		// Use regex pattern ".*" to match all namespaces (empty array watches only installation namespace)
		"configs": pulumi.Map{
			"files": pulumi.Map{
				"config.yaml": pulumi.Map{
					"watch": pulumi.Map{
						"namespaces": pulumi.Array{
							pulumi.String(".*"),
						},
					},
				},
			},
		},
	}

	return l
}
