package azuremonitordatacollectionrulev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorDataCollectionRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorDataCollectionRuleSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func strPtr(v string) *string { return &v }

const (
	testWorkspaceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.OperationalInsights/workspaces/obs-law"
	testEventHubId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.EventHub/namespaces/obs-ns/eventhubs/obs-hub"
	testStorageId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.Storage/storageAccounts/obssa"
	testMonitorId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.Monitor/accounts/obs-prom"
	testDceId       = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.Insights/dataCollectionEndpoints/obs-dce"
	testIdentityId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/obs-uai"
)

// lawDestinations returns a destinations block with one Log Analytics
// workspace destination named "law-dest".
func lawDestinations() *AzureMonitorDataCollectionRuleDestinations {
	return &AzureMonitorDataCollectionRuleDestinations{
		LogAnalytics: []*AzureMonitorDataCollectionRuleLogAnalytics{
			{
				Name:                "law-dest",
				WorkspaceResourceId: literal(testWorkspaceId),
			},
		},
	}
}

// syslogSource returns a valid Linux syslog source.
func syslogSource() *AzureMonitorDataCollectionRuleSyslog {
	return &AzureMonitorDataCollectionRuleSyslog{
		Name:          "linux-syslog",
		FacilityNames: []string{"auth", "daemon"},
		LogLevels:     []string{"Warning", "Error", "Critical", "Alert", "Emergency"},
		Streams:       []string{"Microsoft-Syslog"},
	}
}

