package kubernetesmanifestv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesManifest(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesManifest Suite")
}

var _ = ginkgo.Describe("KubernetesManifest Validation Tests", func() {
	var input *KubernetesManifest

	ginkgo.BeforeEach(func() {
		input = &KubernetesManifest{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesManifest",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-manifest",
			},
			Spec: &KubernetesManifestSpec{
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
				ManifestYaml: `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  key: value`,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with all required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multi-document manifest", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ManifestYaml = `apiVersion: v1
kind: ConfigMap
metadata:
  name: config-1
data:
  key: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config-2
data:
  key: value2`
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("without target_cluster (optional field)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TargetCluster = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Namespace validation", func() {
		ginkgo.Context("when namespace is missing", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when namespace has a value", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-namespace",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("ManifestYaml validation", func() {
		ginkgo.Context("when manifest_yaml is empty", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ManifestYaml = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when manifest_yaml has content", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ManifestYaml = "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: test"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Metadata validation", func() {
		ginkgo.Context("when metadata is missing", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when metadata name is provided", func() {
			ginkgo.It("should pass validation", func() {
				input.Metadata.Name = "my-manifest"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Spec validation", func() {
		ginkgo.Context("when spec is missing", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Namespace creation flag", func() {
		ginkgo.Context("when create_namespace is true", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.CreateNamespace = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when create_namespace is false", func() {
			ginkgo.It("should pass validation", func() {
				input.Spec.CreateNamespace = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("API version and kind validation", func() {
		ginkgo.Context("when api_version is incorrect", func() {
			ginkgo.It("should return a validation error", func() {
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when kind is incorrect", func() {
			ginkgo.It("should return a validation error", func() {
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
