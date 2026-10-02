package gcpvertexaifeatureonlinestorev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiFeatureOnlineStoreSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiFeatureOnlineStoreSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	optimized := func() *GcpVertexAiFeatureOnlineStore {
		return &GcpVertexAiFeatureOnlineStore{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiFeatureOnlineStore",
			Metadata:   &shared.CloudResourceMetadata{Name: "serving-store"},
			Spec: &GcpVertexAiFeatureOnlineStoreSpec{
				Location:             "us-central1",
				FeatureOnlineStoreId: "serving_store",
				Optimized:            true,
			},
		}
	}

	bigtable := func() *GcpVertexAiFeatureOnlineStore {
		msg := optimized()
		msg.Spec.Optimized = false
		msg.Spec.Bigtable = &GcpVertexAiFeatureOnlineStoreBigtable{
			AutoScaling: &GcpVertexAiFeatureOnlineStoreBigtableAutoScaling{MinNodeCount: 1, MaxNodeCount: 3},
		}
		msg.Spec.FeatureViews = []*GcpVertexAiFeatureOnlineStoreFeatureView{
			{
				FeatureViewId: "customer_view",
				FeatureRegistrySource: &GcpVertexAiFeatureOnlineStoreFeatureRegistrySource{
					FeatureGroups: []*GcpVertexAiFeatureOnlineStoreFeatureGroupSelection{
						{FeatureGroupId: litRef("customer_features"), FeatureIds: []string{"age", "lifetime_value"}},
					},
				},
				SyncConfig: &GcpVertexAiFeatureOnlineStoreSyncConfig{Cron: "0 */6 * * *"},
			},
			{
				FeatureViewId: "raw_view",
				BigQuerySource: &GcpVertexAiFeatureOnlineStoreBigQuerySource{
					Uri:             litRef("ml-project.features.customers"),
					EntityIdColumns: []string{"customer_id"},
				},
			},
		}
		return msg
	}

	ginkgo.It("should accept an optimized store alone", func() {
		gomega.Expect(validator.Validate(optimized())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a bigtable store with both view sources", func() {
		gomega.Expect(validator.Validate(bigtable())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := bigtable()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.Labels = map[string]string{"team": "ml"}
		msg.Spec.Bigtable.EnableDirectBigtableAccess = true
		msg.Spec.Bigtable.Zone = "us-central1-a"
		target := int32(60)
		msg.Spec.Bigtable.AutoScaling.CpuUtilizationTarget = &target
		msg.Spec.DedicatedServingEndpoint = &GcpVertexAiFeatureOnlineStoreDedicatedServingEndpoint{
			PrivateServiceConnectConfig: &GcpVertexAiFeatureOnlineStorePrivateServiceConnectConfig{
				EnablePrivateServiceConnect: true,
				ProjectAllowlist:            []string{"consumer-project"},
			},
		}
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.ForceDestroy = true
		msg.Spec.FeatureViews[0].Labels = map[string]string{"tier": "gold"}
		msg.Spec.FeatureViews[0].FeatureRegistrySource.ProjectNumber = litRef("123456789012")
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one storage kind", func() {
		msg := optimized()
		msg.Spec.Optimized = false
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		msg.Spec.Optimized = true
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bound bigtable autoscaling the way Google does", func() {
		msg := bigtable()
		msg.Spec.Bigtable.AutoScaling.MinNodeCount = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		msg.Spec.Bigtable.AutoScaling.MaxNodeCount = 11
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		msg.Spec.Bigtable.AutoScaling.MinNodeCount = 4
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		target := int32(90)
		msg.Spec.Bigtable.AutoScaling.CpuUtilizationTarget = &target
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require store and view ids Google accepts", func() {
		msg := optimized()
		msg.Spec.FeatureOnlineStoreId = "serving-store"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		msg.Spec.FeatureViews[0].FeatureViewId = "9view"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one source per view", func() {
		msg := bigtable()
		msg.Spec.FeatureViews[1].BigQuerySource = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = bigtable()
		msg.Spec.FeatureViews[1].FeatureRegistrySource = msg.Spec.FeatureViews[0].FeatureRegistrySource
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a feature selection to name features", func() {
		msg := bigtable()
		msg.Spec.FeatureViews[0].FeatureRegistrySource.FeatureGroups[0].FeatureIds = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a sync that is both cron and continuous", func() {
		msg := bigtable()
		msg.Spec.FeatureViews[0].SyncConfig.Continuous = true
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a duplicate view id", func() {
		msg := bigtable()
		msg.Spec.FeatureViews[1].FeatureViewId = "customer_view"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := optimized()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