// validResource returns a minimal valid rule (one workspace
// destination, one syslog flow) that individual cases mutate into the
// shape under test.
func validResource() *AzureMonitorDataCollectionRule {
	return &AzureMonitorDataCollectionRule{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorDataCollectionRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-dcr",
		},
		Spec: &AzureMonitorDataCollectionRuleSpec{
			ResourceGroup: literal("obs-rg"),
			Name:          "linux-logs",
			Region:        "eastus",
			DataSources: &AzureMonitorDataCollectionRuleDataSources{
				Syslogs: []*AzureMonitorDataCollectionRuleSyslog{syslogSource()},
			},
			Destinations: lawDestinations(),
			DataFlows: []*AzureMonitorDataCollectionRuleDataFlow{
				{
					Streams:      []string{"Microsoft-Syslog"},
					Destinations: []string{"law-dest"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorDataCollectionRuleSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_monitor_data_collection_rule", func() {

			ginkgo.It("should not return a validation error for the minimal syslog-to-workspace rule", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every rule kind token", func() {
				for _, kind := range []string{"Linux", "Windows", "AgentDirectToStore", "WorkspaceTransforms"} {
					input := validResource()
					input.Spec.Kind = strPtr(kind)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "kind %q should be valid", kind)
				}
			})

			ginkgo.It("should accept a rule without data_sources", func() {
				input := validResource()
				input.Spec.DataSources = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data collection endpoint reference and a description", func() {
				input := validResource()
				input.Spec.DataCollectionEndpointId = literal(testDceId)
				input.Spec.Description = "collects auth and daemon syslog"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("data sources", func() {

			ginkgo.It("should accept every syslog facility token", func() {
				facilities := []string{
					"*", "alert", "audit", "auth", "authpriv", "clock", "cron",
					"daemon", "ftp", "kern", "local0", "local1", "local2",
					"local3", "local4", "local5", "local6", "local7", "lpr",
					"mail", "mark", "news", "nopri", "ntp", "syslog", "user",
					"uucp",
				}
				for _, facility := range facilities {
					input := validResource()
					input.Spec.DataSources.Syslogs[0].FacilityNames = []string{facility}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "facility %q should be valid", facility)
				}
			})

			ginkgo.It("should accept every syslog log level token", func() {
				levels := []string{"*", "Alert", "Critical", "Debug", "Emergency", "Error", "Info", "Notice", "Warning"}
				for _, level := range levels {
					input := validResource()
					input.Spec.DataSources.Syslogs[0].LogLevels = []string{level}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "log level %q should be valid", level)
				}
			})

			ginkgo.It("should accept a performance counter source", func() {
				input := validResource()
				input.Spec.DataSources.PerformanceCounters = []*AzureMonitorDataCollectionRulePerformanceCounter{
					{
						Name:                       "vm-perf",
						SamplingFrequencyInSeconds: int32Ptr(60),
						CounterSpecifiers:          []string{"\\Processor(_Total)\\% Processor Time"},
						Streams:                    []string{"Microsoft-Perf"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept sampling frequency at both bounds", func() {
				for _, frequency := range []int32{1, 1800} {
					input := validResource()
					input.Spec.DataSources.PerformanceCounters = []*AzureMonitorDataCollectionRulePerformanceCounter{
						{
							Name:                       "vm-perf",
							SamplingFrequencyInSeconds: int32Ptr(frequency),
							CounterSpecifiers:          []string{"\\*"},
							Streams:                    []string{"Microsoft-Perf"},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "sampling frequency %d should be valid", frequency)
				}
			})

			ginkgo.It("should accept a Windows event log source", func() {
				input := validResource()
				input.Spec.DataSources.WindowsEventLogs = []*AzureMonitorDataCollectionRuleWindowsEventLog{
					{
						Name:         "win-events",
						XPathQueries: []string{"System!*[System[(Level=1 or Level=2 or Level=3)]]"},
						Streams:      []string{"Microsoft-Event"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an extension source with settings and inputs", func() {
				input := validResource()
				input.Spec.DataSources.Extensions = []*AzureMonitorDataCollectionRuleExtension{
					{
						Name:             "dependency-agent",
						ExtensionName:    "DependencyAgent",
						ExtensionJson:    `{"sampleRate": 5}`,
						InputDataSources: []string{"vm-perf"},
						Streams:          []string{"Microsoft-ServiceMap"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IIS log source with and without directories", func() {
				for _, directories := range [][]string{nil, {"C:\\inetpub\\logs\\LogFiles\\W3SVC1"}} {
					input := validResource()
					input.Spec.DataSources.IisLogs = []*AzureMonitorDataCollectionRuleIisLog{
						{
							Name:           "iis-access",
							LogDirectories: directories,
							Streams:        []string{"Microsoft-W3CIISLog"},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a JSON log file source", func() {
				input := validResource()
				input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
					{
						Name:         "app-json-logs",
						FilePatterns: []string{"/var/log/myapp/*.json"},
						Format:       "json",
						Streams:      []string{"Custom-MyAppLogs"},
					},
				}
				input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{customStream("Custom-MyAppLogs")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every text record timestamp format", func() {
				formats := []string{
					"ISO 8601",
					"YYYY-MM-DD HH:MM:SS",
					"M/D/YYYY HH:MM:SS AM/PM",
					"Mon DD, YYYY HH:MM:SS",
					"yyMMdd HH:mm:ss",
					"ddMMyy HH:mm:ss",
					"MMM d hh:mm:ss",
					"dd/MMM/yyyy:HH:mm:ss zzz",
					"yyyy-MM-ddTHH:mm:ssK",
				}
				for _, format := range formats {
					input := validResource()
					input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
						{
							Name:         "app-text-logs",
							FilePatterns: []string{"/var/log/myapp/*.log"},
							Format:       "text",
							Settings: &AzureMonitorDataCollectionRuleLogFileSettings{
								Text: &AzureMonitorDataCollectionRuleLogFileSettingsText{
									RecordStartTimestampFormat: format,
								},
							},
							Streams: []string{"Custom-MyAppLogs"},
						},
					}
					input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{customStream("Custom-MyAppLogs")}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "timestamp format %q should be valid", format)
				}
			})

			ginkgo.It("should accept a Prometheus forwarder source with a label filter", func() {
				input := validResource()
				input.Spec.DataSources.PrometheusForwarders = []*AzureMonitorDataCollectionRulePrometheusForwarder{
					{
						Name: "prom-forwarder",
						LabelIncludeFilters: []*AzureMonitorDataCollectionRuleLabelIncludeFilter{
							{Label: "microsoft_metrics_include_label", Value: "monitored"},
						},
						Streams: []string{"Microsoft-PrometheusMetrics"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept firewall log and platform telemetry sources", func() {
				input := validResource()
				input.Spec.DataSources.WindowsFirewallLogs = []*AzureMonitorDataCollectionRuleWindowsFirewallLog{
					{Name: "fw-logs", Streams: []string{"Microsoft-WindowsFirewall"}},
				}
				input.Spec.DataSources.PlatformTelemetries = []*AzureMonitorDataCollectionRulePlatformTelemetry{
					{Name: "platform", Streams: []string{"Microsoft.Cache/redis:Metrics-Group-All"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an Event Hub data import source", func() {
				input := validResource()
				input.Spec.DataSources.DataImport = &AzureMonitorDataCollectionRuleDataImport{
					EventHubDataSource: &AzureMonitorDataCollectionRuleDataImportEventHub{
						Name:          "hub-import",
						Stream:        "Custom-HubEvents",
						ConsumerGroup: "$Default",
					},
				}
				input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{customStream("Custom-HubEvents")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("destinations and flows", func() {

			ginkgo.It("should accept every destination arm populated together", func() {
				input := validResource()
				input.Spec.Destinations = &AzureMonitorDataCollectionRuleDestinations{
					LogAnalytics: []*AzureMonitorDataCollectionRuleLogAnalytics{
						{Name: "law-dest", WorkspaceResourceId: literal(testWorkspaceId)},
					},
					AzureMonitorMetrics: &AzureMonitorDataCollectionRuleAzureMonitorMetrics{Name: "metrics-dest"},
					EventHub: &AzureMonitorDataCollectionRuleEventHubDestination{
						Name:       "hub-dest",
						EventHubId: literal(testEventHubId),
					},
					EventHubDirect: &AzureMonitorDataCollectionRuleEventHubDestination{
						Name:       "hub-direct-dest",
						EventHubId: literal(testEventHubId),
					},
					MonitorAccounts: []*AzureMonitorDataCollectionRuleMonitorAccount{
						{Name: "prom-dest", MonitorAccountId: literal(testMonitorId)},
					},
					StorageBlobs: []*AzureMonitorDataCollectionRuleStorageBlobDestination{
						{Name: "blob-dest", ContainerName: "telemetry", StorageAccountId: literal(testStorageId)},
					},
					StorageBlobDirects: []*AzureMonitorDataCollectionRuleStorageBlobDestination{
						{Name: "blob-direct-dest", ContainerName: "telemetry-direct", StorageAccountId: literal(testStorageId)},
					},
					StorageTableDirects: []*AzureMonitorDataCollectionRuleStorageTableDirect{
						{Name: "table-direct-dest", TableName: "telemetry", StorageAccountId: literal(testStorageId)},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a flow carrying a KQL transformation", func() {
				input := validResource()
				input.Spec.DataFlows[0].TransformKql = "source | where SeverityLevel != 'verbose'"
				input.Spec.DataFlows[0].OutputStream = "Microsoft-Syslog"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("identity and stream declarations", func() {

			ginkgo.It("should accept a system-assigned identity without ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMonitorDataCollectionRuleIdentity{
					Type: AzureMonitorDataCollectionRuleIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with one id", func() {
				input := validResource()
				input.Spec.Identity = &AzureMonitorDataCollectionRuleIdentity{
					Type:        AzureMonitorDataCollectionRuleIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every stream declaration column type", func() {
				for _, columnType := range []string{"boolean", "datetime", "dynamic", "int", "long", "real", "string"} {
					input := validResource()
					input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{
						{
							StreamName: "Custom-Typed",
							Columns: []*AzureMonitorDataCollectionRuleStreamDeclarationColumn{
								{Name: "Value", Type: columnType},
							},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "column type %q should be valid", columnType)
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("root fields", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown rule kind", func() {
				input := validResource()
				input.Spec.Kind = strPtr("LinuxAndWindows")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing destinations block", func() {
				input := validResource()
				input.Spec.Destinations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty destinations block", func() {
				input := validResource()
				input.Spec.Destinations = &AzureMonitorDataCollectionRuleDestinations{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one destination"))
			})

			ginkgo.It("should reject a rule without data flows", func() {
				input := validResource()
				input.Spec.DataFlows = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("data flows", func() {

			ginkgo.It("should reject a flow without streams", func() {
				input := validResource()
				input.Spec.DataFlows[0].Streams = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a flow without destinations", func() {
				input := validResource()
				input.Spec.DataFlows[0].Destinations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("data sources", func() {

			ginkgo.It("should reject a source name longer than 32 characters", func() {
				input := validResource()
				input.Spec.DataSources.Syslogs[0].Name = strings.Repeat("x", 33)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a syslog source without facilities", func() {
				input := validResource()
				input.Spec.DataSources.Syslogs[0].FacilityNames = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown syslog facility", func() {
				input := validResource()
				input.Spec.DataSources.Syslogs[0].FacilityNames = []string{"kernel"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown syslog log level", func() {
				input := validResource()
				input.Spec.DataSources.Syslogs[0].LogLevels = []string{"Fatal"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a syslog source without streams", func() {
				input := validResource()
				input.Spec.DataSources.Syslogs[0].Streams = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject sampling frequency outside 1-1800", func() {
				for _, bad := range []int32{0, 1801} {
					input := validResource()
					input.Spec.DataSources.PerformanceCounters = []*AzureMonitorDataCollectionRulePerformanceCounter{
						{
							Name:                       "vm-perf",
							SamplingFrequencyInSeconds: int32Ptr(bad),
							CounterSpecifiers:          []string{"\\*"},
							Streams:                    []string{"Microsoft-Perf"},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil(), "sampling frequency %d should be rejected", bad)
				}
			})

			ginkgo.It("should reject a performance counter source missing its sampling frequency", func() {
				input := validResource()
				input.Spec.DataSources.PerformanceCounters = []*AzureMonitorDataCollectionRulePerformanceCounter{
					{
						Name:              "vm-perf",
						CounterSpecifiers: []string{"\\*"},
						Streams:           []string{"Microsoft-Perf"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a Windows event log source without XPath queries", func() {
				input := validResource()
				input.Spec.DataSources.WindowsEventLogs = []*AzureMonitorDataCollectionRuleWindowsEventLog{
					{Name: "win-events", Streams: []string{"Microsoft-Event"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an extension source without its extension name", func() {
				input := validResource()
				input.Spec.DataSources.Extensions = []*AzureMonitorDataCollectionRuleExtension{
					{Name: "ext", Streams: []string{"Microsoft-ServiceMap"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a log file source with an unknown format", func() {
				input := validResource()
				input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
					{
						Name:         "app-logs",
						FilePatterns: []string{"/var/log/*.log"},
						Format:       "csv",
						Streams:      []string{"Custom-MyAppLogs"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a log file source without file patterns", func() {
				input := validResource()
				input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
					{
						Name:    "app-logs",
						Format:  "json",
						Streams: []string{"Custom-MyAppLogs"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject text settings with an unknown timestamp format", func() {
				input := validResource()
				input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
					{
						Name:         "app-logs",
						FilePatterns: []string{"/var/log/*.log"},
						Format:       "text",
						Settings: &AzureMonitorDataCollectionRuleLogFileSettings{
							Text: &AzureMonitorDataCollectionRuleLogFileSettingsText{
								RecordStartTimestampFormat: "unix-epoch",
							},
						},
						Streams: []string{"Custom-MyAppLogs"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a settings block without its text settings", func() {
				input := validResource()
				input.Spec.DataSources.LogFiles = []*AzureMonitorDataCollectionRuleLogFile{
					{
						Name:         "app-logs",
						FilePatterns: []string{"/var/log/*.log"},
						Format:       "text",
						Settings:     &AzureMonitorDataCollectionRuleLogFileSettings{},
						Streams:      []string{"Custom-MyAppLogs"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a Prometheus forwarder on a non-Prometheus stream", func() {
				input := validResource()
				input.Spec.DataSources.PrometheusForwarders = []*AzureMonitorDataCollectionRulePrometheusForwarder{
					{Name: "prom", Streams: []string{"Microsoft-Perf"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a label filter with an unknown label", func() {
				input := validResource()
				input.Spec.DataSources.PrometheusForwarders = []*AzureMonitorDataCollectionRulePrometheusForwarder{
					{
						Name: "prom",
						LabelIncludeFilters: []*AzureMonitorDataCollectionRuleLabelIncludeFilter{
							{Label: "cluster", Value: "prod"},
						},
						Streams: []string{"Microsoft-PrometheusMetrics"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data import without its event hub source", func() {
				input := validResource()
				input.Spec.DataSources.DataImport = &AzureMonitorDataCollectionRuleDataImport{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an event hub import without its stream", func() {
				input := validResource()
				input.Spec.DataSources.DataImport = &AzureMonitorDataCollectionRuleDataImport{
					EventHubDataSource: &AzureMonitorDataCollectionRuleDataImportEventHub{
						Name: "hub-import",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("destinations", func() {

			ginkgo.It("should reject a workspace destination without the workspace id", func() {
				input := validResource()
				input.Spec.Destinations.LogAnalytics[0].WorkspaceResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a workspace destination without a name", func() {
				input := validResource()
				input.Spec.Destinations.LogAnalytics[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an event hub destination without the hub id", func() {
				input := validResource()
				input.Spec.Destinations.EventHub = &AzureMonitorDataCollectionRuleEventHubDestination{Name: "hub-dest"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a storage blob destination without a container", func() {
				input := validResource()
				input.Spec.Destinations.StorageBlobs = []*AzureMonitorDataCollectionRuleStorageBlobDestination{
					{Name: "blob-dest", StorageAccountId: literal(testStorageId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a storage table destination without a table name", func() {
				input := validResource()
				input.Spec.Destinations.StorageTableDirects = []*AzureMonitorDataCollectionRuleStorageTableDirect{
					{Name: "table-dest", StorageAccountId: literal(testStorageId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a monitor account destination without the account id", func() {
				input := validResource()
				input.Spec.Destinations.MonitorAccounts = []*AzureMonitorDataCollectionRuleMonitorAccount{
					{Name: "prom-dest"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("identity", func() {

			ginkgo.It("should reject an identity block without a flavor", func() {
				input := validResource()
				input.Spec.Identity = &AzureMonitorDataCollectionRuleIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMonitorDataCollectionRuleIdentity{
					Type:        AzureMonitorDataCollectionRuleIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMonitorDataCollectionRuleIdentity{
					Type: AzureMonitorDataCollectionRuleIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("stream declarations", func() {

			ginkgo.It("should reject duplicate stream declaration names", func() {
				input := validResource()
				input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{
					customStream("Custom-Dup"),
					customStream("Custom-Dup"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("unique"))
			})

			ginkgo.It("should reject a stream declaration without columns", func() {
				input := validResource()
				input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{
					{StreamName: "Custom-Empty"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a column with an unknown type", func() {
				input := validResource()
				input.Spec.StreamDeclarations = []*AzureMonitorDataCollectionRuleStreamDeclaration{
					{
						StreamName: "Custom-Bad",
						Columns: []*AzureMonitorDataCollectionRuleStreamDeclarationColumn{
							{Name: "Value", Type: "float"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})

// customStream returns a minimal valid custom stream declaration.
func customStream(name string) *AzureMonitorDataCollectionRuleStreamDeclaration {
	return &AzureMonitorDataCollectionRuleStreamDeclaration{
		StreamName: name,
		Columns: []*AzureMonitorDataCollectionRuleStreamDeclarationColumn{
			{Name: "TimeGenerated", Type: "datetime"},
			{Name: "RawData", Type: "string"},
		},
	}
}
