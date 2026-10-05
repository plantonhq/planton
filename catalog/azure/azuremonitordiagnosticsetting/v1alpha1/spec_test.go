package azuremonitordiagnosticsettingv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorDiagnosticSettingSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorDiagnosticSettingSpec Validation Tests")
}

// buildValidDiagnosticSetting returns a minimal valid resource; tests mutate
// copies of it to probe individual rules.
func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func buildValidDiagnosticSetting() *AzureMonitorDiagnosticSetting {
	return &AzureMonitorDiagnosticSetting{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorDiagnosticSetting",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-diag",
		},
		Spec: &AzureMonitorDiagnosticSettingSpec{
			SettingName: "route-to-law",
			TargetResourceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/my-vault",
				},
			},
			EnabledLogs: []*AzureMonitorDiagnosticSettingLog{
				{Category: "AuditEvent"},
			},
			LogAnalyticsWorkspaceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorDiagnosticSettingSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(buildValidDiagnosticSetting())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a category group entry", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledLogs = []*AzureMonitorDiagnosticSettingLog{
				{CategoryGroup: "allLogs"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a metrics-only setting with a storage destination", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledLogs = nil
			input.Spec.LogAnalyticsWorkspaceId = nil
			input.Spec.EnabledMetrics = []*AzureMonitorDiagnosticSettingMetric{
				{Category: "AllMetrics"},
			}
			input.Spec.StorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/archive",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an event hub destination with a named hub", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.LogAnalyticsWorkspaceId = nil
			input.Spec.EventhubAuthorizationRuleId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/ns/authorizationRules/RootManageSharedAccessKey")
			input.Spec.EventhubName = literal("diagnostics")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a partner solution destination", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.LogAnalyticsWorkspaceId = nil
			input.Spec.PartnerSolutionId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Elastic/monitors/elastic"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the dedicated destination type", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.LogAnalyticsDestinationType = AzureMonitorDiagnosticSettingLogAnalyticsDestinationType_DEDICATED
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a target reference by valueFrom", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.TargetResourceId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-vault", FieldPath: "status.outputs.key_vault_id"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing setting name", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.SettingName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a setting name with forbidden characters", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.SettingName = "route/to/law"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("may not contain"))
		})

		ginkgo.It("should reject a missing target resource", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.TargetResourceId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a setting with no logs and no metrics", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledLogs = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one log or metric"))
		})

		ginkgo.It("should reject a log entry with both category and category group", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledLogs = []*AzureMonitorDiagnosticSettingLog{
				{Category: "AuditEvent", CategoryGroup: "allLogs"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
		})

		ginkgo.It("should reject a log entry with neither category nor category group", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledLogs = []*AzureMonitorDiagnosticSettingLog{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a metric entry without a category", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EnabledMetrics = []*AzureMonitorDiagnosticSettingMetric{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a setting with no destination", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.LogAnalyticsWorkspaceId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("route the telemetry somewhere"))
		})

		ginkgo.It("should reject an event hub name without an authorization rule", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.EventhubName = literal("diagnostics")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("authorization rule"))
		})

		ginkgo.It("should reject an undefined destination type enum number", func() {
			input := buildValidDiagnosticSetting()
			input.Spec.LogAnalyticsDestinationType = AzureMonitorDiagnosticSettingLogAnalyticsDestinationType(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
