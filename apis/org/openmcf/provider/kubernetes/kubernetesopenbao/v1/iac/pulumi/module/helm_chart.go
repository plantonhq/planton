package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerresources"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the upstream OpenBao Helm chart and tailors it to the spec.
func helmChart(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {

	target := locals.KubernetesOpenBao
	spec := target.Spec

	// Build helm values based on spec
	helmValues := pulumi.Map{
		"fullnameOverride": pulumi.String(target.Metadata.Name),
		"global": pulumi.Map{
			"enabled":    pulumi.Bool(true),
			"tlsDisable": pulumi.Bool(!spec.TlsEnabled),
		},
		"server": pulumi.Map{
			"extraLabels": convertstringmaps.ConvertGoStringMapToPulumiMap(locals.Labels),
			"dataStorage": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"size":    pulumi.String(spec.ServerContainer.DataStorageSize),
			},
		},
	}

	// Configure server resources if provided
	if spec.ServerContainer != nil && spec.ServerContainer.Resources != nil {
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["resources"] = containerresources.ConvertToPulumiMap(spec.ServerContainer.Resources)
	}

	// Configure standalone vs HA mode
	if locals.HaEnabled {
		// HA mode with Raft
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["ha"] = pulumi.Map{
			"enabled":  pulumi.Bool(true),
			"replicas": pulumi.Int(locals.HaReplicas),
			"raft": pulumi.Map{
				"enabled":   pulumi.Bool(true),
				"setNodeId": pulumi.Bool(true),
				"config": pulumi.String(`ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "raft" {
  path = "/openbao/data"
}

service_registration "kubernetes" {}
`),
			},
		}
		serverMap["standalone"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	} else {
		// Standalone mode
		serverMap := helmValues["server"].(pulumi.Map)
		serverMap["standalone"] = pulumi.Map{
			"enabled": pulumi.Bool(true),
			"config": pulumi.String(`ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "file" {
  path = "/openbao/data"
}
`),
		}
		serverMap["ha"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	}

	// Configure UI
	uiEnabled := true
	if spec.UiEnabled != nil {
		uiEnabled = *spec.UiEnabled
	}
	helmValues["ui"] = pulumi.Map{
		"enabled": pulumi.Bool(uiEnabled),
	}

	// Configure injector if enabled
	if spec.Injector != nil && spec.Injector.Enabled {
		injectorReplicas := int32(1)
		if spec.Injector.Replicas != nil {
			injectorReplicas = *spec.Injector.Replicas
		}
		helmValues["injector"] = pulumi.Map{
			"enabled":  pulumi.Bool(true),
			"replicas": pulumi.Int(injectorReplicas),
		}
	} else {
		helmValues["injector"] = pulumi.Map{
			"enabled": pulumi.Bool(false),
		}
	}

	// Install helm chart
	chartOpts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	chartOpts = append(chartOpts, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesOpenBao.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.HelmChartName),
			Version:   pulumi.String(locals.HelmChartVersion),
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
