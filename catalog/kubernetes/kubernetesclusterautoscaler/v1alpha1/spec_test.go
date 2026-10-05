package kubernetesclusterautoscalerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesClusterAutoscaler(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClusterAutoscaler Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

// awsArm returns the recommended AWS posture (tag-based auto-discovery) —
// the base every test starts from before swapping in the arm under test.
func awsArm() *KubernetesClusterAutoscalerSpec_Aws {
	return &KubernetesClusterAutoscalerSpec_Aws{
		Aws: &KubernetesClusterAutoscalerAws{
			Region: "us-west-2",
			AutoDiscovery: &KubernetesClusterAutoscalerAwsAutoDiscovery{
				ClusterName: "prod-eks",
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesClusterAutoscaler Validation Tests", func() {
	var input *KubernetesClusterAutoscaler

	ginkgo.BeforeEach(func() {
		input = &KubernetesClusterAutoscaler{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesClusterAutoscaler",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-cluster-autoscaler",
			},
			Spec: &KubernetesClusterAutoscalerSpec{
				Namespace: literal("kube-system"),
				Cloud:     awsArm(),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal aws spec with auto-discovery should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("aws auto-discovery with IRSA should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					AutoDiscovery: &KubernetesClusterAutoscalerAwsAutoDiscovery{
						ClusterName: "prod-eks",
						Tags:        []string{"k8s.io/cluster-autoscaler/enabled"},
					},
					IrsaRoleArn: "arn:aws:iam::123456789012:role/cluster-autoscaler",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("aws static node groups with access keys should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "workers-a", MinSize: 1, MaxSize: 10},
						{Name: "workers-b", MinSize: 0, MaxSize: 5},
					},
					AccessKeys: &KubernetesClusterAutoscalerAwsAccessKeys{
						AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
						SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure auto-discovery with workload identity should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						UseWorkloadIdentity: true,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure static scale sets with managed identity and UAID should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "vmss-workers", MinSize: 1, MaxSize: 20},
					},
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						UseManagedIdentity:     true,
						UserAssignedIdentityId: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ca",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("azure service-principal identity should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						ServicePrincipal: &KubernetesClusterAutoscalerAzureServicePrincipal{
							TenantId:     "66666666-7777-8888-9999-000000000000",
							ClientId:     "11111111-2222-3333-4444-555555555555",
							ClientSecret: "sp-secret-value",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gce instance-group prefixes with workload identity email should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Gce{
				Gce: &KubernetesClusterAutoscalerGce{
					InstanceGroupPrefixes: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "k8s-node-group-", MinSize: 0, MaxSize: 10},
					},
					WorkloadIdentityServiceAccountEmail: "cluster-autoscaler@my-project.iam.gserviceaccount.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cluster-api default mode should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_ClusterApi{
				ClusterApi: &KubernetesClusterAutoscalerClusterApi{},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cluster-api kubeconfig mode with secret should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_ClusterApi{
				ClusterApi: &KubernetesClusterAutoscalerClusterApi{
					Mode:                stringPtr("kubeconfig-incluster"),
					KubeconfigSecret:    "workload-kubeconfig",
					Namespace:           "capi-machines",
					NamespaceScopedRbac: true,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("civo arm with all credentials should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Civo{
				Civo: &KubernetesClusterAutoscalerCivo{
					ClusterId: "8f6b2c1a",
					Region:    "LON1",
					ApiKey:    "civo-api-key",
					ApiUrl:    stringPtr("https://api.civo.com"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kwok arm with default config map should be valid", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Kwok{
				Kwok: &KubernetesClusterAutoscalerKwok{},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("fully tuned scaling block should be valid", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				Expander:                  "priority,least-waste",
				BalanceSimilarNodeGroups:  true,
				ScanInterval:              "30s",
				MaxNodeProvisionTime:      "15m",
				SkipNodesWithLocalStorage: boolPtr(false),
				SkipNodesWithSystemPods:   boolPtr(true),
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{
					Enabled:              boolPtr(true),
					UtilizationThreshold: "0.5",
					UnneededTime:         "10m",
					DelayAfterAdd:        "10m",
					DelayAfterDelete:     "0s",
					DelayAfterFailure:    "3m",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("well-formed extra_args keys should be valid", func() {
			input.Spec.ExtraArgs = map[string]string{
				"max-graceful-termination-sec": "600",
				"v":                            "2",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("deployment sizing and prometheus telemetry should be valid", func() {
			input.Spec.Deployment = &KubernetesClusterAutoscalerDeployment{
				Replicas:          int32Ptr(2),
				PriorityClassName: stringPtr("system-cluster-critical"),
				NodeSelector:      map[string]string{"kubernetes.io/os": "linux"},
			}
			input.Spec.Prometheus = &KubernetesClusterAutoscalerPrometheus{
				ServiceMonitor:                true,
				ServiceMonitorSelectorRelease: "kube-prometheus-stack",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("no cloud arm should fail (oneof required)", func() {
			input.Spec.Cloud = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("aws with both auto_discovery and node_groups should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					AutoDiscovery: &KubernetesClusterAutoscalerAwsAutoDiscovery{
						ClusterName: "prod-eks",
					},
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "workers", MinSize: 1, MaxSize: 5},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("aws with neither auto_discovery nor node_groups should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{Region: "us-west-2"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("aws with both irsa_role_arn and access_keys should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					AutoDiscovery: &KubernetesClusterAutoscalerAwsAutoDiscovery{
						ClusterName: "prod-eks",
					},
					IrsaRoleArn: "arn:aws:iam::123456789012:role/cluster-autoscaler",
					AccessKeys: &KubernetesClusterAutoscalerAwsAccessKeys{
						AccessKeyId:     "AKIAIOSFODNN7EXAMPLE",
						SecretAccessKey: "secret",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed IRSA role ARN should fail", func() {
			input.Spec.Cloud.(*KubernetesClusterAutoscalerSpec_Aws).Aws.IrsaRoleArn = "role/cluster-autoscaler"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("aws access keys without secret_access_key should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "workers", MinSize: 1, MaxSize: 5},
					},
					AccessKeys: &KubernetesClusterAutoscalerAwsAccessKeys{
						AccessKeyId: "AKIAIOSFODNN7EXAMPLE",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure with both cluster_name and node_groups should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "vmss-workers", MinSize: 1, MaxSize: 20},
					},
					Identity: &KubernetesClusterAutoscalerAzureIdentity{UseWorkloadIdentity: true},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure with neither cluster_name nor node_groups should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					Identity:       &KubernetesClusterAutoscalerAzureIdentity{UseWorkloadIdentity: true},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure without identity should fail (required)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure identity with no posture selected should fail (exactly one)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity:       &KubernetesClusterAutoscalerAzureIdentity{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure identity with two postures selected should fail (exactly one)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						UseWorkloadIdentity: true,
						UseManagedIdentity:  true,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("user_assigned_identity_id without use_managed_identity should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						UseWorkloadIdentity:    true,
						UserAssignedIdentityId: "/subscriptions/s/rg/providers/msi/ca",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("azure service principal without client_secret should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Azure{
				Azure: &KubernetesClusterAutoscalerAzure{
					SubscriptionId: "00000000-1111-2222-3333-444444444444",
					ResourceGroup:  "MC_rg_cluster_westus2",
					ClusterName:    "prod-aks",
					Identity: &KubernetesClusterAutoscalerAzureIdentity{
						ServicePrincipal: &KubernetesClusterAutoscalerAzureServicePrincipal{
							TenantId: "66666666-7777-8888-9999-000000000000",
							ClientId: "11111111-2222-3333-4444-555555555555",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("gce with no instance group prefixes should fail (min_items)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Gce{
				Gce: &KubernetesClusterAutoscalerGce{},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed GCP workload identity email should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Gce{
				Gce: &KubernetesClusterAutoscalerGce{
					InstanceGroupPrefixes: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "k8s-node-group-", MinSize: 0, MaxSize: 10},
					},
					WorkloadIdentityServiceAccountEmail: "autoscaler@example.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown cluster-api mode should fail (closed enum)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_ClusterApi{
				ClusterApi: &KubernetesClusterAutoscalerClusterApi{
					Mode: stringPtr("incluster-magic"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("cluster-api kubeconfig mode without kubeconfig_secret should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_ClusterApi{
				ClusterApi: &KubernetesClusterAutoscalerClusterApi{
					Mode: stringPtr("kubeconfig-kubeconfig"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("node group with min_size greater than max_size should fail", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "workers", MinSize: 5, MaxSize: 2},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("node group with zero max_size should fail (gte 1)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Aws{
				Aws: &KubernetesClusterAutoscalerAws{
					Region: "us-west-2",
					NodeGroups: []*KubernetesClusterAutoscalerNodeGroup{
						{Name: "workers", MinSize: 0, MaxSize: 0},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown expander should fail (closed enum list)", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{Expander: "cheapest"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("expander list with one bad entry should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{Expander: "least-waste,cheapest"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed scan_interval should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{ScanInterval: "10 seconds"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed max_node_provision_time should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{MaxNodeProvisionTime: "15"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("scale-down utilization threshold above 1 should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{UtilizationThreshold: "1.5"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed scale-down unneeded_time should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{UnneededTime: "ten-minutes"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed scale-down delay_after_add should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{DelayAfterAdd: "10min"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed scale-down delay_after_delete should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{DelayAfterDelete: "0"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed scale-down delay_after_failure should fail", func() {
			input.Spec.Scaling = &KubernetesClusterAutoscalerScaling{
				ScaleDown: &KubernetesClusterAutoscalerScaleDown{DelayAfterFailure: "3 m"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("extra_args key with leading dashes should fail (flag shape)", func() {
			input.Spec.ExtraArgs = map[string]string{"--expander": "least-waste"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("extra_args key with uppercase should fail (flag shape)", func() {
			input.Spec.ExtraArgs = map[string]string{"scanInterval": "10s"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("civo without api_key should fail (required)", func() {
			input.Spec.Cloud = &KubernetesClusterAutoscalerSpec_Civo{
				Civo: &KubernetesClusterAutoscalerCivo{
					ClusterId: "8f6b2c1a",
					Region:    "LON1",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero deployment replicas should fail (gte 1)", func() {
			input.Spec.Deployment = &KubernetesClusterAutoscalerDeployment{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
