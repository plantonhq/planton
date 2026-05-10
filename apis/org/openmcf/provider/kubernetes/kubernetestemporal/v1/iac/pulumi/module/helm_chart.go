package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetestemporalv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetestemporal/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildResourcesMap converts ContainerResources proto to a pulumi.Map for Helm values
func buildResourcesMap(resources *kubernetes.ContainerResources) pulumi.Map {
	result := pulumi.Map{}

	if resources.Limits != nil {
		limits := pulumi.Map{}
		if resources.Limits.Cpu != "" {
			limits["cpu"] = pulumi.String(resources.Limits.Cpu)
		}
		if resources.Limits.Memory != "" {
			limits["memory"] = pulumi.String(resources.Limits.Memory)
		}
		if len(limits) > 0 {
			result["limits"] = limits
		}
	}

	if resources.Requests != nil {
		requests := pulumi.Map{}
		if resources.Requests.Cpu != "" {
			requests["cpu"] = pulumi.String(resources.Requests.Cpu)
		}
		if resources.Requests.Memory != "" {
			requests["memory"] = pulumi.String(resources.Requests.Memory)
		}
		if len(requests) > 0 {
			result["requests"] = requests
		}
	}

	return result
}

func helmChart(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	values := pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesTemporal.Metadata.Name),
	}

	// ---------------------------------------------------------------- database
	db := locals.KubernetesTemporal.Spec.Database

	if db.ExternalDatabase != nil {
		ext := db.ExternalDatabase

		// disable embedded datastores
		values["cassandra"] = pulumi.Map{"enabled": pulumi.Bool(false)}
		values["mysql"] = pulumi.Map{"enabled": pulumi.Bool(false)}
		values["postgresql"] = pulumi.Map{"enabled": pulumi.Bool(false)}

		// choose sub-driver
		subDriver := "postgres12"
		if db.Backend == kubernetestemporalv1.KubernetesTemporalDatabaseBackend_mysql {
			subDriver = "mysql8"
		}

		// default DB names if not provided
		defaultDB := db.GetDatabaseName()
		visibilityDB := db.GetVisibilityName()

		// username field
		user := ext.Username
		if user == "" {
			user = ext.Username
		}

		// common TLS section (SSL on, host-verification off)
		tls := pulumi.Map{
			"enabled":                pulumi.Bool(true),
			"enableHostVerification": pulumi.Bool(false),
		}

		// Determine which secret to use for database password
		// If secretRef is provided, use the existing secret; otherwise, use the secret we created
		dbSecretName := locals.DatabasePasswordSecretName
		dbSecretKey := vars.DatabasePasswordSecretKey
		if ext.Password != nil && ext.Password.GetSecretRef() != nil {
			secretRef := ext.Password.GetSecretRef()
			dbSecretName = secretRef.Name
			dbSecretKey = secretRef.Key
		}

		defaultSql := pulumi.Map{
			"driver": pulumi.String("sql"),
			"sql": pulumi.Map{
				"driver":            pulumi.String(subDriver),
				"host":              pulumi.String(ext.Host),
				"port":              pulumi.Int(ext.Port),
				"database":          pulumi.String(defaultDB),
				"user":              pulumi.String(user),
				"existingSecret":    pulumi.String(dbSecretName),
				"existingSecretKey": pulumi.String(dbSecretKey),
				"tls":               tls,
			},
		}

		visibilitySql := pulumi.Map{
			"driver": pulumi.String("sql"),
			"sql": pulumi.Map{
				"driver":            pulumi.String(subDriver),
				"host":              pulumi.String(ext.Host),
				"port":              pulumi.Int(ext.Port),
				"database":          pulumi.String(visibilityDB),
				"user":              pulumi.String(user),
				"existingSecret":    pulumi.String(dbSecretName),
				"existingSecretKey": pulumi.String(dbSecretKey),
				"tls":               tls,
			},
		}

		values["server"] = pulumi.Map{
			"config": pulumi.Map{
				"services": pulumi.Map{
					"frontend": pulumi.Map{
						"rpc": pulumi.Map{
							"grpcPort": pulumi.Int(vars.FrontendGrpcPort),
							"httpPort": pulumi.Int(vars.FrontendHttpPort),
						},
					},
				},
				"persistence": pulumi.Map{
					"default":    defaultSql,
					"visibility": visibilitySql,
					"driver":     pulumi.String("sql"),
				},
			},
		}
	} else {
		// embedded datastore paths
		switch db.Backend {
		case kubernetestemporalv1.KubernetesTemporalDatabaseBackend_cassandra:
			values["cassandra"] = pulumi.Map{
				"enabled":      pulumi.Bool(true),
				"replicaCount": pulumi.Int(1),
				"config": pulumi.Map{
					"dev":          pulumi.Bool(true),
					"cluster_size": pulumi.Int(1),
				},
			}
			values["mysql"] = pulumi.Map{"enabled": pulumi.Bool(false)}
			values["postgresql"] = pulumi.Map{"enabled": pulumi.Bool(false)}

		case kubernetestemporalv1.KubernetesTemporalDatabaseBackend_mysql:
			values["cassandra"] = pulumi.Map{"enabled": pulumi.Bool(false)}
			values["mysql"] = pulumi.Map{"enabled": pulumi.Bool(true)}
			values["postgresql"] = pulumi.Map{"enabled": pulumi.Bool(false)}

			defaultDB := db.GetDatabaseName()
			visibilityDB := db.GetVisibilityName()
			values["server"] = pulumi.Map{
				"config": pulumi.Map{
					"persistence": pulumi.Map{
						"driver": pulumi.String("sql"),
						"default": pulumi.Map{
							"driver": pulumi.String("sql"),
							"sql": pulumi.Map{
								"driver":   pulumi.String("mysql8"),
								"host":     pulumi.String(locals.KubernetesTemporal.Metadata.Name + "-mysql"),
								"port":     pulumi.Int(3306),
								"database": pulumi.String(defaultDB),
								"user":     pulumi.String("root"),
							},
						},
						"visibility": pulumi.Map{
							"driver": pulumi.String("sql"),
							"sql": pulumi.Map{
								"driver":   pulumi.String("mysql8"),
								"host":     pulumi.String(locals.KubernetesTemporal.Metadata.Name + "-mysql"),
								"port":     pulumi.Int(3306),
								"database": pulumi.String(visibilityDB),
								"user":     pulumi.String("root"),
							},
						},
					},
				},
			}

		case kubernetestemporalv1.KubernetesTemporalDatabaseBackend_postgresql:
			values["cassandra"] = pulumi.Map{"enabled": pulumi.Bool(false)}
			values["mysql"] = pulumi.Map{"enabled": pulumi.Bool(false)}
			values["postgresql"] = pulumi.Map{"enabled": pulumi.Bool(true)}

			defaultDB := db.GetDatabaseName()
			visibilityDB := db.GetVisibilityName()
			values["server"] = pulumi.Map{
				"config": pulumi.Map{
					"persistence": pulumi.Map{
						"driver": pulumi.String("sql"),
						"default": pulumi.Map{
							"driver": pulumi.String("sql"),
							"sql": pulumi.Map{
								"driver":   pulumi.String("postgres12"),
								"host":     pulumi.String(locals.KubernetesTemporal.Metadata.Name + "-postgresql"),
								"port":     pulumi.Int(5432),
								"database": pulumi.String(defaultDB),
								"user":     pulumi.String("temporal"),
							},
						},
						"visibility": pulumi.Map{
							"driver": pulumi.String("sql"),
							"sql": pulumi.Map{
								"driver":   pulumi.String("postgres12"),
								"host":     pulumi.String(locals.KubernetesTemporal.Metadata.Name + "-postgresql"),
								"port":     pulumi.Int(5432),
								"database": pulumi.String(visibilityDB),
								"user":     pulumi.String("temporal"),
							},
						},
					},
				},
			}
		}
	}

	// ------------------------------------------------------ schema jobs
	values["schema"] = pulumi.Map{
		"createDatabase": pulumi.Map{
			"enabled": pulumi.Bool(!db.DisableAutoSchemaSetup), // true by default
		},
		"setup":  pulumi.Map{"enabled": pulumi.Bool(true)},
		"update": pulumi.Map{"enabled": pulumi.Bool(true)},
	}

	// -------------------------------------------------------------- web-UI
	if locals.KubernetesTemporal.Spec.DisableWebUi {
		values["web"] = pulumi.Map{"enabled": pulumi.Bool(false)}
	}

	// ---------------------------------------------------- monitoring stack
	monitoring := locals.KubernetesTemporal.Spec.EnableMonitoringStack
	if locals.KubernetesTemporal.Spec.ExternalElasticsearch != nil &&
		locals.KubernetesTemporal.Spec.ExternalElasticsearch.Host != "" {
		monitoring = true
	}
	values["prometheus"] = pulumi.Map{"enabled": pulumi.Bool(monitoring)}
	values["grafana"] = pulumi.Map{"enabled": pulumi.Bool(monitoring)}
	values["kubePrometheusStack"] = pulumi.Map{"enabled": pulumi.Bool(monitoring)}

	// -------------------------------------------------------- elasticsearch
	es := locals.KubernetesTemporal.Spec.ExternalElasticsearch
	if es != nil && es.Host != "" {
		esValues := pulumi.Map{
			"enabled":  pulumi.Bool(false),
			"host":     pulumi.String(es.Host),
			"port":     pulumi.Int(es.Port),
			"scheme":   pulumi.String("http"),
			"username": pulumi.String(es.User),
		}

		// Handle password - either as plain string or from existing secret
		if es.Password != nil {
			if es.Password.GetSecretRef() != nil {
				// Use existing Kubernetes secret for password
				secretRef := es.Password.GetSecretRef()
				esValues["existingSecret"] = pulumi.String(secretRef.Name)
				esValues["existingSecretKey"] = pulumi.String(secretRef.Key)
			} else if es.Password.GetValue() != "" {
				// Use plain string password
				esValues["password"] = pulumi.String(es.Password.GetValue())
			}
		}

		values["elasticsearch"] = esValues
	} else if !locals.KubernetesTemporal.Spec.EnableEmbeddedElasticsearch {
		values["elasticsearch"] = pulumi.Map{"enabled": pulumi.Bool(false)}
	}

	// ----------------------------------------------------------- version
	// determine which version to use: spec.version if provided, otherwise default
	chartVersion := vars.HelmChartVersion
	if locals.KubernetesTemporal.Spec.Version != "" {
		chartVersion = locals.KubernetesTemporal.Spec.Version
	}

	// ----------------------------------------------- dynamic configuration
	// Configure runtime behavior settings like history limits
	// Note: The Temporal Helm chart expects server.dynamicConfig (not server.config.dynamicConfigValues)
	// See: https://github.com/temporalio/helm-charts/blob/main/charts/temporal/templates/server-dynamicconfigmap.yaml
	dynamicConfig := locals.KubernetesTemporal.Spec.DynamicConfig
	if dynamicConfig != nil {
		dynamicConfigValues := pulumi.Map{}

		if dynamicConfig.HistorySizeLimitError != nil {
			dynamicConfigValues["limit.historySize.error"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.HistorySizeLimitError)},
			}
		}

		if dynamicConfig.HistoryCountLimitError != nil {
			dynamicConfigValues["limit.historyCount.error"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.HistoryCountLimitError)},
			}
		}

		if dynamicConfig.HistorySizeLimitWarn != nil {
			dynamicConfigValues["limit.historySize.warn"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.HistorySizeLimitWarn)},
			}
		}

		if dynamicConfig.HistoryCountLimitWarn != nil {
			dynamicConfigValues["limit.historyCount.warn"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.HistoryCountLimitWarn)},
			}
		}

		// Blob size limits - controls individual payload sizes (markers, signals, activity I/O)
		// This is different from history limits which control total workflow history size
		if dynamicConfig.BlobSizeLimitError != nil {
			dynamicConfigValues["limit.blobSize.error"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.BlobSizeLimitError)},
			}
		}

		if dynamicConfig.BlobSizeLimitWarn != nil {
			dynamicConfigValues["limit.blobSize.warn"] = pulumi.Array{
				pulumi.Map{"value": pulumi.Int(*dynamicConfig.BlobSizeLimitWarn)},
			}
		}

		// Only add dynamic config if there are values to set
		// Chart expects server.dynamicConfig, NOT server.config.dynamicConfigValues
		if len(dynamicConfigValues) > 0 {
			if _, ok := values["server"]; !ok {
				values["server"] = pulumi.Map{
					"dynamicConfig": dynamicConfigValues,
				}
			} else {
				serverMap := values["server"].(pulumi.Map)
				serverMap["dynamicConfig"] = dynamicConfigValues
			}
		}
	}

	// -------------------------------------------------- num history shards
	// Configure the number of history shards (immutable after initial deployment)
	if locals.KubernetesTemporal.Spec.NumHistoryShards != nil {
		numShards := *locals.KubernetesTemporal.Spec.NumHistoryShards
		if _, ok := values["server"]; !ok {
			values["server"] = pulumi.Map{
				"numHistoryShards": pulumi.Int(numShards),
			}
		} else {
			serverMap := values["server"].(pulumi.Map)
			serverMap["numHistoryShards"] = pulumi.Int(numShards)
		}
	}

	// ----------------------------------------- service-level configuration
	// Configure replicas and resources for each Temporal service
	services := locals.KubernetesTemporal.Spec.Services
	if services != nil {
		// Ensure server map exists
		if _, ok := values["server"]; !ok {
			values["server"] = pulumi.Map{}
		}
		serverMap := values["server"].(pulumi.Map)

		// Frontend service configuration
		if services.Frontend != nil {
			frontendConfig := pulumi.Map{}
			if services.Frontend.Replicas != nil {
				frontendConfig["replicaCount"] = pulumi.Int(*services.Frontend.Replicas)
			}
			if services.Frontend.Resources != nil {
				frontendConfig["resources"] = buildResourcesMap(services.Frontend.Resources)
			}
			if len(frontendConfig) > 0 {
				serverMap["frontend"] = frontendConfig
			}
		}

		// History service configuration
		if services.History != nil {
			historyConfig := pulumi.Map{}
			if services.History.Replicas != nil {
				historyConfig["replicaCount"] = pulumi.Int(*services.History.Replicas)
			}
			if services.History.Resources != nil {
				historyConfig["resources"] = buildResourcesMap(services.History.Resources)
			}
			if len(historyConfig) > 0 {
				serverMap["history"] = historyConfig
			}
		}

		// Matching service configuration
		if services.Matching != nil {
			matchingConfig := pulumi.Map{}
			if services.Matching.Replicas != nil {
				matchingConfig["replicaCount"] = pulumi.Int(*services.Matching.Replicas)
			}
			if services.Matching.Resources != nil {
				matchingConfig["resources"] = buildResourcesMap(services.Matching.Resources)
			}
			if len(matchingConfig) > 0 {
				serverMap["matching"] = matchingConfig
			}
		}

		// Worker service configuration
		if services.Worker != nil {
			workerConfig := pulumi.Map{}
			if services.Worker.Replicas != nil {
				workerConfig["replicaCount"] = pulumi.Int(*services.Worker.Replicas)
			}
			if services.Worker.Resources != nil {
				workerConfig["resources"] = buildResourcesMap(services.Worker.Resources)
			}
			if len(workerConfig) > 0 {
				serverMap["worker"] = workerConfig
			}
		}
	}

	// ------------------------------------------------------- install chart
	chartOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	chartOpts = append(chartOpts, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesTemporal.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(chartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values:    values,
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(vars.HelmChartRepoUrl),
			},
		}, chartOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create temporal helm chart")
	}

	return nil
}
