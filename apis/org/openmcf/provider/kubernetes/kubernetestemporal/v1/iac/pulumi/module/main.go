package module

import (
	"github.com/pkg/errors"
	kubernetestemporalv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestemporal/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the single entry-point consumed by OpenMCF runtime.
// It wires together all noun-style helpers in a Terraform-like, top-down
// order so the flow is easy for DevOps engineers to follow.
func Resources(ctx *pulumi.Context,
	stackInput *kubernetestemporalv1.KubernetesTemporalStackInput) error {

	locals := initializeLocals(ctx, stackInput)

	if locals.KubernetesTemporal.Spec.Database == nil {
		return errors.New("database configuration is required")
	}

	if locals.KubernetesTemporal.Spec.Database.Backend != kubernetestemporalv1.KubernetesTemporalDatabaseBackend_cassandra &&
		locals.KubernetesTemporal.Spec.Database.ExternalDatabase == nil {

		return errors.New("external_database must be provided when backend is not cassandra")
	}

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	if err := dbPasswordSecret(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create database password secret")
	}

	if err := helmChart(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to install Temporal Helm chart")
	}

	if err := frontendIngress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create frontend gRPC ingress")
	}

	if err := frontendHttpIngress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create frontend HTTP ingress")
	}

	if err := webUiIngress(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create web UI ingress")
	}

	return nil
}
