package kubernetesdeploymentv1

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

func TestKubernetesDeployment(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesDeployment Suite")
}

var _ = ginkgo.Describe("KubernetesDeployment Custom Validation Tests", func() {
	var input *KubernetesDeployment

	ginkgo.BeforeEach(func() {
		input = &KubernetesDeployment{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesDeployment",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-deployment",
			},
			Spec: &KubernetesDeploymentSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Version: "main",
				Container: &KubernetesDeploymentContainer{
					App: &KubernetesDeploymentContainerApp{
						Image: &kubernetes.ContainerImage{
							Repo: "nginx",
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
						Ports: []*KubernetesDeploymentContainerAppPort{
							{
								Name:            "http",
								ContainerPort:   8080,
								ServicePort:     80,
								NetworkProtocol: "TCP",
								AppProtocol:     "http",
								IsIngressPort:   true,
							},
						},
					},
				},
				Ingress: &KubernetesDeploymentIngress{
					Enabled:  true,
					Hostname: "myapp.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("deployment_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Ingress validation", func() {
		ginkgo.Context("When ingress is enabled without hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ingress.Hostname = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When ingress is disabled", func() {
			ginkgo.It("should not require hostname", func() {
				input.Spec.Ingress.Enabled = false
				input.Spec.Ingress.Hostname = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Version validation", func() {
		ginkgo.Context("When version contains uppercase", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Version = "Main"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When version ends with hyphen", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Version = "main-"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
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
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "DATABASE_NAME",
							Source: &kubernetes.EnvVar_Value{Value: "mydb"},
						},
					},
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_Value{Value: "my-password"},
						},
						{
							Name: "API_KEY",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "external-secrets",
									Key:  "api-key",
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
							Name:   "DATABASE_HOST",
							Source: &kubernetes.EnvVar_Value{Value: "localhost"},
						},
						{
							Name:   "DATABASE_PORT",
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
							Name: "DATABASE_HOST",
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

		ginkgo.Context("When variables have configMapKeyRef", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name: "APP_SETTING",
							Source: &kubernetes.EnvVar_ConfigMapKeyRef{
								ConfigMapKeyRef: &kubernetes.ConfigMapKeyRef{
									Name: "app-config",
									Key:  "setting-key",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When variables have fieldRef", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name: "POD_NAME",
							Source: &kubernetes.EnvVar_FieldRef{
								FieldRef: &kubernetes.ObjectFieldRef{
									FieldPath: "metadata.name",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When variables have resourceFieldRef", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name: "CPU_LIMIT",
							Source: &kubernetes.EnvVar_ResourceFieldRef{
								ResourceFieldRef: &kubernetes.ResourceFieldRef{
									Resource: "limits.cpu",
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
							Name: "DATABASE_HOST",
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
		ginkgo.Context("When name starts with a digit", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "1INVALID",
							Source: &kubernetes.EnvVar_Value{Value: "test"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When name contains a hyphen", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "INVALID-NAME",
							Source: &kubernetes.EnvVar_Value{Value: "test"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When name is empty", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "",
							Source: &kubernetes.EnvVar_Value{Value: "test"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When secret name starts with a digit", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "1BAD_SECRET",
							Source: &kubernetes.SecretEnvVar_Value{Value: "test"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("When name starts with underscore", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "_VALID_NAME",
							Source: &kubernetes.EnvVar_Value{Value: "test"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("EnvFrom validation", func() {
		ginkgo.Context("When envFrom has configMapRef", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					EnvFrom: []*kubernetes.EnvFromSource{
						{
							Source: &kubernetes.EnvFromSource_ConfigMapRef{
								ConfigMapRef: &kubernetes.ConfigMapRef{
									Name: "app-config",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When envFrom has secretRef with prefix", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					EnvFrom: []*kubernetes.EnvFromSource{
						{
							Prefix: "SECRET_",
							Source: &kubernetes.EnvFromSource_SecretRef{
								SecretRef: &kubernetes.SecretRef{
									Name: "app-secrets",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When envFrom configMapRef name is missing", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Container.App.Env = &kubernetes.ContainerEnv{
					EnvFrom: []*kubernetes.EnvFromSource{
						{
							Source: &kubernetes.EnvFromSource_ConfigMapRef{
								ConfigMapRef: &kubernetes.ConfigMapRef{
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
})
