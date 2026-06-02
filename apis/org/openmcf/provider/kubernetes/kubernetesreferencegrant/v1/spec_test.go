package kubernetesreferencegrantv1

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

func TestKubernetesReferenceGrant(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesReferenceGrant Suite")
}

func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

var _ = ginkgo.Describe("KubernetesReferenceGrant Validation Tests", func() {
	var input *KubernetesReferenceGrant

	ginkgo.BeforeEach(func() {
		input = &KubernetesReferenceGrant{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesReferenceGrant",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-reference-grant",
			},
			Spec: &KubernetesReferenceGrantSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("app-ns"),
				From: []*KubernetesReferenceGrantFrom{
					{
						Group:     stringPtr("gateway.networking.k8s.io"),
						Kind:      "HTTPRoute",
						Namespace: "frontend-ns",
					},
				},
				To: []*KubernetesReferenceGrantTo{
					{
						Group: stringPtr(""),
						Kind:  "Service",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal grant should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("from entry with empty (core) group should be valid", func() {
			input.Spec.From[0].Group = stringPtr("")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("to entry with a specific name should be valid", func() {
			input.Spec.To[0].Name = stringPtr("my-secret")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple from and to entries should be valid", func() {
			input.Spec.From = []*KubernetesReferenceGrantFrom{
				{Group: stringPtr("gateway.networking.k8s.io"), Kind: "HTTPRoute", Namespace: "frontend-ns"},
				{Group: stringPtr("gateway.networking.k8s.io"), Kind: "Gateway", Namespace: "infra-ns"},
			}
			input.Spec.To = []*KubernetesReferenceGrantTo{
				{Group: stringPtr(""), Kind: "Service"},
				{Group: stringPtr(""), Kind: "Secret", Name: stringPtr("tls-cert")},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty from list should fail (min_items=1)", func() {
			input.Spec.From = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty to list should fail (min_items=1)", func() {
			input.Spec.To = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("from entry without a kind should fail (required)", func() {
			input.Spec.From[0].Kind = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("from entry with an invalid kind pattern should fail", func() {
			input.Spec.From[0].Kind = "bad kind"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("from entry without a namespace should fail (required)", func() {
			input.Spec.From[0].Namespace = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("from entry with an invalid namespace pattern should fail", func() {
			input.Spec.From[0].Namespace = "Bad-NS"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("to entry without a kind should fail (required)", func() {
			input.Spec.To[0].Kind = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("to entry with an invalid group pattern should fail", func() {
			input.Spec.To[0].Group = stringPtr("Invalid_Group")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("to entry without a group should fail (presence required: the CRD needs the key, even if empty)", func() {
			input.Spec.To[0].Group = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("from entry without a group should fail (presence required)", func() {
			input.Spec.From[0].Group = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("to entry with an over-long name should fail (max_len=253)", func() {
			long := make([]byte, 254)
			for i := range long {
				long[i] = 'a'
			}
			input.Spec.To[0].Name = stringPtr(string(long))
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than 16 from entries should fail (max_items=16)", func() {
			entries := make([]*KubernetesReferenceGrantFrom, 0, 17)
			for i := 0; i < 17; i++ {
				entries = append(entries, &KubernetesReferenceGrantFrom{
					Group:     stringPtr("gateway.networking.k8s.io"),
					Kind:      "HTTPRoute",
					Namespace: "frontend-ns",
				})
			}
			input.Spec.From = entries
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing spec should fail", func() {
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing metadata should fail", func() {
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
