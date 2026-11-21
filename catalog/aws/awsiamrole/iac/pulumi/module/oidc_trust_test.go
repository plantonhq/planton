package module

// A role's OIDC trust names the issuer the way IAM's condition keys do: the
// host and path, never the https:// in front. A provider URL handed over with
// its scheme would write a condition no token ever matches.

import (
	"encoding/json"
	"testing"

	iamrolev1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func conditionKeys(t *testing.T, providerUrl string) map[string]bool {
	t.Helper()
	document, err := composeOidcTrustPolicy(&iamrolev1.AwsIamRoleOidcTrust{
		ProviderArn:      literal("arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED"),
		ProviderUrl:      literal(providerUrl),
		Subjects:         []string{"system:serviceaccount:orders:orders-api"},
		WildcardSubjects: []string{"repo:acme/orders:*"},
	})
	if err != nil {
		t.Fatalf("composeOidcTrustPolicy: %v", err)
	}
	var parsed struct {
		Statement []struct {
			Condition map[string]map[string]any
		}
	}
	if err := json.Unmarshal([]byte(document), &parsed); err != nil {
		t.Fatalf("the trust document is not JSON: %v", err)
	}
	keys := map[string]bool{}
	for _, statement := range parsed.Statement {
		for _, operator := range statement.Condition {
			for key := range operator {
				keys[key] = true
			}
		}
	}
	return keys
}

func TestTheTrustNamesTheBareIssuer(t *testing.T) {
	for _, url := range []string{
		"oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED",
		"https://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED",
		"http://oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED",
	} {
		keys := conditionKeys(t, url)
		for _, want := range []string{"oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED:sub", "oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED:aud"} {
			if !keys[want] {
				t.Errorf("provider URL %q: the trust's condition keys %v lack %q", url, keys, want)
			}
		}
		if len(keys) != 2 {
			t.Errorf("provider URL %q: the trust's condition keys are %v, want only the issuer's :sub and :aud", url, keys)
		}
	}
}
