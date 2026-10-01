package module

import (
	"testing"

	auth0promptcustomtextv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0promptcustomtext/v1alpha1"
)

func TestRenderBody(t *testing.T) {
	cases := []struct {
		name    string
		screens map[string]*auth0promptcustomtextv1alpha1.Auth0PromptScreenText
		want    string
	}{
		{
			name: "one screen renders as the document Auth0 takes",
			screens: map[string]*auth0promptcustomtextv1alpha1.Auth0PromptScreenText{
				"login": {Texts: map[string]string{"title": "Welcome back"}},
			},
			want: `{"login":{"title":"Welcome back"}}`,
		},
		{
			name: "screens and keys are written in sorted order, as jsonencode writes them",
			screens: map[string]*auth0promptcustomtextv1alpha1.Auth0PromptScreenText{
				"signup-password": {Texts: map[string]string{"title": "Create your password"}},
				"signup-id":       {Texts: map[string]string{"title": "Create your account", "description": "Sign up to Planton."}},
			},
			want: `{"signup-id":{"description":"Sign up to Planton.","title":"Create your account"},"signup-password":{"title":"Create your password"}}`,
		},
		{
			name: "variables pass through, and HTML characters are escaped as jsonencode escapes them",
			screens: map[string]*auth0promptcustomtextv1alpha1.Auth0PromptScreenText{
				"login": {Texts: map[string]string{"description": "Log in to Planton to continue to ${clientName} & <b>more</b>."}},
			},
			want: `{"login":{"description":"Log in to Planton to continue to ${clientName} \u0026 \u003cb\u003emore\u003c/b\u003e."}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := renderBody(tc.screens)
			if err != nil {
				t.Fatalf("renderBody: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
