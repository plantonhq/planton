package module

import (
	"github.com/pkg/errors"
	kubernetesclickhousev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesclickhouse/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesclickhousev1.KubernetesClickHouseStackInput) error {
	// Initialize local variables and exports
	locals := initializeLocals(ctx, stackInput)

	// Create Kubernetes provider from stack-input credentials
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Create or reference namespace for ClickHouse resources based on create_namespace flag
	createdNamespace, err := createOrGetNamespace(ctx, locals, stackInput.Target.Spec, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create or get namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Create password secret for ClickHouse authentication
	createdSecret, err := createPasswordSecret(ctx, locals, kubernetesProvider, namespaceDeps)
	if err != nil {
		return errors.Wrap(err, "failed to create password secret")
	}

	// Create ClickHouseKeeperInstallation if auto-managed Keeper is requested
	if shouldCreateClickHouseKeeper(locals.KubernetesClickHouse.Spec) {
		keeperConfig := getKeeperConfig(locals.KubernetesClickHouse.Spec)
		if err := clickhouseKeeperInstallation(ctx, locals, kubernetesProvider, keeperConfig, namespaceDeps); err != nil {
			return errors.Wrap(err, "failed to create ClickHouseKeeperInstallation")
		}
	}

	// Create ClickHouseInstallation CRD using Altinity operator
	if err := clickhouseInstallation(ctx, locals, kubernetesProvider, createdSecret, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create ClickHouseInstallation")
	}

	// Create ingress LoadBalancer service if enabled
	if err := createIngressLoadBalancer(ctx, locals, kubernetesProvider, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create ingress load balancer")
	}

	return nil
}
