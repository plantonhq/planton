package module

import (
	"strings"

	"github.com/pkg/errors"
	gcpmonitoringslov1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpmonitoringslo/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/monitoring"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// slo provisions the Cloud Monitoring SLO and, when the spec's service arm
// asks for one, the Monitoring service it measures (a custom service or a
// basic/generic service) — one kind, up to two service resources
// count-gated by the arm, mirroring the Terraform module.
//
// Both outputs derive from the SLO's server-assigned resource name
// (projects/{p}/services/{s}/serviceLevelObjectives/{id}) so they are
// correct on every arm — including an existing service in the provider's
// ambient project, where no service resource exists client-side to read a
// name from.
func slo(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpMonitoringSlo.Spec

	// Enable the Cloud Monitoring API so a fresh project can host the SLO.
	// disable_on_destroy stays false (the provider default): tearing down
	// one SLO must never disable monitoring for everything else in the
	// project. Matches the Terraform module.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("monitoring.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"slo-monitoring.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable monitoring.googleapis.com api")
	}

	dependencies := []pulumi.Resource{createdProjectService}

	// Resolve the service the SLO measures: an existing service ID, or a
	// service this module creates (exactly one arm — proto-CEL-enforced).
	sloServiceId := pulumi.String(spec.Service.GetServiceId()).ToStringOutput()

	if custom := spec.Service.GetCustomService(); custom != nil {
		customServiceArgs := &monitoring.CustomServiceArgs{
			ServiceId:   pulumi.String(locals.CreatedServiceId),
			DisplayName: pulumi.String(customServiceDisplayName(locals, custom)),
			UserLabels:  pulumi.ToStringMap(locals.GcpLabels),
		}
		if custom.TelemetryResourceName != "" {
			customServiceArgs.Telemetry = &monitoring.CustomServiceTelemetryArgs{
				ResourceName: pulumi.StringPtr(custom.TelemetryResourceName),
			}
		}
		// The service follows the kind's deletion contract: destroying the
		// SLO component destroys the service it created.
		if spec.DeletionPolicy != "" {
			customServiceArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			customServiceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		createdCustomService, err := monitoring.NewCustomService(ctx, "custom-service", customServiceArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn(dependencies))
		if err != nil {
			return errors.Wrap(err, "failed to create custom service")
		}
		dependencies = append(dependencies, createdCustomService)
		sloServiceId = createdCustomService.ServiceId
	}

	if basic := spec.Service.GetBasicService(); basic != nil {
		genericServiceArgs := &monitoring.GenericServiceArgs{
			ServiceId:   pulumi.String(locals.CreatedServiceId),
			DisplayName: pulumi.String(locals.DisplayName),
			UserLabels:  pulumi.ToStringMap(locals.GcpLabels),
			BasicService: &monitoring.GenericServiceBasicServiceArgs{
				ServiceType:   pulumi.StringPtr(basic.ServiceType),
				ServiceLabels: pulumi.ToStringMap(basic.ServiceLabels),
			},
		}
		if spec.DeletionPolicy != "" {
			genericServiceArgs.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}
		if spec.ProjectId.GetValue() != "" {
			genericServiceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		createdGenericService, err := monitoring.NewGenericService(ctx, "basic-service", genericServiceArgs,
			pulumi.Provider(gcpProvider), pulumi.DependsOn(dependencies))
		if err != nil {
			return errors.Wrap(err, "failed to create basic service")
		}
		dependencies = append(dependencies, createdGenericService)
		sloServiceId = createdGenericService.ServiceId
	}

	args := &monitoring.SloArgs{
		Service:     sloServiceId,
		Goal:        pulumi.Float64(spec.Goal),
		DisplayName: pulumi.String(locals.DisplayName),
		UserLabels:  pulumi.ToStringMap(locals.GcpLabels),
	}

	// Exactly one period form is set (proto-CEL-enforced, mirroring the
	// provider's ExactlyOneOf).
	if spec.CalendarPeriod != "" {
		args.CalendarPeriod = pulumi.StringPtr(spec.CalendarPeriod)
	}
	if spec.RollingPeriodDays != 0 {
		args.RollingPeriodDays = pulumi.IntPtr(int(spec.RollingPeriodDays))
	}

	if spec.SloId != "" {
		args.SloId = pulumi.StringPtr(spec.SloId)
	}

	// Exactly one SLI family is set (proto-CEL-enforced).
	if basicSli := spec.Sli.GetBasicSli(); basicSli != nil {
		args.BasicSli = expandBasicSli(basicSli)
	}
	if requestBased := spec.Sli.GetRequestBasedSli(); requestBased != nil {
		args.RequestBasedSli = expandRequestBasedSli(requestBased)
	}
	if windowsBased := spec.Sli.GetWindowsBasedSli(); windowsBased != nil {
		args.WindowsBasedSli = expandWindowsBasedSli(windowsBased)
	}

	// Unset defers to the provider default (DELETE).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project (empty string would be sent verbatim and
	// rejected).
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}

	createdSlo, err := monitoring.NewSlo(ctx, "slo", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn(dependencies))
	if err != nil {
		return errors.Wrap(err, "failed to create slo")
	}

	ctx.Export(OpSloName, createdSlo.Name)
	// The service segment of the SLO's own resource name — correct on
	// every service arm (see the function comment).
	ctx.Export(OpServiceName, createdSlo.Name.ApplyT(func(name string) string {
		if idx := strings.Index(name, "/serviceLevelObjectives/"); idx > 0 {
			return name[:idx]
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

// customServiceDisplayName resolves the created custom service's display
// name: the arm's own display_name, else the kind's naming basis.
func customServiceDisplayName(locals *Locals, custom *gcpmonitoringslov1alpha1.GcpMonitoringSloCustomService) string {
	if custom.DisplayName != "" {
		return custom.DisplayName
	}
	return locals.DisplayName
}

// expandBasicSli maps the spec's basic SLI onto the provider's basic_sli
// block. Exactly one of availability/latency is present
// (proto-CEL-enforced).
//
// availability.enabled is sent EXPLICITLY: it is Optional in the provider,
// and the GCP API expects true — omitting it would leave the arm's one
// field to a server-side default the spec cannot see.
func expandBasicSli(basicSli *gcpmonitoringslov1alpha1.GcpMonitoringSloBasicSli) *monitoring.SloBasicSliArgs {
	basicSliArgs := &monitoring.SloBasicSliArgs{}
	if len(basicSli.Location) > 0 {
		basicSliArgs.Locations = pulumi.ToStringArray(basicSli.Location)
	}
	if len(basicSli.Method) > 0 {
		basicSliArgs.Methods = pulumi.ToStringArray(basicSli.Method)
	}
	if len(basicSli.Version) > 0 {
		basicSliArgs.Versions = pulumi.ToStringArray(basicSli.Version)
	}
	if availability := basicSli.Availability; availability != nil {
		basicSliArgs.Availability = &monitoring.SloBasicSliAvailabilityArgs{
			Enabled: pulumi.Bool(availability.Enabled == nil || availability.GetEnabled()),
		}
	}
	if latency := basicSli.Latency; latency != nil {
		basicSliArgs.Latency = &monitoring.SloBasicSliLatencyArgs{
			Threshold: pulumi.String(latency.Threshold),
		}
	}
	return basicSliArgs
}

// expandRequestBasedSli maps the spec's request-based SLI onto the
// provider's request_based_sli block. Exactly one of
// distribution_cut/good_total_ratio is present (proto-CEL-enforced).
func expandRequestBasedSli(requestBased *gcpmonitoringslov1alpha1.GcpMonitoringSloRequestBasedSli) *monitoring.SloRequestBasedSliArgs {
	requestBasedArgs := &monitoring.SloRequestBasedSliArgs{}
	if cut := requestBased.DistributionCut; cut != nil {
		cutArgs := &monitoring.SloRequestBasedSliDistributionCutArgs{
			DistributionFilter: pulumi.String(cut.DistributionFilter),
		}
		if cut.Range != nil {
			rangeArgs := &monitoring.SloRequestBasedSliDistributionCutRangeArgs{}
			if cut.Range.Min != nil {
				rangeArgs.Min = pulumi.Float64Ptr(cut.Range.GetMin())
			}
			if cut.Range.Max != nil {
				rangeArgs.Max = pulumi.Float64Ptr(cut.Range.GetMax())
			}
			cutArgs.Range = rangeArgs
		}
		requestBasedArgs.DistributionCut = cutArgs
	}
	if ratio := requestBased.GoodTotalRatio; ratio != nil {
		ratioArgs := &monitoring.SloRequestBasedSliGoodTotalRatioArgs{}
		if ratio.GoodServiceFilter != "" {
			ratioArgs.GoodServiceFilter = pulumi.StringPtr(ratio.GoodServiceFilter)
		}
		if ratio.BadServiceFilter != "" {
			ratioArgs.BadServiceFilter = pulumi.StringPtr(ratio.BadServiceFilter)
		}
		if ratio.TotalServiceFilter != "" {
			ratioArgs.TotalServiceFilter = pulumi.StringPtr(ratio.TotalServiceFilter)
		}
		requestBasedArgs.GoodTotalRatio = ratioArgs
	}
	return requestBasedArgs
}

// expandWindowsBasedSli maps the spec's windows-based SLI onto the
// provider's windows_based_sli block. Exactly one window criterion is
// present (proto-CEL-enforced).
func expandWindowsBasedSli(windowsBased *gcpmonitoringslov1alpha1.GcpMonitoringSloWindowsBasedSli) *monitoring.SloWindowsBasedSliArgs {
	windowsBasedArgs := &monitoring.SloWindowsBasedSliArgs{}

	if windowsBased.WindowPeriod != "" {
		windowsBasedArgs.WindowPeriod = pulumi.StringPtr(windowsBased.WindowPeriod)
	}

	if windowsBased.GoodBadMetricFilter != "" {
		windowsBasedArgs.GoodBadMetricFilter = pulumi.StringPtr(windowsBased.GoodBadMetricFilter)
	}

	if threshold := windowsBased.GoodTotalRatioThreshold; threshold != nil {
		thresholdArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdArgs{
			// threshold 0 is a legal (if degenerate) ratio bound, so it is
			// always sent — zero and unset are deliberately NOT
			// distinguished for this field.
			Threshold: pulumi.Float64Ptr(threshold.Threshold),
		}
		if performance := threshold.BasicSliPerformance; performance != nil {
			performanceArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdBasicSliPerformanceArgs{}
			if len(performance.Location) > 0 {
				performanceArgs.Locations = pulumi.ToStringArray(performance.Location)
			}
			if len(performance.Method) > 0 {
				performanceArgs.Methods = pulumi.ToStringArray(performance.Method)
			}
			if len(performance.Version) > 0 {
				performanceArgs.Versions = pulumi.ToStringArray(performance.Version)
			}
			if availability := performance.Availability; availability != nil {
				performanceArgs.Availability = &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdBasicSliPerformanceAvailabilityArgs{
					Enabled: pulumi.Bool(availability.Enabled == nil || availability.GetEnabled()),
				}
			}
			if latency := performance.Latency; latency != nil {
				performanceArgs.Latency = &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdBasicSliPerformanceLatencyArgs{
					Threshold: pulumi.String(latency.Threshold),
				}
			}
			thresholdArgs.BasicSliPerformance = performanceArgs
		}
		if performance := threshold.Performance; performance != nil {
			performanceArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdPerformanceArgs{}
			if cut := performance.DistributionCut; cut != nil {
				cutArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdPerformanceDistributionCutArgs{
					DistributionFilter: pulumi.String(cut.DistributionFilter),
				}
				if cut.Range != nil {
					rangeArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdPerformanceDistributionCutRangeArgs{}
					if cut.Range.Min != nil {
						rangeArgs.Min = pulumi.Float64Ptr(cut.Range.GetMin())
					}
					if cut.Range.Max != nil {
						rangeArgs.Max = pulumi.Float64Ptr(cut.Range.GetMax())
					}
					cutArgs.Range = rangeArgs
				}
				performanceArgs.DistributionCut = cutArgs
			}
			if ratio := performance.GoodTotalRatio; ratio != nil {
				ratioArgs := &monitoring.SloWindowsBasedSliGoodTotalRatioThresholdPerformanceGoodTotalRatioArgs{}
				if ratio.GoodServiceFilter != "" {
					ratioArgs.GoodServiceFilter = pulumi.StringPtr(ratio.GoodServiceFilter)
				}
				if ratio.BadServiceFilter != "" {
					ratioArgs.BadServiceFilter = pulumi.StringPtr(ratio.BadServiceFilter)
				}
				if ratio.TotalServiceFilter != "" {
					ratioArgs.TotalServiceFilter = pulumi.StringPtr(ratio.TotalServiceFilter)
				}
				performanceArgs.GoodTotalRatio = ratioArgs
			}
			thresholdArgs.Performance = performanceArgs
		}
		windowsBasedArgs.GoodTotalRatioThreshold = thresholdArgs
	}

	if meanInRange := windowsBased.MetricMeanInRange; meanInRange != nil {
		windowsBasedArgs.MetricMeanInRange = &monitoring.SloWindowsBasedSliMetricMeanInRangeArgs{
			TimeSeries: pulumi.String(meanInRange.TimeSeries),
			Range:      expandMeanRange(meanInRange.Range),
		}
	}

	if sumInRange := windowsBased.MetricSumInRange; sumInRange != nil {
		windowsBasedArgs.MetricSumInRange = &monitoring.SloWindowsBasedSliMetricSumInRangeArgs{
			TimeSeries: pulumi.String(sumInRange.TimeSeries),
			Range:      expandSumRange(sumInRange.Range),
		}
	}

	return windowsBasedArgs
}

// expandMeanRange / expandSumRange map the spec's range onto the two
// same-shaped (but distinct) provider range types. The provider REQUIRES a
// range block on mean/sum criteria, so a nil spec range still renders an
// empty block and lets the API report the miss precisely.
func expandMeanRange(specRange *gcpmonitoringslov1alpha1.GcpMonitoringSloRange) *monitoring.SloWindowsBasedSliMetricMeanInRangeRangeArgs {
	rangeArgs := &monitoring.SloWindowsBasedSliMetricMeanInRangeRangeArgs{}
	if specRange != nil {
		if specRange.Min != nil {
			rangeArgs.Min = pulumi.Float64Ptr(specRange.GetMin())
		}
		if specRange.Max != nil {
			rangeArgs.Max = pulumi.Float64Ptr(specRange.GetMax())
		}
	}
	return rangeArgs
}

func expandSumRange(specRange *gcpmonitoringslov1alpha1.GcpMonitoringSloRange) *monitoring.SloWindowsBasedSliMetricSumInRangeRangeArgs {
	rangeArgs := &monitoring.SloWindowsBasedSliMetricSumInRangeRangeArgs{}
	if specRange != nil {
		if specRange.Min != nil {
			rangeArgs.Min = pulumi.Float64Ptr(specRange.GetMin())
		}
		if specRange.Max != nil {
			rangeArgs.Max = pulumi.Float64Ptr(specRange.GetMax())
		}
	}
	return rangeArgs
}
