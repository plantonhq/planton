package module

import (
	"testing"

	gcpprivatecapoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecapool/v1alpha1"
)

// Baseline values that name key usage but no CA options still send the
// ca_options block the provider requires, empty.
func TestX509ParametersWithoutCaOptions(t *testing.T) {
	args := x509Parameters(&gcpprivatecapoolv1alpha1.GcpPrivateCaPoolX509Parameters{
		KeyUsage: &gcpprivatecapoolv1alpha1.GcpPrivateCaPoolKeyUsage{
			BaseKeyUsage: &gcpprivatecapoolv1alpha1.GcpPrivateCaPoolBaseKeyUsage{DigitalSignature: true},
		},
	})
	if args.CaOptions == nil {
		t.Fatal("ca_options was left out; the provider requires it whenever baseline values are sent")
	}
	empty := caOptions(nil)
	if empty.IsCa != nil || empty.NonCa != nil || empty.MaxIssuerPathLength != nil || empty.ZeroMaxIssuerPathLength != nil {
		t.Fatalf("ca_options states something the spec never said: %+v", empty)
	}
}
