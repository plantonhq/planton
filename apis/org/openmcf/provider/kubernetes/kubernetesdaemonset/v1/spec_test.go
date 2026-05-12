package kubernetesdaemonsetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestKubernetesDaemonSet(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesDaemonSet Suite")
}

var _ = ginkgo.Describe("KubernetesDaemonSet Custom Validation Tests", func() {
	var input *KubernetesDaemonSet

	ginkgo.BeforeEach(func() {
		input = &KubernetesDaemonSet{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesDaemonSet",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-daemonset",
			},
			Spec: &KubernetesDaemonSetSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesDaemonSetContainer{
					App: &KubernetesDaemonSetContainerApp{
						Image: &kubernetes.ContainerImage{
							Repo: "fluentd",
							Tag:  "latest",
						},
						Resources: &kubernetes.ContainerResources{
							Limits: &kubernetes.CpuMemory{
								Cpu:    "1000m",
								Memory: "1Gi",
							},
							Requests: &kubernetes.CpuMemory{
								Cpu:    "50m",
								Memory: "100Mi",
							},
						},
						Ports: []*KubernetesDaemonSetContainerAppPort{
							{
								Name:            "metrics",
								ContainerPort:   9090,
								NetworkProtocol: "TCP",
							},
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("daemonset with basic configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Namespace creation flag", func() {
		ginkgo.Context("When create_namespace is true", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.CreateNamespace = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When create_namespace is false", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.CreateNamespace = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Container image validation", func() {
		ginkgo.Context("When image repo is empty", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container.App.Image.Repo = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When image tag is empty", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container.App.Image.Tag = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Port validation", func() {
		ginkgo.Context("When port name contains uppercase", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container.App.Ports[0].Name = "HTTP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When port name starts with hyphen", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container.App.Ports[0].Name = "-metrics"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When network protocol is invalid", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container.App.Ports[0].NetworkProtocol = "HTTP"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Toleration validation", func() {
		ginkgo.Context("When toleration has valid operator", func() {
			ginkgo.It("should pass validation with Exists operator", func() {
				input.Spec.Tolerations = []*KubernetesDaemonSetToleration{
					{
						Key:      "node-role.kubernetes.io/master",
						Operator: "Exists",
						Effect:   "NoSchedule",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should pass validation with Equal operator", func() {
				input.Spec.Tolerations = []*KubernetesDaemonSetToleration{
					{
						Key:      "dedicated",
						Operator: "Equal",
						Value:    "logging",
						Effect:   "NoSchedule",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When toleration has invalid operator", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tolerations = []*KubernetesDaemonSetToleration{
					{
						Key:      "node-role.kubernetes.io/master",
						Operator: "Invalid",
						Effect:   "NoSchedule",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When toleration has invalid effect", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tolerations = []*KubernetesDaemonSetToleration{
					{
						Key:      "node-role.kubernetes.io/master",
						Operator: "Exists",
						Effect:   "InvalidEffect",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Update strategy validation", func() {
		ginkgo.Context("When update strategy is RollingUpdate", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.UpdateStrategy = &KubernetesDaemonSetUpdateStrategy{
					Type: "RollingUpdate",
					RollingUpdate: &KubernetesDaemonSetRollingUpdate{
						MaxUnavailable: "1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When update strategy is OnDelete", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.UpdateStrategy = &KubernetesDaemonSetUpdateStrategy{
					Type: "OnDelete",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When update strategy type is invalid", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.UpdateStrategy = &KubernetesDaemonSetUpdateStrategy{
					Type: "InvalidStrategy",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Node selector validation", func() {
		ginkgo.Context("When node selector is specified", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.NodeSelector = map[string]string{
					"kubernetes.io/os": "linux",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Volume mount validation", func() {
		ginkgo.Context("When volume mount with hostPath is specified", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.VolumeMounts = []*kubernetes.VolumeMount{
					{
						Name:      "varlog",
						MountPath: "/var/log",
						ReadOnly:  true,
						HostPath: &kubernetes.HostPathVolumeSource{
							Path: "/var/log",
							Type: "Directory",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When volume mount with configMap is specified", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.VolumeMounts = []*kubernetes.VolumeMount{
					{
						Name:      "config",
						MountPath: "/etc/app/config.yaml",
						ConfigMap: &kubernetes.ConfigMapVolumeSource{
							Name: "app-config",
							Key:  "config.yaml",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("RBAC validation", func() {
		ginkgo.Context("When RBAC with cluster rules is specified", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.CreateServiceAccount = true
				input.Spec.Rbac = &KubernetesDaemonSetRbac{
					ClusterRules: []*KubernetesDaemonSetRbacRule{
						{
							ApiGroups: []string{""},
							Resources: []string{"pods", "nodes"},
							Verbs:     []string{"get", "list", "watch"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When RBAC rule has empty api_groups", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CreateServiceAccount = true
				input.Spec.Rbac = &KubernetesDaemonSetRbac{
					ClusterRules: []*KubernetesDaemonSetRbacRule{
						{
							ApiGroups: []string{},
							Resources: []string{"pods"},
							Verbs:     []string{"get"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When RBAC rule has empty resources", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CreateServiceAccount = true
				input.Spec.Rbac = &KubernetesDaemonSetRbac{
					ClusterRules: []*KubernetesDaemonSetRbacRule{
						{
							ApiGroups: []string{""},
							Resources: []string{},
							Verbs:     []string{"get"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When RBAC rule has empty verbs", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CreateServiceAccount = true
				input.Spec.Rbac = &KubernetesDaemonSetRbac{
					ClusterRules: []*KubernetesDaemonSetRbacRule{
						{
							ApiGroups: []string{""},
							Resources: []string{"pods"},
							Verbs:     []string{},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("ConfigMaps validation", func() {
		ginkgo.Context("When configMaps is specified", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.ConfigMaps = map[string]string{
					"app-config": "key: value",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Environment secrets validation", func() {
		ginkgo.Context("When secrets have direct string values", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_Value{Value: "my-password"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When secrets have Kubernetes Secret references", func() {
			ginkgo.It("should pass validation with valid secret ref", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name: "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "my-app-secrets",
									Key:  "db-password",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When secrets have mixed types", func() {
			ginkgo.It("should pass validation with both string values and secret refs", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "DEBUG_TOKEN",
							Source: &kubernetes.SecretEnvVar_Value{Value: "debug-only-token"},
						},
						{
							Name: "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "postgres-credentials",
									Key:  "password",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When secret ref is missing required fields", func() {
			ginkgo.It("should fail validation when name is missing", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name: "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "",
									Key:  "db-password",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should fail validation when key is missing", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name: "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "my-app-secrets",
									Key:  "",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Environment variables validation", func() {
		ginkgo.Context("When variables have direct string values", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "LOG_LEVEL",
							Source: &kubernetes.EnvVar_Value{Value: "info"},
						},
						{
							Name:   "NODE_ENV",
							Source: &kubernetes.EnvVar_Value{Value: "production"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When variables have valueFrom references", func() {
			ginkgo.It("should pass validation with valid valueFrom ref", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name: "CLUSTER_NAME",
							Source: &kubernetes.EnvVar_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-cluster",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When variables have mixed types", func() {
			ginkgo.It("should pass validation with both direct values and valueFrom refs", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "LOG_LEVEL",
							Source: &kubernetes.EnvVar_Value{Value: "info"},
						},
						{
							Name: "CLUSTER_NAME",
							Source: &kubernetes.EnvVar_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-cluster",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When valueFrom ref is missing required name", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name: "CLUSTER_NAME",
							Source: &kubernetes.EnvVar_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Environment variable name validation", func() {
		ginkgo.Context("When env var name starts with a digit", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "1BAD_NAME",
							Source: &kubernetes.EnvVar_Value{Value: "value"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When env var name contains a hyphen", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "BAD-NAME",
							Source: &kubernetes.EnvVar_Value{Value: "value"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When secret env var name starts with a digit", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "2BAD_SECRET",
							Source: &kubernetes.SecretEnvVar_Value{Value: "secret"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When secret env var name contains a hyphen", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "BAD-SECRET",
							Source: &kubernetes.SecretEnvVar_Value{Value: "secret"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
