package capacityestimator

import (
	"strings"
	"testing"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	grafanav1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	keycloakv1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskeycloak/v1alpha1"
	kpgv1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
	capacityv1 "github.com/plantonhq/planton/finops/catalogkindcapacityderivation/v1"
	derivationv1 "github.com/plantonhq/planton/finops/catalogkindcostderivation/v1"
	"google.golang.org/protobuf/proto"
)

// postgresBinding is the kubernetespostgres capacity derivation's shape:
// one workload whose count is the instances field (provider default 1),
// with a required data volume and an optional WAL volume.
func postgresBinding() *capacityv1.CatalogKindCapacityDerivationSpec {
	return &capacityv1.CatalogKindCapacityDerivationSpec{
		Workloads: []*capacityv1.WorkloadBinding{{
			Label:         "instance",
			ResourcesPath: "resources",
			Instances: &capacityv1.InstanceCount{Count: &capacityv1.InstanceCount_FieldValue{
				FieldValue: &derivationv1.FieldValue{FieldPath: "instances", DefaultWhenUnset: "1"},
			}},
			Volumes: []*capacityv1.VolumeBinding{
				{Label: "data", SizePath: "storage.size"},
				{Label: "WAL", SizePath: "wal_storage.size"},
			},
		}},
		Exclusions: []*derivationv1.ConditionalText{
			{Text: "the dollar value of this capacity is the target cluster's node and storage-class economics"},
			{
				AppliesWhen: []*derivationv1.Condition{{FieldPath: "backup", Op: derivationv1.Condition_is_set}},
				Text:        "object storage for backups grows with database churn and retention",
			},
		},
	}
}

func postgres(spec *kpgv1.KubernetesPostgresSpec) *kpgv1.KubernetesPostgres {
	return &kpgv1.KubernetesPostgres{Spec: spec}
}

func resources(requestsCpu, requestsMemory, limitsCpu, limitsMemory string) *kubernetes.ContainerResources {
	return &kubernetes.ContainerResources{
		Requests: &kubernetes.CpuMemory{Cpu: requestsCpu, Memory: requestsMemory},
		Limits:   &kubernetes.CpuMemory{Cpu: limitsCpu, Memory: limitsMemory},
	}
}

// TestSumsReplicasVolumesAndRendersBasis pins the whole arithmetic against
// the production-HA shape: 3 instances x (1 CPU / 2Gi requests, 2 CPU /
// 4Gi limits) + 3 x (100Gi data + 20Gi WAL) volumes.
func TestSumsReplicasVolumesAndRendersBasis(t *testing.T) {
	manifest := postgres(&kpgv1.KubernetesPostgresSpec{
		Instances:  proto.Int32(3),
		Resources:  resources("1", "2Gi", "2", "4Gi"),
		Storage:    &kpgv1.KubernetesPostgresStorage{Size: "100Gi"},
		WalStorage: &kpgv1.KubernetesPostgresStorage{Size: "20Gi"},
	})

	preset, refusal, err := Evaluate(manifest, postgresBinding())
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	footprint := preset.GetCapacityFootprint()
	for _, check := range []struct{ name, got, want string }{
		{"cpu_requests", footprint.GetCpuRequests(), "3"},
		{"memory_requests", footprint.GetMemoryRequests(), "6Gi"},
		{"cpu_limits", footprint.GetCpuLimits(), "6"},
		{"memory_limits", footprint.GetMemoryLimits(), "12Gi"},
		{"persistent_storage", footprint.GetPersistentStorage(), "360Gi"},
	} {
		if check.got != check.want {
			t.Errorf("%s: got %q, want %q", check.name, check.got, check.want)
		}
	}
	wantBasis := "3 instances x (1 CPU / 2Gi memory requests, 2 CPU / 4Gi limits) + 3 x (100Gi data + 20Gi WAL) volumes"
	if footprint.GetBasis() != wantBasis {
		t.Errorf("basis:\n got %q\nwant %q", footprint.GetBasis(), wantBasis)
	}
	if len(preset.GetExclusions()) != 1 {
		t.Errorf("exclusions: got %d, want 1 (no backup configured)", len(preset.GetExclusions()))
	}
}

