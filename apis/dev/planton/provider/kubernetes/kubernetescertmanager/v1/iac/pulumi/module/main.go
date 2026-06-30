package module

import (
	"github.com/pkg/errors"
	kubernetescertmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetescertmanager/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetescertmanagerv1.KubernetesCertManagerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}

	spec := stackInput.Target.Spec

	chartVersion := spec.GetHelmChartVersion()

	createdNamespace, err := namespace(ctx, stackInput, locals, kubeProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	annotations := pulumi.StringMap{}
	if wi := spec.WorkloadIdentity; wi != nil {
		if gke := wi.GetGke(); gke != nil {
			annotations["iam.gke.io/gcp-service-account"] = pulumi.String(gke.ServiceAccountEmail)
		} else if eks := wi.GetEks(); eks != nil {
			annotations["eks.amazonaws.com/role-arn"] = pulumi.String(eks.RoleArn)
		} else if aks := wi.GetAks(); aks != nil {
			annotations["azure.workload.identity/client-id"] = pulumi.String(aks.ClientId)
		}
	}

	saOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	_, err = corev1.NewServiceAccount(ctx, locals.ServiceAccountName,
		&corev1.ServiceAccountArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:        pulumi.String(locals.ServiceAccountName),
				Namespace:   pulumi.String(locals.Namespace),
				Annotations: annotations,
			},
		},
		saOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create service account")
	}

	helmValues := pulumi.Map{
		"installCRDs": pulumi.Bool(true),
		"serviceAccount": pulumi.Map{
			"create": pulumi.Bool(false),
			"name":   pulumi.String(locals.ServiceAccountName),
		},
		"extraArgs": pulumi.Array{
			pulumi.String("--dns01-recursive-nameservers-only"),
			pulumi.String("--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"),
		},
	}

	certManagerVersion := spec.GetKubernetesCertManagerVersion()
	if certManagerVersion != "" {
		helmValues["image"] = pulumi.Map{
			"tag": pulumi.String(certManagerVersion),
		}
	}

	if spec.SkipInstallSelfSignedIssuer {
		helmValues["startupapicheck"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	}

	helmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubeProvider)}, namespaceDeps...)
	_, err = helm.NewRelease(ctx, "cert-manager",
		&helm.ReleaseArgs{
			Name:            pulumi.String(vars.HelmChartName),
			Namespace:       pulumi.String(locals.Namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(chartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(true),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values:          helmValues,
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to install cert-manager helm release")
	}

	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpReleaseName, pulumi.String(vars.HelmChartName))
	ctx.Export(OpServiceAccountName, pulumi.String(locals.ServiceAccountName))

	return nil
}
