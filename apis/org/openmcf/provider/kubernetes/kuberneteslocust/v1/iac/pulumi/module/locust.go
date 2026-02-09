package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/mergestringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func locust(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// Create a ConfigMap for the main.py file
	cmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := kubernetescorev1.NewConfigMap(ctx, locals.MainPyConfigMapName, &kubernetescorev1.ConfigMapArgs{
		Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.MainPyConfigMapName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		}),
		Data: pulumi.StringMap{
			"main.py": pulumi.String(locals.KubernetesLocust.Spec.LoadTest.MainPyContent),
		},
	}, cmOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create main py configmap")
	}

	// Create a ConfigMap for the lib files
	_, err = kubernetescorev1.NewConfigMap(ctx, locals.LibFilesConfigMapName, &kubernetescorev1.ConfigMapArgs{
		Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.LibFilesConfigMapName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		}),
		Data: pulumi.ToStringMap(locals.KubernetesLocust.Spec.LoadTest.LibFilesContent),
	}, cmOpts...)

	if err != nil {
		return errors.Wrap(err, "failed to create lib files configmap")
	}

	//https://github.com/deliveryhero/helm-charts/tree/master/stable/locust#values
	// helm values
	var helmValues = pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesLocust.Metadata.Name),
		"master": pulumi.Map{
			"replicas":  pulumi.Int(locals.KubernetesLocust.Spec.MasterContainer.Replicas),
			"resources": containerresources.ConvertToPulumiMap(locals.KubernetesLocust.Spec.MasterContainer.Resources),
		},
		"worker": pulumi.Map{
			"replicas":  pulumi.Int(locals.KubernetesLocust.Spec.WorkerContainer.Replicas),
			"resources": containerresources.ConvertToPulumiMap(locals.KubernetesLocust.Spec.WorkerContainer.Resources),
		},
		"loadtest": pulumi.Map{
			"name":                        pulumi.String(locals.KubernetesLocust.Spec.LoadTest.Name),
			"locust_locustfile_configmap": pulumi.String(locals.MainPyConfigMapName),
			"locust_lib_configmap":        pulumi.String(locals.LibFilesConfigMapName),
		},
	}
	mergestringmaps.MergeMapToPulumiMap(helmValues, locals.KubernetesLocust.Spec.HelmValues)

	// Deploying a Locust Helm chart from the Helm repository.
	helmOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err = helmv3.NewChart(ctx,
		locals.KubernetesLocust.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String("locust"),
			Version:   pulumi.String("0.31.5"), // Use the Helm chart version you want to install
			Namespace: pulumi.String(locals.Namespace),
			Values:    helmValues,
			//if you need to add the repository, you can specify `repo url`:
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String("https://charts.deliveryhero.io"), // The URL for the Helm chart repository
			},
		}, helmOpts...)

	if err != nil {
		return errors.Wrap(err, "failed to create locust resource")
	}
	return nil
}