// TestMilliArithmeticAndSingleVolume pins the millicore path (500m x 3
// stays in millicores) and the single-volume basis grammar.
func TestMilliArithmeticAndSingleVolume(t *testing.T) {
	manifest := postgres(&kpgv1.KubernetesPostgresSpec{
		Instances: proto.Int32(3),
		Resources: resources("500m", "1Gi", "2", "2Gi"),
		Storage:   &kpgv1.KubernetesPostgresStorage{Size: "50Gi"},
	})

	preset, refusal, err := Evaluate(manifest, postgresBinding())
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	footprint := preset.GetCapacityFootprint()
	if footprint.GetCpuRequests() != "1500m" {
		t.Errorf("cpu_requests: got %q, want 1500m (3 x 500m never rounds to cores)", footprint.GetCpuRequests())
	}
	if footprint.GetPersistentStorage() != "150Gi" {
		t.Errorf("persistent_storage: got %q, want 150Gi", footprint.GetPersistentStorage())
	}
	wantBasis := "3 instances x (500m CPU / 1Gi memory requests, 2 CPU / 2Gi limits) + 3 x 50Gi data volumes"
	if footprint.GetBasis() != wantBasis {
		t.Errorf("basis:\n got %q\nwant %q", footprint.GetBasis(), wantBasis)
	}
}

// TestDefaultsAndConditionalProse pins the instances default (unset reads
// as the provider's documented 1) and the conditional exclusion riding
// exactly the configurations that match.
func TestDefaultsAndConditionalProse(t *testing.T) {
	manifest := postgres(&kpgv1.KubernetesPostgresSpec{
		Resources: resources("100m", "256Mi", "500m", "512Mi"),
		Storage:   &kpgv1.KubernetesPostgresStorage{Size: "5Gi"},
		Backup:    &kpgv1.KubernetesPostgresBackup{},
	})

	preset, refusal, err := Evaluate(manifest, postgresBinding())
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate: err=%v refusal=%+v", err, refusal)
	}
	footprint := preset.GetCapacityFootprint()
	if footprint.GetCpuRequests() != "100m" || footprint.GetPersistentStorage() != "5Gi" {
		t.Errorf("defaulted single instance: got cpu %q storage %q, want 100m and 5Gi",
			footprint.GetCpuRequests(), footprint.GetPersistentStorage())
	}
	wantBasis := "1 instance x (100m CPU / 256Mi memory requests, 500m CPU / 512Mi limits) + 1 x 5Gi data volume"
	if footprint.GetBasis() != wantBasis {
		t.Errorf("basis:\n got %q\nwant %q", footprint.GetBasis(), wantBasis)
	}
	if len(preset.GetExclusions()) != 2 {
		t.Errorf("exclusions: got %d, want 2 (backup is configured)", len(preset.GetExclusions()))
	}
}

// TestNoReservationRefuses pins the honest refusal: a manifest stating no
// requests, limits, or volumes gets no footprint, never an empty one.
func TestNoReservationRefuses(t *testing.T) {
	manifest := postgres(&kpgv1.KubernetesPostgresSpec{})
	_, refusal, err := Evaluate(manifest, postgresBinding())
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if refusal == nil || !strings.Contains(refusal.Reason, "no resource requests") {
		t.Fatalf("want the no-reservation refusal, got %+v", refusal)
	}
}

