package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream Redis Helm chart and tailors it to the spec.
func helmChart(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// install helm-chart
	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesRedis.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(vars.HelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			//https://github.com/bitnami/charts/blob/main/bitnami/redis/values.yaml
			Values: pulumi.Map{
				"fullnameOverride": pulumi.String(locals.KubernetesRedis.Metadata.Name),
				"architecture":     pulumi.String("standalone"),
				"image": pulumi.Map{
					"registry":   pulumi.String(vars.RedisImageRegistry),
					"repository": pulumi.String(vars.RedisImageRepository),
					"tag":        pulumi.String(vars.RedisImageTag),
				},
				"master": pulumi.Map{
					"podLabels": convertstringmaps.ConvertGoStringMapToPulumiMap(locals.Labels),
					"resources": containerresources.ConvertToPulumiMap(locals.KubernetesRedis.Spec.Container.Resources),
					"persistence": pulumi.Map{
						"enabled": pulumi.Bool(locals.KubernetesRedis.Spec.Container.PersistenceEnabled),
						"size":    pulumi.String(locals.KubernetesRedis.Spec.Container.DiskSize),
					},
				},
				"replica": pulumi.Map{
					"podLabels":    convertstringmaps.ConvertGoStringMapToPulumiMap(locals.Labels),
					"replicaCount": pulumi.Int(locals.KubernetesRedis.Spec.Container.Replicas),
					"resources":    containerresources.ConvertToPulumiMap(locals.KubernetesRedis.Spec.Container.Resources),
					"persistence": pulumi.Map{
						"enabled": pulumi.Bool(locals.KubernetesRedis.Spec.Container.PersistenceEnabled),
						"size":    pulumi.String(locals.KubernetesRedis.Spec.Container.DiskSize),
					},
				},
				"auth": pulumi.Map{
					"existingSecret":            pulumi.String(locals.PasswordSecretName),
					"existingSecretPasswordKey": pulumi.String(vars.RedisPasswordSecretKey),
				},
			},
			// if you need to add the repository, you can specify `repo url`:
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(vars.HelmChartRepoUrl),
			},
		}, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create helm chart")
	}

	return nil
}
