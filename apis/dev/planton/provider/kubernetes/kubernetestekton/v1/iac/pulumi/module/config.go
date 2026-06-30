package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// configureCloudEvents patches the config-defaults ConfigMap to set the cloud events sink URL.
// This enables Tekton to send CloudEvents for TaskRun and PipelineRun lifecycle events.
//
// The patch updates the 'default-cloud-events-sink' key in the config-defaults ConfigMap
// located in the tekton-pipelines namespace.
func configureCloudEvents(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider *kubernetes.Provider, dependsOn []pulumi.Resource) error {

	// Create/update the config-defaults ConfigMap with cloud events sink
	// We use a ConfigMap patch to add our configuration without replacing
	// the entire ConfigMap that was created by the Tekton manifests.
	_, err := corev1.NewConfigMapPatch(ctx, "tekton-config-defaults-patch", &corev1.ConfigMapPatchArgs{
		Metadata: metav1.ObjectMetaPatchArgs{
			Name:      pulumi.String("config-defaults"),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Data: pulumi.StringMap{
			"default-cloud-events-sink": pulumi.String(locals.CloudEventsSinkURL),
		},
	}, pulumi.Provider(kubernetesProvider), pulumi.DependsOn(dependsOn))

	if err != nil {
		return errors.Wrap(err, "failed to configure cloud events sink")
	}

	return nil
}
