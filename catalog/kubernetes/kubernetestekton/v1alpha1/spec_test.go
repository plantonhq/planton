package kubernetestektonv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestKubernetesTekton(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTekton Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }
func strPtr(s string) *string { return &s }

var _ = ginkgo.Describe("KubernetesTekton Validation Tests", func() {
	var input *KubernetesTekton

	ginkgo.BeforeEach(func() {
		input = &KubernetesTekton{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTekton",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "tekton",
			},
			Spec: &KubernetesTektonSpec{},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("an empty spec should be valid (profile all, upstream defaults)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every profile value should be valid", func() {
			for _, p := range []string{"lite", "basic", "all"} {
				input.Spec.Profile = strPtr(p)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("a custom target namespace with metadata should be valid", func() {
			input.Spec.TargetNamespace = strPtr("ci-system")
			input.Spec.TargetNamespaceMetadata = &KubernetesTektonNamespaceMetadata{
				Labels:      map[string]string{"pod-security.kubernetes.io/enforce": "baseline"},
				Annotations: map[string]string{"team": "platform"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a full pipeline block should be valid", func() {
			input.Spec.Pipeline = &KubernetesTektonPipeline{
				CloudEventsSinkUrl:    "http://receiver.ci.svc.cluster.local/events",
				EnableApiFields:       strPtr("beta"),
				DefaultTimeoutMinutes: int32Ptr(120),
				DefaultServiceAccount: "pipelines-runner",
				Features: &KubernetesTektonPipelineFeatures{
					DisableCredsInit:                         boolPtr(true),
					AwaitSidecarReadiness:                    boolPtr(true),
					RequireGitSshSecretKnownHosts:            boolPtr(true),
					EnableCustomTasks:                        boolPtr(true),
					EnableProvenanceInStatus:                 boolPtr(true),
					SetSecurityContext:                       boolPtr(true),
					EnableCelInWhenexpression:                boolPtr(true),
					EnableStepActions:                        boolPtr(true),
					EnableParamEnum:                          boolPtr(true),
					ResultsFrom:                              strPtr("sidecar-logs"),
					MaxResultSize:                            int32Ptr(8192),
					Coschedule:                               strPtr("workspaces"),
					RunningInEnvironmentWithInjectedSidecars: boolPtr(false),
				},
				Resolvers: &KubernetesTektonPipelineResolvers{
					EnableBundlesResolver: boolPtr(true),
					EnableHubResolver:     boolPtr(false),
					EnableGitResolver:     boolPtr(true),
					EnableClusterResolver: boolPtr(true),
				},
				Metrics: &KubernetesTektonPipelineMetrics{
					TaskrunLevel:            strPtr("task"),
					TaskrunDurationType:     strPtr("histogram"),
					PipelinerunLevel:        strPtr("pipeline"),
					PipelinerunDurationType: strPtr("histogram"),
					CountWithReason:         boolPtr(true),
				},
				Performance: &KubernetesTektonPipelinePerformance{
					Replicas:             int32Ptr(2),
					Buckets:              int32Ptr(2),
					ThreadsPerController: int32Ptr(4),
					KubeApiQps:           int32Ptr(50),
					KubeApiBurst:         int32Ptr(100),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("trigger, dashboard and chain blocks should be valid", func() {
			input.Spec.Trigger = &KubernetesTektonTrigger{
				EnableApiFields:       strPtr("stable"),
				DefaultServiceAccount: "triggers-sa",
			}
			input.Spec.Dashboard = &KubernetesTektonDashboard{
				Readonly:     true,
				ExternalLogs: "https://logs.example.com",
			}
			input.Spec.Chain = &KubernetesTektonChain{Disabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a pruner with keep should be valid", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Schedule:  "0 8 * * *",
				Resources: []string{"pipelinerun", "taskrun"},
				Keep:      int32Ptr(100),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a pruner with keep_since should be valid", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Schedule:         "0 */6 * * *",
				Resources:        []string{"pipelinerun"},
				KeepSince:        int32Ptr(1440),
				PrunePerResource: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("placement and additional params should be valid", func() {
			input.Spec.Placement = &KubernetesTektonPlacement{
				NodeSelector:      map[string]string{"role": "ci"},
				PriorityClassName: "platform-critical",
			}
			input.Spec.AdditionalParams = []*KubernetesTektonParam{
				{Name: "createRbacResource", Value: "false"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("an unknown profile should fail", func() {
			input.Spec.Profile = strPtr("full")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase target namespace should fail", func() {
			input.Spec.TargetNamespace = strPtr("Tekton")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a cloud events sink without a scheme should fail", func() {
			input.Spec.Pipeline = &KubernetesTektonPipeline{
				CloudEventsSinkUrl: "receiver.ci.svc.cluster.local/events",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown enable_api_fields value should fail", func() {
			input.Spec.Pipeline = &KubernetesTektonPipeline{
				EnableApiFields: strPtr("experimental"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pruner with BOTH keep and keep_since should fail", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Schedule:  "0 8 * * *",
				Resources: []string{"pipelinerun"},
				Keep:      int32Ptr(100),
				KeepSince: int32Ptr(1440),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pruner with NEITHER keep nor keep_since should fail", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Schedule:  "0 8 * * *",
				Resources: []string{"pipelinerun"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pruner without a schedule should fail", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Resources: []string{"pipelinerun"},
				Keep:      int32Ptr(100),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a pruner with an unknown resource should fail", func() {
			input.Spec.Pruner = &KubernetesTektonPruner{
				Schedule:  "0 8 * * *",
				Resources: []string{"deployments"},
				Keep:      int32Ptr(100),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("performance buckets above the upstream maximum should fail", func() {
			input.Spec.Pipeline = &KubernetesTektonPipeline{
				Performance: &KubernetesTektonPipelinePerformance{Buckets: int32Ptr(11)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown metrics level should fail", func() {
			input.Spec.Pipeline = &KubernetesTektonPipeline{
				Metrics: &KubernetesTektonPipelineMetrics{TaskrunLevel: strPtr("verbose")},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
