package gcpcolabruntimetemplatev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpColabRuntimeTemplateSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpColabRuntimeTemplateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpColabRuntimeTemplate {
		return &GcpColabRuntimeTemplate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpColabRuntimeTemplate",
			Metadata:   &shared.CloudResourceMetadata{Name: "standard-runtime"},
			Spec:       &GcpColabRuntimeTemplateSpec{Location: "us-central1"},
		}
	}

	ginkgo.It("should accept the smallest template", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.RuntimeTemplateId = "gpu-t4"
		msg.Spec.DisplayName = "GPU (T4)"
		msg.Spec.Description = "One T4 GPU for the vision team"
		msg.Spec.Labels = map[string]string{"team": "vision"}
		msg.Spec.MachineSpec = &GcpColabRuntimeTemplateMachineSpec{MachineType: "n1-standard-8", AcceleratorType: "NVIDIA_TESLA_T4", AcceleratorCount: 1}
		msg.Spec.DataPersistentDiskSpec = &GcpColabRuntimeTemplateDataPersistentDiskSpec{DiskType: "pd-balanced", DiskSizeGb: 200}
		msg.Spec.NetworkSpec = &GcpColabRuntimeTemplateNetworkSpec{
			EnableInternetAccess: false,
			Network:              litRef("projects/ml-project/global/networks/ml-vpc"),
			Subnetwork:           litRef("projects/ml-project/regions/us-central1/subnetworks/notebooks"),
		}
		msg.Spec.IdleTimeout = "3600s"
		msg.Spec.EucDisabled = true
		msg.Spec.EnableSecureBoot = true
		msg.Spec.NetworkTags = []string{"notebooks"}
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.SoftwareConfig = &GcpColabRuntimeTemplateSoftwareConfig{
			Env:                     []*GcpColabRuntimeTemplateEnvVar{{Name: "DATA_BUCKET", Value: "gs://ml-data"}},
			PostStartupScriptConfig: &GcpColabRuntimeTemplatePostStartupScriptConfig{PostStartupScriptUrl: "gs://ml-scripts/setup.sh", PostStartupScriptBehavior: "RUN_EVERY_START"},
			ColabImage:              &GcpColabRuntimeTemplateColabImage{ReleaseName: "py310"},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a regional location", func() {
		msg := minimal()
		msg.Spec.Location = "us"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an upper-case template id and an over-long display name", func() {
		msg := minimal()
		msg.Spec.RuntimeTemplateId = "GPU"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DisplayName = strings.Repeat("d", 129)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require an accelerator count with an accelerator type", func() {
		msg := minimal()
		msg.Spec.MachineSpec = &GcpColabRuntimeTemplateMachineSpec{AcceleratorType: "NVIDIA_L4"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a disk type with a disk size and keep the size in range", func() {
		msg := minimal()
		msg.Spec.DataPersistentDiskSpec = &GcpColabRuntimeTemplateDataPersistentDiskSpec{DiskSizeGb: 100}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.DataPersistentDiskSpec = &GcpColabRuntimeTemplateDataPersistentDiskSpec{DiskType: "pd-ssd", DiskSizeGb: 5}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.DataPersistentDiskSpec = &GcpColabRuntimeTemplateDataPersistentDiskSpec{DiskType: "ssd"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a non-second idle timeout", func() {
		msg := minimal()
		msg.Spec.IdleTimeout = "1h"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an environment variable that is not a C identifier", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpColabRuntimeTemplateSoftwareConfig{Env: []*GcpColabRuntimeTemplateEnvVar{{Name: "1-BAD"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown post-startup behavior and deletion policy", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = &GcpColabRuntimeTemplateSoftwareConfig{PostStartupScriptConfig: &GcpColabRuntimeTemplatePostStartupScriptConfig{PostStartupScriptBehavior: "ALWAYS"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