// TestSpecDeclaredDefaults pins the annotation fallbacks: a manifest
// omitting resources reserves the spec's own
// (default_container_resources) values -- the sizing the modules apply
// at deploy time -- with the defaults' origin named in the basis; a
// PRESENT storage block omitting its size reserves the field's
// (options.default) size; and an ABSENT storage block reserves nothing
// (the block is the volume's existence switch -- a default applied to an
// unconfigured volume would fabricate a reservation).
func TestSpecDeclaredDefaults(t *testing.T) {
	// kuberneteskeycloak's resources field carries the annotation
	// (requests 250m/768Mi, limits 1/1Gi) -- the paired fixture from the
	// committed 02-dev-sandbox preset, which omits resources entirely.
	keycloakBinding := &capacityv1.CatalogKindCapacityDerivationSpec{
		Workloads: []*capacityv1.WorkloadBinding{{
			Label:         "instance",
			ResourcesPath: "resources",
			Instances: &capacityv1.InstanceCount{Count: &capacityv1.InstanceCount_FieldValue{
				FieldValue: &derivationv1.FieldValue{FieldPath: "instances", DefaultWhenUnset: "1"},
			}},
		}},
		Exclusions: []*derivationv1.ConditionalText{
			{Text: "the dollar value of this capacity is the target cluster's node economics"},
		},
	}
	manifest := &keycloakv1.KubernetesKeycloak{Spec: &keycloakv1.KubernetesKeycloakSpec{}}
	preset, refusal, err := Evaluate(manifest, keycloakBinding)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(defaults): err=%v refusal=%+v", err, refusal)
	}
	footprint := preset.GetCapacityFootprint()
	for _, check := range []struct{ name, got, want string }{
		{"cpu_requests", footprint.GetCpuRequests(), "250m"},
		{"memory_requests", footprint.GetMemoryRequests(), "768Mi"},
		{"cpu_limits", footprint.GetCpuLimits(), "1"},
		{"memory_limits", footprint.GetMemoryLimits(), "1Gi"},
	} {
		if check.got != check.want {
			t.Errorf("defaults %s: got %q, want %q (the spec's own annotation)", check.name, check.got, check.want)
		}
	}
	if !strings.Contains(footprint.GetBasis(), "spec-declared defaults") {
		t.Errorf("basis %q does not name the defaults' origin", footprint.GetBasis())
	}

	// Explicit manifest values still win over the annotation.
	manifest.Spec.Resources = resources("2", "4Gi", "4", "8Gi")
	preset, refusal, err = Evaluate(manifest, keycloakBinding)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(explicit): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetCapacityFootprint().GetCpuRequests(); got != "2" {
		t.Errorf("explicit resources: got cpu_requests %q, want 2 (manifest wins over the annotation)", got)
	}
	if strings.Contains(preset.GetCapacityFootprint().GetBasis(), "spec-declared defaults") {
		t.Error("explicit resources wear the defaults marker -- the basis lies about the values' origin")
	}

	// kubernetesgrafana's storage.size carries (options.default) "10Gi":
	// a PRESENT block without a size reserves the default; an ABSENT
	// block reserves nothing.
	grafanaBinding := &capacityv1.CatalogKindCapacityDerivationSpec{
		Workloads: []*capacityv1.WorkloadBinding{{
			Label:     "state volume",
			Instances: &capacityv1.InstanceCount{Count: &capacityv1.InstanceCount_Constant{Constant: "1"}},
			Volumes:   []*capacityv1.VolumeBinding{{Label: "state", SizePath: "storage.size"}},
		}},
		Exclusions: []*derivationv1.ConditionalText{
			{Text: "the dollar value of this capacity is the target cluster's node economics"},
		},
	}
	grafana := &grafanav1.KubernetesGrafana{Spec: &grafanav1.KubernetesGrafanaSpec{
		Storage: &grafanav1.KubernetesGrafanaStorage{},
	}}
	preset, refusal, err = Evaluate(grafana, grafanaBinding)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(size default): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetCapacityFootprint().GetPersistentStorage(); got != "10Gi" {
		t.Errorf("present block, omitted size: got %q, want the 10Gi spec-declared default", got)
	}
	if !strings.Contains(preset.GetCapacityFootprint().GetBasis(), "spec-declared default size") {
		t.Errorf("basis %q does not name the size default's origin", preset.GetCapacityFootprint().GetBasis())
	}

	grafana.Spec.Storage = nil
	_, refusal, err = Evaluate(grafana, grafanaBinding)
	if err != nil {
		t.Fatalf("Evaluate(absent block): %v", err)
	}
	if refusal == nil {
		t.Fatal("an absent storage block must reserve NOTHING -- the size default fabricated a volume that never binds")
	}
}

