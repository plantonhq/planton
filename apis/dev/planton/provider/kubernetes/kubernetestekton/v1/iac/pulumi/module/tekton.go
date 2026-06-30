package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deployTektonPipelines deploys Tekton Pipelines using the official release manifest.
// This creates the tekton-pipelines namespace, CRDs, and all pipeline components.
func deployTektonPipelines(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider) (pulumi.Resource, error) {

	// Deploy Tekton Pipelines using ConfigFile
	// This applies the manifest from:
	// https://storage.googleapis.com/tekton-releases/pipeline/{version}/release.yaml
	pipelineManifests, err := yamlv2.NewConfigFile(ctx, "tekton-pipelines", &yamlv2.ConfigFileArgs{
		File: pulumi.String(locals.PipelineManifestURL),
	}, pulumi.Provider(kubernetesProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy Tekton Pipelines manifests")
	}

	return pipelineManifests, nil
}

// deployTektonDashboard deploys Tekton Dashboard using the official release manifest.
// This adds the web UI for viewing pipelines, tasks, and runs.
func deployTektonDashboard(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider, dependsOn pulumi.Resource) (pulumi.Resource, error) {

	// Deploy Tekton Dashboard using ConfigFile
	// This applies the manifest from:
	// https://infra.tekton.dev/tekton-releases/dashboard/{version}/release.yaml
	dashboardManifests, err := yamlv2.NewConfigFile(ctx, "tekton-dashboard", &yamlv2.ConfigFileArgs{
		File: pulumi.String(locals.DashboardManifestURL),
	}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn([]pulumi.Resource{dependsOn}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to deploy Tekton Dashboard manifests")
	}

	return dashboardManifests, nil
}
