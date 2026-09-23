package component

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The store's sizing resolves the way every volume does: the component's own
// field, then the platform-wide block where one applies, then the documented
// default -- and the defaults are the ones the crash loop taught (a ceiling
// inside the limit, eviction, persistence for the log stream).
func TestRedisHelmOptions_DefaultsAreTheBoundedPersistedStore(t *testing.T) {
	opts := redisHelmOptions(ownershipPlatform())
	if !opts.Persistence {
		t.Error("persistence defaults on: the store carries the live build-log stream")
	}
	if opts.MaxMemory != resources.ValkeyDefaultMaxMemory || opts.MaxMemoryPolicy != resources.ValkeyDefaultMaxMemoryPolicy {
		t.Errorf("expected the default ceiling %s/%s, got %s/%s",
			resources.ValkeyDefaultMaxMemory, resources.ValkeyDefaultMaxMemoryPolicy, opts.MaxMemory, opts.MaxMemoryPolicy)
	}
	if opts.StorageSize != defaultRedisStorageSize || opts.StorageClass != "" {
		t.Errorf("expected the default volume %s with the cluster's class, got %s/%q", defaultRedisStorageSize, opts.StorageSize, opts.StorageClass)
	}
	if opts.Resources.Limits.Memory().IsZero() || opts.Resources.Requests.Cpu().IsZero() || !opts.Resources.Limits.Cpu().IsZero() {
		t.Errorf("expected requests, a memory limit, and no CPU limit; got %v", opts.Resources)
	}
}

func TestRedisHelmOptions_DeclaredFieldsWinOverDefaultsAndTheStorageBlock(t *testing.T) {
	off := false
	platform := ownershipPlatform()
	platform.Spec.Storage = &v1.StorageSpec{Size: resource.MustParse("50Gi"), StorageClassName: "platform-wide"}
	platform.Spec.Database = &v1.DatabaseSpec{Redis: &v1.RedisSpec{
		StorageSize:     resource.MustParse("8Gi"),
		Persistence:     &off,
		MaxMemory:       "2gb",
		MaxMemoryPolicy: "volatile-ttl",
		Resources: &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("3Gi")},
		},
	}}

	opts := redisHelmOptions(platform)
	if opts.Persistence {
		t.Error("an explicit false must turn persistence off")
	}
	if opts.MaxMemory != "2gb" || opts.MaxMemoryPolicy != "volatile-ttl" {
		t.Errorf("declared ceiling and policy must win, got %s/%s", opts.MaxMemory, opts.MaxMemoryPolicy)
	}
	if opts.StorageSize != "8Gi" {
		t.Errorf("the component's own size wins over the platform-wide block, got %s", opts.StorageSize)
	}
	if opts.StorageClass != "platform-wide" {
		t.Errorf("an unset component class falls back to the platform-wide block, got %q", opts.StorageClass)
	}
	if opts.Resources.Limits.Memory().String() != "3Gi" || !opts.Resources.Requests.Cpu().IsZero() {
		t.Errorf("declared resources are honored verbatim, not merged with the defaults; got %v", opts.Resources)
	}
}

// A StatefulSet's volume claim templates cannot change in place, so flipping
// persistence is answered before the refusal: the live StatefulSet is deleted
// (and the claim nothing will mount again, when persistence turned off) and
// the apply that follows recreates it in the declared shape. When the shape
// did not change, nothing is touched.
func TestRedis_ReplacesTheStatefulSetOnlyWhenPersistenceFlipped(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	persisted := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-redis-primary", Namespace: "planton"},
		Spec: appsv1.StatefulSetSpec{VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{ObjectMeta: metav1.ObjectMeta{Name: redisVolumeClaimTemplateName}},
		}},
	}
	claim := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "valkey-data-planton-redis-primary-0", Namespace: "planton"},
	}

	newClient := func(objs ...client.Object) (client.Client, *[]string) {
		var deleted []string
		c := interceptor.NewClient(fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build(), interceptor.Funcs{
			Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
				deleted = append(deleted, obj.GetName())
				return cl.Delete(ctx, obj, opts...)
			},
		})
		return c, &deleted
	}
	r := &Redis{}
	platform := ownershipPlatform()

	// Same shape: a persisted live StatefulSet under a persisted declaration.
	c, deleted := newClient(persisted.DeepCopy(), claim.DeepCopy())
	if err := r.replaceStatefulSetIfPersistenceFlipped(context.Background(), c, platform, persisted.Name, true); err != nil {
		t.Fatal(err)
	}
	if len(*deleted) != 0 {
		t.Fatalf("an unchanged shape must delete nothing, got %v", *deleted)
	}

	// Turned off: the StatefulSet and the orphaned claim go.
	c, deleted = newClient(persisted.DeepCopy(), claim.DeepCopy())
	if err := r.replaceStatefulSetIfPersistenceFlipped(context.Background(), c, platform, persisted.Name, false); err != nil {
		t.Fatal(err)
	}
	if len(*deleted) != 2 || (*deleted)[0] != persisted.Name || (*deleted)[1] != claim.Name {
		t.Fatalf("turning persistence off must delete the StatefulSet then its claim, got %v", *deleted)
	}

	// Turned on from an in-memory store: the StatefulSet goes, there is no claim to delete.
	inMemory := persisted.DeepCopy()
	inMemory.Spec.VolumeClaimTemplates = nil
	c, deleted = newClient(inMemory)
	if err := r.replaceStatefulSetIfPersistenceFlipped(context.Background(), c, platform, persisted.Name, true); err != nil {
		t.Fatal(err)
	}
	if len(*deleted) != 1 || (*deleted)[0] != persisted.Name {
		t.Fatalf("turning persistence on must delete only the StatefulSet, got %v", *deleted)
	}

	// Nothing live yet (a first install): nothing to replace.
	c, deleted = newClient()
	if err := r.replaceStatefulSetIfPersistenceFlipped(context.Background(), c, platform, persisted.Name, false); err != nil {
		t.Fatal(err)
	}
	if len(*deleted) != 0 {
		t.Fatalf("a first install has nothing to replace, got %v", *deleted)
	}
}
