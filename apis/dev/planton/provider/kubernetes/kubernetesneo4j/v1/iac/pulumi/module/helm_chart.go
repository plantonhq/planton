package module

import (
	"github.com/pkg/errors"
	kubernetesneo4jv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesneo4j/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// helmChart installs the Neo4j Helm chart with values derived from the spec.
func helmChart(
	ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	namespaceDeps []pulumi.ResourceOption,
) error {
	container := locals.KubernetesNeo4J.Spec.Container

	// honour ingress settings
	ingressEnabled := locals.KubernetesNeo4J.Spec.Ingress != nil &&
		locals.KubernetesNeo4J.Spec.Ingress.Enabled &&
		locals.KubernetesNeo4J.Spec.Ingress.Hostname != ""

	// Build the services.neo4j block that controls the chart's LoadBalancer
	// service. The chart (2025.03.0) defaults to services.neo4j.enabled: true
	// with type: LoadBalancer. We must explicitly disable it when ingress is off
	// to prevent an unprovisionable LB on clusters without a cloud LB controller.
	neo4jSvc := pulumi.Map{
		"enabled": pulumi.Bool(ingressEnabled),
	}
	if ingressEnabled {
		neo4jSvc["spec"] = pulumi.Map{
			"type": pulumi.String("LoadBalancer"),
		}
		neo4jSvc["annotations"] = pulumi.StringMap{
			"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressExternalHostname),
		}
	}

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := helmv3.NewChart(ctx,
		locals.KubernetesNeo4J.Metadata.Name,
		helmv3.ChartArgs{
			Chart:     pulumi.String(vars.Neo4jHelmChartName),
			Version:   pulumi.String(vars.Neo4jHelmChartVersion),
			Namespace: pulumi.String(locals.Namespace),
			Values: pulumi.Map{
				"neo4j": pulumi.Map{
					"name": pulumi.String(locals.KubernetesNeo4J.Metadata.Name),
					"resources": pulumi.Map{
						"cpu":    pulumi.String(container.Resources.Limits.Cpu),
						"memory": pulumi.String(container.Resources.Limits.Memory),
					},
					"acceptLicenseAgreement": pulumi.String("yes"),
				},

				"services": pulumi.Map{
					"neo4j": neo4jSvc,
				},

				"volumes": pulumi.Map{
					"data": pulumi.Map{
						"mode": pulumi.String("defaultStorageClass"),
						"size": pulumi.String(container.DiskSize),
					},
				},

				"config": memoryConfigValues(locals.KubernetesNeo4J.Spec.MemoryConfig),

				"podLabels": convertstringmaps.ConvertGoStringMapToPulumiMap(locals.Labels),
			},
			FetchArgs: helmv3.FetchArgs{
				Repo: pulumi.String(vars.Neo4jHelmChartRepoUrl),
			},
		},
		opts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to deploy neo4j helm chart")
	}

	// ---------------------------------------------------------------------
	// Export outputs
	// ---------------------------------------------------------------------
	ctx.Export(OpUsername, pulumi.String("neo4j"))

	// The Helm chart creates: <release>-auth with key "neo4j-password"
	// Using locals.PasswordSecretName to ensure consistency and avoid conflicts
	ctx.Export(OpPasswordSecretName, pulumi.String(locals.PasswordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.Neo4jPasswordSecretKey))

	return nil
}

// memoryConfigValues returns Helm config map values for Neo4j memory settings.
// Returns an empty map when memory_config is nil, allowing Neo4j to use its
// internal auto-detection defaults.
func memoryConfigValues(mc *kubernetesneo4jv1.KubernetesNeo4JMemoryConfig) pulumi.Map {
	if mc == nil {
		return pulumi.Map{}
	}
	m := pulumi.Map{}
	if mc.HeapMax != "" {
		m["server.memory.heap.initial_size"] = pulumi.String(mc.HeapMax)
	}
	if mc.PageCache != "" {
		m["server.memory.pagecache.size"] = pulumi.String(mc.PageCache)
	}
	return m
}
