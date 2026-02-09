package module

import (
	"github.com/pkg/errors"
	kuberneteselasticsearchv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kuberneteselasticsearch/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kuberneteselasticsearchv1.KubernetesElasticsearchStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if err := elasticsearch(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create elastic search resources")
	}

	if (locals.KubernetesElasticsearch.Spec.Elasticsearch.Ingress != nil &&
		locals.KubernetesElasticsearch.Spec.Elasticsearch.Ingress.Enabled) ||
		(locals.KubernetesElasticsearch.Spec.Kibana != nil &&
			locals.KubernetesElasticsearch.Spec.Kibana.Enabled &&
			locals.KubernetesElasticsearch.Spec.Kibana.Ingress != nil &&
			locals.KubernetesElasticsearch.Spec.Kibana.Ingress.Enabled) {
		if err := ingress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create ingress resources")
		}
	}

	return nil
}
