package gcpvertexaisearchenginev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiSearchEngineSpec Suite")
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

const storeName = "projects/ai-project/locations/global/collections/default_collection/dataStores/product-catalog"

var _ = ginkgo.Describe("GcpVertexAiSearchEngineSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The smallest search engine: an id and one data store by reference.
	search := func() *GcpVertexAiSearchEngine {
		return &GcpVertexAiSearchEngine{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiSearchEngine",
			Metadata:   &shared.CloudResourceMetadata{Name: "product-search"},
			Spec: &GcpVertexAiSearchEngineSpec{
				Location:     "global",
				DataStoreIds: []*foreignkeyv1.StringValueOrRef{nameRef("product-catalog")},
			},
		}
	}

	chat := func() *GcpVertexAiSearchEngine {
		msg := search()
		msg.Spec.EngineType = "CHAT"
		msg.Spec.ChatEngineConfig = &GcpVertexAiSearchEngineChatEngineConfig{
			AgentCreationConfig: &GcpVertexAiSearchEngineAgentCreationConfig{
				Business:            "Acme",
				DefaultLanguageCode: "en",
				TimeZone:            "America/Los_Angeles",
			},
		}
		return msg
	}

	recommendation := func() *GcpVertexAiSearchEngine {
		msg := search()
		msg.Spec.EngineType = "RECOMMENDATION"
		msg.Spec.IndustryVertical = "MEDIA"
		msg.Spec.MediaRecommendationEngineConfig = &GcpVertexAiSearchEngineMediaRecommendationEngineConfig{
			Type:                  "recommended-for-you",
			OptimizationObjective: "ctr",
			TrainingState:         "PAUSED",
			EngineFeaturesConfig: &GcpVertexAiSearchEngineEngineFeaturesConfig{
				RecommendedForYouConfig: &GcpVertexAiSearchEngineRecommendedForYouConfig{ContextEventType: "generic"},
			},
		}
		return msg
	}

	synonyms := func(id string) *GcpVertexAiSearchEngineControl {
		return &GcpVertexAiSearchEngineControl{
			ControlId:      id,
			DisplayName:    "Synonyms",
			UseCases:       []string{"SEARCH_USE_CASE_SEARCH"},
			SynonymsAction: &GcpVertexAiSearchEngineSynonymsAction{Synonyms: []string{"laptop", "notebook"}},
		}
	}

	ginkgo.It("should accept the smallest search engine, and an explicit SEARCH type", func() {
		gomega.Expect(validator.Validate(search())).To(gomega.Succeed())
		msg := search()
		msg.Spec.EngineType = "SEARCH"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every search-arm lever, controls wired through the serving config, a widget, and an assistant", func() {
		msg := search()
		msg.Spec.ProjectId = litRef("ai-project")
		msg.Spec.EngineId = "product-search-v2"
		msg.Spec.DisplayName = "Product search"
		msg.Spec.CollectionId = litRef("default_collection")
		msg.Spec.IndustryVertical = "GENERIC"
		msg.Spec.CommonConfig = &GcpVertexAiSearchEngineCommonConfig{CompanyName: "Acme"}
		msg.Spec.SearchEngineConfig = &GcpVertexAiSearchEngineSearchEngineConfig{
			SearchTier:               "SEARCH_TIER_ENTERPRISE",
			SearchAddOns:             []string{"SEARCH_ADD_ON_LLM"},
			RequiredSubscriptionTier: "SUBSCRIPTION_TIER_SEARCH",
		}
		msg.Spec.AppType = "APP_TYPE_INTRANET"
		msg.Spec.DisableAnalytics = true
		msg.Spec.Features = map[string]string{"disable-agent-sharing": "FEATURE_STATE_OFF"}
		msg.Spec.KmsKeyName = litRef("projects/ai-project/locations/global/keyRings/ai/cryptoKeys/search")
		msg.Spec.KnowledgeGraphConfig = &GcpVertexAiSearchEngineKnowledgeGraphConfig{
			EnableCloudKnowledgeGraph: proto.Bool(true),
			FeatureConfig:             &GcpVertexAiSearchEngineKnowledgeGraphFeatureConfig{DisablePrivateKgEnrichment: true},
		}
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{
			synonyms("syn"),
			{
				ControlId:   "boost-docs",
				DisplayName: "Boost docs",
				Conditions: []*GcpVertexAiSearchEngineControlCondition{{
					QueryTerms: []*GcpVertexAiSearchEngineQueryTerm{{Value: "install", FullMatch: true}},
				}},
				BoostAction: &GcpVertexAiSearchEngineBoostAction{
					DataStore:  litRef(storeName),
					Filter:     `(category: ANY("docs"))`,
					FixedBoost: proto.Float32(0.5),
				},
			},
			{
				ControlId:   "fresh",
				DisplayName: "Fresh first",
				BoostAction: &GcpVertexAiSearchEngineBoostAction{
					DataStore: litRef(storeName),
					Filter:    `(category: ANY("news"))`,
					InterpolationBoostSpec: &GcpVertexAiSearchEngineInterpolationBoostSpec{
						FieldName:         "published",
						AttributeType:     "FRESHNESS",
						InterpolationType: "LINEAR",
						ControlPoint:      &GcpVertexAiSearchEngineControlPoint{AttributeValue: "7d", BoostAmount: proto.Float32(0.8)},
					},
				},
			},
			{
				ControlId:    "only-public",
				DisplayName:  "Only public",
				FilterAction: &GcpVertexAiSearchEngineFilterAction{DataStore: litRef(storeName), Filter: `(visibility: ANY("public"))`},
			},
			{
				ControlId:   "promo",
				DisplayName: "Promo",
				PromoteAction: &GcpVertexAiSearchEnginePromoteAction{
					DataStore:           litRef(storeName),
					SearchLinkPromotion: &GcpVertexAiSearchEngineSearchLinkPromotion{Title: "Getting started", Uri: "https://example.com/start", Enabled: true},
				},
			},
			{
				ControlId:      "redirect-pricing",
				DisplayName:    "Redirect pricing",
				Conditions:     []*GcpVertexAiSearchEngineControlCondition{{QueryRegex: "^pricing$"}},
				RedirectAction: &GcpVertexAiSearchEngineRedirectAction{RedirectUri: "https://example.com/pricing"},
			},
		}
		msg.Spec.ServingConfig = &GcpVertexAiSearchEngineServingConfig{
			BoostControlIds:    []string{"boost-docs", "fresh"},
			FilterControlIds:   []string{"only-public"},
			PromoteControlIds:  []string{"promo"},
			RedirectControlIds: []string{"redirect-pricing"},
			SynonymsControlIds: []string{"syn"},
		}
		msg.Spec.WidgetConfig = &GcpVertexAiSearchEngineWidgetConfig{
			AccessSettings: &GcpVertexAiSearchEngineWidgetAccessSettings{AllowPublicAccess: true, AllowlistedDomains: []string{"www.example.com"}},
			UiBranding:     &GcpVertexAiSearchEngineWidgetUiBranding{LogoUrl: "https://example.com/logo.png"},
			UiSettings: &GcpVertexAiSearchEngineWidgetUiSettings{
				InteractionType:       "SEARCH_WITH_ANSWER",
				ResultDescriptionType: "SNIPPET",
				EnableAutocomplete:    true,
				DataStoreUiConfigs: []*GcpVertexAiSearchEngineWidgetDataStoreUiConfig{{
					Name:        litRef(storeName),
					FacetFields: []*GcpVertexAiSearchEngineWidgetFacetField{{Field: "category", DisplayName: "Category"}},
					FieldsUiComponentsMap: []*GcpVertexAiSearchEngineWidgetFieldUiComponent{{
						UiComponent: "title", Field: "title", DeviceVisibility: []string{"DESKTOP", "MOBILE"},
					}},
				}},
				GenerativeAnswerConfig: &GcpVertexAiSearchEngineWidgetGenerativeAnswerConfig{
					ResultCount:      proto.Int32(5),
					MaxRephraseSteps: proto.Int32(2),
					ImageSource:      "CORPUS_IMAGE_ONLY",
				},
			},
		}
		msg.Spec.Assistants = []*GcpVertexAiSearchEngineAssistant{{
			AssistantId:      "default_assistant",
			DisplayName:      "Acme assistant",
			WebGroundingType: "WEB_GROUNDING_TYPE_GOOGLE_SEARCH",
			CustomerPolicy: &GcpVertexAiSearchEngineCustomerPolicy{
				BannedPhrases: []*GcpVertexAiSearchEngineBannedPhrase{{Phrase: "confidential", MatchType: "WORD_BOUNDARY_STRING_MATCH"}},
				ModelArmorConfig: &GcpVertexAiSearchEngineModelArmorConfig{
					UserPromptTemplate: "projects/ai-project/locations/global/templates/prompts",
					ResponseTemplate:   "projects/ai-project/locations/global/templates/responses",
					FailureMode:        "FAIL_CLOSED",
				},
			},
			GenerationConfig: &GcpVertexAiSearchEngineGenerationConfig{DefaultLanguage: "en", AdditionalSystemInstruction: "Answer briefly."},
		}}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a chat engine that creates its agent, or links one", func() {
		gomega.Expect(validator.Validate(chat())).To(gomega.Succeed())
		msg := chat()
		msg.Spec.ChatEngineConfig = &GcpVertexAiSearchEngineChatEngineConfig{
			DialogflowAgentToLink: "projects/ai-project/locations/global/agents/abc-123",
			AllowCrossRegion:      true,
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a media recommendation engine", func() {
		gomega.Expect(validator.Validate(recommendation())).To(gomega.Succeed())
	})

	ginkgo.It("should require location from Google's list and at least one data store", func() {
		msg := search()
		msg.Spec.Location = "us-central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.DataStoreIds = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown engine type and a bad engine id", func() {
		msg := search()
		msg.Spec.EngineType = "ANSWER"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.EngineId = "Product_Search"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require the chat config on the chat arm and forbid it elsewhere", func() {
		msg := chat()
		msg.Spec.ChatEngineConfig = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.ChatEngineConfig = chat().Spec.ChatEngineConfig
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one agent source on the chat arm and a well-formed agent name", func() {
		msg := chat()
		msg.Spec.ChatEngineConfig.DialogflowAgentToLink = "projects/ai-project/locations/global/agents/abc-123"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = chat()
		msg.Spec.ChatEngineConfig = &GcpVertexAiSearchEngineChatEngineConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = chat()
		msg.Spec.ChatEngineConfig = &GcpVertexAiSearchEngineChatEngineConfig{DialogflowAgentToLink: "abc-123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = chat()
		msg.Spec.ChatEngineConfig.AgentCreationConfig.TimeZone = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should wall search-only fields off the chat and recommendation arms", func() {
		msg := chat()
		msg.Spec.SearchEngineConfig = &GcpVertexAiSearchEngineSearchEngineConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = chat()
		msg.Spec.AppType = "APP_TYPE_INTRANET"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = recommendation()
		msg.Spec.KmsKeyName = litRef("projects/p/locations/global/keyRings/r/cryptoKeys/k")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = recommendation()
		msg.Spec.Features = map[string]string{"x": "FEATURE_STATE_ON"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = chat()
		msg.Spec.KnowledgeGraphConfig = &GcpVertexAiSearchEngineKnowledgeGraphConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should keep a recommendation engine in default_collection with one data store and no FHIR vertical", func() {
		msg := recommendation()
		msg.Spec.CollectionId = litRef("jira-collection")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = recommendation()
		msg.Spec.CollectionId = litRef("default_collection")
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg = recommendation()
		msg.Spec.DataStoreIds = append(msg.Spec.DataStoreIds, nameRef("second"))
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = recommendation()
		msg.Spec.IndustryVertical = "HEALTHCARE_FHIR"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = recommendation()
		msg.Spec.MediaRecommendationEngineConfig.EngineFeaturesConfig.MostPopularConfig = &GcpVertexAiSearchEngineMostPopularConfig{TimeWindowDays: proto.Int32(7)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.MediaRecommendationEngineConfig = &GcpVertexAiSearchEngineMediaRecommendationEngineConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should keep a chat engine GENERIC", func() {
		msg := chat()
		msg.Spec.IndustryVertical = "MEDIA"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one action per control and unique ids", func() {
		msg := search()
		ctl := synonyms("dup")
		ctl.RedirectAction = &GcpVertexAiSearchEngineRedirectAction{RedirectUri: "https://example.com"}
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{ctl}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{{ControlId: "none", DisplayName: "None"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{synonyms("same"), synonyms("same")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one boost form and a bounded boost", func() {
		msg := search()
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{{
			ControlId: "b", DisplayName: "B",
			BoostAction: &GcpVertexAiSearchEngineBoostAction{DataStore: litRef(storeName), Filter: "x"},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Controls[0].BoostAction.FixedBoost = proto.Float32(1.5)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Controls[0].BoostAction.FixedBoost = proto.Float32(0.2)
		msg.Spec.Controls[0].BoostAction.InterpolationBoostSpec = &GcpVertexAiSearchEngineInterpolationBoostSpec{FieldName: "f"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require every serving config id to name a declared control with the matching action", func() {
		msg := search()
		msg.Spec.Controls = []*GcpVertexAiSearchEngineControl{synonyms("syn")}
		msg.Spec.ServingConfig = &GcpVertexAiSearchEngineServingConfig{SynonymsControlIds: []string{"syn"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.ServingConfig = &GcpVertexAiSearchEngineServingConfig{SynonymsControlIds: []string{"missing"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ServingConfig = &GcpVertexAiSearchEngineServingConfig{BoostControlIds: []string{"syn"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject widget, feature, and assistant values outside Google's lists", func() {
		msg := search()
		msg.Spec.Features = map[string]string{"x": "ON"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.WidgetConfig = &GcpVertexAiSearchEngineWidgetConfig{UiSettings: &GcpVertexAiSearchEngineWidgetUiSettings{InteractionType: "CHAT"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.WidgetConfig = &GcpVertexAiSearchEngineWidgetConfig{UiSettings: &GcpVertexAiSearchEngineWidgetUiSettings{
			GenerativeAnswerConfig: &GcpVertexAiSearchEngineWidgetGenerativeAnswerConfig{ResultCount: proto.Int32(11)},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.Assistants = []*GcpVertexAiSearchEngineAssistant{{AssistantId: "a", DisplayName: "A", WebGroundingType: "BING"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = search()
		msg.Spec.Assistants = []*GcpVertexAiSearchEngineAssistant{{AssistantId: "a", DisplayName: "A"}, {AssistantId: "a", DisplayName: "B"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := search()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
