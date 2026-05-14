//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	auth0v1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0resourceserver/v1"
	awsvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsvpc/v1"
	gcpdnsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpdnszone/v1"
	k8spgv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1"
)

func TestPopulate_StringFields(t *testing.T) {
	msg := &auth0v1.Auth0ResourceServerStackOutputs{}
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
	msg := &gcpdnsv1.GcpDnsZoneStackOutputs{}
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

func TestPopulate_NestedMessageDotPath(t *testing.T) {
	msg := &k8spgv1.KubernetesPostgresStackOutputs{}
	outputs := map[string]string{
		"namespace":            "db-namespace",
		"service":              "postgres-svc",
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
	msg := &k8spgv1.KubernetesPostgresStackOutputs{}
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
	msg := &awsvpcv1.AwsVpcStackOutputs{}
	outputs := map[string]string{
		"vpc_id":                  "vpc-abc",
		"private_subnets[0].id":   "subnet-001",
		"private_subnets[0].cidr": "10.0.1.0/24",
		"private_subnets[0].nat_gateway.public_ip": "34.56.78.90",
		"private_subnets[1].id":                    "subnet-002",
		"private_subnets[1].cidr":                  "10.0.2.0/24",
	}

	if err := populateMessage(msg, outputs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msg.GetVpcId() != "vpc-abc" {
		t.Errorf("vpc_id: expected %q, got %q", "vpc-abc", msg.GetVpcId())
	}
	subnets := msg.GetPrivateSubnets()
	if len(subnets) != 2 {
		t.Fatalf("private_subnets: expected 2, got %d", len(subnets))
	}
	if subnets[0].GetId() != "subnet-001" {
		t.Errorf("private_subnets[0].id: expected %q, got %q", "subnet-001", subnets[0].GetId())
	}
	if subnets[0].GetCidr() != "10.0.1.0/24" {
		t.Errorf("private_subnets[0].cidr: expected %q, got %q", "10.0.1.0/24", subnets[0].GetCidr())
	}
	if subnets[0].GetNatGateway() == nil {
		t.Fatal("private_subnets[0].nat_gateway: expected non-nil")
	}
	if subnets[0].GetNatGateway().GetPublicIp() != "34.56.78.90" {
		t.Errorf("private_subnets[0].nat_gateway.public_ip: expected %q, got %q",
			"34.56.78.90", subnets[0].GetNatGateway().GetPublicIp())
	}
	if subnets[1].GetId() != "subnet-002" {
		t.Errorf("private_subnets[1].id: expected %q, got %q", "subnet-002", subnets[1].GetId())
	}
}

func TestPopulate_UnknownFieldSkipped(t *testing.T) {
	msg := &auth0v1.Auth0ResourceServerStackOutputs{}
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
	msg := &auth0v1.Auth0ResourceServerStackOutputs{}
	err := populateMessage(msg, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.GetId() != "" {
		t.Errorf("expected empty id, got %q", msg.GetId())
	}
}

func TestPopulate_EmptyRepeatedField(t *testing.T) {
	msg := &gcpdnsv1.GcpDnsZoneStackOutputs{}
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
