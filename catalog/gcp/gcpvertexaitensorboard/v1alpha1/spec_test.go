package gcpvertexaitensorboardv1alpha1

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
	ginkgo.RunSpecs(t, "GcpVertexAiTensorboardSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiTensorboardSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiTensorboard {
		return &GcpVertexAiTensorboard{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiTensorboard",
			Metadata:   &shared.CloudResourceMetadata{Name: "training-metrics"},
			Spec: &GcpVertexAiTensorboardSpec{
				Location: "us-central1",
			},
		}
	}

	withExperiments := func() *GcpVertexAiTensorboard {
		msg := minimal()
		msg.Spec.Experiments = []*GcpVertexAiTensorboardExperiment{
			{
				ExperimentId: "churn-model",
				DisplayName:  "Churn model",
				Source:       "custom training job",
				Runs: []*GcpVertexAiTensorboardRun{
					{RunId: "baseline"},
					{RunId: "lr-sweep-1", DisplayName: "Learning-rate sweep 1"},
				},
			},
			{ExperimentId: "ranking-v2"},
		}
		return msg
	}

	ginkgo.It("should accept a TensorBoard alone", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept experiments and runs with every field set", func() {
		msg := withExperiments()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.DisplayName = "Training metrics"
		msg.Spec.Description = "Shared TensorBoard"
		msg.Spec.Labels = map[string]string{"team": "ml"}
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a regional location", func() {
		msg := minimal()
		msg.Spec.Location = "us"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject ids outside lowercase letters, digits, and hyphens", func() {
		msg := withExperiments()
		msg.Spec.Experiments[0].ExperimentId = "Churn_Model"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withExperiments()
		msg.Spec.Experiments[0].Runs[0].RunId = "base line"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a duplicate experiment id", func() {
		msg := withExperiments()
		msg.Spec.Experiments[1].ExperimentId = "churn-model"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a duplicate run id within an experiment", func() {
		msg := withExperiments()
		msg.Spec.Experiments[0].Runs[1].RunId = "baseline"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a run display name that collides with another run's default", func() {
		msg := withExperiments()
		msg.Spec.Experiments[0].Runs[1].DisplayName = "baseline"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
