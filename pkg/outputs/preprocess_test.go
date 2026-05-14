package outputs

import (
	"testing"
)

func TestPreprocessKeys_DotBeforeBracket(t *testing.T) {
	input := map[string]string{
		"subnets.[0].id":   "subnet-abc",
		"subnets.[1].cidr": "10.0.0.0/24",
	}
	got := preprocessKeys(input)

	assertKey(t, got, "subnets[0].id", "subnet-abc")
	assertKey(t, got, "subnets[1].cidr", "10.0.0.0/24")
}

func TestPreprocessKeys_HyphensToUnderscores(t *testing.T) {
	input := map[string]string{
		"load-balancer-arn": "arn:aws:elasticloadbalancing:...",
	}
	got := preprocessKeys(input)

	assertKey(t, got, "load_balancer_arn", "arn:aws:elasticloadbalancing:...")
}

func TestPreprocessKeys_Combined(t *testing.T) {
	input := map[string]string{
		"private-subnets.[0].nat-gateway.public-ip": "34.56.78.90",
	}
	got := preprocessKeys(input)

	assertKey(t, got, "private_subnets[0].nat_gateway.public_ip", "34.56.78.90")
}

func TestPreprocessKeys_NoChanges(t *testing.T) {
	input := map[string]string{
		"vpc_id": "vpc-123",
		"name":   "test",
	}
	got := preprocessKeys(input)

	assertKey(t, got, "vpc_id", "vpc-123")
	assertKey(t, got, "name", "test")
}

func TestPreprocessKeys_EmptyMap(t *testing.T) {
	got := preprocessKeys(map[string]string{})
	if len(got) != 0 {
		t.Errorf("expected empty map, got %d entries", len(got))
	}
}

func assertKey(t *testing.T, m map[string]string, key, expectedValue string) {
	t.Helper()
	val, ok := m[key]
	if !ok {
		t.Errorf("expected key %q not found in map; keys: %v", key, mapKeys(m))
		return
	}
	if val != expectedValue {
		t.Errorf("key %q: expected %q, got %q", key, expectedValue, val)
	}
}

func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
