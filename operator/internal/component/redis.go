package component

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

const (
	defaultRedisStorageSize = "1Gi"

	// Persistence is on unless the platform says otherwise: the one store
	// carries the live build-log stream beside the cache, and a restart
	// mid-build must not blank the log a person is tailing. The ceiling
	// (resources.ValkeyDefaultMaxMemory) is what makes a persisted store safe
	// to reload -- see the RedisSpec field docs.
	defaultRedisPersistence = true

	// The chart names the volume claim template "valkey-data"; the claim a
	// StatefulSet creates from it is "<template>-<statefulset>-<ordinal>".
	redisVolumeClaimTemplateName = "valkey-data"
)

// Redis deploys and monitors the platform's redis-protocol store via the
// Bitnami Valkey Helm chart -- the engine is Valkey (BSD-3-Clause), not
// Redis 8+ (RSALv2/SSPLv1/AGPLv3); see resources/valkey_helm.go for the
// rationale. The component keeps the "redis" role name across the CRD, status,
// and connection surfaces. There is no operator mode -- scaling is handled
// through chart values.
type Redis struct{ Base }

func (r *Redis) Name() string                                { return "redis" }
func (r *Redis) Dependencies(_ *v1.PlantonPlatform) []string { return nil }
func (r *Redis) IsEnabled(_ *v1.PlantonPlatform) bool        { return true }

func (r *Redis) Reconcile(ctx context.Context, c client.Client, _ *runtime.Scheme, planton *v1.PlantonPlatform) (Result, error) {
	log := logf.FromContext(ctx).WithValues("component", r.Name())

	opts := redisHelmOptions(planton)

	if err := r.EnsureCredentialSecret(ctx, c,
		resources.RedisSecretName(planton.Name), planton.Namespace,
		resources.RedisSecretKey, r.OwnerReferenceFor(planton)); err != nil {
		return Result{}, fmt.Errorf("ensuring Redis credentials: %w", err)
	}

	chartData := resources.LoadValkeyChart()
	values := resources.ValkeyHelmValues(opts)

	rendered, err := resources.RenderHelmChart(
		chartData,
		fmt.Sprintf("%s-redis", planton.Name),
		planton.Namespace,
		values,
	)
	if err != nil {
		return Result{}, fmt.Errorf("rendering Redis chart: %w", err)
	}

	// The Valkey chart names the data-serving StatefulSet "primary" (the Redis
	// chart said "master").
	stsName := fmt.Sprintf("%s-redis-primary", planton.Name)

	if err := r.replaceStatefulSetIfPersistenceFlipped(ctx, c, planton, stsName, opts.Persistence); err != nil {
		return Result{}, err
	}

	if err := r.ApplyManifests(ctx, c, planton, rendered); err != nil {
		return Result{}, fmt.Errorf("applying Redis manifests: %w", err)
	}

	ready, err := r.IsStatefulSetReady(ctx, c, stsName, planton.Namespace)
	if err != nil {
		return Result{}, fmt.Errorf("checking Redis readiness: %w", err)
	}
	if !ready {
		log.Info("Redis not ready")
		return r.NotReady(ctx, c, planton.Namespace, StatefulSetRef(stsName), "Waiting for Redis"), nil
	}

	log.Info("Redis ready")
	return Result{Ready: true, Message: "Redis healthy"}, nil
}

// redisHelmOptions resolves every sizing knob the store's render needs: the
// component's own field wins, then the platform-wide storage block where one
// applies, then the documented default -- the same order every volume the
// operator creates resolves through (storage.go).
func redisHelmOptions(planton *v1.PlantonPlatform) resources.ValkeyHelmOptions {
	var spec v1.RedisSpec
	if planton.Spec.Database != nil && planton.Spec.Database.Redis != nil {
		spec = *planton.Spec.Database.Redis
	}

	persistence := defaultRedisPersistence
	if spec.Persistence != nil {
		persistence = *spec.Persistence
	}
	maxMemory := resources.ValkeyDefaultMaxMemory
	if spec.MaxMemory != "" {
		maxMemory = spec.MaxMemory
	}
	maxMemoryPolicy := resources.ValkeyDefaultMaxMemoryPolicy
	if spec.MaxMemoryPolicy != "" {
		maxMemoryPolicy = spec.MaxMemoryPolicy
	}
	containerResources := resources.ValkeyDefaultResources()
	if spec.Resources != nil {
		containerResources = *spec.Resources
	}

	return resources.ValkeyHelmOptions{
		CRName:          planton.Name,
		Persistence:     persistence,
		StorageSize:     effectiveStorageSize(planton, spec.StorageSize, defaultRedisStorageSize),
		StorageClass:    effectiveStorageClass(planton, spec.StorageClassName),
		MaxMemory:       maxMemory,
		MaxMemoryPolicy: maxMemoryPolicy,
		Resources:       containerResources,
	}
}

// replaceStatefulSetIfPersistenceFlipped handles the one change a StatefulSet
// cannot take in place: its volume claim templates. Turning persistence off
// removes the claim template, turning it on adds one, and the API server
// refuses both as an edit. The generic apply deliberately does not answer
// that refusal for StatefulSets (the vault's claim must never be replaced by
// accident); this component knows its store is a cache and a log stream whose
// durable copy is the archive, so it makes the decision before the refusal,
// in words: delete the StatefulSet (its pod follows), and the claim too when
// persistence was turned off -- a volume nothing will mount again is exactly
// the kind of leftover the platform should not leave behind. The apply that
// follows recreates the StatefulSet in the declared shape; the store restarts
// empty in the off direction and empty-then-persisted in the on direction.
func (r *Redis) replaceStatefulSetIfPersistenceFlipped(ctx context.Context, c client.Client, planton *v1.PlantonPlatform, stsName string, persistence bool) error {
	log := logf.FromContext(ctx).WithValues("component", r.Name())

	live := &appsv1.StatefulSet{}
	if err := c.Get(ctx, types.NamespacedName{Name: stsName, Namespace: planton.Namespace}, live); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("reading the live Redis StatefulSet: %w", err)
	}

	liveHasClaim := len(live.Spec.VolumeClaimTemplates) > 0
	if liveHasClaim == persistence {
		return nil
	}

	log.Info("Replacing the Redis StatefulSet: its persistence changed, and a StatefulSet's volume claim templates cannot be edited in place",
		"name", stsName, "persistenceWas", liveHasClaim, "persistenceNow", persistence)
	if err := c.Delete(ctx, live, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("replacing the Redis StatefulSet %s/%s after its persistence changed: deleting the old one: %w",
			planton.Namespace, stsName, err)
	}

	if !persistence {
		// The claim the old StatefulSet created outlives it by design
		// (Kubernetes retains claims on delete). Nothing will mount it again.
		claim := &corev1.PersistentVolumeClaim{}
		claim.Name = fmt.Sprintf("%s-%s-0", redisVolumeClaimTemplateName, stsName)
		claim.Namespace = planton.Namespace
		log.Info("Deleting the Redis volume claim: persistence was turned off and nothing will mount it again", "claim", claim.Name)
		if err := c.Delete(ctx, claim); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("deleting the Redis volume claim %s/%s after persistence was turned off: %w",
				planton.Namespace, claim.Name, err)
		}
	}
	return nil
}
