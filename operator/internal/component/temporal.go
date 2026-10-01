package component

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// Temporal deploys and monitors Temporal via its official Helm chart.
// Depends on PostgreSQL because Temporal uses it for persistence and
// visibility storage.
type Temporal struct{ Base }

func (t *Temporal) Name() string                                { return "temporal" }
func (t *Temporal) Dependencies(_ *v1.PlantonPlatform) []string { return []string{"postgresql"} }
func (t *Temporal) IsEnabled(_ *v1.PlantonPlatform) bool        { return true }

func (t *Temporal) Reconcile(ctx context.Context, c client.Client, _ *runtime.Scheme, planton *v1.PlantonPlatform) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", t.Name())

	chartData := resources.LoadTemporalChart()
	values := resources.TemporalHelmValues(temporalHelmOptions(planton))

	rendered, err := resources.RenderHelmChart(
		chartData,
		fmt.Sprintf("%s-temporal", planton.Name),
		planton.Namespace,
		values,
	)
	if err != nil {
		return Result{}, fmt.Errorf("rendering Temporal chart: %w", err)
	}

	if err := t.ApplyManifests(ctx, c, planton, rendered); err != nil {
		return Result{}, fmt.Errorf("applying Temporal manifests: %w", err)
	}

	// Clients dial the frontend, so it alone gates readiness; every server
	// service is read for a memory kill it recovered from, since history --
	// the busiest -- would otherwise hide behind a healthy frontend.
	workloads := temporalWorkloads(planton.Name)
	frontend := workloads[0]
	ready, err := t.IsDeploymentReady(ctx, c, frontend.Name, planton.Namespace)
	if err != nil {
		return Result{}, fmt.Errorf("checking Temporal readiness: %w", err)
	}
	if !ready {
		log.Info("Temporal not ready")
		// The chart's schema Job creates Temporal's databases and tables;
		// until it finishes, every server pod exits on "schema_version does
		// not exist" and restarts. Read the Job FIRST so those restarts are
		// reported as the wait they are, not as a crash loop -- and a Job
		// that failed is named as the cause instead of the pods it starves.
		if expl := t.explainJobs(ctx, c, planton.Namespace, map[string]string{
			"app.kubernetes.io/instance":  fmt.Sprintf("%s-temporal", planton.Name),
			"app.kubernetes.io/component": "database",
		}, "Temporal schema setup"); expl != nil {
			return Result{Ready: false, Reason: expl.Reason, Object: expl.Object, Message: expl.Message}, nil
		}
		return t.NotReady(ctx, c, planton.Namespace, frontend, "Waiting for Temporal frontend"), nil
	}

	log.Info("Temporal ready")
	return t.Ready(ctx, c, planton.Namespace, "Temporal healthy", workloads...), nil
}

// temporalWorkloads references the chart's four server Deployments, each
// sized by its own field, frontend first.
func temporalWorkloads(crName string) []WorkloadRef {
	refs := make([]WorkloadRef, 0, len(resources.TemporalServerServices))
	for _, svc := range resources.TemporalServerServices {
		refs = append(refs, DeploymentRef(resources.TemporalServiceDeploymentName(crName, svc.Name)).Sized(svc.SizedBy))
	}
	return refs
}

// temporalHelmOptions resolves each server service's sizing: the service's own
// override laid over its measured default, quantity by quantity.
func temporalHelmOptions(planton *v1.PlantonPlatform) resources.TemporalHelmOptions {
	return resources.TemporalHelmOptions{
		CRName:    planton.Name,
		Namespace: planton.Namespace,
		Frontend:  resources.EffectiveFor(resources.SizingTemporalFrontend, &planton.Spec),
		History:   resources.EffectiveFor(resources.SizingTemporalHistory, &planton.Spec),
		Matching:  resources.EffectiveFor(resources.SizingTemporalMatching, &planton.Spec),
		Worker:    resources.EffectiveFor(resources.SizingTemporalWorker, &planton.Spec),
	}
}
