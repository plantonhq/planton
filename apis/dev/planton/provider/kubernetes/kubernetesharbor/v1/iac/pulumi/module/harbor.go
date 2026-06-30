package module

import (
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesharbor/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func harbor(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	// https://github.com/goharbor/harbor-helm/blob/main/values.yaml
	helmValues := pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesHarbor.Metadata.Name),
		"commonLabels":     pulumi.ToStringMap(locals.KubernetesLabels),
	}

	// Configure Harbor Core container
	if locals.KubernetesHarbor.Spec.CoreContainer != nil {
		coreValues := pulumi.Map{
			"replicas": pulumi.Int(int(locals.KubernetesHarbor.Spec.CoreContainer.Replicas)),
		}

		if locals.KubernetesHarbor.Spec.CoreContainer.Resources != nil {
			coreValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesHarbor.Spec.CoreContainer.Resources)
		}

		if locals.KubernetesHarbor.Spec.CoreContainer.Image != nil {
			coreValues["image"] = pulumi.Map{
				"repository": pulumi.String(locals.KubernetesHarbor.Spec.CoreContainer.Image.Repo),
				"tag":        pulumi.String(locals.KubernetesHarbor.Spec.CoreContainer.Image.Tag),
			}
		}

		helmValues["core"] = coreValues
	}

	// Configure Harbor Portal container
	if locals.KubernetesHarbor.Spec.PortalContainer != nil {
		portalValues := pulumi.Map{
			"replicas": pulumi.Int(int(locals.KubernetesHarbor.Spec.PortalContainer.Replicas)),
		}

		if locals.KubernetesHarbor.Spec.PortalContainer.Resources != nil {
			portalValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesHarbor.Spec.PortalContainer.Resources)
		}

		helmValues["portal"] = portalValues
	}

	// Configure Harbor Registry container
	if locals.KubernetesHarbor.Spec.RegistryContainer != nil {
		registryValues := pulumi.Map{
			"replicas": pulumi.Int(int(locals.KubernetesHarbor.Spec.RegistryContainer.Replicas)),
		}

		if locals.KubernetesHarbor.Spec.RegistryContainer.Resources != nil {
			registryValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesHarbor.Spec.RegistryContainer.Resources)
		}

		helmValues["registry"] = registryValues
	}

	// Configure Harbor Jobservice container
	if locals.KubernetesHarbor.Spec.JobserviceContainer != nil {
		jobserviceValues := pulumi.Map{
			"replicas": pulumi.Int(int(locals.KubernetesHarbor.Spec.JobserviceContainer.Replicas)),
		}

		if locals.KubernetesHarbor.Spec.JobserviceContainer.Resources != nil {
			jobserviceValues["resources"] = containerresources.ConvertToPulumiMap(
				locals.KubernetesHarbor.Spec.JobserviceContainer.Resources)
		}

		helmValues["jobservice"] = jobserviceValues
	}

	// Configure database (PostgreSQL)
	if locals.KubernetesHarbor.Spec.Database != nil {
		if locals.KubernetesHarbor.Spec.Database.IsExternal {
			// External PostgreSQL configuration
			if locals.KubernetesHarbor.Spec.Database.ExternalDatabase != nil {
				ext := locals.KubernetesHarbor.Spec.Database.ExternalDatabase
				helmValues["database"] = pulumi.Map{
					"type": pulumi.String("external"),
					"external": pulumi.Map{
						"host":                 pulumi.String(ext.Host),
						"port":                 pulumi.String(ext.GetPort()),
						"username":             pulumi.String(ext.Username),
						"password":             pulumi.String(ext.Password),
						"coreDatabase":         pulumi.String(ext.GetCoreDatabase()),
						"clairDatabase":        pulumi.String(ext.GetClairDatabase()),
						"notaryServerDatabase": pulumi.String(ext.GetNotaryServerDatabase()),
						"notarySignerDatabase": pulumi.String(ext.GetNotarySignerDatabase()),
						"sslmode":              pulumi.String(map[bool]string{true: "require", false: "disable"}[ext.UseSsl]),
					},
				}
				// Disable internal PostgreSQL
				helmValues["postgresql"] = pulumi.Map{
					"enabled": pulumi.Bool(false),
				}
			}
		} else {
			// Self-managed PostgreSQL configuration
			if locals.KubernetesHarbor.Spec.Database.ManagedDatabase != nil {
				managed := locals.KubernetesHarbor.Spec.Database.ManagedDatabase
				postgresValues := pulumi.Map{
					"enabled": pulumi.Bool(true),
				}

				if managed.Container != nil {
					if managed.Container.Resources != nil {
						postgresValues["resources"] = containerresources.ConvertToPulumiMap(
							managed.Container.Resources)
					}

					if managed.Container.PersistenceEnabled {
						postgresValues["persistence"] = pulumi.Map{
							"enabled": pulumi.Bool(true),
							"size":    pulumi.String(managed.Container.DiskSize),
						}
					} else {
						postgresValues["persistence"] = pulumi.Map{
							"enabled": pulumi.Bool(false),
						}
					}
				}

				helmValues["postgresql"] = postgresValues
			}
		}
	}

	// Configure cache (Redis)
	if locals.KubernetesHarbor.Spec.Cache != nil {
		if locals.KubernetesHarbor.Spec.Cache.IsExternal {
			// External Redis configuration
			if locals.KubernetesHarbor.Spec.Cache.ExternalCache != nil {
				ext := locals.KubernetesHarbor.Spec.Cache.ExternalCache
				redisConfig := pulumi.Map{
					"type": pulumi.String("external"),
					"external": pulumi.Map{
						"addr":     pulumi.String(ext.Host),
						"password": pulumi.String(ext.Password),
					},
				}

				if ext.UseSentinel {
					redisConfig["external"].(pulumi.Map)["sentinelMasterSet"] = pulumi.String(ext.SentinelMasterSet)
				}

				helmValues["redis"] = redisConfig

				// Disable internal Redis
				helmValues["redis"].(pulumi.Map)["internal"] = pulumi.Map{
					"enabled": pulumi.Bool(false),
				}
			}
		} else {
			// Self-managed Redis configuration
			if locals.KubernetesHarbor.Spec.Cache.ManagedCache != nil {
				managed := locals.KubernetesHarbor.Spec.Cache.ManagedCache
				redisValues := pulumi.Map{
					"type": pulumi.String("internal"),
					"internal": pulumi.Map{
						"enabled": pulumi.Bool(true),
					},
				}

				if managed.Container != nil {
					if managed.Container.Resources != nil {
						redisValues["internal"].(pulumi.Map)["resources"] = containerresources.ConvertToPulumiMap(
							managed.Container.Resources)
					}
				}

				helmValues["redis"] = redisValues
			}
		}
	}

	// Configure storage
	if locals.KubernetesHarbor.Spec.Storage != nil {
		storageConfig := pulumi.Map{}

		switch locals.KubernetesHarbor.Spec.Storage.Type {
		case kubernetesharborv1.KubernetesHarborStorageType_s3:
			if locals.KubernetesHarbor.Spec.Storage.S3 != nil {
				s3 := locals.KubernetesHarbor.Spec.Storage.S3
				storageConfig = pulumi.Map{
					"type": pulumi.String("s3"),
					"s3": pulumi.Map{
						"bucket":         pulumi.String(s3.Bucket),
						"region":         pulumi.String(s3.Region),
						"accesskey":      pulumi.String(s3.AccessKey),
						"secretkey":      pulumi.String(s3.SecretKey),
						"regionendpoint": pulumi.String(s3.EndpointUrl),
						"encrypt":        pulumi.Bool(s3.Encrypt),
						"secure":         pulumi.Bool(s3.Secure),
					},
				}
			}
		case kubernetesharborv1.KubernetesHarborStorageType_gcs:
			if locals.KubernetesHarbor.Spec.Storage.Gcs != nil {
				gcs := locals.KubernetesHarbor.Spec.Storage.Gcs
				storageConfig = pulumi.Map{
					"type": pulumi.String("gcs"),
					"gcs": pulumi.Map{
						"bucket":    pulumi.String(gcs.Bucket),
						"keydata":   pulumi.String(gcs.KeyData),
						"chunksize": pulumi.Int(int(gcs.GetChunkSize())),
					},
				}
			}
		case kubernetesharborv1.KubernetesHarborStorageType_azure:
			if locals.KubernetesHarbor.Spec.Storage.Azure != nil {
				azure := locals.KubernetesHarbor.Spec.Storage.Azure
				storageConfig = pulumi.Map{
					"type": pulumi.String("azure"),
					"azure": pulumi.Map{
						"accountname": pulumi.String(azure.AccountName),
						"accountkey":  pulumi.String(azure.AccountKey),
						"container":   pulumi.String(azure.Container),
					},
				}
			}
		case kubernetesharborv1.KubernetesHarborStorageType_oss:
			if locals.KubernetesHarbor.Spec.Storage.Oss != nil {
				oss := locals.KubernetesHarbor.Spec.Storage.Oss
				storageConfig = pulumi.Map{
					"type": pulumi.String("oss"),
					"oss": pulumi.Map{
						"bucket":          pulumi.String(oss.Bucket),
						"endpoint":        pulumi.String(oss.Endpoint),
						"accesskeyid":     pulumi.String(oss.AccessKeyId),
						"accesskeysecret": pulumi.String(oss.AccessKeySecret),
						"secure":          pulumi.Bool(oss.Secure),
					},
				}
			}
		case kubernetesharborv1.KubernetesHarborStorageType_filesystem:
			if locals.KubernetesHarbor.Spec.Storage.Filesystem != nil {
				fs := locals.KubernetesHarbor.Spec.Storage.Filesystem
				storageConfig = pulumi.Map{
					"type": pulumi.String("filesystem"),
					"filesystem": pulumi.Map{
						"size": pulumi.String(fs.DiskSize),
					},
				}
			}
		}

		helmValues["persistence"] = pulumi.Map{
			"imageChartStorage": storageConfig,
		}
	}

	// Disable ingress in Helm chart (we manage it separately using Gateway API)
	helmValues["expose"] = pulumi.Map{
		"type": pulumi.String("clusterIP"),
	}

	// Merge with custom helm values if provided
	if locals.KubernetesHarbor.Spec.HelmValues != nil {
		for k, v := range locals.KubernetesHarbor.Spec.HelmValues {
			helmValues[k] = pulumi.String(v)
		}
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)

	// Deploy Harbor using Helm chart
	_, err := helmv3.NewRelease(ctx,
		locals.KubernetesHarbor.Metadata.Name,
		&helmv3.ReleaseArgs{
			Name:      pulumi.String(locals.KubernetesHarbor.Metadata.Name),
			Namespace: pulumi.String(locals.Namespace),
			Chart:     pulumi.String("harbor"),
			RepositoryOpts: helmv3.RepositoryOptsArgs{
				Repo: pulumi.String("https://helm.goharbor.io"),
			},
			Values: helmValues,
		},
		opts...,
	)

	return err
}
