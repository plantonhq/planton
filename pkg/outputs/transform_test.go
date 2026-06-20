//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	auth0v1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0resourceserver/v1"
	awsvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsvpc/v1"
	gcpdnsv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpdnszone/v1"
	gcpsubnetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpsubnetwork/v1"
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
		"vpc_id":                    "vpc-0abc123",
		"vpc_arn":                   "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123",
		"cidr_block":                "10.0.0.0/16",
		"ipv6_cidr_block":           "2600:1f18:abcd:1200::/56",
		"owner_id":                  "123456789012",
		"main_route_table_id":       "rtb-0abc123",
		"default_security_group_id": "sg-0abc123",
		"default_network_acl_id":    "acl-0abc123",
		"default_route_table_id":    "rtb-0abc123",
		"region":                    "us-west-2",
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
	if typed.GetVpcArn() != "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123" {
		t.Errorf("vpc_arn: got %q", typed.GetVpcArn())
	}
	if typed.GetCidrBlock() != "10.0.0.0/16" {
		t.Errorf("cidr_block: expected %q, got %q", "10.0.0.0/16", typed.GetCidrBlock())
	}
	if typed.GetIpv6CidrBlock() != "2600:1f18:abcd:1200::/56" {
		t.Errorf("ipv6_cidr_block: got %q", typed.GetIpv6CidrBlock())
	}
	if typed.GetOwnerId() != "123456789012" {
		t.Errorf("owner_id: expected %q, got %q", "123456789012", typed.GetOwnerId())
	}
	if typed.GetMainRouteTableId() != "rtb-0abc123" {
		t.Errorf("main_route_table_id: expected %q, got %q", "rtb-0abc123", typed.GetMainRouteTableId())
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
// and hyphens are normalized before field lookup, exercised against a repeated
// nested-message output (GcpSubnetwork.secondary_ranges).
func TestTransform_KeyPreprocessing(t *testing.T) {
	outputs := map[string]string{
		"secondary_ranges.[0].range_name": "pods",
		"subnetwork_name":                 "my-subnet",
	}

	msg, err := Transform(cloudresourcekind.CloudResourceKind_GcpSubnetwork, outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	typed := msg.(*gcpsubnetworkv1.GcpSubnetworkStackOutputs)
	if len(typed.GetSecondaryRanges()) != 1 {
		t.Fatalf("secondary_ranges: expected 1, got %d", len(typed.GetSecondaryRanges()))
	}
	if typed.GetSecondaryRanges()[0].GetRangeName() != "pods" {
		t.Errorf("secondary_ranges[0].range_name: expected %q, got %q",
			"pods", typed.GetSecondaryRanges()[0].GetRangeName())
	}
}
