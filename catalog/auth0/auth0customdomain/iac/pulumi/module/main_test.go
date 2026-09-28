package module

import (
	"testing"

	auth0customdomainv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0customdomain/v1alpha1"
	"github.com/plantonhq/planton/shared"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// mocks answers the program's resource registrations the way Auth0 answers a
// new Auth0-managed domain: its inputs, plus the CNAME verification method.
type mocks struct{}

func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	state := args.Inputs.Copy()
	if args.TypeToken == "auth0:index/customDomain:CustomDomain" {
		state["verifications"] = resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{
				"methods": resource.NewArrayProperty([]resource.PropertyValue{
					resource.NewObjectProperty(resource.PropertyMap{
						"name":   resource.NewStringProperty("cname"),
						"record": resource.NewStringProperty("e2e-cd-abc123.edge.tenants.us.auth0.com"),
					}),
				}),
			}),
		})
	}
	return args.Name + "-id", state, nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

// TestResources runs the whole program against mocks: a program that type-checks
// can still panic when its outputs are wired (an applier whose parameter the
// input output cannot supply), and only running it proves the wiring.
func TestResources(t *testing.T) {
	stackInput := &auth0customdomainv1alpha1.Auth0CustomDomainStackInput{
		Target: &auth0customdomainv1alpha1.Auth0CustomDomain{
			Metadata: &shared.CloudResourceMetadata{Name: "sign-in"},
			Spec: &auth0customdomainv1alpha1.Auth0CustomDomainSpec{
				Domain: "id.example.com",
				Type:   "auth0_managed_certs",
			},
		},
	}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		return Resources(ctx, stackInput)
	}, pulumi.WithMocks("auth0-custom-domain", "test", mocks{}))
	if err != nil {
		t.Fatalf("the program must run to completion: %v", err)
	}
}
