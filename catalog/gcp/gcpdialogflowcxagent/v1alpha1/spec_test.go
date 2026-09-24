package gcpdialogflowcxagentv1alpha1

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
	ginkgo.RunSpecs(t, "GcpDialogflowCxAgentSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

const secretVersion = "projects/support-project/secrets/webhook-password/versions/latest"

var _ = ginkgo.Describe("GcpDialogflowCxAgentSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDialogflowCxAgent {
		return &GcpDialogflowCxAgent{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDialogflowCxAgent",
			Metadata:   &shared.CloudResourceMetadata{Name: "support-agent"},
			Spec: &GcpDialogflowCxAgentSpec{
				Location:            "global",
				DefaultLanguageCode: "en",
				TimeZone:            "America/New_York",
			},
		}
	}

	webService := func() *GcpDialogflowCxAgentGenericWebService {
		return &GcpDialogflowCxAgentGenericWebService{Uri: "https://orders.example.com/webhook"}
	}

	full := func() *GcpDialogflowCxAgent {
		msg := minimal()
		msg.Spec.Webhooks = []*GcpDialogflowCxAgentWebhook{
			{DisplayName: "order-lookup", GenericWebService: webService()},
			{
				DisplayName: "private-billing",
				ServiceDirectory: &GcpDialogflowCxAgentServiceDirectory{
					Service:           "projects/p/locations/us-central1/namespaces/billing/services/api",
					GenericWebService: webService(),
				},
			},
		}
		msg.Spec.Tools = []*GcpDialogflowCxAgentTool{
			{
				DisplayName:  "orders-api",
				Description:  "Looks up orders by id.",
				OpenApiSpec:  &GcpDialogflowCxAgentToolOpenApiSpec{TextSchema: "openapi: 3.0.0"},
				Versions: []*GcpDialogflowCxAgentToolVersion{{
					DisplayName: "v1",
					Tool: &GcpDialogflowCxAgentToolSnapshot{
						DisplayName: "orders-api",
						Description: "Looks up orders by id.",
						OpenApiSpec: &GcpDialogflowCxAgentToolOpenApiSpec{TextSchema: "openapi: 3.0.0"},
					},
				}},
			},
			{
				DisplayName: "policy-search",
				Description: "Searches the policy handbook.",
				DataStoreSpec: &GcpDialogflowCxAgentToolDataStoreSpec{
					DataStoreConnections: []*GcpDialogflowCxAgentToolDataStoreConnection{{
						DataStore:     nameRef("policy-handbook"),
						DataStoreType: "UNSTRUCTURED",
					}},
				},
			},
		}
		msg.Spec.Versions = []*GcpDialogflowCxAgentVersion{{DisplayName: "release-1"}}
		msg.Spec.Environments = []*GcpDialogflowCxAgentEnvironment{{
			DisplayName:    "production",
			VersionConfigs: []*GcpDialogflowCxAgentEnvironmentVersionConfig{{Version: "release-1"}},
		}}
		msg.Spec.GenerativeSettings = []*GcpDialogflowCxAgentGenerativeSettings{{
			LanguageCode: "en",
			KnowledgeConnectorSettings: &GcpDialogflowCxAgentKnowledgeConnectorSettings{
				Business: "Acme",
			},
		}}
		return msg
	}

	ginkgo.It("should accept an agent alone", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an agent with every folded child", func() {
		gomega.Expect(validator.Validate(full())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every agent-level setting", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("support-project")
		msg.Spec.Location = "us-central1"
		msg.Spec.DisplayName = "Support agent"
		msg.Spec.Description = "Answers order questions."
		msg.Spec.AvatarUri = "https://example.com/avatar.png"
		msg.Spec.SupportedLanguageCodes = []string{"es", "fr"}
		msg.Spec.EnableMultiLanguageTraining = true
		msg.Spec.EnableSpellCorrection = true
		msg.Spec.SecuritySettings = nameRef("pii-redaction")
		msg.Spec.StartWithDefaultPlaybook = true
		msg.Spec.AdvancedSettings = &GcpDialogflowCxAgentAdvancedSettings{
			AudioExportGcsDestination: &GcpDialogflowCxAgentAudioExportGcsDestination{Uri: "gs://support-audio/calls"},
			DtmfSettings:              &GcpDialogflowCxAgentDtmfSettings{Enabled: true, FinishDigit: "#", MaxDigits: 6},
			LoggingSettings:           &GcpDialogflowCxAgentLoggingSettings{EnableInteractionLogging: true},
			SpeechSettings: &GcpDialogflowCxAgentSpeechSettings{
				EndpointerSensitivity: 30,
				Models:                map[string]string{"en": "phone_call"},
				NoSpeechTimeout:       "3.5s",
			},
		}
		msg.Spec.EnableAnswerFeedback = true
		msg.Spec.ClientCertificateSettings = &GcpDialogflowCxAgentClientCertificateSettings{
			SslCertificate: "-----BEGIN CERTIFICATE-----",
			PrivateKey:     "projects/p/secrets/client-key/versions/1",
		}
		msg.Spec.GenAppBuilderSettings = &GcpDialogflowCxAgentGenAppBuilderSettings{
			Engine: litRef("projects/p/locations/global/collections/default_collection/engines/support"),
		}
		msg.Spec.DeleteChatEngineOnDestroy = true
		msg.Spec.GitIntegrationSettings = &GcpDialogflowCxAgentGitIntegrationSettings{
			GithubSettings: &GcpDialogflowCxAgentGithubSettings{
				RepositoryUri:  "https://api.github.com/repos/acme/support-bot",
				TrackingBranch: "main",
				AccessToken:    "ghp_example",
			},
		}
		msg.Spec.DefaultEndUserMetadata = `{"tier": "$session.params.tier"}`
		msg.Spec.EnableSpeechAdaptation = true
		msg.Spec.SynthesizeSpeechConfigs = `{"en": {"speakingRate": 1.1}}`
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location, a default language, and a time zone", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DefaultLanguageCode = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.TimeZone = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject the default language among the supported languages", func() {
		msg := minimal()
		msg.Spec.SupportedLanguageCodes = []string{"es", "en"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse to delete an engine another block owns, and allow a literal one", func() {
		msg := minimal()
		msg.Spec.GenAppBuilderSettings = &GcpDialogflowCxAgentGenAppBuilderSettings{Engine: nameRef("support-chat")}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.DeleteChatEngineOnDestroy = true
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a webhook to be exactly one endpoint", func() {
		msg := full()
		msg.Spec.Webhooks[0].ServiceDirectory = msg.Spec.Webhooks[1].ServiceDirectory
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Webhooks[0].GenericWebService = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require https and reject values outside Google's lists on a web service", func() {
		msg := full()
		msg.Spec.Webhooks[0].GenericWebService.Uri = "http://orders.example.com/webhook"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Webhooks[0].GenericWebService.WebhookType = "CUSTOM"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Webhooks[0].GenericWebService.HttpMethod = "TRACE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Webhooks[0].GenericWebService.ServiceAgentAuth = "SAML"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept every web service authentication and header setting", func() {
		msg := full()
		service := msg.Spec.Webhooks[0].GenericWebService
		service.WebhookType = "FLEXIBLE"
		service.HttpMethod = "GET"
		service.RequestBody = `{"order": "$session.params.order_id"}`
		service.ParameterMapping = map[string]string{"status": "$.status"}
		service.RequestHeaders = map[string]string{"X-Source": "dialogflow"}
		service.SecretVersionsForRequestHeaders = []*GcpDialogflowCxAgentSecretHeader{{Key: "X-Api-Key", SecretVersion: secretVersion}}
		service.SecretVersionForUsernamePassword = secretVersion
		service.OauthConfig = &GcpDialogflowCxAgentOauthConfig{
			ClientId:                     "client",
			TokenEndpoint:                "https://auth.example.com/token",
			SecretVersionForClientSecret: secretVersion,
			Scopes:                       []string{"orders.read"},
		}
		service.ServiceAgentAuth = "ID_TOKEN"
		service.ServiceAccount = nameRef("webhook-caller")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed Secret Manager version or Service Directory service", func() {
		msg := full()
		msg.Spec.Webhooks[0].GenericWebService.SecretVersionForUsernamePassword = "webhook-password"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Webhooks[1].ServiceDirectory.Service = "billing-api"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a tool and its snapshot to be exactly one specification", func() {
		msg := full()
		msg.Spec.Tools[0].FunctionSpec = &GcpDialogflowCxAgentToolFunctionSpec{InputSchema: "{}"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Tools[0].OpenApiSpec = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Tools[0].Versions[0].Tool.DataStoreSpec = msg.Spec.Tools[1].DataStoreSpec
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept one authentication method and reject two", func() {
		msg := full()
		msg.Spec.Tools[0].OpenApiSpec.Authentication = &GcpDialogflowCxAgentToolAuthentication{
			ApiKeyConfig: &GcpDialogflowCxAgentToolApiKeyConfig{KeyName: "X-Api-Key", RequestLocation: "HEADER", SecretVersionForApiKey: secretVersion},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Tools[0].OpenApiSpec.Authentication.BearerTokenConfig = &GcpDialogflowCxAgentToolBearerTokenConfig{Token: "t"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a data store tool to search at least one store", func() {
		msg := full()
		msg.Spec.Tools[1].DataStoreSpec.DataStoreConnections = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject duplicate keys among webhooks, tools, versions, environments, and languages", func() {
		msg := full()
		msg.Spec.Webhooks[1].DisplayName = "order-lookup"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Tools[1].DisplayName = "orders-api"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Tools[0].Versions = append(msg.Spec.Tools[0].Versions, msg.Spec.Tools[0].Versions[0])
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Versions = append(msg.Spec.Versions, &GcpDialogflowCxAgentVersion{
			FlowId:      "00000000-0000-0000-0000-000000000000",
			DisplayName: "release-1",
		})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Environments = append(msg.Spec.Environments, msg.Spec.Environments[0])
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.GenerativeSettings = append(msg.Spec.GenerativeSettings, &GcpDialogflowCxAgentGenerativeSettings{LanguageCode: "en"})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should allow the same version name on different flows", func() {
		msg := full()
		msg.Spec.Versions = append(msg.Spec.Versions, &GcpDialogflowCxAgentVersion{
			FlowId:      "4f1c8a2e-0000-4000-8000-000000000001",
			DisplayName: "release-1",
		})
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require an environment to name declared versions or outside version ids", func() {
		msg := full()
		msg.Spec.Environments[0].VersionConfigs[0].Version = "release-2"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Environments[0].VersionConfigs[0].FlowId = "4f1c8a2e-0000-4000-8000-000000000001"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Environments[0].VersionConfigs[0] = &GcpDialogflowCxAgentEnvironmentVersionConfig{VersionId: "3"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Environments[0].VersionConfigs[0].Version = "release-1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = full()
		msg.Spec.Environments[0].VersionConfigs = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should hold Google's length and range limits", func() {
		msg := full()
		msg.Spec.Versions[0].DisplayName = "a-version-display-name-that-runs-well-past-the-sixty-four-character-limit"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AdvancedSettings = &GcpDialogflowCxAgentAdvancedSettings{
			SpeechSettings: &GcpDialogflowCxAgentSpeechSettings{EndpointerSensitivity: 101},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AdvancedSettings = &GcpDialogflowCxAgentAdvancedSettings{
			SpeechSettings: &GcpDialogflowCxAgentSpeechSettings{NoSpeechTimeout: "3.5"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
