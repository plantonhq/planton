package module

import (
	"fmt"

	"github.com/pkg/errors"
	kubernetesmongodbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesmongodb/v1"
	psmdbv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/perconamongodb/kubernetes/psmdb/v1"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// mongodb creates a PerconaServerMongoDB custom resource using the Percona operator
func mongodb(
	ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	createdSecret *kubernetescorev1.Secret,
	namespaceDeps []pulumi.ResourceOption,
) error {
	spec := locals.KubernetesMongodb.Spec

	// Determine replica set size from container replicas
	replicaSetSize := int(spec.Container.Replicas)
	if replicaSetSize < 1 {
		replicaSetSize = 1
	}

	// Build the PerconaServerMongoDB CRD
	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	_, err := psmdbv1.NewPerconaServerMongoDB(ctx,
		locals.KubernetesMongodb.Metadata.Name,
		&psmdbv1.PerconaServerMongoDBArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KubernetesMongodb.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			Spec: &psmdbv1.PerconaServerMongoDBSpecArgs{
				CrVersion: pulumi.String(vars.CRVersion),
				Image:     pulumi.String(fmt.Sprintf("percona/percona-server-mongodb:%s", vars.MongoDBVersion)),
				Replsets:  buildReplicaSets(spec, replicaSetSize),
				Secrets:   buildSecrets(createdSecret),
				// Allow replica sets with less than 3 members for dev/test environments
				UnsafeFlags: &psmdbv1.PerconaServerMongoDBSpecUnsafeFlagsArgs{
					ReplsetSize: pulumi.Bool(true),
				},
			},
		}, opts...,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create PerconaServerMongoDB")
	}

	return nil
}

// buildReplicaSets creates the replica set configuration
func buildReplicaSets(
	spec *kubernetesmongodbv1.KubernetesMongodbSpec,
	replicaSetSize int,
) psmdbv1.PerconaServerMongoDBSpecReplsetsArray {
	replset := &psmdbv1.PerconaServerMongoDBSpecReplsetsArgs{
		Name: pulumi.String(vars.ReplicaSetName),
		Size: pulumi.Int(replicaSetSize),
	}

	// Configure resources
	replset.Resources = buildResources(spec.Container)

	// Configure persistence if enabled
	if spec.Container.PersistenceEnabled && spec.Container.DiskSize != "" {
		replset.VolumeSpec = buildVolumeSpec(spec.Container.DiskSize)
	}

	return psmdbv1.PerconaServerMongoDBSpecReplsetsArray{replset}
}

// buildResources maps container resources to Percona format
func buildResources(container *kubernetesmongodbv1.KubernetesMongodbContainer) *psmdbv1.PerconaServerMongoDBSpecReplsetsResourcesArgs {
	if container == nil || container.Resources == nil {
		return nil
	}

	return &psmdbv1.PerconaServerMongoDBSpecReplsetsResourcesArgs{
		Limits: pulumi.Map{
			"cpu":    pulumi.String(container.Resources.Limits.Cpu),
			"memory": pulumi.String(container.Resources.Limits.Memory),
		},
		Requests: pulumi.Map{
			"cpu":    pulumi.String(container.Resources.Requests.Cpu),
			"memory": pulumi.String(container.Resources.Requests.Memory),
		},
	}
}

// buildVolumeSpec creates the persistent volume configuration
func buildVolumeSpec(diskSize string) *psmdbv1.PerconaServerMongoDBSpecReplsetsVolumeSpecArgs {
	return &psmdbv1.PerconaServerMongoDBSpecReplsetsVolumeSpecArgs{
		PersistentVolumeClaim: &psmdbv1.PerconaServerMongoDBSpecReplsetsVolumeSpecPersistentVolumeClaimArgs{
			Resources: &psmdbv1.PerconaServerMongoDBSpecReplsetsVolumeSpecPersistentVolumeClaimResourcesArgs{
				Requests: pulumi.Map{
					"storage": pulumi.String(diskSize),
				},
			},
		},
	}
}

// buildSecrets creates the secrets reference for MongoDB authentication
func buildSecrets(createdSecret *kubernetescorev1.Secret) *psmdbv1.PerconaServerMongoDBSpecSecretsArgs {
	return &psmdbv1.PerconaServerMongoDBSpecSecretsArgs{
		Users: createdSecret.Metadata.Name(),
	}
}
