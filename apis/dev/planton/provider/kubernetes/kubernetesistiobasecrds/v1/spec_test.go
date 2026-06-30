package kubernetesistiobasecrdsv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestKubernetesIstioBaseCrdsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIstioBaseCrdsSpec Validation Suite")
}

// The spec is intentionally minimal: the CRD version is pinned to the typed SDK (no
// user version field), so there are no value-validation rules to exercise here. The
// cases below assert the spec accepts both a set and an unset (optional) target_cluster.
var _ = ginkgo.Describe("KubernetesIstioBaseCrdsSpec validations", func() {
	var spec *KubernetesIstioBaseCrdsSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesIstioBaseCrdsSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "test-cluster",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with target cluster set", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("without target cluster (optional field)", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.TargetCluster = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
