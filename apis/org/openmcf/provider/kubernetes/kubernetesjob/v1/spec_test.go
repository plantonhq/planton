package kubernetesjobv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestKubernetesJob(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesJob Suite")
}

var _ = ginkgo.Describe("KubernetesJob Custom Validation Tests", func() {
	var input *KubernetesJob

	ginkgo.BeforeEach(func() {
		input = &KubernetesJob{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesJob",
			Metadata: &shared.CloudResourceMetadata{
				Name: "my-batch-job",
			},
			Spec: &KubernetesJobSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				CreateNamespace: true,
				Image: &kubernetes.ContainerImage{
					Repo: "busybox",
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
				Env: &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "BATCH_SIZE",
							Source: &kubernetes.EnvVar_Value{Value: "1000"},
						},
					},
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name:   "API_KEY",
							Source: &kubernetes.SecretEnvVar_Value{Value: "secret_value"},
						},
					},
				},
				CompletionMode: proto.String("NonIndexed"),
				RestartPolicy:  proto.String("Never"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("kubernetes_job with create_namespace=true", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with create_namespace=false", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.CreateNamespace = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with restart_policy=OnFailure", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.RestartPolicy = proto.String("OnFailure")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with completion_mode=Indexed", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.CompletionMode = proto.String("Indexed")
				input.Spec.Completions = proto.Uint32(5)
				input.Spec.Parallelism = proto.Uint32(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with parallel execution", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Parallelism = proto.Uint32(10)
				input.Spec.Completions = proto.Uint32(100)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with active_deadline_seconds", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ActiveDeadlineSeconds = proto.Uint64(3600)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with ttl_seconds_after_finished", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TtlSecondsAfterFinished = proto.Uint32(86400)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("kubernetes_job with invalid restart_policy", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.RestartPolicy = proto.String("Always")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("kubernetes_job with invalid completion_mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.CompletionMode = proto.String("Invalid")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Environment secrets validation", func() {
		ginkgo.Context("When secrets have direct string values", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
					Secrets: []*kubernetes.SecretEnvVar{
						{
							Name: "DATABASE_PASSWORD",
							Source: &kubernetes.SecretEnvVar_SecretRef{
								SecretRef: &kubernetes.KubernetesSecretKeyRef{
									Name: "my-secret",
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
				input.Spec.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "INPUT_FILE",
							Source: &kubernetes.EnvVar_Value{Value: "/data/input.csv"},
						},
						{
							Name:   "OUTPUT_FILE",
							Source: &kubernetes.EnvVar_Value{Value: "/data/output.csv"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("When variables have valueFrom references", func() {
			ginkgo.It("should pass validation with valid valueFrom ref", func() {
				input.Spec.Env = &kubernetes.ContainerEnv{
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

		ginkgo.Context("When variables have mixed types", func() {
			ginkgo.It("should pass validation with both direct values and valueFrom refs", func() {
				input.Spec.Env = &kubernetes.ContainerEnv{
					Variables: []*kubernetes.EnvVar{
						{
							Name:   "BATCH_SIZE",
							Source: &kubernetes.EnvVar_Value{Value: "1000"},
						},
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

		ginkgo.Context("When valueFrom ref is missing required name", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Env = &kubernetes.ContainerEnv{
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
		ginkgo.Context("When env var name starts with a digit", func() {
			ginkgo.It("should fail validation", func() {
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
				input.Spec.Env = &kubernetes.ContainerEnv{
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
