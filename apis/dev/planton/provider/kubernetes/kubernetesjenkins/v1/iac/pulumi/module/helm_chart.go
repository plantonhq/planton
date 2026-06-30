package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/mergestringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func helmChart(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	createdAdminPasswordSecret *kubernetescorev1.Secret,
	namespaceDeps []pulumi.ResourceOption) error {

	// https://github.com/jenkinsci/helm-charts/blob/main/charts/jenkins/values.yaml
	var helmValues = pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesJenkins.Metadata.Name),
		"controller": pulumi.Map{
			"image": pulumi.Map{
				"tag": pulumi.String(vars.JenkinsDockerImageTag),
			},
			"resources": containerresources.ConvertToPulumiMap(locals.KubernetesJenkins.Spec.ContainerResources),
			"admin": pulumi.Map{
				"passwordKey":    pulumi.String(vars.JenkinsAdminPasswordSecretKey),
				"existingSecret": createdAdminPasswordSecret.Metadata.Name(),
			},
		},
	}

	//merge extra helm values provided in the spec with base values
	mergestringmaps.MergeMapToPulumiMap(helmValues, locals.KubernetesJenkins.Spec.HelmValues)

	//install jenkins helm-chart
	chartOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
	}, namespaceDeps...)

	_, err := helmv3.NewChart(ctx,
		locals.KubernetesJenkins.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(vars.HelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values:    helmValues,
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(vars.HelmChartRepoUrl),
			},
		}, chartOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create helm chart")
	}

	return nil
}
