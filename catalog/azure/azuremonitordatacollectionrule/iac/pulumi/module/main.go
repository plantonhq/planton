package module

import (
	"github.com/pkg/errors"
	azuremonitordatacollectionrulev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitordatacollectionrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// identityTypeStrings maps the spec enum's values to the provider's
// identity tokens.
var identityTypeStrings = map[azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleIdentityType]string{
	azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleIdentityType_USER_ASSIGNED:   "UserAssigned",
}

func Resources(ctx *pulumi.Context, iacInput *azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorDataCollectionRule.Spec

	ruleArgs := &monitoring.DataCollectionRuleArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(spec.ResourceGroup.GetValue()),
		Location:          pulumi.String(spec.Region),
		Destinations:      buildDestinations(spec.Destinations),
		DataFlows:         buildDataFlows(spec.DataFlows),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted when unset -- the default rule kind accepts every
	// platform's sources. Once set, changing (or clearing) the kind
	// forces a new rule (provider lifecycle).
	if spec.Kind != nil && *spec.Kind != "" {
		ruleArgs.Kind = pulumi.String(*spec.Kind)
	}

	// Sent only when non-empty for a clean plan; Azure treats an absent
	// and an empty description identically.
	if spec.Description != "" {
		ruleArgs.Description = pulumi.String(spec.Description)
	}

	// The DCE the rule ingests through -- required by Azure when the rule
	// declares custom streams; sent only when set.
	if spec.DataCollectionEndpointId.GetValue() != "" {
		ruleArgs.DataCollectionEndpointId = pulumi.String(spec.DataCollectionEndpointId.GetValue())
	}

	if spec.Identity != nil {
		identityArgs := &monitoring.DataCollectionRuleIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		ruleArgs.Identity = identityArgs
	}

	if spec.DataSources != nil {
		ruleArgs.DataSources = buildDataSources(spec.DataSources)
	}

	if len(spec.StreamDeclarations) > 0 {
		declarations := monitoring.DataCollectionRuleStreamDeclarationArray{}
		for _, declaration := range spec.StreamDeclarations {
			columns := monitoring.DataCollectionRuleStreamDeclarationColumnArray{}
			for _, column := range declaration.Columns {
				columns = append(columns, &monitoring.DataCollectionRuleStreamDeclarationColumnArgs{
					Name: pulumi.String(column.Name),
					Type: pulumi.String(column.Type),
				})
			}
			declarations = append(declarations, &monitoring.DataCollectionRuleStreamDeclarationArgs{
				StreamName: pulumi.String(declaration.StreamName),
				Columns:    columns,
			})
		}
		ruleArgs.StreamDeclarations = declarations
	}

	createdRule, err := monitoring.NewDataCollectionRule(ctx,
		spec.Name,
		ruleArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create data collection rule %s", spec.Name)
	}

	ctx.Export(OpDataCollectionRuleId, createdRule.ID())
	ctx.Export(OpDataCollectionRuleName, createdRule.Name)
	ctx.Export(OpImmutableId, createdRule.ImmutableId)
	// Empty unless SYSTEM_ASSIGNED is enabled -- mirrors the TF module's
	// try(identity[0].principal_id, "").
	ctx.Export(OpIdentityPrincipalId, createdRule.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

	return nil
}

// buildDestinations mirrors the destinations block. ENGINE-SHAPE: the
// classic SDK types azure_monitor_metrics / event_hub /
// event_hub_direct as singular pointers -- the same one-entry cap the
// Terraform provider enforces with MaxItems, so both engines carry the
// identical shape.
func buildDestinations(destinations *azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleDestinations) *monitoring.DataCollectionRuleDestinationsArgs {
	destinationsArgs := &monitoring.DataCollectionRuleDestinationsArgs{}

	if len(destinations.LogAnalytics) > 0 {
		lawDestinations := monitoring.DataCollectionRuleDestinationsLogAnalyticArray{}
		for _, law := range destinations.LogAnalytics {
			lawDestinations = append(lawDestinations, &monitoring.DataCollectionRuleDestinationsLogAnalyticArgs{
				Name:                pulumi.String(law.Name),
				WorkspaceResourceId: pulumi.String(law.WorkspaceResourceId.GetValue()),
			})
		}
		destinationsArgs.LogAnalytics = lawDestinations
	}

	if destinations.AzureMonitorMetrics != nil {
		destinationsArgs.AzureMonitorMetrics = &monitoring.DataCollectionRuleDestinationsAzureMonitorMetricsArgs{
			Name: pulumi.String(destinations.AzureMonitorMetrics.Name),
		}
	}

	if destinations.EventHub != nil {
		destinationsArgs.EventHub = &monitoring.DataCollectionRuleDestinationsEventHubArgs{
			Name:       pulumi.String(destinations.EventHub.Name),
			EventHubId: pulumi.String(destinations.EventHub.EventHubId.GetValue()),
		}
	}

	if destinations.EventHubDirect != nil {
		destinationsArgs.EventHubDirect = &monitoring.DataCollectionRuleDestinationsEventHubDirectArgs{
			Name:       pulumi.String(destinations.EventHubDirect.Name),
			EventHubId: pulumi.String(destinations.EventHubDirect.EventHubId.GetValue()),
		}
	}

	if len(destinations.MonitorAccounts) > 0 {
		monitorAccounts := monitoring.DataCollectionRuleDestinationsMonitorAccountArray{}
		for _, account := range destinations.MonitorAccounts {
			monitorAccounts = append(monitorAccounts, &monitoring.DataCollectionRuleDestinationsMonitorAccountArgs{
				Name:             pulumi.String(account.Name),
				MonitorAccountId: pulumi.String(account.MonitorAccountId.GetValue()),
			})
		}
		destinationsArgs.MonitorAccounts = monitorAccounts
	}

	if len(destinations.StorageBlobs) > 0 {
		storageBlobs := monitoring.DataCollectionRuleDestinationsStorageBlobArray{}
		for _, blob := range destinations.StorageBlobs {
			storageBlobs = append(storageBlobs, &monitoring.DataCollectionRuleDestinationsStorageBlobArgs{
				Name:             pulumi.String(blob.Name),
				ContainerName:    pulumi.String(blob.ContainerName),
				StorageAccountId: pulumi.String(blob.StorageAccountId.GetValue()),
			})
		}
		destinationsArgs.StorageBlobs = storageBlobs
	}

	if len(destinations.StorageBlobDirects) > 0 {
		storageBlobDirects := monitoring.DataCollectionRuleDestinationsStorageBlobDirectArray{}
		for _, blob := range destinations.StorageBlobDirects {
			storageBlobDirects = append(storageBlobDirects, &monitoring.DataCollectionRuleDestinationsStorageBlobDirectArgs{
				Name:             pulumi.String(blob.Name),
				ContainerName:    pulumi.String(blob.ContainerName),
				StorageAccountId: pulumi.String(blob.StorageAccountId.GetValue()),
			})
		}
		destinationsArgs.StorageBlobDirects = storageBlobDirects
	}

	if len(destinations.StorageTableDirects) > 0 {
		storageTableDirects := monitoring.DataCollectionRuleDestinationsStorageTableDirectArray{}
		for _, table := range destinations.StorageTableDirects {
			storageTableDirects = append(storageTableDirects, &monitoring.DataCollectionRuleDestinationsStorageTableDirectArgs{
				Name:             pulumi.String(table.Name),
				TableName:        pulumi.String(table.TableName),
				StorageAccountId: pulumi.String(table.StorageAccountId.GetValue()),
			})
		}
		destinationsArgs.StorageTableDirects = storageTableDirects
	}

	return destinationsArgs
}

func buildDataFlows(dataFlows []*azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleDataFlow) monitoring.DataCollectionRuleDataFlowArray {
	flows := monitoring.DataCollectionRuleDataFlowArray{}
	for _, flow := range dataFlows {
		flowArgs := &monitoring.DataCollectionRuleDataFlowArgs{
			Streams:      pulumi.ToStringArray(flow.Streams),
			Destinations: pulumi.ToStringArray(flow.Destinations),
		}
		// The provider validates non-empty strings on all three -- each is
		// sent only when set.
		if flow.BuiltInTransform != "" {
			flowArgs.BuiltInTransform = pulumi.String(flow.BuiltInTransform)
		}
		if flow.OutputStream != "" {
			flowArgs.OutputStream = pulumi.String(flow.OutputStream)
		}
		if flow.TransformKql != "" {
			flowArgs.TransformKql = pulumi.String(flow.TransformKql)
		}
		flows = append(flows, flowArgs)
	}
	return flows
}

func buildDataSources(dataSources *azuremonitordatacollectionrulev1alpha1.AzureMonitorDataCollectionRuleDataSources) *monitoring.DataCollectionRuleDataSourcesArgs {
	dataSourcesArgs := &monitoring.DataCollectionRuleDataSourcesArgs{}

	if len(dataSources.Syslogs) > 0 {
		syslogs := monitoring.DataCollectionRuleDataSourcesSyslogArray{}
		for _, syslog := range dataSources.Syslogs {
			syslogs = append(syslogs, &monitoring.DataCollectionRuleDataSourcesSyslogArgs{
				Name:          pulumi.String(syslog.Name),
				FacilityNames: pulumi.ToStringArray(syslog.FacilityNames),
				LogLevels:     pulumi.ToStringArray(syslog.LogLevels),
				Streams:       pulumi.ToStringArray(syslog.Streams),
			})
		}
		dataSourcesArgs.Syslogs = syslogs
	}

	if len(dataSources.PerformanceCounters) > 0 {
		counters := monitoring.DataCollectionRuleDataSourcesPerformanceCounterArray{}
		for _, counter := range dataSources.PerformanceCounters {
			counters = append(counters, &monitoring.DataCollectionRuleDataSourcesPerformanceCounterArgs{
				Name:                       pulumi.String(counter.Name),
				SamplingFrequencyInSeconds: pulumi.Int(int(*counter.SamplingFrequencyInSeconds)),
				CounterSpecifiers:          pulumi.ToStringArray(counter.CounterSpecifiers),
				Streams:                    pulumi.ToStringArray(counter.Streams),
			})
		}
		dataSourcesArgs.PerformanceCounters = counters
	}

	if len(dataSources.WindowsEventLogs) > 0 {
		eventLogs := monitoring.DataCollectionRuleDataSourcesWindowsEventLogArray{}
		for _, eventLog := range dataSources.WindowsEventLogs {
			eventLogs = append(eventLogs, &monitoring.DataCollectionRuleDataSourcesWindowsEventLogArgs{
				Name:         pulumi.String(eventLog.Name),
				XPathQueries: pulumi.ToStringArray(eventLog.XPathQueries),
				Streams:      pulumi.ToStringArray(eventLog.Streams),
			})
		}
		dataSourcesArgs.WindowsEventLogs = eventLogs
	}

	if len(dataSources.Extensions) > 0 {
		extensions := monitoring.DataCollectionRuleDataSourcesExtensionArray{}
		for _, extension := range dataSources.Extensions {
			extensionArgs := &monitoring.DataCollectionRuleDataSourcesExtensionArgs{
				Name:          pulumi.String(extension.Name),
				ExtensionName: pulumi.String(extension.ExtensionName),
				Streams:       pulumi.ToStringArray(extension.Streams),
			}
			// The provider validates a non-empty JSON string -- sent only
			// when set.
			if extension.ExtensionJson != "" {
				extensionArgs.ExtensionJson = pulumi.String(extension.ExtensionJson)
			}
			if len(extension.InputDataSources) > 0 {
				extensionArgs.InputDataSources = pulumi.ToStringArray(extension.InputDataSources)
			}
			extensions = append(extensions, extensionArgs)
		}
		dataSourcesArgs.Extensions = extensions
	}

	if len(dataSources.IisLogs) > 0 {
		iisLogs := monitoring.DataCollectionRuleDataSourcesIisLogArray{}
		for _, iisLog := range dataSources.IisLogs {
			iisLogArgs := &monitoring.DataCollectionRuleDataSourcesIisLogArgs{
				Name:    pulumi.String(iisLog.Name),
				Streams: pulumi.ToStringArray(iisLog.Streams),
			}
			// Omitted when empty -- the agent then reads the server's
			// configured IIS log location.
			if len(iisLog.LogDirectories) > 0 {
				iisLogArgs.LogDirectories = pulumi.ToStringArray(iisLog.LogDirectories)
			}
			iisLogs = append(iisLogs, iisLogArgs)
		}
		dataSourcesArgs.IisLogs = iisLogs
	}

	if len(dataSources.LogFiles) > 0 {
		logFiles := monitoring.DataCollectionRuleDataSourcesLogFileArray{}
		for _, logFile := range dataSources.LogFiles {
			logFileArgs := &monitoring.DataCollectionRuleDataSourcesLogFileArgs{
				Name:         pulumi.String(logFile.Name),
				FilePatterns: pulumi.ToStringArray(logFile.FilePatterns),
				Format:       pulumi.String(logFile.Format),
				Streams:      pulumi.ToStringArray(logFile.Streams),
			}
			if logFile.Settings != nil {
				logFileArgs.Settings = &monitoring.DataCollectionRuleDataSourcesLogFileSettingsArgs{
					Text: &monitoring.DataCollectionRuleDataSourcesLogFileSettingsTextArgs{
						RecordStartTimestampFormat: pulumi.String(logFile.Settings.Text.RecordStartTimestampFormat),
					},
				}
			}
			logFiles = append(logFiles, logFileArgs)
		}
		dataSourcesArgs.LogFiles = logFiles
	}

	if len(dataSources.PrometheusForwarders) > 0 {
		forwarders := monitoring.DataCollectionRuleDataSourcesPrometheusForwarderArray{}
		for _, forwarder := range dataSources.PrometheusForwarders {
			forwarderArgs := &monitoring.DataCollectionRuleDataSourcesPrometheusForwarderArgs{
				Name:    pulumi.String(forwarder.Name),
				Streams: pulumi.ToStringArray(forwarder.Streams),
			}
			if len(forwarder.LabelIncludeFilters) > 0 {
				filters := monitoring.DataCollectionRuleDataSourcesPrometheusForwarderLabelIncludeFilterArray{}
				for _, filter := range forwarder.LabelIncludeFilters {
					filters = append(filters, &monitoring.DataCollectionRuleDataSourcesPrometheusForwarderLabelIncludeFilterArgs{
						Label: pulumi.String(filter.Label),
						Value: pulumi.String(filter.Value),
					})
				}
				forwarderArgs.LabelIncludeFilters = filters
			}
			forwarders = append(forwarders, forwarderArgs)
		}
		dataSourcesArgs.PrometheusForwarders = forwarders
	}

	if len(dataSources.WindowsFirewallLogs) > 0 {
		firewallLogs := monitoring.DataCollectionRuleDataSourcesWindowsFirewallLogArray{}
		for _, firewallLog := range dataSources.WindowsFirewallLogs {
			firewallLogs = append(firewallLogs, &monitoring.DataCollectionRuleDataSourcesWindowsFirewallLogArgs{
				Name:    pulumi.String(firewallLog.Name),
				Streams: pulumi.ToStringArray(firewallLog.Streams),
			})
		}
		dataSourcesArgs.WindowsFirewallLogs = firewallLogs
	}

	if len(dataSources.PlatformTelemetries) > 0 {
		telemetries := monitoring.DataCollectionRuleDataSourcesPlatformTelemetryArray{}
		for _, telemetry := range dataSources.PlatformTelemetries {
			telemetries = append(telemetries, &monitoring.DataCollectionRuleDataSourcesPlatformTelemetryArgs{
				Name:    pulumi.String(telemetry.Name),
				Streams: pulumi.ToStringArray(telemetry.Streams),
			})
		}
		dataSourcesArgs.PlatformTelemetries = telemetries
	}

	if dataSources.DataImport != nil {
		// Azure's rule model carries exactly ONE event-hub import; the
		// spec models it singular (the provider would silently drop extra
		// entries), so the SDK's array receives exactly one element.
		eventHubSource := dataSources.DataImport.EventHubDataSource
		eventHubArgs := &monitoring.DataCollectionRuleDataSourcesDataImportEventHubDataSourceArgs{
			Name:   pulumi.String(eventHubSource.Name),
			Stream: pulumi.String(eventHubSource.Stream),
		}
		// Sent only when set -- the provider omits an empty consumer
		// group (Azure then reads $Default).
		if eventHubSource.ConsumerGroup != "" {
			eventHubArgs.ConsumerGroup = pulumi.String(eventHubSource.ConsumerGroup)
		}
		dataSourcesArgs.DataImport = &monitoring.DataCollectionRuleDataSourcesDataImportArgs{
			EventHubDataSources: monitoring.DataCollectionRuleDataSourcesDataImportEventHubDataSourceArray{eventHubArgs},
		}
	}

	return dataSourcesArgs
}
