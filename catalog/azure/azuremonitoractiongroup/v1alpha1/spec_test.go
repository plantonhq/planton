package azuremonitoractiongroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorActionGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorActionGroupSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// buildValidActionGroup returns a minimal valid resource; tests mutate copies
// of it to probe individual rules. An action group with zero receivers is
// legal (a "null" routing target).
func buildValidActionGroup() *AzureMonitorActionGroup {
	return &AzureMonitorActionGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorActionGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ag",
		},
		Spec: &AzureMonitorActionGroupSpec{
			ResourceGroup: literal("test-resource-group"),
			ShortName:     "TestAG",
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorActionGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for a receiver-less group", func() {
			err := protovalidate.Validate(buildValidActionGroup())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept every receiver type together", func() {
			input := buildValidActionGroup()
			input.Spec.EmailReceivers = []*AzureMonitorActionGroupEmailReceiver{
				{Name: "oncall-email", EmailAddress: "oncall@example.com", UseCommonAlertSchema: true},
			}
			input.Spec.SmsReceivers = []*AzureMonitorActionGroupSmsReceiver{
				{Name: "oncall-sms", CountryCode: "1", PhoneNumber: "5555550100"},
			}
			input.Spec.VoiceReceivers = []*AzureMonitorActionGroupVoiceReceiver{
				{Name: "oncall-voice", CountryCode: "1", PhoneNumber: "5555550100"},
			}
			input.Spec.WebhookReceivers = []*AzureMonitorActionGroupWebhookReceiver{
				{
					Name:                 "pager",
					ServiceUri:           "https://events.example.com/hook",
					UseCommonAlertSchema: true,
					AadAuth: &AzureMonitorActionGroupWebhookAadAuth{
						ObjectId: "11111111-2222-3333-4444-555555555555",
						TenantId: "99999999-8888-7777-6666-555555555555",
					},
				},
			}
			input.Spec.AzureAppPushReceivers = []*AzureMonitorActionGroupAzureAppPushReceiver{
				{Name: "oncall-push", EmailAddress: "oncall@example.com"},
			}
			input.Spec.AutomationRunbookReceivers = []*AzureMonitorActionGroupAutomationRunbookReceiver{
				{
					Name:                 "remediate",
					AutomationAccountId:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Automation/automationAccounts/aa",
					RunbookName:          "restart-app",
					WebhookResourceId:    "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Automation/automationAccounts/aa/webHooks/wh",
					IsGlobalRunbook:      false,
					ServiceUri:           "https://s1events.azure-automation.net/webhooks?token=abc",
					UseCommonAlertSchema: true,
				},
			}
			input.Spec.LogicAppReceivers = []*AzureMonitorActionGroupLogicAppReceiver{
				{
					Name:        "workflow",
					ResourceId:  "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Logic/workflows/wf",
					CallbackUrl: "https://prod-1.eastus.logic.azure.com/workflows/abc/triggers/manual/paths/invoke",
				},
			}
			input.Spec.AzureFunctionReceivers = []*AzureMonitorActionGroupAzureFunctionReceiver{
				{
					Name:                  "func",
					FunctionAppResourceId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/sites/fn"),
					FunctionName:          "HandleAlert",
					HttpTriggerUrl:        "https://fn.azurewebsites.net/api/HandleAlert",
				},
			}
			input.Spec.ArmRoleReceivers = []*AzureMonitorActionGroupArmRoleReceiver{
				{Name: "owners", RoleId: literal("8e3af657-a8ff-443c-a75c-2fe8c4bcb635"), UseCommonAlertSchema: true},
			}
			input.Spec.EventHubReceivers = []*AzureMonitorActionGroupEventHubReceiver{
				{
					Name:              "siem",
					EventHubName:      literal("alerts"),
					EventHubNamespace: literal("alerts-ns"),
					SubscriptionId:    "00000000-0000-0000-0000-000000000000",
				},
			}
			input.Spec.ItsmReceivers = []*AzureMonitorActionGroupItsmReceiver{
				{
					Name:                "servicenow",
					WorkspaceId:         "00000000-0000-0000-0000-000000000000|11111111-1111-1111-1111-111111111111",
					ConnectionId:        "22222222-2222-2222-2222-222222222222",
					TicketConfiguration: "{\"PayloadRevision\":0,\"WorkItemType\":\"Incident\"}",
					Region:              "eastus",
				},
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a 12-character short name", func() {
			input := buildValidActionGroup()
			input.Spec.ShortName = "abcdefghijkl"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing resource group", func() {
			input := buildValidActionGroup()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing short name", func() {
			input := buildValidActionGroup()
			input.Spec.ShortName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a short name longer than 12 characters", func() {
			input := buildValidActionGroup()
			input.Spec.ShortName = "abcdefghijklm"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an email receiver without an address", func() {
			input := buildValidActionGroup()
			input.Spec.EmailReceivers = []*AzureMonitorActionGroupEmailReceiver{
				{Name: "oncall-email"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an SMS receiver without a phone number", func() {
			input := buildValidActionGroup()
			input.Spec.SmsReceivers = []*AzureMonitorActionGroupSmsReceiver{
				{Name: "oncall-sms", CountryCode: "1"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a managed-secret reference as a webhook URI", func() {
			// The receiver URLs are sensitive (they embed their credential), so
			// consuming platforms store a secret reference in them -- the
			// http(s) shape is taught in the field comment, never rule-enforced.
			input := buildValidActionGroup()
			input.Spec.WebhookReceivers = []*AzureMonitorActionGroupWebhookReceiver{
				{Name: "pager", ServiceUri: "${secrets-group/alerting/pagerduty-webhook-url}"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject webhook Entra auth with a non-UUID object id", func() {
			input := buildValidActionGroup()
			input.Spec.WebhookReceivers = []*AzureMonitorActionGroupWebhookReceiver{
				{
					Name:       "pager",
					ServiceUri: "https://events.example.com/hook",
					AadAuth:    &AzureMonitorActionGroupWebhookAadAuth{ObjectId: "not-a-uuid"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject webhook Entra auth with a non-UUID tenant id", func() {
			input := buildValidActionGroup()
			input.Spec.WebhookReceivers = []*AzureMonitorActionGroupWebhookReceiver{
				{
					Name:       "pager",
					ServiceUri: "https://events.example.com/hook",
					AadAuth: &AzureMonitorActionGroupWebhookAadAuth{
						ObjectId: "11111111-2222-3333-4444-555555555555",
						TenantId: "not-a-uuid",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a runbook receiver missing its webhook resource", func() {
			input := buildValidActionGroup()
			input.Spec.AutomationRunbookReceivers = []*AzureMonitorActionGroupAutomationRunbookReceiver{
				{
					Name:                "remediate",
					AutomationAccountId: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Automation/automationAccounts/aa",
					RunbookName:         "restart-app",
					ServiceUri:          "https://s1events.azure-automation.net/webhooks?token=abc",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a managed-secret reference as a function trigger URL", func() {
			// The trigger URL's code parameter IS the function key, so the field
			// is sensitive and platforms store a secret reference in it.
			input := buildValidActionGroup()
			input.Spec.AzureFunctionReceivers = []*AzureMonitorActionGroupAzureFunctionReceiver{
				{
					Name:                  "func",
					FunctionAppResourceId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Web/sites/fn"),
					FunctionName:          "HandleAlert",
					HttpTriggerUrl:        "${secrets-group/alerting/handle-alert-trigger-url}",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject an ARM role receiver without a role", func() {
			input := buildValidActionGroup()
			input.Spec.ArmRoleReceivers = []*AzureMonitorActionGroupArmRoleReceiver{
				{Name: "owners"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an event hub receiver with a non-UUID subscription", func() {
			input := buildValidActionGroup()
			input.Spec.EventHubReceivers = []*AzureMonitorActionGroupEventHubReceiver{
				{
					Name:              "siem",
					EventHubName:      literal("alerts"),
					EventHubNamespace: literal("alerts-ns"),
					SubscriptionId:    "not-a-uuid",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an ITSM receiver whose ticket configuration lacks the required keys", func() {
			input := buildValidActionGroup()
			input.Spec.ItsmReceivers = []*AzureMonitorActionGroupItsmReceiver{
				{
					Name:                "servicenow",
					WorkspaceId:         "sub|workspace",
					ConnectionId:        "22222222-2222-2222-2222-222222222222",
					TicketConfiguration: "{\"WorkItemType\":\"Incident\"}",
					Region:              "eastus",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("PayloadRevision"))
		})
	})
})
