package gcpmodelarmorfloorsettingv1alpha1

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
	ginkgo.RunSpecs(t, "GcpModelArmorFloorSettingSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpModelArmorFloorSettingSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpModelArmorFloorSetting {
		return &GcpModelArmorFloorSetting{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpModelArmorFloorSetting",
			Metadata:   &shared.CatalogObjectMetadata{Name: "project-floor"},
			Spec: &GcpModelArmorFloorSettingSpec{
				FilterConfig: &GcpModelArmorFloorSettingFilterConfig{
					PiAndJailbreakFilterSettings: &GcpModelArmorFloorSettingPiAndJailbreakFilterSettings{FilterEnforcement: "ENABLED"},
				},
			},
		}
	}

	ginkgo.It("should accept the smallest floor on the provider's project", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{FolderId: litRef("123456789012")}
		msg.Spec.Location = "global"
		msg.Spec.EnableFloorSettingEnforcement = true
		msg.Spec.IntegratedServices = []string{"AI_PLATFORM", "GOOGLE_MCP_SERVER"}
		msg.Spec.FilterConfig.MaliciousUriFilterSettings = &GcpModelArmorFloorSettingMaliciousUriFilterSettings{FilterEnforcement: "ENABLED"}
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorFloorSettingRaiSettings{RaiFilters: []*GcpModelArmorFloorSettingRaiFilter{
			{FilterType: "DANGEROUS", ConfidenceLevel: "HIGH"},
		}}
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorFloorSettingSdpSettings{BasicConfig: &GcpModelArmorFloorSettingSdpBasicConfig{FilterEnforcement: "ENABLED"}}
		msg.Spec.AiPlatformFloorSetting = &GcpModelArmorFloorSettingServiceSetting{EnforcementType: "INSPECT_ONLY", EnableCloudLogging: true}
		msg.Spec.GoogleMcpServerFloorSetting = &GcpModelArmorFloorSettingServiceSetting{EnforcementType: "INSPECT_AND_BLOCK"}
		msg.Spec.EnableMultiLanguageDetection = true
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a filter configuration", func() {
		msg := minimal()
		msg.Spec.FilterConfig = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two scope arms", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{ProjectId: litRef("p"), OrganizationId: "123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a prefixed organization id", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{OrganizationId: "organizations/123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown or repeated integrated service", func() {
		msg := minimal()
		msg.Spec.IntegratedServices = []string{"GEMINI"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.IntegratedServices = []string{"AI_PLATFORM", "AI_PLATFORM"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require an enforcement type for an integrated service setting", func() {
		msg := minimal()
		msg.Spec.IntegratedServices = []string{"AI_PLATFORM"}
		msg.Spec.AiPlatformFloorSetting = &GcpModelArmorFloorSettingServiceSetting{EnableCloudLogging: true}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.AiPlatformFloorSetting.EnforcementType = "BLOCK"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject basic and advanced Sensitive Data Protection together", func() {
		msg := minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorFloorSettingSdpSettings{
			BasicConfig:    &GcpModelArmorFloorSettingSdpBasicConfig{FilterEnforcement: "ENABLED"},
			AdvancedConfig: &GcpModelArmorFloorSettingSdpAdvancedConfig{InspectTemplate: "projects/p/locations/l/inspectTemplates/t"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed location", func() {
		msg := minimal()
		msg.Spec.Location = "Global"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should pair each integrated service with its setting", func() {
		msg := minimal()
		msg.Spec.IntegratedServices = []string{"AI_PLATFORM"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.AiPlatformFloorSetting = &GcpModelArmorFloorSettingServiceSetting{EnforcementType: "INSPECT_ONLY"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.IntegratedServices = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.IntegratedServices = []string{"GOOGLE_MCP_SERVER"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.GoogleMcpServerFloorSetting = &GcpModelArmorFloorSettingServiceSetting{EnforcementType: "INSPECT_AND_BLOCK"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.IntegratedServices = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require an enforcement type for the MCP server setting", func() {
		msg := minimal()
		msg.Spec.IntegratedServices = []string{"GOOGLE_MCP_SERVER"}
		msg.Spec.GoogleMcpServerFloorSetting = &GcpModelArmorFloorSettingServiceSetting{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.GoogleMcpServerFloorSetting.EnforcementType = "BLOCK"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject every pair of scope arms and accept each arm alone", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{ProjectId: litRef("p"), FolderId: litRef("123")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{FolderId: litRef("123"), OrganizationId: "456"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		for _, scope := range []*GcpModelArmorFloorSettingScope{
			{ProjectId: litRef("p")}, {FolderId: litRef("123")}, {OrganizationId: "456"}, {},
		} {
			msg.Spec.Scope = scope
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should require a numeric folder id without its prefix", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{FolderId: litRef("folders/123")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{FolderId: litRef("engineering")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Scope = &GcpModelArmorFloorSettingScope{FolderId: &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "eng"}},
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject unknown enforcement and confidence values", func() {
		msg := minimal()
		msg.Spec.FilterConfig.PiAndJailbreakFilterSettings.FilterEnforcement = "ON"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig.PiAndJailbreakFilterSettings.ConfidenceLevel = "MEDIUM"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig.MaliciousUriFilterSettings = &GcpModelArmorFloorSettingMaliciousUriFilterSettings{FilterEnforcement: "ON"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorFloorSettingSdpSettings{BasicConfig: &GcpModelArmorFloorSettingSdpBasicConfig{FilterEnforcement: "ON"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an empty, duplicated, or unknown Responsible AI filter list", func() {
		msg := minimal()
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorFloorSettingRaiSettings{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorFloorSettingRaiSettings{RaiFilters: []*GcpModelArmorFloorSettingRaiFilter{
			{FilterType: "HARASSMENT"}, {FilterType: "HARASSMENT"},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorFloorSettingRaiSettings{RaiFilters: []*GcpModelArmorFloorSettingRaiFilter{{FilterType: "VIOLENCE"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.FilterConfig.RaiSettings = &GcpModelArmorFloorSettingRaiSettings{RaiFilters: []*GcpModelArmorFloorSettingRaiFilter{{FilterType: "DANGEROUS", ConfidenceLevel: "MEDIUM"}}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should check both Sensitive Data Protection template names", func() {
		msg := minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorFloorSettingSdpSettings{AdvancedConfig: &GcpModelArmorFloorSettingSdpAdvancedConfig{InspectTemplate: "pii"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.FilterConfig.SdpSettings.AdvancedConfig = &GcpModelArmorFloorSettingSdpAdvancedConfig{
			InspectTemplate:    "projects/p/locations/l/inspectTemplates/t",
			DeidentifyTemplate: "redact",
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.FilterConfig.SdpSettings.AdvancedConfig.DeidentifyTemplate = "projects/p/locations/l/deidentifyTemplates/d"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a de-identify template only beside an inspect template", func() {
		msg := minimal()
		msg.Spec.FilterConfig.SdpSettings = &GcpModelArmorFloorSettingSdpSettings{AdvancedConfig: &GcpModelArmorFloorSettingSdpAdvancedConfig{
			DeidentifyTemplate: "projects/p/locations/l/deidentifyTemplates/d",
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
