package kubernetesgharunnerscalesetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesGhaRunnerScaleSet(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGhaRunnerScaleSet Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }

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

func patAuth() *KubernetesGhaRunnerScaleSetAuth {
	return &KubernetesGhaRunnerScaleSetAuth{
		Method: &KubernetesGhaRunnerScaleSetAuth_ExistingSecretName{
			ExistingSecretName: "github-credential",
		},
	}
}

var _ = ginkgo.Describe("KubernetesGhaRunnerScaleSet Validation Tests", func() {
	var input *KubernetesGhaRunnerScaleSet

	ginkgo.BeforeEach(func() {
		input = &KubernetesGhaRunnerScaleSet{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGhaRunnerScaleSet",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "build-runners",
			},
			Spec: &KubernetesGhaRunnerScaleSetSpec{
				Namespace:       literal("ci-runners"),
				GithubConfigUrl: "https://github.com/my-org/my-repo",
				Auth:            patAuth(),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec with an existing credential Secret should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an organization URL and enterprise URL should be valid", func() {
			input.Spec.GithubConfigUrl = "https://github.com/my-org"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.GithubConfigUrl = "https://github.com/enterprises/my-enterprise"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a declared PAT should be valid", func() {
			input.Spec.Auth = &KubernetesGhaRunnerScaleSetAuth{
				Method: &KubernetesGhaRunnerScaleSetAuth_Pat{
					Pat: &KubernetesGhaRunnerScaleSetAuthPat{Token: "ghp_example"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a declared GitHub App should be valid", func() {
			input.Spec.Auth = &KubernetesGhaRunnerScaleSetAuth{
				Method: &KubernetesGhaRunnerScaleSetAuth_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetAuthGithubApp{
						AppId:          "123456",
						InstallationId: "654321",
						PrivateKey:     "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("scaling bounds and identity should be valid", func() {
			input.Spec.RunnerScaleSetName = "org-build-runners"
			input.Spec.RunnerGroup = "platform"
			input.Spec.MinRunners = int32Ptr(2)
			input.Spec.MaxRunners = int32Ptr(20)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dind container mode should be valid", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{Mode: "dind"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kubernetes container mode with a work volume should be valid", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
				Mode: "kubernetes",
				KubernetesWorkVolume: &KubernetesGhaRunnerScaleSetWorkVolume{
					StorageClass: valueFrom(catalogkind.CatalogKind_KubernetesStorageClass, "fast-ssd", "metadata.name"),
					Size:         "2Gi",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kubernetes-novolume mode without a work volume should be valid", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{Mode: "kubernetes-novolume"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("runner, proxy, TLS and controller reference should be valid", func() {
			input.Spec.Runner = &KubernetesGhaRunnerScaleSetRunner{
				Image: "ghcr.io/actions/actions-runner:2.321.0",
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "500m", Memory: "1Gi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "2000m", Memory: "4Gi"},
				},
			}
			input.Spec.Proxy = &KubernetesGhaRunnerScaleSetProxy{
				Https:   &KubernetesGhaRunnerScaleSetProxyServer{Url: "http://proxy.example.com:8080", CredentialSecretName: "proxy-auth"},
				NoProxy: []string{"svc.cluster.local"},
			}
			input.Spec.GithubServerTls = &KubernetesGhaRunnerScaleSetGithubServerTls{
				ConfigMapName:   literal("ghes-ca"),
				RunnerMountPath: "/usr/local/share/ca-certificates/",
			}
			input.Spec.ControllerServiceAccount = &KubernetesGhaRunnerScaleSetControllerRef{
				Namespace: "arc-system",
				Name:      "arc-gha-rs-controller",
			}
			input.Spec.HelmValues = "listenerTemplate:\n  spec:\n    containers:\n      - name: listener\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing github_config_url should fail", func() {
			input.Spec.GithubConfigUrl = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a github_config_url without a scheme should fail", func() {
			input.Spec.GithubConfigUrl = "github.com/my-org/my-repo"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a missing auth block should fail", func() {
			input.Spec.Auth = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an empty auth block should fail", func() {
			input.Spec.Auth = &KubernetesGhaRunnerScaleSetAuth{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a PAT without a token should fail", func() {
			input.Spec.Auth = &KubernetesGhaRunnerScaleSetAuth{
				Method: &KubernetesGhaRunnerScaleSetAuth_Pat{Pat: &KubernetesGhaRunnerScaleSetAuthPat{}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a GitHub App missing the installation id should fail", func() {
			input.Spec.Auth = &KubernetesGhaRunnerScaleSetAuth{
				Method: &KubernetesGhaRunnerScaleSetAuth_GithubApp{
					GithubApp: &KubernetesGhaRunnerScaleSetAuthGithubApp{
						AppId:      "123456",
						PrivateKey: "-----BEGIN RSA PRIVATE KEY-----",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a 46-character runner scale set name should fail", func() {
			input.Spec.RunnerScaleSetName = "a123456789b123456789c123456789d123456789e12345"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("max_runners below min_runners should fail", func() {
			input.Spec.MinRunners = int32Ptr(10)
			input.Spec.MaxRunners = int32Ptr(5)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown container mode should fail", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{Mode: "docker"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("kubernetes mode without a work volume should fail", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{Mode: "kubernetes"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("dind mode WITH a work volume should fail", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
				Mode: "dind",
				KubernetesWorkVolume: &KubernetesGhaRunnerScaleSetWorkVolume{
					StorageClass: literal("fast-ssd"),
					Size:         "1Gi",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a work volume with a malformed size should fail", func() {
			input.Spec.ContainerMode = &KubernetesGhaRunnerScaleSetContainerMode{
				Mode: "kubernetes",
				KubernetesWorkVolume: &KubernetesGhaRunnerScaleSetWorkVolume{
					StorageClass: literal("fast-ssd"),
					Size:         "2 gigabytes",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a controller reference missing the name should fail", func() {
			input.Spec.ControllerServiceAccount = &KubernetesGhaRunnerScaleSetControllerRef{Namespace: "arc-system"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
