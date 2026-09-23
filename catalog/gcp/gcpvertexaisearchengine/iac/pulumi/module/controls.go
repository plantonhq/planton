package module

import (
	"fmt"

	"github.com/pkg/errors"
	gcpvertexaisearchenginev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaisearchengine/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/discoveryengine"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// controls creates one `google_discovery_engine_control` per
// spec.controls[] entry, keyed by control_id. solution_type is derived
// from the arm -- the provider hard-codes each engine resource's solution,
// and a control must carry the same one. A control takes effect only when
// the serving config lists it. Names are exported in manifest order; the
// resources are returned so the serving config can depend on them.
func controls(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider,
	created *createdEngine) ([]pulumi.Resource, pulumi.StringArray, error) {
	spec := locals.GcpVertexAiSearchEngine.Spec
	names := pulumi.StringArray{}
	resources := []pulumi.Resource{}

	for _, control := range spec.Controls {
		args := &discoveryengine.ControlArgs{
			ControlId:    pulumi.String(control.ControlId),
			DisplayName:  pulumi.String(control.DisplayName),
			EngineId:     created.EngineId,
			Location:     pulumi.String(spec.Location),
			CollectionId: pulumi.String(locals.CollectionId),
			SolutionType: pulumi.String(locals.SolutionType),
		}
		if spec.ProjectId.GetValue() != "" {
			args.Project = pulumi.String(spec.ProjectId.GetValue())
		}
		if len(control.UseCases) > 0 {
			args.UseCases = pulumi.ToStringArray(control.UseCases)
		}
		if len(control.Conditions) > 0 {
			args.Conditions = buildConditions(control.Conditions)
		}
		switch {
		case control.BoostAction != nil:
			args.BoostAction = buildBoostAction(control.BoostAction)
		case control.FilterAction != nil:
			args.FilterAction = &discoveryengine.ControlFilterActionArgs{
				DataStore: pulumi.String(control.FilterAction.DataStore.GetValue()),
				Filter:    pulumi.String(control.FilterAction.Filter),
			}
		case control.PromoteAction != nil:
			args.PromoteAction = buildPromoteAction(control.PromoteAction)
		case control.RedirectAction != nil:
			args.RedirectAction = &discoveryengine.ControlRedirectActionArgs{
				RedirectUri: pulumi.String(control.RedirectAction.RedirectUri),
			}
		case control.SynonymsAction != nil:
			args.SynonymsAction = &discoveryengine.ControlSynonymsActionArgs{
				Synonyms: pulumi.ToStringArray(control.SynonymsAction.Synonyms),
			}
		}
		if spec.DeletionPolicy != "" {
			args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
		}

		createdControl, err := discoveryengine.NewControl(ctx,
			fmt.Sprintf("%s-%s", locals.GcpVertexAiSearchEngine.Metadata.Name, control.ControlId), args,
			pulumi.Provider(gcpProvider),
			pulumi.Parent(created.Resource))
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to create control %s", control.ControlId)
		}
		names = append(names, createdControl.Name)
		resources = append(resources, createdControl)
	}
	return resources, names, nil
}

// buildConditions maps when a control is active: query terms, time
// windows, and a query regex, each sent only when set.
func buildConditions(conditions []*gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineControlCondition) discoveryengine.ControlConditionArray {
	out := discoveryengine.ControlConditionArray{}
	for _, condition := range conditions {
		args := &discoveryengine.ControlConditionArgs{}
		if condition.QueryRegex != "" {
			args.QueryRegex = pulumi.String(condition.QueryRegex)
		}
		if len(condition.QueryTerms) > 0 {
			terms := discoveryengine.ControlConditionQueryTermArray{}
			for _, term := range condition.QueryTerms {
				terms = append(terms, &discoveryengine.ControlConditionQueryTermArgs{
					Value:     pulumi.String(term.Value),
					FullMatch: pulumi.BoolPtr(term.FullMatch),
				})
			}
			args.QueryTerms = terms
		}
		if len(condition.ActiveTimeRanges) > 0 {
			ranges := discoveryengine.ControlConditionActiveTimeRangeArray{}
			for _, window := range condition.ActiveTimeRanges {
				rangeArgs := &discoveryengine.ControlConditionActiveTimeRangeArgs{}
				if window.StartTime != "" {
					rangeArgs.StartTime = pulumi.String(window.StartTime)
				}
				if window.EndTime != "" {
					rangeArgs.EndTime = pulumi.String(window.EndTime)
				}
				ranges = append(ranges, rangeArgs)
			}
			args.ActiveTimeRanges = ranges
		}
		out = append(out, args)
	}
	return out
}

// buildBoostAction maps a boost: exactly one of a fixed amount or an
// interpolation curve (a spec rule).
func buildBoostAction(boost *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEngineBoostAction) *discoveryengine.ControlBoostActionArgs {
	args := &discoveryengine.ControlBoostActionArgs{
		DataStore: pulumi.String(boost.DataStore.GetValue()),
		Filter:    pulumi.String(boost.Filter),
	}
	if boost.FixedBoost != nil {
		args.FixedBoost = pulumi.Float64(float64(boost.GetFixedBoost()))
	}
	if ib := boost.InterpolationBoostSpec; ib != nil {
		ibArgs := &discoveryengine.ControlBoostActionInterpolationBoostSpecArgs{}
		if ib.FieldName != "" {
			ibArgs.FieldName = pulumi.String(ib.FieldName)
		}
		if ib.AttributeType != "" {
			ibArgs.AttributeType = pulumi.String(ib.AttributeType)
		}
		if ib.InterpolationType != "" {
			ibArgs.InterpolationType = pulumi.String(ib.InterpolationType)
		}
		if cp := ib.ControlPoint; cp != nil {
			cpArgs := &discoveryengine.ControlBoostActionInterpolationBoostSpecControlPointArgs{}
			if cp.AttributeValue != "" {
				cpArgs.AttributeValue = pulumi.String(cp.AttributeValue)
			}
			if cp.BoostAmount != nil {
				cpArgs.BoostAmount = pulumi.Float64(float64(cp.GetBoostAmount()))
			}
			ibArgs.ControlPoint = cpArgs
		}
		args.InterpolationBoostSpec = ibArgs
	}
	return args
}

// buildPromoteAction maps a pinned link; optional strings are sent only
// when set.
func buildPromoteAction(promote *gcpvertexaisearchenginev1alpha1.GcpVertexAiSearchEnginePromoteAction) *discoveryengine.ControlPromoteActionArgs {
	link := promote.SearchLinkPromotion
	linkArgs := &discoveryengine.ControlPromoteActionSearchLinkPromotionArgs{
		Title:   pulumi.String(link.Title),
		Enabled: pulumi.BoolPtr(link.Enabled),
	}
	if link.Uri != "" {
		linkArgs.Uri = pulumi.String(link.Uri)
	}
	if link.Document != "" {
		linkArgs.Document = pulumi.String(link.Document)
	}
	if link.Description != "" {
		linkArgs.Description = pulumi.String(link.Description)
	}
	if link.ImageUri != "" {
		linkArgs.ImageUri = pulumi.String(link.ImageUri)
	}
	return &discoveryengine.ControlPromoteActionArgs{
		DataStore:           pulumi.String(promote.DataStore.GetValue()),
		SearchLinkPromotion: linkArgs,
	}
}
