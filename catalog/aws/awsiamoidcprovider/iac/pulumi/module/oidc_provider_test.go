package module

// provider_url is the issuer as IAM's condition keys name it, so a role trust
// that references it matches: the host and path, never the scheme the URL was
// configured with.

import "testing"

func TestTheProviderUrlIsTheBareIssuer(t *testing.T) {
	cases := map[string]string{
		"https://dev.planton.ai":                               "dev.planton.ai",
		"https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED": "oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED",
		"http://issuer.example.com":                            "issuer.example.com",
		"token.actions.githubusercontent.com":                  "token.actions.githubusercontent.com",
	}
	for configured, want := range cases {
		if got := issuerWithoutScheme(configured); got != want {
			t.Errorf("issuerWithoutScheme(%q) = %q, want %q", configured, got, want)
		}
	}
}
