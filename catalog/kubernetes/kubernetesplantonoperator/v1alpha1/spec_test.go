package kubernetesplantonoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesPlantonOperatorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPlantonOperatorSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalValidOperator is the common case: the operator in its
// conventional namespace, everything else riding the chart defaults.
func minimalValidOperator() *KubernetesPlantonOperator {
	return &KubernetesPlantonOperator{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPlantonOperator",
		Metadata: &shared.CloudResourceMetadata{
			Name: "planton-operator",
		},
		Spec: &KubernetesPlantonOperatorSpec{
			Namespace:       literalRef("planton-operator"),
			CreateNamespace: true,
		},
	}
}

var _ = ginkgo.Describe("KubernetesPlantonOperatorSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for a minimal operator", func() {
			err := protovalidate.Validate(minimalValidOperator())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an exact-semver chart version pin", func() {
			input := minimalValidOperator()
			input.Spec.ChartVersion = stringPtr("0.8.0")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a mirrored chart repository", func() {
			input := minimalValidOperator()
			input.Spec.ChartRepository = stringPtr("oci://asia-south1-docker.pkg.dev/plantonhq/charts")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept sizing, scheduling, and image overrides", func() {
			input := minimalValidOperator()
			replicas := int32(2)
			leaderElection := true
			input.Spec.Replicas = &replicas
			input.Spec.LeaderElection = &leaderElection
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "10m", Memory: "256Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "500m", Memory: "512Mi"},
			}
			input.Spec.NodeSelector = map[string]string{"role": "platform"}
			input.Spec.Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "platform", Operator: "Exists", Effect: "NoSchedule"},
			}
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Image = &KubernetesPlantonOperatorImage{
				Repository: "mirror.example.com/planton/operator",
				Tag:        "v0.0.41-selfhosted-preview",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a brought ServiceAccount", func() {
			input := minimalValidOperator()
			create := false
			input.Spec.ServiceAccount = &KubernetesPlantonOperatorServiceAccount{
				Create: &create,
				Name:   "platform-operator",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept every position of the two CRD dials", func() {
			for _, dials := range []*KubernetesPlantonOperatorCrds{
				{},
				{Install: boolPtr(false)},
				{KeepOnUninstall: boolPtr(false)},
				{Install: boolPtr(true), KeepOnUninstall: boolPtr(true)},
			} {
				input := minimalValidOperator()
				input.Spec.Crds = dials
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should fail when spec is missing", func() {
			input := minimalValidOperator()
			input.Spec = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail when namespace is missing", func() {
			input := minimalValidOperator()
			input.Spec.Namespace = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a wrong kind constant", func() {
			input := minimalValidOperator()
			input.Kind = "PlantonOperator"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a chart repository that is not an oci:// path", func() {
			for _, repo := range []string{"ghcr.io/plantonhq/charts", "oci://ghcr.io/plantonhq/charts/"} {
				input := minimalValidOperator()
				input.Spec.ChartRepository = stringPtr(repo)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), repo)
			}
		})

		ginkgo.It("should fail on a chart version range (not reproducible)", func() {
			input := minimalValidOperator()
			input.Spec.ChartVersion = stringPtr("^0.8.0")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on zero replicas", func() {
			input := minimalValidOperator()
			replicas := int32(0)
			input.Spec.Replicas = &replicas
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})

func boolPtr(b bool) *bool { return &b }

func stringPtr(v string) *string { return &v }
