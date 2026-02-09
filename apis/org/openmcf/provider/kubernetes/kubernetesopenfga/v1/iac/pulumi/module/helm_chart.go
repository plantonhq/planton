package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream OpenFGA Helm chart and tailors it to the spec.
func helmChart(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	ds := locals.KubernetesOpenFga.Spec.Datastore

	// Determine port - use default based on engine if not specified
	port := ds.GetPort()
	if port == 0 {
		if ds.Engine == "postgres" {
			port = 5432
		} else if ds.Engine == "mysql" {
			port = 3306
		}
	}

	// Build connection string options based on engine and is_secure flag
	connOptions := ""
	if ds.Engine == "postgres" && ds.IsSecure {
		connOptions = "?sslmode=require"
	} else if ds.Engine == "mysql" {
		// MySQL driver requires parseTime=true for proper time handling
		if ds.IsSecure {
			connOptions = "?parseTime=true&tls=true"
		} else {
			connOptions = "?parseTime=true"
		}
	}

	// https://github.com/openfga/helm-charts/blob/main/charts/openfga/values.yaml
	var helmValues = pulumi.Map{
		"fullnameOverride": pulumi.String(locals.KubernetesOpenFga.Metadata.Name),
		"replicaCount":     pulumi.Int(locals.KubernetesOpenFga.Spec.Container.Replicas),
		"resources":        containerresources.ConvertToPulumiMap(locals.KubernetesOpenFga.Spec.Container.Resources),
	}

	// Handle password - either as plain string or from existing secret
	if ds.Password != nil {
		if ds.Password.GetSecretRef() != nil {
			// Use existing Kubernetes secret for password
			// We construct the URI with a placeholder and use extraEnvVars to inject the actual password
			secretRef := ds.Password.GetSecretRef()

			// Construct URI with environment variable placeholder
			// The shell will expand $(OPENFGA_DATASTORE_PASSWORD) when the container starts
			datastoreUri := fmt.Sprintf("%s://%s:$(OPENFGA_DATASTORE_PASSWORD)@%s:%d/%s%s",
				ds.Engine,
				ds.Username,
				ds.Host,
				port,
				ds.Database,
				connOptions,
			)

			helmValues["datastore"] = pulumi.Map{
				"engine": pulumi.String(ds.Engine),
				"uri":    pulumi.String(datastoreUri),
			}

			// Add extraEnvVars to inject the password from secret
			helmValues["extraEnvVars"] = pulumi.Array{
				pulumi.Map{
					"name": pulumi.String("OPENFGA_DATASTORE_PASSWORD"),
					"valueFrom": pulumi.Map{
						"secretKeyRef": pulumi.Map{
							"name": pulumi.String(secretRef.Name),
							"key":  pulumi.String(secretRef.Key),
						},
					},
				},
			}
		} else if ds.Password.GetValue() != "" {
			// Use plain string password - construct full URI
			datastoreUri := fmt.Sprintf("%s://%s:%s@%s:%d/%s%s",
				ds.Engine,
				ds.Username,
				ds.Password.GetValue(),
				ds.Host,
				port,
				ds.Database,
				connOptions,
			)

			helmValues["datastore"] = pulumi.Map{
				"engine": pulumi.String(ds.Engine),
				"uri":    pulumi.String(datastoreUri),
			}
		}
	}

	// Install openfga helm-chart
	chartOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	chartOpts = append(chartOpts, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesOpenFga.Metadata.Name,
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
