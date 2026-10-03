package module

import (
	"sync"
	"testing"

	auth0tenantsettingsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0tenantsettings/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// recordingMocks answers the program's registrations with their inputs and
// its tenant lookup with an empty tenant, recording each resource type the
// program registers.
type recordingMocks struct {
	mu    sync.Mutex
	types []string
}

func (m *recordingMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	m.types = append(m.types, args.TypeToken)
	m.mu.Unlock()
	return args.Name + "-id", args.Inputs, nil
}

func (m *recordingMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.PropertyMap{}, nil
}

func (m *recordingMocks) registered(typeToken string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.types {
		if t == typeToken {
			return true
		}
	}
	return false
}

// TestResources runs the whole program under mocks. Auth0 refuses a tenant
// update that carries no setting ("Too few properties defined (0)"), so a spec
// that sets only the default domain must not declare the tenant resource at
// all: its settings are read through the provider's lookup instead.
func TestResources(t *testing.T) {
	const tenantType = "auth0:index/tenant:Tenant"
	const defaultDomainType = "auth0:index/customDomainDefault:CustomDomainDefault"
	friendly := "Planton"

	cases := []struct {
		name              string
		spec              *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec
		wantTenant        bool
		wantDefaultDomain bool
	}{
		{
			name: "only the default domain: no tenant resource, the default domain alone",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{
				DefaultCustomDomain: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "id.example.com"},
				},
			},
			wantDefaultDomain: true,
		},
		{
			name:       "a declared setting: the tenant resource",
			spec:       &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{FriendlyName: friendly},
			wantTenant: true,
		},
		{
			name: "an empty spec: nothing is written",
			spec: &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mocks := &recordingMocks{}
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				return Resources(ctx, iacInput(tc.spec))
			}, pulumi.WithMocks("auth0-tenant-settings", "test", mocks))
			if err != nil {
				t.Fatalf("the program must run to completion: %v", err)
			}
			if got := mocks.registered(tenantType); got != tc.wantTenant {
				t.Errorf("tenant resource registered = %v, want %v", got, tc.wantTenant)
			}
			if got := mocks.registered(defaultDomainType); got != tc.wantDefaultDomain {
				t.Errorf("default-domain resource registered = %v, want %v", got, tc.wantDefaultDomain)
			}
		})
	}
}
