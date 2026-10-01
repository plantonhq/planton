package module

import (
	"reflect"
	"testing"

	"github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
)

// The sizing contract: a component's resources render only the quantities the
// manifest set, because the operator merges each one over its own measured
// default -- an unset quantity arriving empty would be read as a choice. The
// Terraform module renders the same maps (iac/tf/locals.tf).

func TestPlatformSpecBody_NoSizingRendersNoResources(t *testing.T) {
	body := platformSpecBody(localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		ControlPlane: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformControlPlane{},
		Temporal:     &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformTemporal{},
	}))
	for _, key := range []string{"controlPlane", "temporal", "openfga"} {
		if _, ok := body[key]; ok {
			t.Errorf("%s rendered with nothing declared: %#v", key, body[key])
		}
	}
}

func TestPlatformSpecBody_OneQuantityRendersAlone(t *testing.T) {
	body := platformSpecBody(localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		ControlPlane: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformControlPlane{
			Resources: &kubernetes.ContainerResources{Limits: &kubernetes.CpuMemory{Memory: "8Gi"}},
		},
	}))
	want := map[string]interface{}{"resources": map[string]interface{}{"limits": map[string]interface{}{"memory": "8Gi"}}}
	if got := body["controlPlane"]; !reflect.DeepEqual(got, want) {
		t.Errorf("controlPlane rendered\n got: %#v\nwant: %#v", got, want)
	}
}

// Each component's resources land at the operator's path for it, including
// the store's ceiling beside its size, the gateway without a local port, and
// Temporal's services by name.
func TestPlatformSpecBody_EveryComponentRendersAtItsOperatorPath(t *testing.T) {
	sized := &kubernetes.ContainerResources{Requests: &kubernetes.CpuMemory{Cpu: "500m"}}
	res := map[string]interface{}{"requests": map[string]interface{}{"cpu": "500m"}}
	body := platformSpecBody(localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		ControlPlane: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformControlPlane{Resources: sized},
		Console:      &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformConsole{Resources: sized},
		Runner:       &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformRunner{Resources: sized},
		Gateway:      &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformGateway{Resources: sized},
		Identity:     &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformIdentity{Resources: sized},
		Database: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformDatabase{
			Postgresql: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresql{Resources: sized},
			Redis:      &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformRedis{Resources: sized, MaxMemory: "2gb"},
		},
		Vault:      &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{Resources: sized},
		Components: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformComponents{Graph: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformGraph{Resources: sized}},
		Temporal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformTemporal{
			History: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformTemporalService{Resources: sized},
		},
		Openfga: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformOpenFga{Resources: sized},
	}))
	at := func(path ...string) interface{} {
		var node interface{} = body
		for _, key := range path {
			m, ok := node.(map[string]interface{})
			if !ok {
				return nil
			}
			node = m[key]
		}
		return node
	}
	for _, path := range [][]string{
		{"controlPlane", "resources"}, {"console", "resources"}, {"runner", "resources"}, {"gateway", "resources"},
		{"identity", "resources"}, {"database", "postgresql", "resources"}, {"database", "redis", "resources"},
		{"vault", "resources"}, {"components", "graph", "resources"}, {"temporal", "history", "resources"},
		{"openfga", "resources"},
	} {
		if got := at(path...); !reflect.DeepEqual(got, res) {
			t.Errorf("%v rendered %#v, want %#v", path, got, res)
		}
	}
	if got := at("database", "redis", "maxMemory"); got != "2gb" {
		t.Errorf("the store's ceiling must render beside its size, got %#v", got)
	}
	if _, ok := at("gateway").(map[string]interface{})["localPort"]; ok {
		t.Error("a gateway sized without a local port must not render one")
	}
	if got := at("temporal", "frontend"); got != nil {
		t.Errorf("an unsized Temporal service must not render, got %#v", got)
	}
}
