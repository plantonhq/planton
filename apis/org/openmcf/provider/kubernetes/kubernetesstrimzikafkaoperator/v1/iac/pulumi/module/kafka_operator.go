package module

import (
	"github.com/pkg/errors"
	kubernetesstrimzikafkaoperatorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesstrimzikafkaoperator/v1"
	pulumikubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// kafkaOperator installs the Strimzi Kafka Operator on the target cluster.
//
// The function:
//  1. Deploys the Helm chart (watch‑any‑namespace=true so one install can
//     manage topics/streams across all namespaces).
//  2. Exports the namespace name so other stacks can import it later.
func kafkaOperator(
	ctx *pulumi.Context,
	target *kubernetesstrimzikafkaoperatorv1.KubernetesStrimziKafkaOperator,
	l *locals,
	kubernetesProvider *pulumikubernetes.Provider,
	namespaceDeps []pulumi.ResourceOption,
) error {
	// ---------------------------------------------------------------------
	// 1. Helm release
	// ---------------------------------------------------------------------
	helmOpts := append([]pulumi.ResourceOption{
		pulumi.Provider(kubernetesProvider),
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}, namespaceDeps...)
	_, err := helm.NewRelease(
		ctx,
		l.HelmReleaseName,
		&helm.ReleaseArgs{
			Name:            pulumi.String(l.HelmReleaseName),
			Namespace:       pulumi.String(l.namespace),
			Chart:           pulumi.String(vars.HelmChartName),
			Version:         pulumi.String(vars.HelmChartVersion),
			CreateNamespace: pulumi.Bool(false),
			Atomic:          pulumi.Bool(false),
			CleanupOnFail:   pulumi.Bool(true),
			WaitForJobs:     pulumi.Bool(true),
			Timeout:         pulumi.Int(180),
			Values: pulumi.Map{
				"watchAnyNamespace": pulumi.Bool(true),
			},
			RepositoryOpts: helm.RepositoryOptsArgs{
				Repo: pulumi.String(vars.HelmChartRepo),
			},
		},
		helmOpts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Strimzi Helm release")
	}

	// ---------------------------------------------------------------------
	// 2. Stack output
	// ---------------------------------------------------------------------
	ctx.Export(OpNamespace, pulumi.String(l.namespace))

	return nil
}
