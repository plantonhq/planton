package gcpmodelarmortemplatev1alpha1

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
	ginkgo.RunSpecs(t, "GcpModelArmorTemplateSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpModelArmorTemplateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpModelArmorTemplate {
		return &GcpModelArmorTemplate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpModelArmorTemplate",
			Metadata:   &shared.CloudResourceMetadata{Name: "prompt-guard"},
			Spec: &GcpModelArmorTemplateSpec{
				Location: "us-central1",
				FilterConfig: &GcpModelArmorTemplateFilterConfig{
					PiAndJailbreakFilterSettings: &GcpModelArmorTemplatePiAndJailbreakFilterSettings{
						FilterEnforcement: "ENABLED",
						ConfidenceLevel:   "MEDIUM_AND_ABOVE",
					},
				},
			},
		}
	}

	ginkgo.It("should accept the smallest template", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ai-project")
		msg.Spec.Location = "us"
		msg.Spec.TemplateId = "prompt_guard-01"
		msg.Spec.Labels = map[string]string{"team": "platform"}
		msg.Spec.FilterConfig.MaliciousUriFilterSettings = &GcpModelArmorTemplateMaliciousUriFilterSettings{FilterEnforcement: "ENABLED"}
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorTemplateRaiSettings{RaiFilters: []*GcpModelArmorTemplateRaiFilter{
			{FilterType: "HATE_SPEECH", ConfidenceLevel: "LOW_AND_ABOVE"},
			{FilterType: "DANGEROUS"},
		}}
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorTemplateSdpSettings{AdvancedConfig: &GcpModelArmorTemplateSdpAdvancedConfig{
			InspectTemplate:    "projects/ai-project/locations/us-central1/inspectTemplates/pii",
			DeidentifyTemplate: "projects/ai-project/locations/us-central1/deidentifyTemplates/redact",
		}}
		msg.Spec.TemplateMetadata = &GcpModelArmorTemplateMetadata{
			LogTemplateOperations:               true,
			LogSanitizeOperations:               true,
			EnableMultiLanguageDetection:        true,
			IgnorePartialInvocationFailures:     true,
			CustomPromptSafetyErrorCode:         403,
			CustomPromptSafetyErrorMessage:      "Prompt blocked",
			CustomLlmResponseSafetyErrorCode:    451,
			CustomLlmResponseSafetyErrorMessage: "Response blocked",
			EnforcementType:                     "INSPECT_AND_BLOCK",
			FilterVersionSelector:               &GcpModelArmorTemplateFilterVersionSelector{Version: "v2"},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location and a filter configuration", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a template id with a slash", func() {
		msg := minimal()
		msg.Spec.TemplateId = "a/b"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject unknown enforcement and confidence values", func() {
		msg := minimal()
		msg.Spec.FilterConfig.PiAndJailbreakFilterSettings.FilterEnforcement = "ON"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig.PiAndJailbreakFilterSettings.ConfidenceLevel = "MEDIUM"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an empty or duplicated Responsible AI filter list", func() {
		msg := minimal()
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorTemplateRaiSettings{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorTemplateRaiSettings{RaiFilters: []*GcpModelArmorTemplateRaiFilter{
			{FilterType: "HARASSMENT"}, {FilterType: "HARASSMENT"},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject basic and advanced Sensitive Data Protection together", func() {
		msg := minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorTemplateSdpSettings{
			BasicConfig:    &GcpModelArmorTemplateSdpBasicConfig{FilterEnforcement: "ENABLED"},
			AdvancedConfig: &GcpModelArmorTemplateSdpAdvancedConfig{InspectTemplate: "projects/p/locations/l/inspectTemplates/t"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed Sensitive Data Protection template name", func() {
		msg := minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorTemplateSdpSettings{
			AdvancedConfig: &GcpModelArmorTemplateSdpAdvancedConfig{InspectTemplate: "pii"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of filter version alias or version", func() {
		msg := minimal()
		msg.Spec.TemplateMetadata = &GcpModelArmorTemplateMetadata{FilterVersionSelector: &GcpModelArmorTemplateFilterVersionSelector{}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.TemplateMetadata.FilterVersionSelector = &GcpModelArmorTemplateFilterVersionSelector{Alias: "FILTER_VERSION_ALIAS_STABLE", Version: "v1"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.TemplateMetadata.FilterVersionSelector = &GcpModelArmorTemplateFilterVersionSelector{Alias: "FILTER_VERSION_ALIAS_LATEST"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown enforcement type and deletion policy", func() {
		msg := minimal()
		msg.Spec.TemplateMetadata = &GcpModelArmorTemplateMetadata{EnforcementType: "BLOCK"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