// TestVolumeAppliesWhen pins the volume-existence gate: a mode that
// provisions no volume (ephemeral on emptyDir) keeps its size value and
// its spec default while binding NOTHING -- a size without a volume must
// contribute nothing, never a fabricated reservation.
func TestVolumeAppliesWhen(t *testing.T) {
	binding := &capacityv1.CatalogKindCapacityDerivationSpec{
		Workloads: []*capacityv1.WorkloadBinding{{
			Label:     "state volume",
			Instances: &capacityv1.InstanceCount{Count: &capacityv1.InstanceCount_Constant{Constant: "1"}},
			Volumes: []*capacityv1.VolumeBinding{{
				Label:    "state",
				SizePath: "storage.size",
				AppliesWhen: []*derivationv1.Condition{{
					FieldPath: "database", Op: derivationv1.Condition_is_unset,
				}},
			}},
		}},
		Exclusions: []*derivationv1.ConditionalText{
			{Text: "the dollar value of this capacity is the target cluster's node economics"},
		},
	}
	grafana := &grafanav1.KubernetesGrafana{Spec: &grafanav1.KubernetesGrafanaSpec{
		Storage: &grafanav1.KubernetesGrafanaStorage{Size: proto.String("25Gi")},
	}}
	preset, refusal, err := Evaluate(grafana, binding)
	if err != nil || refusal != nil {
		t.Fatalf("Evaluate(volume exists): err=%v refusal=%+v", err, refusal)
	}
	if got := preset.GetCapacityFootprint().GetPersistentStorage(); got != "25Gi" {
		t.Errorf("gated volume: got %q, want 25Gi", got)
	}

	// The gate condition failing means NO volume -- even with an
	// explicit size on the manifest.
	grafana.Spec.Database = &grafanav1.KubernetesGrafanaDatabase{}
	_, refusal, err = Evaluate(grafana, binding)
	if err != nil {
		t.Fatalf("Evaluate(volume gated off): %v", err)
	}
	if refusal == nil {
		t.Fatal("a gated-off volume still contributed -- the size fabricated a reservation the mode never binds")
	}
}

// TestQuantityGrammar pins the exact parse/render pairs the footprints
// depend on -- binary units survive, mixed totals fall to the largest
// dividing unit, and rounding is refused.
func TestQuantityGrammar(t *testing.T) {
	cases := []struct{ in, want string }{
		{"256Mi", "256Mi"},
		{"1536Mi", "1536Mi"}, // 1.5Gi does not divide Gi -- stays Mi
		{"2G", "2G"},
	}
	for _, c := range cases {
		bytes, err := parseQuantity(c.in, false)
		if err != nil {
			t.Fatalf("parseQuantity(%q): %v", c.in, err)
		}
		if got := renderBytes(bytes); got != c.want {
			t.Errorf("render(parse(%q)): got %q, want %q", c.in, got, c.want)
		}
	}
	if milli, err := parseQuantity("0.5", true); err != nil || renderCpu(milli) != "500m" {
		t.Errorf("0.5 cores: got %v/%v, want 500m", milli, err)
	}
	if _, err := parseQuantity("0.3m", true); err == nil {
		t.Error("0.3m parsed -- fractional millicores must refuse, never round")
	}
}
