package module

import (
	"github.com/pkg/errors"
	kubernetesargocdv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesargocd/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesargocdv1.KubernetesArgocdStackInput) error {
	// Initialize local values for computed data transformations
	locals := initializeLocals(ctx, stackInput)

	// Create kubernetes-provider from the credential in the stack-input
	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubeProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// Get resource specifications
	containerResources := stackInput.Target.Spec.Container.Resources

	// Prepare Helm chart values
	helmValues := pulumi.Map{
		"server": pulumi.Map{
			"resources": pulumi.Map{
				"requests": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Requests.Cpu),
					"memory": pulumi.String(containerResources.Requests.Memory),
				},
				"limits": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Limits.Cpu),
					"memory": pulumi.String(containerResources.Limits.Memory),
				},
			},
			"extraArgs": pulumi.StringArray{
				pulumi.String("--insecure"), // Allow HTTP access (use ingress for TLS termination)
			},
		},
		"controller": pulumi.Map{
			"resources": pulumi.Map{
				"requests": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Requests.Cpu),
					"memory": pulumi.String(containerResources.Requests.Memory),
				},
				"limits": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Limits.Cpu),
					"memory": pulumi.String(containerResources.Limits.Memory),
				},
			},
		},
		"repoServer": pulumi.Map{
			"resources": pulumi.Map{
				"requests": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Requests.Cpu),
					"memory": pulumi.String(containerResources.Requests.Memory),
				},
				"limits": pulumi.Map{
					"cpu":    pulumi.String(containerResources.Limits.Cpu),
					"memory": pulumi.String(containerResources.Limits.Memory),
				},
			},
		},
		"redis": pulumi.Map{
			"resources": pulumi.Map{
				"requests": pulumi.Map{
					"cpu":    pulumi.String("50m"),
					"memory": pulumi.String("64Mi"),
				},
				"limits": pulumi.Map{
					"cpu":    pulumi.String("100m"),
					"memory": pulumi.String("128Mi"),
				},
			},
		},
		"global": pulumi.Map{
			"image": pulumi.Map{
				"repository": pulumi.String("quay.io/argoproj/argocd"),
			},
		},
	}

	// Deploy Argo CD using the official Helm chart
	resourceId := stackInput.Target.Metadata.Name
	if stackInput.Target.Metadata.Id != "" {
		resourceId = stackInput.Target.Metadata.Id
	}

	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubeProvider),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, "argocd",
		&helm.ReleaseArgs{
			Name:      pulumi.String(resourceId),
			Namespace: pulumi.String(locals.Namespace),
			Chart:     pulumi.String("argo-cd"),
			Version:   pulumi.String("7.7.12"), // Pin to stable version
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://argoproj.github.io/argo-helm"),
			},
			Values:        helmValues,
			WaitForJobs:   pulumi.Bool(true),
			Timeout:       pulumi.Int(600), // 10 minutes
			Atomic:        pulumi.Bool(true),
			CleanupOnFail: pulumi.Bool(true),
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install Argo CD helm release")
	}

	return nil
}
