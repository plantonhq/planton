package awssesaccountsettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

func TestAwsSesAccountSettingsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSesAccountSettingsSpec Validation Suite")
}

// suppressionOnly manages just the suppression list.
func suppressionOnly() *AwsSesAccountSettingsSpec {
	return &AwsSesAccountSettingsSpec{
		Region: "us-west-2",
		Suppression: &AwsSesAccountSettingsSuppression{
			Reasons: []string{"BOUNCE", "COMPLAINT"},
		},
	}
}

// vdmOnly manages just the VDM posture.
func vdmOnly() *AwsSesAccountSettingsSpec {
	return &AwsSesAccountSettingsSpec{
		Region: "us-west-2",
		Vdm: &AwsSesAccountSettingsVdm{
			Enabled:           proto.Bool(true),
			EngagementMetrics: proto.Bool(true),
		},
	}
}

var _ = ginkgo.Describe("AwsSesAccountSettingsSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with the suppression arm only", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(suppressionOnly())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the VDM arm only", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(vdmOnly())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with suppression switched off (auto-suppression off)", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := suppressionOnly()
				spec.Suppression = &AwsSesAccountSettingsSuppression{Enabled: proto.Bool(false)}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with both arms", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := suppressionOnly()
				spec.Vdm = &AwsSesAccountSettingsVdm{Enabled: proto.Bool(true), OptimizedSharedDelivery: proto.Bool(false)}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("with a suppression arm that names no events and does not say off", func() {
			ginkgo.It("should return a validation error naming the switch", func() {
				spec := suppressionOnly()
				spec.Suppression = &AwsSesAccountSettingsSuppression{}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("enabled: false"))
			})
		})

		ginkgo.Context("with suppression switched off but still naming events", func() {
			ginkgo.It("should return a validation error", func() {
				spec := suppressionOnly()
				spec.Suppression = &AwsSesAccountSettingsSuppression{Enabled: proto.Bool(false), Reasons: []string{"BOUNCE"}}
				gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a vdm arm that does not say on or off", func() {
			ginkgo.It("should return a validation error", func() {
				spec := vdmOnly()
				spec.Vdm = &AwsSesAccountSettingsVdm{EngagementMetrics: proto.Bool(true)}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("vdm.enabled"))
			})
		})

		ginkgo.It("rejects an instance managing neither arm", func() {
			spec := &AwsSesAccountSettingsSpec{Region: "us-west-2"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one"))
		})

		ginkgo.It("rejects an unknown suppression reason", func() {
			spec := suppressionOnly()
			spec.Suppression.Reasons = []string{"BOUNCE", "UNSUBSCRIBE"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate suppression reasons", func() {
			spec := suppressionOnly()
			spec.Suppression.Reasons = []string{"BOUNCE", "BOUNCE"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing region", func() {
			spec := suppressionOnly()
			spec.Region = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
