package gcpdeliverypipelinev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDeliveryPipelineSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

func smokeTask() *GcpDeliveryPipelineTask {
	return &GcpDeliveryPipelineTask{Container: &GcpDeliveryPipelineContainerTask{
		Image:   "us-docker.pkg.dev/acme/tools/smoke:1",
		Command: []string{"/bin/smoke"},
		Args:    []string{"--fast"},
		Env:     map[string]string{"MODE": "ci"},
	}}
}

func analysis() *GcpDeliveryPipelineAnalysis {
	return &GcpDeliveryPipelineAnalysis{
		Duration: "600s",
		GoogleCloud: &GcpDeliveryPipelineGoogleCloudAnalysis{AlertPolicyChecks: []*GcpDeliveryPipelineAlertPolicyCheck{
			{Id: "errors", AlertPolicies: []*foreignkeyv1.StringValueOrRef{reference(catalogkind.CatalogKind_GcpMonitoringAlertPolicy, "web-errors")}},
		}},
		CustomChecks: []*GcpDeliveryPipelineCustomCheck{{Id: "latency", Frequency: "60s", Task: smokeTask()}},
	}
}

var _ = ginkgo.Describe("GcpDeliveryPipelineSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDeliveryPipeline {
		return &GcpDeliveryPipeline{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDeliveryPipeline",
			Metadata:   &shared.CatalogObjectMetadata{Name: "web"},
			Spec:       &GcpDeliveryPipelineSpec{Location: "us-central1"},
		}
	}

	full := func() *GcpDeliveryPipeline {
		msg := minimal()
		s := msg.Spec
		s.ProjectId = reference(catalogkind.CatalogKind_GcpProject, "apps")
		s.DeliveryPipelineId = "web"
		s.Description = "web app promotion"
		s.Labels = map[string]string{"team": "web"}
		s.Annotations = map[string]string{"owner": "web"}
		s.SerialPipeline = &GcpDeliveryPipelineSerialPipeline{Stages: []*GcpDeliveryPipelineStage{
			{
				TargetId: reference(catalogkind.CatalogKind_GcpDeployTarget, "web-dev"),
				Profiles: []string{"dev"},
				DeployParameters: []*GcpDeliveryPipelineDeployParameters{
					{Values: map[string]string{"replicas": "1"}, MatchTargetLabels: map[string]string{"tier": "dev"}},
				},
				Strategy: &GcpDeliveryPipelineStrategy{Standard: &GcpDeliveryPipelineStandard{
					Verify:       true,
					Predeploy:    &GcpDeliveryPipelineStandardDeployHook{Actions: []string{"migrate"}},
					Postdeploy:   &GcpDeliveryPipelineStandardDeployHook{Tasks: []*GcpDeliveryPipelineTask{smokeTask()}},
					VerifyConfig: &GcpDeliveryPipelineVerifyConfig{Tasks: []*GcpDeliveryPipelineTask{smokeTask()}},
					Analysis:     analysis(),
				}},
			},
			{
				TargetId: literal("web-staging"),
				Strategy: &GcpDeliveryPipelineStrategy{Canary: &GcpDeliveryPipelineCanary{
					CanaryDeployment: &GcpDeliveryPipelineCanaryDeployment{
						Percentages:  []int32{25, 50},
						Verify:       true,
						Predeploy:    &GcpDeliveryPipelineCanaryDeployHook{Actions: []string{"warm"}},
						VerifyConfig: &GcpDeliveryPipelineVerifyConfig{Tasks: []*GcpDeliveryPipelineTask{smokeTask()}},
						Analysis:     analysis(),
					},
					RuntimeConfig: &GcpDeliveryPipelineRuntimeConfig{CloudRun: &GcpDeliveryPipelineCloudRunConfig{
						AutomaticTrafficControl: true,
						CanaryRevisionTags:      []string{"canary"},
					}},
				}},
			},
			{
				TargetId: reference(catalogkind.CatalogKind_GcpDeployTarget, "web-prod"),
				Strategy: &GcpDeliveryPipelineStrategy{Canary: &GcpDeliveryPipelineCanary{
					CustomCanaryDeployment: &GcpDeliveryPipelineCustomCanaryDeployment{PhaseConfigs: []*GcpDeliveryPipelinePhaseConfig{
						{PhaseId: "canary-10", Percentage: 10, Profiles: []string{"canary"}, Verify: true, Analysis: analysis()},
						{PhaseId: "stable", Percentage: 100},
					}},
					RuntimeConfig: &GcpDeliveryPipelineRuntimeConfig{Kubernetes: &GcpDeliveryPipelineKubernetesConfig{
						GatewayServiceMesh: &GcpDeliveryPipelineGatewayServiceMesh{
							HttpRoute: "web", Service: "web", Deployment: "web",
							RouteUpdateWaitTime: "60s", StableCutbackDuration: "300s", PodSelectorLabel: "app",
							RouteDestinations: &GcpDeliveryPipelineRouteDestinations{DestinationIds: []string{"@self", "east"}, PropagateService: true},
						},
					}},
				}},
			},
		}}
		s.Automations = []*GcpDeliveryPipelineAutomation{{
			AutomationId:   "promote-and-repair",
			Description:    "promote dev to staging, repair failures",
			Labels:         map[string]string{"team": "web"},
			ServiceAccount: reference(catalogkind.CatalogKind_GcpServiceAccount, "web-deployer"),
			Selector: &GcpDeliveryPipelineAutomationSelector{Targets: []*GcpDeliveryPipelineAutomationTarget{
				{Id: reference(catalogkind.CatalogKind_GcpDeployTarget, "web-dev")},
				{Id: literal("*"), Labels: map[string]string{"tier": "prod"}},
			}},
			Rules: []*GcpDeliveryPipelineAutomationRule{
				{PromoteReleaseRule: &GcpDeliveryPipelinePromoteReleaseRule{Id: "promote", Wait: "600s", DestinationTargetId: literal("@next")}},
				{AdvanceRolloutRule: &GcpDeliveryPipelineAdvanceRolloutRule{Id: "advance", SourcePhases: []string{"canary-10"}}},
				{RepairRolloutRule: &GcpDeliveryPipelineRepairRolloutRule{Id: "repair", RepairPhases: []*GcpDeliveryPipelineRepairPhase{
					{Retry: &GcpDeliveryPipelineRetry{Attempts: "3", Wait: "60s", BackoffMode: "BACKOFF_MODE_EXPONENTIAL"}},
					{Rollback: &GcpDeliveryPipelineRollback{}},
				}}},
				{TimedPromoteReleaseRule: &GcpDeliveryPipelineTimedPromoteReleaseRule{Id: "weekly", Schedule: "0 9 * * 1", TimeZone: "America/New_York",
					DestinationTargetId: reference(catalogkind.CatalogKind_GcpDeployTarget, "web-prod")}},
			},
		}}
		s.DeletionPolicy = "PREVENT"
		return msg
	}

	expectInvalid := func(mutate func(*GcpDeliveryPipelineSpec)) {
		msg := full()
		mutate(msg.Spec)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	}

	stage := func(s *GcpDeliveryPipelineSpec, i int) *GcpDeliveryPipelineStage {
		return s.SerialPipeline.Stages[i]
	}
	automation := func(s *GcpDeliveryPipelineSpec) *GcpDeliveryPipelineAutomation {
		return s.Automations[0]
	}

	ginkgo.It("should accept a minimal pipeline and a fully declared one", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(full())).To(gomega.Succeed())
	})

	ginkgo.It("should require a location and reject a malformed pipeline ID, a long description, or an unknown deletion policy", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.Location = "" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.DeliveryPipelineId = "Web" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.DeliveryPipelineId = "1web" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.DeliveryPipelineId = "web-" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.Description = string(make([]byte, 256)) })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { s.DeletionPolicy = "KEEP" })
	})

	ginkgo.It("should take a stage target as a bare ID only", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 1).TargetId = literal("projects/apps/locations/us-central1/targets/web-staging")
		})
	})

	ginkgo.It("should require deploy-parameter values", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { stage(s, 0).DeployParameters[0].Values = nil })
	})

	ginkgo.It("should accept a strategy with both arms left to Google, but refuse actions and tasks on one job", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 0).Strategy.Standard.Predeploy.Tasks = []*GcpDeliveryPipelineTask{smokeTask()}
		})
		msg := full()
		stage(msg.Spec, 0).Strategy.Standard.Postdeploy = &GcpDeliveryPipelineStandardDeployHook{Actions: []string{"notify"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should refuse both canary deployment forms", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.CanaryDeployment = &GcpDeliveryPipelineCanaryDeployment{Percentages: []int32{50}}
		})
	})

	ginkgo.It("should require automatic traffic control for a Cloud Run canary_deployment only", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 1).Strategy.Canary.RuntimeConfig.CloudRun.AutomaticTrafficControl = false
		})
		msg := full()
		stage(msg.Spec, 2).Strategy.Canary.RuntimeConfig = &GcpDeliveryPipelineRuntimeConfig{CloudRun: &GcpDeliveryPipelineCloudRunConfig{}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should refuse two runtimes or two Kubernetes networking arms", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 1).Strategy.Canary.RuntimeConfig.Kubernetes = &GcpDeliveryPipelineKubernetesConfig{}
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.RuntimeConfig.Kubernetes.ServiceNetworking = &GcpDeliveryPipelineServiceNetworking{Service: "web", Deployment: "web"}
		})
		msg := full()
		stage(msg.Spec, 2).Strategy.Canary.RuntimeConfig.Kubernetes = &GcpDeliveryPipelineKubernetesConfig{
			ServiceNetworking: &GcpDeliveryPipelineServiceNetworking{Service: "web", Deployment: "web", DisablePodOverprovisioning: true},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require the Kubernetes object names and route destinations", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.RuntimeConfig.Kubernetes.GatewayServiceMesh.HttpRoute = ""
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.RuntimeConfig.Kubernetes.GatewayServiceMesh.RouteDestinations.DestinationIds = nil
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.RuntimeConfig.Kubernetes = &GcpDeliveryPipelineKubernetesConfig{
				ServiceNetworking: &GcpDeliveryPipelineServiceNetworking{Service: "web"},
			}
		})
	})

	ginkgo.It("should bound canary percentages and phases", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { stage(s, 1).Strategy.Canary.CanaryDeployment.Percentages = nil })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 1).Strategy.Canary.CanaryDeployment.Percentages = []int32{25, 101}
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.CustomCanaryDeployment.PhaseConfigs = nil
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.CustomCanaryDeployment.PhaseConfigs[1].PhaseId = "canary-10"
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.CustomCanaryDeployment.PhaseConfigs[0].PhaseId = "Canary"
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 2).Strategy.Canary.CustomCanaryDeployment.PhaseConfigs[0].Percentage = 120
		})
	})

	ginkgo.It("should require a task image, an analysis duration, and check IDs", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 0).Strategy.Standard.VerifyConfig.Tasks[0].Container.Image = ""
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { stage(s, 0).Strategy.Standard.Analysis.Duration = "" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 0).Strategy.Standard.Analysis.GoogleCloud.AlertPolicyChecks[0].Id = ""
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			stage(s, 0).Strategy.Standard.Analysis.GoogleCloud.AlertPolicyChecks[0].AlertPolicies = nil
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { stage(s, 0).Strategy.Standard.Analysis.CustomChecks[0].Id = "" })
		msg := full()
		stage(msg.Spec, 0).Strategy.Standard.VerifyConfig.Tasks = []*GcpDeliveryPipelineTask{{}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should key automations by a unique, required ID", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).AutomationId = "" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			s.Automations = append(s.Automations, &GcpDeliveryPipelineAutomation{
				AutomationId:   "promote-and-repair",
				ServiceAccount: literal("deployer@apps.iam.gserviceaccount.com"),
				Selector:       &GcpDeliveryPipelineAutomationSelector{Targets: []*GcpDeliveryPipelineAutomationTarget{{Id: literal("*")}}},
				Rules:          []*GcpDeliveryPipelineAutomationRule{{AdvanceRolloutRule: &GcpDeliveryPipelineAdvanceRolloutRule{Id: "advance"}}},
			})
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Description = string(make([]byte, 256)) })
	})

	ginkgo.It("should require an automation's service account, selector, targets, and rules", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).ServiceAccount = nil })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Selector = nil })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Selector.Targets = nil })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules = nil })
	})

	ginkgo.It("should take selector and destination targets as bare IDs, \"*\", or \"@next\"", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Selector.Targets[1].Id = literal("projects/apps/locations/us-central1/targets/web-prod")
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[0].PromoteReleaseRule.DestinationTargetId = literal("projects/apps/locations/us-central1/targets/web-prod")
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[3].TimedPromoteReleaseRule.DestinationTargetId = literal("locations/us-central1/targets/web-prod")
		})
	})

	ginkgo.It("should require exactly one kind per rule and unique, well-formed rule IDs", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[0].AdvanceRolloutRule = &GcpDeliveryPipelineAdvanceRolloutRule{Id: "advance-2"}
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[0] = &GcpDeliveryPipelineAutomationRule{} })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[1].AdvanceRolloutRule.Id = "promote" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[0].PromoteReleaseRule.Id = "Promote" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[1].AdvanceRolloutRule.Id = "" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[2].RepairRolloutRule.Id = "repair-" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[3].TimedPromoteReleaseRule.Id = "9weekly" })
	})

	ginkgo.It("should require a repair rule's phases, each exactly one of retry or rollback", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[2].RepairRolloutRule.RepairPhases = nil })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[2].RepairRolloutRule.RepairPhases[0].Rollback = &GcpDeliveryPipelineRollback{}
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[2].RepairRolloutRule.RepairPhases[1] = &GcpDeliveryPipelineRepairPhase{}
		})
	})

	ginkgo.It("should require a whole-number retry count and a known backoff mode", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[2].RepairRolloutRule.RepairPhases[0].Retry.Attempts = ""
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[2].RepairRolloutRule.RepairPhases[0].Retry.Attempts = "three"
		})
		expectInvalid(func(s *GcpDeliveryPipelineSpec) {
			automation(s).Rules[2].RepairRolloutRule.RepairPhases[0].Retry.BackoffMode = "EXPONENTIAL"
		})
	})

	ginkgo.It("should require a timed promotion's schedule and time zone", func() {
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[3].TimedPromoteReleaseRule.Schedule = "" })
		expectInvalid(func(s *GcpDeliveryPipelineSpec) { automation(s).Rules[3].TimedPromoteReleaseRule.TimeZone = "" })
	})
})
