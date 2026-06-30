package kubernetesgharunnerscalesetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesGhaRunnerScaleSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGhaRunnerScaleSetSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesGhaRunnerScaleSetSpec Validation Tests", func() {
	var spec *KubernetesGhaRunnerScaleSetSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesGhaRunnerScaleSetSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "gha-runners",
				},
			},
			Github: &KubernetesGhaRunnerScaleSetGitHubConfig{
				ConfigUrl: "https://github.com/myorg/myrepo",
				Auth: &KubernetesGhaRunnerScaleSetGitHubConfig_PatToken{
					PatToken: &KubernetesGhaRunnerScaleSetPatToken{
						Token: "ghp_sample_token",
					},
				},
			},
			ContainerMode: &KubernetesGhaRunnerScaleSetContainerMode{
				Type: KubernetesGhaRunnerScaleSetContainerMode_DIND,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with minimal configuration (PAT token + DIND)", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with GitHub App authentication", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetGitHubApp{
						AppId:            "123456",
						InstallationId:   "654321",
						PrivateKeyBase64: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with existing secret name", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_ExistingSecretName{
					ExistingSecretName: "github-credentials",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with organization-level runners", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Github.ConfigUrl = "https://github.com/myorg"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with enterprise-level runners", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Github.ConfigUrl = "https://github.com/enterprises/myenterprise"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with kubernetes container mode", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES,
					WorkVolumeClaim: &KubernetesGhaRunnerScaleSetWorkVolumeClaim{
						StorageClass: "standard",
						Size:         "10Gi",
						AccessModes:  []string{"ReadWriteOnce"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with kubernetes-novolume container mode", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES_NO_VOLUME,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with default container mode", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_DEFAULT,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with scaling configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				minRunners := int32(2)
				maxRunners := int32(10)
				spec.Scaling = &KubernetesGhaRunnerScaleSetScaling{
					MinRunners: &minRunners,
					MaxRunners: &maxRunners,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with scale-to-zero", func() {
			ginkgo.It("should not return a validation error", func() {
				minRunners := int32(0)
				maxRunners := int32(5)
				spec.Scaling = &KubernetesGhaRunnerScaleSetScaling{
					MinRunners: &minRunners,
					MaxRunners: &maxRunners,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with runner group", func() {
			ginkgo.It("should not return a validation error", func() {
				runnerGroup := "production"
				spec.RunnerGroup = &runnerGroup
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom runner scale set name", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.RunnerScaleSetName = "my-custom-runners"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom runner image", func() {
			ginkgo.It("should not return a validation error", func() {
				repository := "my-registry.com/custom-runner"
				tag := "v1.0.0"
				pullPolicy := "IfNotPresent"
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					Image: &KubernetesGhaRunnerScaleSetRunnerImage{
						Repository: &repository,
						Tag:        &tag,
						PullPolicy: &pullPolicy,
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with runner resources", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					Resources: &kubernetes.ContainerResources{
						Requests: &kubernetes.CpuMemory{
							Cpu:    "1",
							Memory: "2Gi",
						},
						Limits: &kubernetes.CpuMemory{
							Cpu:    "4",
							Memory: "8Gi",
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with runner environment variables", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					Env: []*KubernetesGhaRunnerScaleSetEnvVar{
						{Name: "RUNNER_DEBUG", Value: "1"},
						{Name: "CUSTOM_VAR", Value: "custom_value"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with persistent volumes", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:         "npm-cache",
						StorageClass: "standard",
						Size:         "20Gi",
						AccessModes:  []string{"ReadWriteOnce"},
						MountPath:    "/home/runner/.npm",
					},
					{
						Name:      "gradle-cache",
						Size:      "50Gi",
						MountPath: "/home/runner/.gradle",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with controller service account", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ControllerServiceAccount = &KubernetesGhaRunnerScaleSetControllerServiceAccount{
					Namespace: "arc-system",
					Name:      "arc-controller",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with image pull secrets", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ImagePullSecrets = []string{"ghcr-secret", "docker-secret"}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with labels and annotations", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Labels = map[string]string{
					"team":        "platform",
					"environment": "production",
				}
				spec.Annotations = map[string]string{
					"description": "production runners",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with create namespace enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.CreateNamespace = true
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with GitHub Enterprise Server URL", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Github.ConfigUrl = "https://github.mycompany.com/myorg/myrepo"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("without namespace", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Namespace = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without GitHub config", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without GitHub config URL", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.ConfigUrl = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid GitHub config URL", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.ConfigUrl = "not-a-github-url"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without container mode", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ContainerMode = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with unspecified container mode type", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_container_mode_type_unspecified,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid volume size format", func() {
			ginkgo.It("should return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:      "cache",
						Size:      "invalid-size",
						MountPath: "/cache",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid access modes", func() {
			ginkgo.It("should return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:        "cache",
						Size:        "10Gi",
						MountPath:   "/cache",
						AccessModes: []string{"InvalidMode"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with persistent volume missing name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:      "",
						Size:      "10Gi",
						MountPath: "/cache",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with persistent volume missing size", func() {
			ginkgo.It("should return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:      "cache",
						Size:      "",
						MountPath: "/cache",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with persistent volume missing mount path", func() {
			ginkgo.It("should return a validation error", func() {
				spec.PersistentVolumes = []*KubernetesGhaRunnerScaleSetPersistentVolume{
					{
						Name:      "cache",
						Size:      "10Gi",
						MountPath: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with PAT token missing token", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_PatToken{
					PatToken: &KubernetesGhaRunnerScaleSetPatToken{
						Token: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with GitHub App missing app_id", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetGitHubApp{
						AppId:            "",
						InstallationId:   "654321",
						PrivateKeyBase64: "key",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with GitHub App missing installation_id", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetGitHubApp{
						AppId:            "123456",
						InstallationId:   "",
						PrivateKeyBase64: "key",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with GitHub App missing private_key", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Github.Auth = &KubernetesGhaRunnerScaleSetGitHubConfig_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetGitHubApp{
						AppId:            "123456",
						InstallationId:   "654321",
						PrivateKeyBase64: "",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with env var missing name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					Env: []*KubernetesGhaRunnerScaleSetEnvVar{
						{Name: "", Value: "value"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with volume mount missing name", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					VolumeMounts: []*KubernetesGhaRunnerScaleSetVolumeMount{
						{Name: "", MountPath: "/data"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with volume mount missing mount path", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
					VolumeMounts: []*KubernetesGhaRunnerScaleSetVolumeMount{
						{Name: "cache", MountPath: ""},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with negative min runners", func() {
			ginkgo.It("should return a validation error", func() {
				minRunners := int32(-1)
				spec.Scaling = &KubernetesGhaRunnerScaleSetScaling{
					MinRunners: &minRunners,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with max runners less than 1", func() {
			ginkgo.It("should return a validation error", func() {
				maxRunners := int32(0)
				spec.Scaling = &KubernetesGhaRunnerScaleSetScaling{
					MaxRunners: &maxRunners,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid work volume claim size", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES,
					WorkVolumeClaim: &KubernetesGhaRunnerScaleSetWorkVolumeClaim{
						Size: "not-valid",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid work volume claim access modes", func() {
			ginkgo.It("should return a validation error", func() {
				spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
					Type: KubernetesGhaRunnerScaleSetContainerMode_KUBERNETES,
					WorkVolumeClaim: &KubernetesGhaRunnerScaleSetWorkVolumeClaim{
						Size:        "10Gi",
						AccessModes: []string{"InvalidMode"},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
