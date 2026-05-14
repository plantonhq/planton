package outputs

import (
	"encoding/json"
	"testing"
)

func TestFlatten_NilMap(t *testing.T) {
	got := Flatten(nil)
	if len(got) != 0 {
		t.Errorf("expected empty map for nil input, got %d entries", len(got))
	}
}

func TestFlatten_EmptyMap(t *testing.T) {
	got := Flatten(map[string]interface{}{})
	if len(got) != 0 {
		t.Errorf("expected empty map for empty input, got %d entries", len(got))
	}
}

func TestFlatten_StringValues(t *testing.T) {
	input := map[string]interface{}{
		"vpc_id":     "vpc-0abc123",
		"identifier": "https://api.example.com",
		"name":       "my-resource",
	}
	got := Flatten(input)

	assertFlatKey(t, got, "vpc_id", "vpc-0abc123")
	assertFlatKey(t, got, "identifier", "https://api.example.com")
	assertFlatKey(t, got, "name", "my-resource")
	assertFlatLen(t, got, 3)
}

func TestFlatten_EmptyString(t *testing.T) {
	got := Flatten(map[string]interface{}{"empty": ""})
	assertFlatKey(t, got, "empty", "")
}

func TestFlatten_Float64_WholeNumbers(t *testing.T) {
	input := map[string]interface{}{
		"token_lifetime":         float64(3600),
		"zero":                   float64(0),
		"negative":               float64(-1),
		"large":                  float64(86400),
		"token_lifetime_for_web": float64(7200),
	}
	got := Flatten(input)

	assertFlatKey(t, got, "token_lifetime", "3600")
	assertFlatKey(t, got, "zero", "0")
	assertFlatKey(t, got, "negative", "-1")
	assertFlatKey(t, got, "large", "86400")
	assertFlatKey(t, got, "token_lifetime_for_web", "7200")
}

func TestFlatten_Float64_Fractional(t *testing.T) {
	input := map[string]interface{}{
		"ratio":    float64(3.14),
		"half":     float64(0.5),
		"small":    float64(0.001),
		"neg_frac": float64(-2.5),
	}
	got := Flatten(input)

	assertFlatKey(t, got, "ratio", "3.14")
	assertFlatKey(t, got, "half", "0.5")
	assertFlatKey(t, got, "small", "0.001")
	assertFlatKey(t, got, "neg_frac", "-2.5")
}

func TestFlatten_BoolValues(t *testing.T) {
	input := map[string]interface{}{
		"allow_offline_access": true,
		"is_system":            false,
		"enforce_policies":     true,
	}
	got := Flatten(input)

	assertFlatKey(t, got, "allow_offline_access", "true")
	assertFlatKey(t, got, "is_system", "false")
	assertFlatKey(t, got, "enforce_policies", "true")
}

func TestFlatten_NilValues(t *testing.T) {
	input := map[string]interface{}{
		"signing_secret": nil,
		"client_id":      nil,
	}
	got := Flatten(input)

	assertFlatKey(t, got, "signing_secret", "")
	assertFlatKey(t, got, "client_id", "")
}

func TestFlatten_JsonNumber(t *testing.T) {
	input := map[string]interface{}{
		"count": json.Number("42"),
		"big":   json.Number("9999999999999999"),
	}
	got := Flatten(input)

	assertFlatKey(t, got, "count", "42")
	assertFlatKey(t, got, "big", "9999999999999999")
}

func TestFlatten_SliceOfStrings(t *testing.T) {
	input := map[string]interface{}{
		"nameservers": []interface{}{
			"ns-1234.awsdns-56.org",
			"ns-5678.awsdns-78.co.uk",
			"ns-9012.awsdns-34.net",
		},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "nameservers.0", "ns-1234.awsdns-56.org")
	assertFlatKey(t, got, "nameservers.1", "ns-5678.awsdns-78.co.uk")
	assertFlatKey(t, got, "nameservers.2", "ns-9012.awsdns-34.net")
	assertFlatLen(t, got, 3)
}

func TestFlatten_EmptySlice(t *testing.T) {
	input := map[string]interface{}{
		"tags": []interface{}{},
	}
	got := Flatten(input)
	assertFlatKey(t, got, "tags", "")
	assertFlatLen(t, got, 1)
}

func TestFlatten_SliceOfMixedScalars(t *testing.T) {
	input := map[string]interface{}{
		"ports": []interface{}{float64(8080), float64(443), "custom"},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "ports.0", "8080")
	assertFlatKey(t, got, "ports.1", "443")
	assertFlatKey(t, got, "ports.2", "custom")
}

func TestFlatten_SliceWithNil(t *testing.T) {
	input := map[string]interface{}{
		"items": []interface{}{"a", nil, "c"},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "items.0", "a")
	assertFlatKey(t, got, "items.1", "")
	assertFlatKey(t, got, "items.2", "c")
}

func TestFlatten_NestedMap(t *testing.T) {
	input := map[string]interface{}{
		"username_secret": map[string]interface{}{
			"name": "postgres.db-xyz.credentials.username",
			"key":  "username",
		},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "username_secret.name", "postgres.db-xyz.credentials.username")
	assertFlatKey(t, got, "username_secret.key", "username")
	assertFlatLen(t, got, 2)
}

func TestFlatten_EmptyNestedMap(t *testing.T) {
	input := map[string]interface{}{
		"metadata": map[string]interface{}{},
	}
	got := Flatten(input)
	assertFlatKey(t, got, "metadata", "")
	assertFlatLen(t, got, 1)
}

