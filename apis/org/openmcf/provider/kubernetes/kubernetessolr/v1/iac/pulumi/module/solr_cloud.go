package module

import (
	"github.com/pkg/errors"
	kubernetessolrv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetessolr/v1"
	"github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/solroperator/kubernetes/solr/v1beta1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func solrCloud(ctx *pulumi.Context, locals *Locals,
	kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	//create solr-operator's solrcloud resource
	_, err := v1beta1.NewSolrCloud(ctx, "solr-cloud",
		&v1beta1.SolrCloudArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.KubernetesSolr.Metadata.Name),
				Namespace: pulumi.String(locals.Namespace),
				Labels:    pulumi.ToStringMap(locals.Labels),
				// CRITICAL: Background Deletion Propagation Policy
				//
				// This annotation is required to prevent a 10-minute timeout during `pulumi destroy`.
				//
				// Problem: By default, Pulumi uses "Foreground" cascading deletion, which keeps the
				// SolrCloud CR alive (with a foregroundDeletion finalizer) while waiting for all child
				// resources to be deleted first. However, the Solr Operator's reconciliation loop
				// continues to run and recreates any child resources (StatefulSets, Services, ConfigMaps)
				// that the Kubernetes Garbage Collector deletes. This creates an infinite loop where:
				//   1. GC deletes child resources
				//   2. Operator recreates them (because the SolrCloud CR still exists)
				//   3. GC deletes them again
				//   4. Repeat until timeout
				//
				// Solution: Using "background" propagation policy causes Pulumi to delete the SolrCloud
				// CR immediately. Once the CR is gone, the Solr Operator stops reconciling, and the
				// Kubernetes Garbage Collector can clean up all child resources (ZookeeperCluster,
				// StatefulSets, Services, PVCs, etc.) asynchronously without interference.
				//
				// Reference: https://www.pulumi.com/registry/packages/kubernetes/installation-configuration/
				Annotations: pulumi.StringMap{
					"pulumi.com/deletionPropagationPolicy": pulumi.String("background"),
				},
			},
			Spec: v1beta1.SolrCloudSpecArgs{
				Replicas: pulumi.Int(locals.KubernetesSolr.Spec.SolrContainer.Replicas),
				SolrImage: v1beta1.SolrCloudSpecSolrImageArgs{
					Repository: pulumi.String(locals.KubernetesSolr.Spec.SolrContainer.Image.Repo),
					Tag:        pulumi.String(locals.KubernetesSolr.Spec.SolrContainer.Image.Tag),
				},
			SolrJavaMem: pulumi.String(solrConfigJavaMem(locals.KubernetesSolr.Spec)),
			SolrOpts:    pulumi.String(solrConfigOpts(locals.KubernetesSolr.Spec)),
			SolrGCTune:  pulumi.String(solrConfigGCTune(locals.KubernetesSolr.Spec)),
				SolrModules: pulumi.ToStringArray(vars.SolrCloudSolrModules),
				CustomSolrKubeOptions: v1beta1.SolrCloudSpecCustomSolrKubeOptionsArgs{
					PodOptions: v1beta1.SolrCloudSpecCustomSolrKubeOptionsPodOptionsArgs{
						Resources: v1beta1.SolrCloudSpecCustomSolrKubeOptionsPodOptionsResourcesArgs{
							Limits: pulumi.ToMap(map[string]interface{}{
								//"cpu":    locals.KubernetesSolr.Spec.SolrContainer.Resources.Limits.Cpu,
								"memory": locals.KubernetesSolr.Spec.SolrContainer.Resources.Limits.Memory,
							}),
							Requests: pulumi.ToMap(map[string]interface{}{
								"cpu":    locals.KubernetesSolr.Spec.SolrContainer.Resources.Requests.Cpu,
								"memory": locals.KubernetesSolr.Spec.SolrContainer.Resources.Requests.Memory,
							}),
						},
					},
				},
				DataStorage: v1beta1.SolrCloudSpecDataStorageArgs{
					Ephemeral: nil,
					Persistent: v1beta1.SolrCloudSpecDataStoragePersistentArgs{
						ReclaimPolicy: pulumi.String("Delete"),
						PvcTemplate: v1beta1.SolrCloudSpecDataStoragePersistentPvcTemplateArgs{
							Spec: v1beta1.SolrCloudSpecDataStoragePersistentPvcTemplateSpecArgs{
								Resources: v1beta1.SolrCloudSpecDataStoragePersistentPvcTemplateSpecResourcesArgs{
									Requests: pulumi.ToMap(map[string]interface{}{
										"storage": locals.KubernetesSolr.Spec.SolrContainer.DiskSize,
									}),
								},
							},
						},
					},
				},
				ZookeeperRef: v1beta1.SolrCloudSpecZookeeperRefArgs{
					Provided: v1beta1.SolrCloudSpecZookeeperRefProvidedArgs{
						Replicas: pulumi.Int(locals.KubernetesSolr.Spec.ZookeeperContainer.Replicas),
						Persistence: v1beta1.SolrCloudSpecZookeeperRefProvidedPersistenceArgs{
							Spec: v1beta1.SolrCloudSpecZookeeperRefProvidedPersistenceSpecArgs{
								Resources: v1beta1.SolrCloudSpecZookeeperRefProvidedPersistenceSpecResourcesArgs{
									Requests: pulumi.Map{
										"storage": pulumi.String(locals.KubernetesSolr.Spec.ZookeeperContainer.DiskSize),
									},
								},
							},
						},
						ZookeeperPodPolicy: v1beta1.SolrCloudSpecZookeeperRefProvidedZookeeperPodPolicyArgs{
							Resources: v1beta1.SolrCloudSpecZookeeperRefProvidedZookeeperPodPolicyResourcesArgs{
								Limits: pulumi.ToMap(map[string]interface{}{
									//"cpu":    locals.KubernetesSolr.Spec.ZookeeperContainer.Resources.Limits.Cpu,
									"memory": locals.KubernetesSolr.Spec.ZookeeperContainer.Resources.Limits.Memory,
								}),
								Requests: pulumi.Map{
									"cpu":    pulumi.String(locals.KubernetesSolr.Spec.ZookeeperContainer.Resources.Requests.Cpu),
									"memory": pulumi.String(locals.KubernetesSolr.Spec.ZookeeperContainer.Resources.Requests.Memory),
								},
							},
						},
					},
				},
			},
		}, append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create solr-cloud resource")
	}
	return nil
}

func solrConfigJavaMem(spec *kubernetessolrv1.KubernetesSolrSpec) string {
	if spec.Config == nil {
		return ""
	}
	return spec.Config.JavaMem
}

func solrConfigOpts(spec *kubernetessolrv1.KubernetesSolrSpec) string {
	if spec.Config == nil {
		return ""
	}
	return spec.Config.Opts
}

func solrConfigGCTune(spec *kubernetessolrv1.KubernetesSolrSpec) string {
	if spec.Config == nil {
		return ""
	}
	return spec.Config.GarbageCollectionTuning
}
