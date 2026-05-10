package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/mergestringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func signoz(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// https://github.com/SigNoz/charts/blob/main/charts/signoz/values.yaml
	helmValues := pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesSignoz.Metadata.Name),
		"podLabels":        pulumi.ToStringMap(locals.KubernetesLabels),
		"commonLabels":     pulumi.ToStringMap(locals.KubernetesLabels),
	}

	// Configure SigNoz container (main binary with UI, API, Ruler, Alertmanager)
	if locals.KubernetesSignoz.Spec.SignozContainer != nil {
		signozValues := pulumi.Map{
			"replicaCount": pulumi.Int(int(locals.KubernetesSignoz.Spec.SignozContainer.Replicas)),
		}

		if locals.KubernetesSignoz.Spec.SignozContainer.Resources != nil {
			signozValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesSignoz.Spec.SignozContainer.Resources)
		}

		if locals.KubernetesSignoz.Spec.SignozContainer.Image != nil {
			signozValues["image"] = pulumi.Map{
				"repository": pulumi.String(locals.KubernetesSignoz.Spec.SignozContainer.Image.Repo),
				"tag":        pulumi.String(locals.KubernetesSignoz.Spec.SignozContainer.Image.Tag),
			}
		}

		helmValues["signoz"] = signozValues
	}

	// Configure OpenTelemetry Collector
	if locals.KubernetesSignoz.Spec.OtelCollectorContainer != nil {
		otelValues := pulumi.Map{
			"replicaCount": pulumi.Int(int(locals.KubernetesSignoz.Spec.OtelCollectorContainer.Replicas)),
		}

		if locals.KubernetesSignoz.Spec.OtelCollectorContainer.Resources != nil {
			otelValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesSignoz.Spec.OtelCollectorContainer.Resources)
		}

		if locals.KubernetesSignoz.Spec.OtelCollectorContainer.Image != nil {
			otelValues["image"] = pulumi.Map{
				"repository": pulumi.String(locals.KubernetesSignoz.Spec.OtelCollectorContainer.Image.Repo),
				"tag":        pulumi.String(locals.KubernetesSignoz.Spec.OtelCollectorContainer.Image.Tag),
			}
		}

		helmValues["otelCollector"] = otelValues
	}

	// Configure database (ClickHouse)
	if locals.KubernetesSignoz.Spec.Database != nil {
		if locals.KubernetesSignoz.Spec.Database.IsExternal {
			// External ClickHouse configuration
			if locals.KubernetesSignoz.Spec.Database.ExternalDatabase != nil {
				ext := locals.KubernetesSignoz.Spec.Database.ExternalDatabase
				helmValues["clickhouse"] = pulumi.Map{
					"enabled": pulumi.Bool(false),
				}
				externalClickhouseValues := pulumi.Map{
					"host":     pulumi.String(ext.Host),
					"httpPort": pulumi.Int(int(ext.GetHttpPort())),
					"tcpPort":  pulumi.Int(int(ext.GetTcpPort())),
					"cluster":  pulumi.String(ext.GetClusterName()),
					"secure":   pulumi.Bool(ext.IsSecure),
					"user":     pulumi.String(ext.Username),
				}

				// Handle password - either as plain string or from existing secret
				if ext.Password != nil {
					if ext.Password.GetSecretRef() != nil {
						// Use existing Kubernetes secret for password
						secretRef := ext.Password.GetSecretRef()
						externalClickhouseValues["existingSecret"] = pulumi.String(secretRef.Name)
						externalClickhouseValues["existingSecretPasswordKey"] = pulumi.String(secretRef.Key)
					} else if ext.Password.GetValue() != "" {
						// Use plain string password
						externalClickhouseValues["password"] = pulumi.String(ext.Password.GetValue())
					}
				}

				helmValues["externalClickhouse"] = externalClickhouseValues
			}
		} else {
			// Self-managed ClickHouse configuration
			if locals.KubernetesSignoz.Spec.Database.ManagedDatabase != nil {
				managed := locals.KubernetesSignoz.Spec.Database.ManagedDatabase
			clickhouseValues := pulumi.Map{
				"enabled": pulumi.Bool(true),
			}

				// ClickHouse container configuration
				if managed.Container != nil {
					clickhouseValues["replicaCount"] = pulumi.Int(int(managed.Container.Replicas))

					if managed.Container.Resources != nil {
						clickhouseValues["resources"] = containerresources.ConvertToPulumiMap(
							managed.Container.Resources)
					}

					if managed.Container.Image != nil {
						clickhouseValues["image"] = pulumi.Map{
							"registry":   pulumi.String("docker.io"),
							"repository": pulumi.String(managed.Container.Image.Repo),
							"tag":        pulumi.String(managed.Container.Image.Tag),
						}
					}

					clickhouseValues["persistence"] = pulumi.Map{
						"enabled": pulumi.Bool(managed.Container.PersistenceEnabled),
						"size":    pulumi.String(managed.Container.DiskSize),
					}
				}

				// ClickHouse clustering configuration
				if managed.Cluster != nil && managed.Cluster.IsEnabled {
					clickhouseValues["layout"] = pulumi.Map{
						"shardsCount":   pulumi.Int(int(managed.Cluster.ShardCount)),
						"replicasCount": pulumi.Int(int(managed.Cluster.ReplicaCount)),
					}
				}

				helmValues["clickhouse"] = clickhouseValues

				// Zookeeper configuration (required for distributed ClickHouse)
				// Note: In SigNoz, ZooKeeper settings must be configured under clickhouse.zookeeper
				if managed.Zookeeper != nil && managed.Zookeeper.IsEnabled {
					// Get existing clickhouse.zookeeper map or create new one
					zkConfig, hasZk := clickhouseValues["zookeeper"].(pulumi.Map)
					if !hasZk {
						zkConfig = pulumi.Map{}
					}

					zkConfig["enabled"] = pulumi.Bool(true)

					if managed.Zookeeper.Container != nil {
						zkConfig["replicaCount"] = pulumi.Int(int(managed.Zookeeper.Container.Replicas))

						if managed.Zookeeper.Container.Resources != nil {
							zkConfig["resources"] = containerresources.ConvertToPulumiMap(
								managed.Zookeeper.Container.Resources)
						}

						if managed.Zookeeper.Container.Image != nil {
							zkConfig["image"] = pulumi.Map{
								"registry":   pulumi.String("docker.io"),
								"repository": pulumi.String(managed.Zookeeper.Container.Image.Repo),
								"tag":        pulumi.String(managed.Zookeeper.Container.Image.Tag),
							}
						}

						zkConfig["persistence"] = pulumi.Map{
							"size": pulumi.String(managed.Zookeeper.Container.DiskSize),
						}
					}

					// Update the clickhouse.zookeeper configuration
					clickhouseValues["zookeeper"] = zkConfig
				} else {
					// Get existing clickhouse.zookeeper map or create new one
					zkConfig, hasZk := clickhouseValues["zookeeper"].(pulumi.Map)
					if !hasZk {
						zkConfig = pulumi.Map{}
					}
					// Explicitly disable Zookeeper if not needed
					zkConfig["enabled"] = pulumi.Bool(false)
					clickhouseValues["zookeeper"] = zkConfig
				}
			}
		}
	}

	// Note: Ingress is NOT configured via Helm chart values
	// We use Kubernetes Gateway API for ingress (see ingress.go)
	// This provides better control and consistency with other workloads

	// Merge custom Helm values
	mergestringmaps.MergeMapToPulumiMap(helmValues, locals.KubernetesSignoz.Spec.HelmValues)

	// Install SigNoz Helm chart
	chartOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	chartOpts = append(chartOpts, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesSignoz.Metadata.Name,
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
		return errors.Wrap(err, "failed to create signoz helm-chart")
	}

	return nil
}
