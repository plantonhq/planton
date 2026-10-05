package module

import (
	"fmt"
	"strconv"

	kubernetesflinkdeploymentv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesflinkdeployment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep: same rendered CR body, same
// outputs.
type Locals struct {
	KubernetesFlinkDeployment *kubernetesflinkdeploymentv1alpha1.KubernetesFlinkDeployment
	Spec                      *kubernetesflinkdeploymentv1alpha1.KubernetesFlinkDeploymentSpec

	// ResourceName is metadata.name — the CR name and the operator's
	// naming root for every derived object (`<name>-rest`,
	// `<name>-taskmanager-N-M`).
	ResourceName string

	// Namespace the cluster deploys into (resolved literal from the
	// spec's value-or-ref).
	Namespace string

	// Labels tie the module-created objects back to the Planton
	// resource.
	Labels map[string]string

	// Output handles per the operator naming contract.
	RestService        string
	RestEndpoint       string
	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesflinkdeploymentv1alpha1.KubernetesFlinkDeploymentIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesFlinkDeployment.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	resourceName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	// The JobManager REST Service is `<name>-rest` — the Flink REST API
	// and web UI on 8081; where session-mode jobs submit and where job
	// status reads from.
	restService := resourceName + vars.RestServiceSuffix

	return &Locals{
		KubernetesFlinkDeployment: target,
		Spec:                      spec,
		ResourceName:              resourceName,
		Namespace:                 namespace,
		Labels:                    labels,
		RestService:               restService,
		RestEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			restService, namespace, vars.RestPort),
		PortForwardCommand: fmt.Sprintf("kubectl port-forward -n %s service/%s %d:%d",
			namespace, restService, vars.RestPort, vars.RestPort),
	}
}
