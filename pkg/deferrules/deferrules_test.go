package deferrules

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	netpolv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnetworkpolicy/v1alpha1"
	secretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessecret/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/stretchr/testify/assert"
)

// isToken stands in for a host's grammar: the Planton Platform's two
// reference prefixes.
func isToken(s string) bool {
	return strings.HasPrefix(s, "$secret/") || strings.HasPrefix(s, "$var/")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func binarySecret(name string, binary map[string]string) *secretv1alpha1.KubernetesSecret {
	return &secretv1alpha1.KubernetesSecret{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesSecret",
		Metadata:   &shared.CatalogObjectMetadata{Name: name},
		Spec: &secretv1alpha1.KubernetesSecretSpec{
			Name:      "runner-ca",
			Namespace: literal("builds"),
			SecretData: &secretv1alpha1.KubernetesSecretSpec_Opaque{
				Opaque: &secretv1alpha1.KubernetesSecretOpaqueData{BinaryData: binary},
			},
		},
	}
}

func egressPolicy(cidr string) *netpolv1alpha1.KubernetesNetworkPolicy {
	return &netpolv1alpha1.KubernetesNetworkPolicy{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesNetworkPolicy",
		Metadata:   &shared.CatalogObjectMetadata{Name: "build-egress"},
		Spec: &netpolv1alpha1.KubernetesNetworkPolicySpec{
			Namespace: literal("builds"),
			Name:      "build-egress",
			PolicyTypes: []netpolv1alpha1.KubernetesNetworkPolicySpec_KubernetesNetworkPolicyType{
				netpolv1alpha1.KubernetesNetworkPolicySpec_egress,
			},
			EgressRules: []*netpolv1alpha1.KubernetesNetworkPolicyEgressRule{{
				To: []*netpolv1alpha1.KubernetesNetworkPolicyPeer{{
					IpBlock: &netpolv1alpha1.KubernetesNetworkPolicyIpBlock{Cidr: literal(cidr)},
				}},
			}},
		},
	}
}

func TestSplit_DefersAShapeRuleOnAMapValueToken(t *testing.T) {
	err := protovalidate.Validate(binarySecret("runner-ca", map[string]string{"ca.crt": "$secret/runner-ca"}))
	if !assert.Error(t, err, "the base64 rule judges the token itself") {
		return
	}
	kept, deferred := Split(err, isToken)
	assert.NoError(t, kept)
	assert.Len(t, deferred, 1)
}

func TestSplit_DefersAShapeRuleOnAReferenceWrappersLiteralToken(t *testing.T) {
	err := protovalidate.Validate(egressPolicy("$var/build-nodes-range"))
	if !assert.Error(t, err, "the CIDR rule judges the token itself") {
		return
	}
	kept, deferred := Split(err, isToken)
	assert.NoError(t, kept)
	assert.Len(t, deferred, 1)
}

func TestSplit_KeepsAMalformedLiteral(t *testing.T) {
	err := protovalidate.Validate(binarySecret("runner-ca", map[string]string{"ca.crt": "not base64!"}))
	kept, deferred := Split(err, isToken)
	assert.Error(t, kept)
	assert.Empty(t, deferred)
}

func TestSplit_KeepsATokenOutsideTheSpec(t *testing.T) {
	// metadata is never resolved later: a token there is a malformed slug.
	secret := binarySecret("runner-ca", map[string]string{"ca.crt": "AQIDBA=="})
	secret.Metadata.Slug = "$secret/runner-ca"
	err := protovalidate.Validate(secret)
	if !assert.Error(t, err) {
		return
	}
	kept, deferred := Split(err, isToken)
	assert.Error(t, kept)
	assert.Empty(t, deferred)
}

func TestSplit_KeepsTheOtherViolationsBesideADeferredOne(t *testing.T) {
	err := protovalidate.Validate(binarySecret("runner-ca", map[string]string{
		"ca.crt":  "$secret/runner-ca",
		"key.pem": "not base64!",
	}))
	kept, deferred := Split(err, isToken)
	assert.Len(t, deferred, 1)
	if assert.Error(t, kept) {
		var validationErr *protovalidate.ValidationError
		assert.ErrorAs(t, kept, &validationErr)
		assert.Len(t, validationErr.Violations, 1)
	}
}

func TestSplit_WithoutAClassifierKeepsEverything(t *testing.T) {
	err := protovalidate.Validate(binarySecret("runner-ca", map[string]string{"ca.crt": "$secret/runner-ca"}))
	kept, deferred := Split(err, nil)
	assert.Equal(t, err, kept)
	assert.Empty(t, deferred)
}
