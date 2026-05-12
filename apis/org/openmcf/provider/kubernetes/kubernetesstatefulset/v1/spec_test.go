package kubernetesstatefulsetv1

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

func TestKubernetesStatefulSet(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesStatefulSet Suite")
}

var _ = ginkgo.Describe("KubernetesStatefulSet Custom Validation Tests", func() {
	var input *KubernetesStatefulSet

	ginkgo.BeforeEach(func() {
		input = &KubernetesStatefulSet{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesStatefulSet",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-statefulset",
			},
			Spec: &KubernetesStatefulSetSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesStatefulSetContainer{
					App: &KubernetesStatefulSetContainerApp{
						Image: &kubernetes.ContainerImage{
							Repo: "postgres",
							Tag:  "15",
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
						Ports: []*KubernetesStatefulSetContainerAppPort{
							{
								Name:            "postgres",
								ContainerPort:   5432,
								ServicePort:     5432,
								NetworkProtocol: "TCP",
								AppProtocol:     "tcp",
							},
						},
						VolumeMounts: []*kubernetes.VolumeMount{
							{
								Name:      "data",
								MountPath: "/var/lib/postgresql/data",
								Pvc: &kubernetes.PvcVolumeSource{
									ClaimName: "data",
								},
							},
						},
					},
				},
				VolumeClaimTemplates: []*KubernetesStatefulSetVolumeClaimTemplate{
					{
						Name:        "data",
						Size:        "10Gi",
						AccessModes: []string{"ReadWriteOnce"},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("statefulset_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Ingress validation", func() {
		ginkgo.Context("When ingress is enabled without hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ingress = &KubernetesStatefulSetIngress{
					Enabled:  true,
					Hostname: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When ingress is disabled", func() {
			ginkgo.It("should not require hostname", func() {
				input.Spec.Ingress = &KubernetesStatefulSetIngress{
					Enabled:  false,
					Hostname: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When ingress is enabled with hostname", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Ingress = &KubernetesStatefulSetIngress{
					Enabled:  true,
					Hostname: "myapp.example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Volume claim template validation", func() {
		ginkgo.Context("When size is invalid", func() {
			ginkgo.It("should return a validation error for invalid size format", func() {
				input.Spec.VolumeClaimTemplates[0].Size = "invalid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When access mode is invalid", func() {
			ginkgo.It("should return a validation error for invalid access mode", func() {
				input.Spec.VolumeClaimTemplates[0].AccessModes = []string{"InvalidMode"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When access mode is valid", func() {
			ginkgo.It("should not return a validation error for ReadWriteMany", func() {
				input.Spec.VolumeClaimTemplates[0].AccessModes = []string{"ReadWriteMany"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Pod management policy validation", func() {
		ginkgo.Context("When pod management policy is OrderedReady", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.PodManagementPolicy = "OrderedReady"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When pod management policy is Parallel", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.PodManagementPolicy = "Parallel"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When pod management policy is invalid", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.PodManagementPolicy = "Invalid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When pod management policy is empty", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.PodManagementPolicy = ""
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

	ginkgo.Describe("Port validation", func() {
		ginkgo.Context("When port name is invalid", func() {
			ginkgo.It("should return a validation error for underscore in name", func() {
				input.Spec.Container.App.Ports[0].Name = "my_port"
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
									Key:  "password",
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
									Name: "my-secrets",
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
							Name:   "PGDATA",
							Source: &kubernetes.EnvVar_Value{Value: "/var/lib/postgresql/data"},
						},
						{
							Name:   "POSTGRES_PORT",
							Source: &kubernetes.EnvVar_Value{Value: "5432"},
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
							Name: "POSTGRES_HOST",
							Source: &kubernetes.EnvVar_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-postgres",
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
							Name:   "PGDATA",
							Source: &kubernetes.EnvVar_Value{Value: "/var/lib/postgresql/data"},
						},
						{
							Name: "POSTGRES_HOST",
							Source: &kubernetes.EnvVar_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-postgres",
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
							Name: "POSTGRES_HOST",
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
