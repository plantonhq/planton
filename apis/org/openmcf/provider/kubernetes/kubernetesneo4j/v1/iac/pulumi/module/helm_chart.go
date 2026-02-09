package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
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

	// optional external LB
	externalSvc := pulumi.Map{
		"enabled": pulumi.Bool(ingressEnabled),
	}
	if ingressEnabled {
		externalSvc["type"] = pulumi.String("LoadBalancer")
		externalSvc["annotations"] = pulumi.StringMap{
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

					// let the chart create its own secret + password
					// (no passwordFromSecret / passwordKey provided)

					"resources": pulumi.Map{
						"cpu":    pulumi.String(container.Resources.Limits.Cpu),
						"memory": pulumi.String(container.Resources.Limits.Memory),
					},
					"acceptLicenseAgreement": pulumi.String("yes"),
				},

				"externalService": externalSvc,

				// persistence
				"volumes": pulumi.Map{
					"data": pulumi.Map{
						"mode": pulumi.String("defaultStorageClass"),
						"size": pulumi.String(container.DiskSize),
					},
				},

				// neo4j.conf overrides
				"config": pulumi.Map{
					"server.memory.heap.initial_size": pulumi.String(locals.KubernetesNeo4J.Spec.MemoryConfig.HeapMax),
					"server.memory.pagecache.size":    pulumi.String(locals.KubernetesNeo4J.Spec.MemoryConfig.PageCache),
				},

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