func TestFlatten_SliceOfMaps(t *testing.T) {
	input := map[string]interface{}{
		"private_subnets": []interface{}{
			map[string]interface{}{
				"id":   "subnet-abc",
				"cidr": "10.0.1.0/24",
			},
			map[string]interface{}{
				"id":   "subnet-def",
				"cidr": "10.0.2.0/24",
			},
		},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "private_subnets.0.id", "subnet-abc")
	assertFlatKey(t, got, "private_subnets.0.cidr", "10.0.1.0/24")
	assertFlatKey(t, got, "private_subnets.1.id", "subnet-def")
	assertFlatKey(t, got, "private_subnets.1.cidr", "10.0.2.0/24")
	assertFlatLen(t, got, 4)
}

func TestFlatten_DeepNesting(t *testing.T) {
	input := map[string]interface{}{
		"private_subnets": []interface{}{
			map[string]interface{}{
				"id":   "subnet-abc",
				"cidr": "10.0.1.0/24",
				"nat_gateway": map[string]interface{}{
					"id":        "nat-123",
					"public_ip": "34.56.78.90",
				},
			},
		},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "private_subnets.0.id", "subnet-abc")
	assertFlatKey(t, got, "private_subnets.0.cidr", "10.0.1.0/24")
	assertFlatKey(t, got, "private_subnets.0.nat_gateway.id", "nat-123")
	assertFlatKey(t, got, "private_subnets.0.nat_gateway.public_ip", "34.56.78.90")
	assertFlatLen(t, got, 4)
}

func TestFlatten_Auth0ResourceServerOutputs(t *testing.T) {
	// Simulates the real Auth0ResourceServer Pulumi module output shape.
	// token_lifetime, allow_offline_access, is_system, etc. are non-string
	// types that the old converter turned into "unknown".
	input := map[string]interface{}{
		"id":                     "auth0|abc123",
		"identifier":             "https://api.example.com",
		"name":                   "Example API",
		"signing_alg":            "RS256",
		"signing_secret":         nil,
		"token_lifetime":         float64(3600),
		"token_lifetime_for_web": float64(7200),
		"allow_offline_access":   true,
		"skip_consent_for_verifiable_first_party_clients": false,
		"enforce_policies": true,
		"token_dialect":    "access_token_authz",
		"is_system":        false,
		"client_id":        nil,
	}
	got := Flatten(input)

	assertFlatKey(t, got, "id", "auth0|abc123")
	assertFlatKey(t, got, "identifier", "https://api.example.com")
	assertFlatKey(t, got, "name", "Example API")
	assertFlatKey(t, got, "signing_alg", "RS256")
	assertFlatKey(t, got, "signing_secret", "")
	assertFlatKey(t, got, "token_lifetime", "3600")
	assertFlatKey(t, got, "token_lifetime_for_web", "7200")
	assertFlatKey(t, got, "allow_offline_access", "true")
	assertFlatKey(t, got, "skip_consent_for_verifiable_first_party_clients", "false")
	assertFlatKey(t, got, "enforce_policies", "true")
	assertFlatKey(t, got, "token_dialect", "access_token_authz")
	assertFlatKey(t, got, "is_system", "false")
	assertFlatKey(t, got, "client_id", "")
	assertFlatLen(t, got, 13)
}

func TestFlatten_MixedTopLevel(t *testing.T) {
	input := map[string]interface{}{
		"vpc_id":       "vpc-123",
		"cidr_block":   "10.0.0.0/16",
		"enable_dns":   true,
		"subnet_count": float64(3),
		"tags":         nil,
		"public_subnets": []interface{}{
			"subnet-aaa",
			"subnet-bbb",
		},
	}
	got := Flatten(input)

	assertFlatKey(t, got, "vpc_id", "vpc-123")
	assertFlatKey(t, got, "cidr_block", "10.0.0.0/16")
	assertFlatKey(t, got, "enable_dns", "true")
	assertFlatKey(t, got, "subnet_count", "3")
	assertFlatKey(t, got, "tags", "")
	assertFlatKey(t, got, "public_subnets.0", "subnet-aaa")
	assertFlatKey(t, got, "public_subnets.1", "subnet-bbb")
	assertFlatLen(t, got, 7)
}

func TestFlatten_Deterministic(t *testing.T) {
	input := map[string]interface{}{
		"z_key": "z",
		"a_key": "a",
		"m_key": "m",
	}
	first := Flatten(input)
	for i := 0; i < 10; i++ {
		again := Flatten(input)
		for k, v := range first {
			if again[k] != v {
				t.Fatalf("iteration %d: key %q differs: %q vs %q", i, k, v, again[k])
			}
		}
	}
}

// assertFlatKey checks that a key exists with the expected value in the flattened map.
func assertFlatKey(t *testing.T, m map[string]string, key, expected string) {
	t.Helper()
	val, ok := m[key]
	if !ok {
		t.Errorf("expected key %q not found; available keys: %v", key, flatMapKeys(m))
		return
	}
	if val != expected {
		t.Errorf("key %q: expected %q, got %q", key, expected, val)
	}
}

// assertFlatLen checks the total number of entries in the flattened map.
func assertFlatLen(t *testing.T, m map[string]string, expected int) {
	t.Helper()
	if len(m) != expected {
		t.Errorf("expected %d entries, got %d; keys: %v", expected, len(m), flatMapKeys(m))
	}
}

func flatMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
