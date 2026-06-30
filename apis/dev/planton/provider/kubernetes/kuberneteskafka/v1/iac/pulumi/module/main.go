package module

import (
	"github.com/pkg/errors"
	kuberneteskafkav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteskafka/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kuberneteskafkav1.KubernetesKafkaStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	//create kubernetes-provider from the credential in the stack-kowlConfigTemplateInput
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// ------------------------------ namespace ----------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	//create kafka cluster custom resource
	createdKafkaCluster, err := kafkaCluster(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create kafka-cluster resources")
	}

	//create kafka admin user
	if err := kafkaAdminUser(ctx, locals, kubernetesProvider, createdKafkaCluster, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create kafka admin user")
	}

	//create kafka topics
	if err := kafkaTopics(ctx, locals, kubernetesProvider, createdKafkaCluster, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create kafka topics")
	}

	//create schema-registry
	if locals.KubernetesKafka.Spec.SchemaRegistryContainer != nil &&
		locals.KubernetesKafka.Spec.SchemaRegistryContainer.IsEnabled {
		if err := schemaRegistry(ctx, locals, kubernetesProvider, createdKafkaCluster, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create schema registry deployment")
		}
	}

	//create kowl
	if locals.KubernetesKafka.Spec.IsDeployKafkaUi {
		if err := kowl(ctx, locals, kubernetesProvider, createdKafkaCluster, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create kowl deployment")
		}
	}
	return nil
}
