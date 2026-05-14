//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	auth0v1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0resourceserver/v1"
	awsvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsvpc/v1"
	gcpdnsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpdnszone/v1"
	k8spgv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetespostgres/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
)

func TestTransform_Auth0ResourceServer(t *testing.T) {
	outputs := map[string]string{
		"id":               "6832a1b0c4e5f7d9e0a1b2c3",
		"identifier":       "https://api.leftbin.ai/",
		"name":             "Leftbin OS API",
		"signing_alg":      "RS256",
		"token_lifetime":   "86400",
		"enforce_policies": "true",
		"token_dialect":    "access_token_authz",
		"is_system":        "false",
		"client_id":        "abc123xyz",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_Auth0ResourceServer, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := msg.(*auth0v1.Auth0ResourceServerStackOutputs)
	if !ok {
		t.Fatalf("expected *Auth0ResourceServerStackOutputs, got %T", msg)
	}

	if typed.GetId() != "6832a1b0c4e5f7d9e0a1b2c3" {
		t.Errorf("id: expected %q, got %q", "6832a1b0c4e5f7d9e0a1b2c3", typed.GetId())
	}
	if typed.GetIdentifier() != "https://api.leftbin.ai/" {
		t.Errorf("identifier: expected %q, got %q", "https://api.leftbin.ai/", typed.GetIdentifier())
	}
	if typed.GetName() != "Leftbin OS API" {
		t.Errorf("name: expected %q, got %q", "Leftbin OS API", typed.GetName())
	}
	if typed.GetSigningAlg() != "RS256" {
		t.Errorf("signing_alg: expected %q, got %q", "RS256", typed.GetSigningAlg())
	}
	if typed.GetTokenLifetime() != "86400" {
		t.Errorf("token_lifetime: expected %q, got %q", "86400", typed.GetTokenLifetime())
	}
	if typed.GetEnforcePolicies() != "true" {
		t.Errorf("enforce_policies: expected %q, got %q", "true", typed.GetEnforcePolicies())
	}
	if typed.GetClientId() != "abc123xyz" {
		t.Errorf("client_id: expected %q, got %q", "abc123xyz", typed.GetClientId())
	}
}

func TestTransform_GcpDnsZone(t *testing.T) {
	outputs := map[string]string{
		"zone_id":       "123456789",
		"zone_name":     "example-zone",
		"nameservers.0": "ns-cloud-a1.googledomains.com",
		"nameservers.1": "ns-cloud-a2.googledomains.com",
		"nameservers.2": "ns-cloud-a3.googledomains.com",
		"nameservers.3": "ns-cloud-a4.googledomains.com",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_GcpDnsZone, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := msg.(*gcpdnsv1.GcpDnsZoneStackOutputs)
	if !ok {
		t.Fatalf("expected *GcpDnsZoneStackOutputs, got %T", msg)
	}

	if typed.GetZoneId() != "123456789" {
		t.Errorf("zone_id: expected %q, got %q", "123456789", typed.GetZoneId())
	}
	ns := typed.GetNameservers()
	if len(ns) != 4 {
		t.Fatalf("nameservers: expected 4, got %d", len(ns))
	}
	if ns[0] != "ns-cloud-a1.googledomains.com" {
		t.Errorf("nameservers[0]: expected %q, got %q", "ns-cloud-a1.googledomains.com", ns[0])
	}
	if ns[3] != "ns-cloud-a4.googledomains.com" {
		t.Errorf("nameservers[3]: expected %q, got %q", "ns-cloud-a4.googledomains.com", ns[3])
	}
}

func TestTransform_AwsVpc(t *testing.T) {
	outputs := map[string]string{
		"vpc_id":                                    "vpc-0abc123",
		"internet_gateway_id":                       "igw-xyz789",
		"vpc_cidr":                                  "10.0.0.0/16",
		"private_subnets[0].id":                     "subnet-priv-001",
		"private_subnets[0].name":                   "private-a",
		"private_subnets[0].cidr":                   "10.0.1.0/24",
		"private_subnets[0].nat_gateway.id":         "nat-001",
		"private_subnets[0].nat_gateway.public_ip":  "34.56.78.90",
		"private_subnets[0].nat_gateway.private_ip": "10.0.1.5",
		"private_subnets[1].id":                     "subnet-priv-002",
		"private_subnets[1].name":                   "private-b",
		"private_subnets[1].cidr":                   "10.0.2.0/24",
		"public_subnets[0].id":                      "subnet-pub-001",
		"public_subnets[0].name":                    "public-a",
		"public_subnets[0].cidr":                    "10.0.100.0/24",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_AwsVpc, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := msg.(*awsvpcv1.AwsVpcStackOutputs)
	if !ok {
		t.Fatalf("expected *AwsVpcStackOutputs, got %T", msg)
	}

	if typed.GetVpcId() != "vpc-0abc123" {
		t.Errorf("vpc_id: expected %q, got %q", "vpc-0abc123", typed.GetVpcId())
	}
	if typed.GetVpcCidr() != "10.0.0.0/16" {
		t.Errorf("vpc_cidr: expected %q, got %q", "10.0.0.0/16", typed.GetVpcCidr())
	}

	privSubnets := typed.GetPrivateSubnets()
	if len(privSubnets) != 2 {
		t.Fatalf("private_subnets: expected 2, got %d", len(privSubnets))
	}
	if privSubnets[0].GetId() != "subnet-priv-001" {
		t.Errorf("private_subnets[0].id: expected %q, got %q", "subnet-priv-001", privSubnets[0].GetId())
	}
	if privSubnets[0].GetNatGateway() == nil {
		t.Fatal("private_subnets[0].nat_gateway: expected non-nil")
	}
	if privSubnets[0].GetNatGateway().GetPublicIp() != "34.56.78.90" {
		t.Errorf("private_subnets[0].nat_gateway.public_ip: expected %q, got %q",
			"34.56.78.90", privSubnets[0].GetNatGateway().GetPublicIp())
	}

	pubSubnets := typed.GetPublicSubnets()
	if len(pubSubnets) != 1 {
		t.Fatalf("public_subnets: expected 1, got %d", len(pubSubnets))
	}
	if pubSubnets[0].GetName() != "public-a" {
		t.Errorf("public_subnets[0].name: expected %q, got %q", "public-a", pubSubnets[0].GetName())
	}
}

func TestTransform_KubernetesPostgres(t *testing.T) {
	outputs := map[string]string{
		"namespace":            "db-prod",
		"service":              "postgres-master",
		"kube_endpoint":        "postgres-master.db-prod.svc.cluster.local:5432",
		"external_hostname":    "postgres.example.com",
		"username_secret.name": "postgres.db-prod.credentials.postgresql.acid.zalan.do",
		"username_secret.key":  "username",
		"password_secret.name": "postgres.db-prod.credentials.postgresql.acid.zalan.do",
		"password_secret.key":  "password",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_KubernetesPostgres, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed, ok := msg.(*k8spgv1.KubernetesPostgresStackOutputs)
	if !ok {
		t.Fatalf("expected *KubernetesPostgresStackOutputs, got %T", msg)
	}

	if typed.GetNamespace() != "db-prod" {
		t.Errorf("namespace: expected %q, got %q", "db-prod", typed.GetNamespace())
	}
	if typed.GetService() != "postgres-master" {
		t.Errorf("service: expected %q, got %q", "postgres-master", typed.GetService())
	}
	if typed.GetUsernameSecret() == nil {
		t.Fatal("username_secret: expected non-nil")
	}
	if typed.GetUsernameSecret().GetName() != "postgres.db-prod.credentials.postgresql.acid.zalan.do" {
		t.Errorf("username_secret.name: got %q", typed.GetUsernameSecret().GetName())
	}
	if typed.GetUsernameSecret().GetKey() != "username" {
		t.Errorf("username_secret.key: got %q", typed.GetUsernameSecret().GetKey())
	}
	if typed.GetPasswordSecret().GetKey() != "password" {
		t.Errorf("password_secret.key: got %q", typed.GetPasswordSecret().GetKey())
	}
}

func TestTransform_EmptyOutputs(t *testing.T) {
	msg, err := Transform(cloudresourcekind.CloudResourceKind_Auth0ResourceServer, map[string]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	typed, ok := msg.(*auth0v1.Auth0ResourceServerStackOutputs)
	if !ok {
		t.Fatalf("expected *Auth0ResourceServerStackOutputs, got %T", msg)
	}
	if typed.GetId() != "" {
		t.Errorf("expected empty id, got %q", typed.GetId())
	}
}

func TestTransform_UnknownKind(t *testing.T) {
	_, err := Transform(cloudresourcekind.CloudResourceKind_unspecified, map[string]string{"id": "test"})
	if err == nil {
		t.Fatal("expected error for unspecified kind, got nil")
	}
}

// TestTransform_KeyPreprocessing verifies that keys with dots-before-brackets
// and hyphens are normalized before field lookup.
func TestTransform_KeyPreprocessing(t *testing.T) {
	outputs := map[string]string{
		"private_subnets.[0].id": "subnet-001",
		"vpc_id":                 "vpc-abc",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_AwsVpc, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed := msg.(*awsvpcv1.AwsVpcStackOutputs)
	if len(typed.GetPrivateSubnets()) != 1 {
		t.Fatalf("private_subnets: expected 1, got %d", len(typed.GetPrivateSubnets()))
	}
	if typed.GetPrivateSubnets()[0].GetId() != "subnet-001" {
		t.Errorf("private_subnets[0].id: expected %q, got %q",
			"subnet-001", typed.GetPrivateSubnets()[0].GetId())
	}
}
