package module

import (
	"github.com/pkg/errors"
	kubernetesingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesingressnginx/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates all Pulumi resources for the Ingress‑Nginx add‑on.
func Resources(ctx *pulumi.Context,
	stackInput *kubernetesingressnginxv1.KubernetesIngressNginxStackInput) error {

	locals := initializeLocals(ctx, stackInput)

	// Build Kubernetes provider from the target‑cluster credential
	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	spec := stackInput.Target.Spec

	// Build service annotations based on cluster flavour + internal/external flag
	var serviceAnnotations map[string]string
	if gke := spec.GetGke(); gke != nil {
		if spec.Internal {
			serviceAnnotations = map[string]string{
				"cloud.google.com/load-balancer-type": "internal",
			}
		} else {
			serviceAnnotations = map[string]string{
				"cloud.google.com/load-balancer-type": "external",
			}
		}
	} else if eks := spec.GetEks(); eks != nil {
		if spec.Internal {
			serviceAnnotations = map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal",
			}
		} else {
			serviceAnnotations = map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-scheme": "internet-facing",
			}
		}
		_ = eks // mute unused‑var if eks has no fields accessed here
	} else if aks := spec.GetAks(); aks != nil {
		if spec.Internal {
			serviceAnnotations = map[string]string{
				"service.beta.kubernetes.io/azure-load-balancer-internal": "true",
			}
		}
		_ = aks
	}

	// ---------------------------------------------------------------------
	// Namespace - conditionally create based on create_namespace flag
	// ---------------------------------------------------------------------
	createdNamespace, err := namespace(ctx, stackInput, locals, kubeProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	// ---------------------------------------------------------------------
	// Helm chart
	// ---------------------------------------------------------------------
	// Build values dynamically so that annotations are only set when needed
	serviceVals := pulumi.Map{
		"type": pulumi.String("LoadBalancer"),
	}
	if len(serviceAnnotations) > 0 {
		serviceVals["annotations"] = pulumi.ToStringMap(serviceAnnotations)
	}

	values := pulumi.Map{
		"controller": pulumi.Map{
			"service":                  serviceVals,
			"ingressClassResource":     pulumi.Map{"default": pulumi.Bool(true)},
			"watchIngressWithoutClass": pulumi.Bool(true),
		},
	}

	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubeProvider),
	}, namespaceDeps...)

	_, err = helm.NewRelease(ctx, locals.ReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(locals.ReleaseName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(locals.ChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          values,
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install kubernetes-ingress-nginx helm release")
	}

	return nil
}
