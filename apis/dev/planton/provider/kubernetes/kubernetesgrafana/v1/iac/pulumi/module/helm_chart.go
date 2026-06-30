package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	grafanaHelmChartName    = "grafana"
	grafanaHelmChartVersion = "8.7.0"
	grafanaHelmChartRepoUrl = "https://grafana.github.io/helm-charts"
)

func helmChart(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption) error {

	// https://github.com/grafana/helm-charts/blob/main/charts/grafana/values.yaml
	var helmValues = pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesGrafana.Metadata.Name),
		"resources":        containerresources.ConvertToPulumiMap(locals.KubernetesGrafana.Spec.Container.Resources),
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
		},
		"adminUser":     pulumi.String("admin"),
		"adminPassword": pulumi.String("admin"),
		"persistence": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	//install grafana helm-chart
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesGrafana.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(grafanaHelmChartName),
			Version:   pulumi.String(grafanaHelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values:    helmValues,
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(grafanaHelmChartRepoUrl),
			},
		}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create helm chart")
	}

	return nil
}
