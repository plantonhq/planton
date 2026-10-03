//go:build !codegen
// +build !codegen

package outputs

import (
	"strings"
	"testing"

	auth0v1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
	azurelbv1 "github.com/plantonhq/planton/catalog/azure/azureloadbalancer/v1alpha1"
	gcpdnsv1 "github.com/plantonhq/planton/catalog/gcp/gcpdnszone/v1alpha1"
	gcpsubnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsubnetwork/v1alpha1"
	k8spgv1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetespostgres/v1alpha1"
)

func TestPopulate_StringFields(t *testing.T) {
	msg := &auth0v1.Auth0ResourceServerOutputs{}
	outputs := map[string]string{
		"id":         "abc123",
		"identifier": "https://api.example.com/",
		"name":       "Example API",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetId() != "abc123" {
		t.Errorf("id: expected %q, got %q", "abc123", msg.GetId())
	}
	if msg.GetIdentifier() != "https://api.example.com/" {
		t.Errorf("identifier: expected %q, got %q", "https://api.example.com/", msg.GetIdentifier())
	}
	if msg.GetName() != "Example API" {
		t.Errorf("name: expected %q, got %q", "Example API", msg.GetName())
	}
}

func TestPopulate_RepeatedString(t *testing.T) {
	msg := &gcpdnsv1.GcpDnsZoneOutputs{}
	outputs := map[string]string{
		"zone_id":       "zone-123",
		"zone_name":     "example-zone",
		"nameservers.0": "ns-1234.awsdns-56.org",
		"nameservers.1": "ns-5678.awsdns-78.co.uk",
		"nameservers.2": "ns-9012.awsdns-34.net",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetZoneId() != "zone-123" {
		t.Errorf("zone_id: expected %q, got %q", "zone-123", msg.GetZoneId())
	}
	if len(msg.GetNameservers()) != 3 {
		t.Fatalf("nameservers: expected 3, got %d", len(msg.GetNameservers()))
	}
	if msg.GetNameservers()[0] != "ns-1234.awsdns-56.org" {
		t.Errorf("nameservers[0]: expected %q, got %q", "ns-1234.awsdns-56.org", msg.GetNameservers()[0])
	}
	if msg.GetNameservers()[2] != "ns-9012.awsdns-34.net" {
		t.Errorf("nameservers[2]: expected %q, got %q", "ns-9012.awsdns-34.net", msg.GetNameservers()[2])
	}
}

func TestPopulate_HyphenatedFieldNames(t *testing.T) {
	// Field-name normalization happens per segment at lookup time; a
	// hyphenated IaC output name still lands on its snake_case proto field.
	msg := &gcpdnsv1.GcpDnsZoneOutputs{}
	outputs := map[string]string{
		"zone-id":   "zone-123",
		"zone-name": "example-zone",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.GetZoneId() != "zone-123" {
		t.Errorf("zone_id: expected %q, got %q", "zone-123", msg.GetZoneId())
	}
	if msg.GetZoneName() != "example-zone" {
		t.Errorf("zone_name: expected %q, got %q", "example-zone", msg.GetZoneName())
	}
}

func TestPopulate_MapEntriesFromDottedKeys(t *testing.T) {
	// The flattener emits one dotted key per map entry
	// ("backend_pool_ids.web" = id). Map keys are user data -- pool or
	// rule names -- and must be preserved verbatim, hyphens included.
	msg := &azurelbv1.AzureLoadBalancerOutputs{}
	outputs := map[string]string{
		"load_balancer_id":              "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/app-lb",
		"backend_pool_ids.web":          "/subscriptions/s/.../backendAddressPools/web",
		"backend_pool_ids.api":          "/subscriptions/s/.../backendAddressPools/api",
		"nat_rule_ids.ssh-admin":        "/subscriptions/s/.../inboundNatRules/ssh-admin",
		"frontend_ip_configuration_ids": "",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := msg.GetBackendPoolIds()["web"]; got != "/subscriptions/s/.../backendAddressPools/web" {
		t.Errorf("backend_pool_ids[web]: got %q", got)
	}
	if got := msg.GetBackendPoolIds()["api"]; got != "/subscriptions/s/.../backendAddressPools/api" {
		t.Errorf("backend_pool_ids[api]: got %q", got)
	}
	if got := msg.GetNatRuleIds()["ssh-admin"]; got != "/subscriptions/s/.../inboundNatRules/ssh-admin" {
		t.Errorf("nat_rule_ids[ssh-admin]: map key must be preserved verbatim, got map %v", msg.GetNatRuleIds())
	}
	if len(msg.GetFrontendIpConfigurationIds()) != 0 {
		t.Errorf("empty map sentinel: expected no entries, got %v", msg.GetFrontendIpConfigurationIds())
	}
}

func TestPopulate_NestedMessageDotPath(t *testing.T) {
	msg := &k8spgv1.KubernetesPostgresOutputs{}
	outputs := map[string]string{
		"namespace":            "db-namespace",
		"rw_service":           "postgres-svc-rw",
		"username_secret.name": "postgres.db-xyz.credentials",
		"username_secret.key":  "username",
		"password_secret.name": "postgres.db-xyz.credentials",
		"password_secret.key":  "password",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetNamespace() != "db-namespace" {
		t.Errorf("namespace: expected %q, got %q", "db-namespace", msg.GetNamespace())
	}
	if msg.GetUsernameSecret() == nil {
		t.Fatal("username_secret: expected non-nil")
	}
	if msg.GetUsernameSecret().GetName() != "postgres.db-xyz.credentials" {
		t.Errorf("username_secret.name: expected %q, got %q",
			"postgres.db-xyz.credentials", msg.GetUsernameSecret().GetName())
	}
	if msg.GetUsernameSecret().GetKey() != "username" {
		t.Errorf("username_secret.key: expected %q, got %q",
			"username", msg.GetUsernameSecret().GetKey())
	}
	if msg.GetPasswordSecret().GetKey() != "password" {
		t.Errorf("password_secret.key: expected %q, got %q",
			"password", msg.GetPasswordSecret().GetKey())
	}
}

func TestPopulate_NestedMessageJSON(t *testing.T) {
	msg := &k8spgv1.KubernetesPostgresOutputs{}
	outputs := map[string]string{
		"username_secret": `{"name":"pg-secret","key":"user"}`,
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetUsernameSecret() == nil {
		t.Fatal("username_secret: expected non-nil")
	}
	if msg.GetUsernameSecret().GetName() != "pg-secret" {
		t.Errorf("username_secret.name: expected %q, got %q",
			"pg-secret", msg.GetUsernameSecret().GetName())
	}
	if msg.GetUsernameSecret().GetKey() != "user" {
		t.Errorf("username_secret.key: expected %q, got %q",
			"user", msg.GetUsernameSecret().GetKey())
	}
}

func TestPopulate_RepeatedMessageWithBracketIndex(t *testing.T) {
	msg := &gcpsubnetworkv1alpha1.GcpSubnetworkOutputs{}
	outputs := map[string]string{
		"subnetwork_name":                   "my-subnet",
		"secondary_ranges[0].range_name":    "pods",
		"secondary_ranges[0].ip_cidr_range": "10.1.0.0/16",
		"secondary_ranges[1].range_name":    "services",
		"secondary_ranges[1].ip_cidr_range": "10.2.0.0/20",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetSubnetworkName() != "my-subnet" {
		t.Errorf("subnetwork_name: expected %q, got %q", "my-subnet", msg.GetSubnetworkName())
	}
	ranges := msg.GetSecondaryRanges()
	if len(ranges) != 2 {
		t.Fatalf("secondary_ranges: expected 2, got %d", len(ranges))
	}
	if ranges[0].GetRangeName() != "pods" {
		t.Errorf("secondary_ranges[0].range_name: expected %q, got %q", "pods", ranges[0].GetRangeName())
	}
	if ranges[0].GetIpCidrRange() != "10.1.0.0/16" {
		t.Errorf("secondary_ranges[0].ip_cidr_range: expected %q, got %q", "10.1.0.0/16", ranges[0].GetIpCidrRange())
	}
	if ranges[1].GetRangeName() != "services" {
		t.Errorf("secondary_ranges[1].range_name: expected %q, got %q", "services", ranges[1].GetRangeName())
	}
}

func TestPopulate_UnknownFieldSkipped(t *testing.T) {
	msg := &auth0v1.Auth0ResourceServerOutputs{}
	outputs := map[string]string{
		"id":                     "abc123",
		"nonexistent_field":      "should-be-skipped",
		"another_unknown.nested": "also-skipped",
	}

	err := populateMessage(msg, outputs)
	if err != nil {
		t.Fatalf("expected no error (unknown fields should be skipped), got: %v", err)
	}
	if msg.GetId() != "abc123" {
		t.Errorf("id: expected %q, got %q", "abc123", msg.GetId())
	}
}

func TestPopulate_EmptyMap(t *testing.T) {
	msg := &auth0v1.Auth0ResourceServerOutputs{}
	err := populateMessage(msg, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.GetId() != "" {
		t.Errorf("expected empty id, got %q", msg.GetId())
	}
}

func TestPopulate_EmptyRepeatedField(t *testing.T) {
	msg := &gcpdnsv1.GcpDnsZoneOutputs{}
	outputs := map[string]string{
		"nameservers": "",
	}

	err := populateMessage(msg, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msg.GetNameservers()) != 0 {
		t.Errorf("expected empty nameservers, got %d entries", len(msg.GetNameservers()))
	}
}

func TestPopulate_MaskedSecretRepeatedFieldIsDiagnosed(t *testing.T) {
	// A secret-marked repeated output read WITHOUT secrets shown arrives as
	// the literal "[secret]" on the bare key (no indexes to flatten) — e.g.
	// awssesemailidentity's dkim_tokens, observed live 2026-08-12. Every
	// sanctioned reader now unmasks; this locks the diagnosis so a reader
	// regression names its real cause instead of "no array index provided".
	msg := &gcpdnsv1.GcpDnsZoneOutputs{}

	err := setFieldRecursively(msg.ProtoReflect(), []string{"nameservers"}, "[secret]", 0)
	if err == nil {
		t.Fatal("expected a masked repeated value to be rejected, got nil")
	}
	if !strings.Contains(err.Error(), "--show-secrets") {
		t.Errorf("the error must teach the reader contract (--show-secrets); got: %v", err)
	}
	if len(msg.GetNameservers()) != 0 {
		t.Errorf("the masked placeholder must never populate the list; got %d entries", len(msg.GetNameservers()))
	}

	// populateMessage stays non-fatal for the whole map: the masked field is
	// skipped with the diagnosis, sibling fields still populate.
	msg2 := &gcpdnsv1.GcpDnsZoneOutputs{}
	outputs := map[string]string{
		"zone_id":     "zone-123",
		"nameservers": "[secret]",
	}
	if err := populateMessage(msg2, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg2.GetZoneId() != "zone-123" {
		t.Errorf("sibling field must still populate; got %q", msg2.GetZoneId())
	}
}
