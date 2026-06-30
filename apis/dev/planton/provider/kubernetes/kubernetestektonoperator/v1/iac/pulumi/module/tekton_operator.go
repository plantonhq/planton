package module

import (
	"github.com/pkg/errors"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tektonOperator installs the Tekton Operator using release manifests.
// Uses yaml/v2 for better CRD ordering and await behavior - the v2 implementation
// is designed to apply CRDs first and wait for reconciliation, preventing the
// "no matches for kind TektonConfig" error that occurs with v1.
func tektonOperator(ctx *pulumi.Context, locals *Locals,
	k8sProvider *pulumikubernetes.Provider) error {

	// --------------------------------------------------------------------
	// 1. Install Tekton Operator from release manifests
	// Using yaml/v2 which handles CRD registration timing properly
	// --------------------------------------------------------------------
	operatorManifests, err := yamlv2.NewConfigFile(ctx, "tekton-operator", &yamlv2.ConfigFileArgs{
		File: pulumi.String(locals.OperatorReleaseURL),
	}, pulumi.Provider(k8sProvider))
	if err != nil {
		return errors.Wrap(err, "install tekton operator manifests")
	}

	// --------------------------------------------------------------------
	// 2. Create TektonConfig to configure components
	// --------------------------------------------------------------------
	// Determine the profile based on enabled components
	profile := "lite" // minimal profile
	if locals.EnablePipelines && locals.EnableTriggers && locals.EnableDashboard {
		profile = "all"
	} else if locals.EnablePipelines && locals.EnableTriggers {
		profile = "basic"
	}

	// Build TektonConfig YAML dynamically
	tektonConfigYAML := buildTektonConfigYAML(locals, profile)

	// Using yaml/v2 ConfigGroup with explicit dependency on operator manifests
	// to ensure CRDs are registered before TektonConfig is created
	_, err = yamlv2.NewConfigGroup(ctx, "tekton-config", &yamlv2.ConfigGroupArgs{
		Yaml: pulumi.StringPtr(tektonConfigYAML),
	}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{operatorManifests}))
	if err != nil {
		return errors.Wrap(err, "create tekton config")
	}

	return nil
}

// buildTektonConfigYAML constructs the TektonConfig YAML based on component settings.
func buildTektonConfigYAML(locals *Locals, profile string) string {
	// TektonConfig CRD that tells the operator which components to install
	// Note: Do not set fields that the operator manages automatically (e.g., pipeline.enable-api-fields)
	// to avoid Server-Side Apply field conflicts
	yaml := `apiVersion: operator.tekton.dev/v1alpha1
kind: TektonConfig
metadata:
  name: ` + locals.TektonConfigName + `
spec:
  profile: ` + profile + `
  targetNamespace: ` + locals.ComponentsNamespace

	// Add cloud events sink if configured
	if locals.CloudEventsSinkURL != "" {
		yaml += `
  pipeline:
    default-cloud-events-sink: ` + locals.CloudEventsSinkURL
	}

	yaml += "\n"
	return yaml
}
