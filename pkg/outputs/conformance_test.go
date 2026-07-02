//go:build !codegen
// +build !codegen

package outputs

import (
	"path/filepath"
	"testing"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

// TestStackOutputsConformance is the standing guard against the systemic IaC
// output-drift class: an engine emits output names/shapes that do not flatten
// onto the kind's StackOutputs proto, silently leaving those proto fields empty.
// (The original bug: the Postgres tofu module emitted a flat
// "password_secret_name" output, which flattens to the key "password_secret_name"
// -- with no dot -- and therefore never populated the proto's nested
// password_secret{name,key} field, while the Pulumi module emitted the correct
// "password_secret.name". See the planton-postgres-iac-parity work.)
//
// Why this also enforces tofu<->pulumi parity: both engines feed the SAME generic
// transformer (TransformRaw -> Flatten -> populateMessage). So a single
// conformance bar per kind -- "this representative output set fully populates the
// proto with nothing left unmapped" -- when satisfied by each engine's emitted
// output set, guarantees the two engines produce the same typed StackOutputs.
//
// To extend coverage: add a case with the raw output shape an engine emits (scalars
// as strings; nested objects as map[string]interface{}, exactly how Terraform state
// and the Pulumi automation API surface them) and the proto fields it must populate.
func TestStackOutputsConformance(t *testing.T) {
	// A module dir with no transform override forces the generic reflection path,
	// which is the convention every in-repo module relies on (0 of 364 use an override).
	genericModuleDir := filepath.Join("testdata", "modules", "empty")

	cases := []struct {
		name string
		kind cloudresourcekind.CloudResourceKind
		// rawOutputs mirrors the post-Flatten-input shape both engines emit.
		rawOutputs map[string]interface{}
		// mustPopulate lists StackOutputs proto fields that MUST be set.
		mustPopulate []string
	}{
		{
			name: "KubernetesPostgres",
			kind: cloudresourcekind.CloudResourceKind_KubernetesPostgres,
			rawOutputs: map[string]interface{}{
				"namespace":            "gosilver-prod",
				"service":              "gosilver-prod-postgres-master",
				"port_forward_command": "kubectl port-forward -n gosilver-prod service/gosilver-prod-postgres-master 8080:8080",
				"kube_endpoint":        "gosilver-prod-postgres-master.gosilver-prod.svc.cluster.local",
				"external_hostname":    "gosilver-prod-postgres.planton.live",
				// Nested objects -- the shape that flattens to password_secret.name etc.
				"password_secret": map[string]interface{}{
					"name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
					"key":  "password",
				},
				"username_secret": map[string]interface{}{
					"name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
					"key":  "username",
				},
			},
			mustPopulate: []string{
				"namespace", "service", "port_forward_command", "kube_endpoint",
				"external_hostname", "password_secret", "username_secret",
			},
		},
		{
			// AwsSubnet: flat scalar outputs from both engines (subnet id/arn, AZ,
			// CIDR, route table id, region) must each land on the StackOutputs proto.
			name: "AwsSubnet",
			kind: cloudresourcekind.CloudResourceKind_AwsSubnet,
			rawOutputs: map[string]interface{}{
				"subnet_id":         "subnet-0abc123",
				"subnet_arn":        "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-0abc123",
				"availability_zone": "us-west-2a",
				"cidr_block":        "10.0.1.0/24",
				"route_table_id":    "rtb-0abc123",
				"region":            "us-west-2",
			},
			mustPopulate: []string{
				"subnet_id", "subnet_arn", "availability_zone",
				"cidr_block", "route_table_id", "region",
			},
		},
		{
			// AwsInternetGateway: flat scalar outputs from both engines (gateway
			// id/arn, attached vpc id, region) must each land on the StackOutputs proto.
			name: "AwsInternetGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsInternetGateway,
			rawOutputs: map[string]interface{}{
				"internet_gateway_id":  "igw-0abc123",
				"internet_gateway_arn": "arn:aws:ec2:us-west-2:123456789012:internet-gateway/igw-0abc123",
				"vpc_id":               "vpc-0abc123",
				"region":               "us-west-2",
			},
			mustPopulate: []string{
				"internet_gateway_id", "internet_gateway_arn", "vpc_id", "region",
			},
		},
		{
			// AwsEgressOnlyInternetGateway: flat scalar outputs from both engines
			// (gateway id, attached vpc id, region) must each land on the StackOutputs
			// proto. An egress-only gateway has no ARN, so none is emitted.
			name: "AwsEgressOnlyInternetGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsEgressOnlyInternetGateway,
			rawOutputs: map[string]interface{}{
				"egress_only_internet_gateway_id": "eigw-0abc123",
				"vpc_id":                          "vpc-0abc123",
				"region":                          "us-west-2",
			},
			mustPopulate: []string{
				"egress_only_internet_gateway_id", "vpc_id", "region",
			},
		},
		{
			// AwsNatGateway: flat scalar outputs from both engines (gateway id,
			// public/private ip, ENI id, subnet id, region) must each land on the
			// StackOutputs proto. A NAT gateway has no ARN, so none is emitted.
			name: "AwsNatGateway",
			kind: cloudresourcekind.CloudResourceKind_AwsNatGateway,
			rawOutputs: map[string]interface{}{
				"nat_gateway_id":       "nat-0abc123",
				"public_ip":            "52.10.20.30",
				"private_ip":           "10.0.0.10",
				"network_interface_id": "eni-0abc123",
				"subnet_id":            "subnet-0abc123",
				"region":               "us-west-2",
			},
			mustPopulate: []string{
				"nat_gateway_id", "public_ip", "private_ip",
				"network_interface_id", "subnet_id", "region",
			},
		},
		{
			// AwsVpc: flat scalar outputs from both engines (vpc id/arn, primary and
			// IPv6 CIDR, owner, the route-table/default-resource ids, region) must
			// each land on the thin StackOutputs proto.
			name: "AwsVpc",
			kind: cloudresourcekind.CloudResourceKind_AwsVpc,
			rawOutputs: map[string]interface{}{
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
			},
			mustPopulate: []string{
				"vpc_id", "vpc_arn", "cidr_block", "ipv6_cidr_block", "owner_id",
				"main_route_table_id", "default_security_group_id",
				"default_network_acl_id", "default_route_table_id", "region",
			},
		},
		{
			// AwsIamPolicy: flat scalar outputs from both engines (policy arn/id/name)
			// must each land on the StackOutputs proto -- policy_arn is what role/user
			// attachments and permissions boundaries reference.
			name: "AwsIamPolicy",
			kind: cloudresourcekind.CloudResourceKind_AwsIamPolicy,
			rawOutputs: map[string]interface{}{
				"policy_arn":  "arn:aws:iam::123456789012:policy/s3-read-only",
				"policy_id":   "ANPAEXAMPLEID12345678",
				"policy_name": "s3-read-only",
			},
			mustPopulate: []string{"policy_arn", "policy_id", "policy_name"},
		},
		{
			// AwsIamInstanceProfile: flat scalar outputs from both engines (profile
			// arn/name/id and the carried role's name) must each land on the
			// StackOutputs proto -- instance_profile_arn is what EC2-shaped resources
			// reference.
			name: "AwsIamInstanceProfile",
			kind: cloudresourcekind.CloudResourceKind_AwsIamInstanceProfile,
			rawOutputs: map[string]interface{}{
				"instance_profile_arn":  "arn:aws:iam::123456789012:instance-profile/web-server",
				"instance_profile_name": "web-server",
				"instance_profile_id":   "AIPAEXAMPLEID12345678",
				"role_name":             "web-server-role",
			},
			mustPopulate: []string{
				"instance_profile_arn", "instance_profile_name",
				"instance_profile_id", "role_name",
			},
		},
		{
			// AwsIamRole: flat scalar outputs from both engines (role arn/name/id)
			// must each land on the StackOutputs proto. Guards the removal of the
			// role's former instance-profile outputs: EC2 delivery now composes
			// through AwsIamInstanceProfile, so the role emits only role-shaped
			// outputs.
			name: "AwsIamRole",
			kind: cloudresourcekind.CloudResourceKind_AwsIamRole,
			rawOutputs: map[string]interface{}{
				"role_arn":  "arn:aws:iam::123456789012:role/lambda-exec",
				"role_name": "lambda-exec",
				"role_id":   "AROAEXAMPLEID12345678",
			},
			mustPopulate: []string{"role_arn", "role_name", "role_id"},
		},
		{
			// AwsIamUser: flat scalar outputs from both engines (user arn/name/id,
			// access key id + base64 secret, console url) must each land on the
			// StackOutputs proto. The secret is base64-encoded by BOTH engines so the
			// emitted values are byte-identical.
			name: "AwsIamUser",
			kind: cloudresourcekind.CloudResourceKind_AwsIamUser,
			rawOutputs: map[string]interface{}{
				"user_arn":          "arn:aws:iam::123456789012:user/ci-deploy",
				"user_name":         "ci-deploy",
				"user_id":           "AIDAEXAMPLEID12345678",
				"access_key_id":     "AKIAEXAMPLEID1234567",
				"secret_access_key": "c2VjcmV0LWtleS1tYXRlcmlhbA==",
				"console_url":       "https://signin.aws.amazon.com/console",
			},
			mustPopulate: []string{
				"user_arn", "user_name", "user_id",
				"access_key_id", "secret_access_key", "console_url",
			},
		},
		{
			// Guards the externaldns tofu module's output rename to solver_sa: the
			// module previously emitted "service_account_name", which does not flatten
			// onto the KubernetesExternalDnsStackOutputs.solver_sa proto field (the
			// Pulumi module already exported "solver_sa"). Both engines now emit the
			// same three outputs.
			name: "KubernetesExternalDns",
			kind: cloudresourcekind.CloudResourceKind_KubernetesExternalDns,
			rawOutputs: map[string]interface{}{
				"namespace":    "external-dns",
				"release_name": "gosilver-in-external-dns",
				"solver_sa":    "gosilver-in-external-dns",
			},
			mustPopulate: []string{"namespace", "release_name", "solver_sa"},
		},
		{
			// CloudflareR2Bucket: both engines emit the same outputs -- bucket name,
			// path-style S3 URL, the list of custom-domain URLs, and the managed
			// r2.dev public URL -- each of which must land on the StackOutputs proto.
			name: "CloudflareR2Bucket",
			kind: cloudresourcekind.CloudResourceKind_CloudflareR2Bucket,
			rawOutputs: map[string]interface{}{
				"bucket_name":        "media-assets",
				"bucket_url":         "https://00000000000000000000000000000000.r2.cloudflarestorage.com/media-assets",
				"custom_domain_urls": []interface{}{"https://media.example.com", "https://cdn.example.com"},
				"public_url":         "https://pub-0123456789abcdef.r2.dev",
			},
			mustPopulate: []string{"bucket_name", "bucket_url", "custom_domain_urls", "public_url"},
		},
		{
			// CloudflareD1Database: both engines emit the database id and name as
			// flat scalars (a Worker reaches D1 through its binding; no DSN exists).
			name: "CloudflareD1Database",
			kind: cloudresourcekind.CloudResourceKind_CloudflareD1Database,
			rawOutputs: map[string]interface{}{
				"database_id":   "9a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d",
				"database_name": "app-prod-db",
			},
			mustPopulate: []string{"database_id", "database_name"},
		},
		{
			// CloudflareKvNamespace: both engines emit the namespace id and the
			// url-encoding support flag.
			name: "CloudflareKvNamespace",
			kind: cloudresourcekind.CloudResourceKind_CloudflareKvNamespace,
			rawOutputs: map[string]interface{}{
				"namespace_id":          "0f1e2d3c4b5a69788796a5b4c3d2e1f0",
				"supports_url_encoding": true,
			},
			mustPopulate: []string{"namespace_id", "supports_url_encoding"},
		},
		{
			// CloudflareWorkersKvPair: both engines emit the entry key and the
			// namespace it was written to.
			name: "CloudflareWorkersKvPair",
			kind: cloudresourcekind.CloudResourceKind_CloudflareWorkersKvPair,
			rawOutputs: map[string]interface{}{
				"key_name":     "feature.new-dashboard",
				"namespace_id": "0f1e2d3c4b5a69788796a5b4c3d2e1f0",
			},
			mustPopulate: []string{"key_name", "namespace_id"},
		},
		{
			// CloudflareHyperdriveConfig: both engines emit the config id and name.
			name: "CloudflareHyperdriveConfig",
			kind: cloudresourcekind.CloudResourceKind_CloudflareHyperdriveConfig,
			rawOutputs: map[string]interface{}{
				"hyperdrive_id": "a1b2c3d4e5f60718293a4b5c6d7e8f90",
				"name":          "app-prod-pg",
			},
			mustPopulate: []string{"hyperdrive_id", "name"},
		},
		{
			// CloudflareQueue: both engines emit the queue id and name (referenced by
			// consumers, worker producer bindings, and R2 event notifications).
			name: "CloudflareQueue",
			kind: cloudresourcekind.CloudResourceKind_CloudflareQueue,
			rawOutputs: map[string]interface{}{
				"queue_id":    "a1b2c3d4e5f60718293a4b5c6d7e8f90",
				"queue_name":  "orders-queue",
				"created_on":  "2026-06-25T00:00:00Z",
				"modified_on": "2026-06-25T00:00:00Z",
			},
			mustPopulate: []string{"queue_id", "queue_name"},
		},
		{
			// CloudflarePagesProject: both engines emit the project name, its
			// pages.dev subdomain, attached custom domains, and creation time.
			name: "CloudflarePagesProject",
			kind: cloudresourcekind.CloudResourceKind_CloudflarePagesProject,
			rawOutputs: map[string]interface{}{
				"project_name": "marketing-site",
				"subdomain":    "marketing-site.pages.dev",
				"domains":      []interface{}{"www.example.com"},
				"created_on":   "2026-06-25T00:00:00Z",
			},
			mustPopulate: []string{"project_name", "subdomain"},
		},
		{
			// CloudflareDnsRecord: both engines emit the record id, name, type and
			// proxied flag as flat scalars onto the StackOutputs proto.
			name: "CloudflareDnsRecord",
			kind: cloudresourcekind.CloudResourceKind_CloudflareDnsRecord,
			rawOutputs: map[string]interface{}{
				"record_id":   "372e67954025e0ba6aaa6d586b9e0b59",
				"record_name": "www",
				"record_type": "A",
				"proxied":     true,
			},
			mustPopulate: []string{"record_id", "record_name", "record_type", "proxied"},
		},
		{
			// CloudflareDnsZone: both engines emit the zone id (scalar) and the
			// assigned nameservers (repeated string) onto the StackOutputs proto.
			name: "CloudflareDnsZone",
			kind: cloudresourcekind.CloudResourceKind_CloudflareDnsZone,
			rawOutputs: map[string]interface{}{
				"zone_id":                 "023e105f4ecef8ad9ca31a8372d0c353",
				"nameservers":             []interface{}{"ns1.cloudflare.com", "ns2.cloudflare.com"},
				"status":                  "active",
				"dnssec_status":           "active",
				"dnssec_ds":               "example.com. 3600 IN DS 2371 13 2 ABCDEF",
				"dnssec_digest":           "abcdef0123456789",
				"dnssec_digest_type":      "2",
				"dnssec_digest_algorithm": "SHA256",
				"dnssec_algorithm":        "13",
				"dnssec_key_tag":          "2371",
				"dnssec_public_key":       "mdsswUyr3DPW132mOi8V9xESWE8jTo0d",
				"dnssec_flags":            "257",
			},
			mustPopulate: []string{"zone_id", "nameservers", "status", "dnssec_ds", "dnssec_key_tag"},
		},
		{
			// CloudflareRuleset: both engines emit ruleset id, version, and the
			// zone_id/phase pass-throughs as flat scalars onto the proto.
			name: "CloudflareRuleset",
			kind: cloudresourcekind.CloudResourceKind_CloudflareRuleset,
			rawOutputs: map[string]interface{}{
				"ruleset_id": "2f2feab2026849078ba485f918791bdc",
				"version":    "3",
				"zone_id":    "023e105f4ecef8ad9ca31a8372d0c353",
				"phase":      "http_request_origin",
			},
			mustPopulate: []string{"ruleset_id", "version", "zone_id", "phase"},
		},
		{
			// CloudflareLoadBalancer: both engines emit the load balancer id,
			// hostname, and cname target as flat scalars onto the proto.
			name: "CloudflareLoadBalancer",
			kind: cloudresourcekind.CloudResourceKind_CloudflareLoadBalancer,
			rawOutputs: map[string]interface{}{
				"load_balancer_id":              "699d98642c564d2e855e9661899b7252",
				"load_balancer_dns_record_name": "lb.example.com",
				"load_balancer_cname_target":    "699d98642c564d2e855e9661899b7252",
			},
			mustPopulate: []string{"load_balancer_id", "load_balancer_dns_record_name", "load_balancer_cname_target"},
		},
		{
			// CloudflareLoadBalancerPool: both engines emit the pool id and name
			// (account-scoped pool referenced by load balancers).
			name: "CloudflareLoadBalancerPool",
			kind: cloudresourcekind.CloudResourceKind_CloudflareLoadBalancerPool,
			rawOutputs: map[string]interface{}{
				"pool_id":   "17b5962d775c646f3f9725cbc7a53df4",
				"pool_name": "web-pool",
			},
			mustPopulate: []string{"pool_id", "pool_name"},
		},
		{
			// CloudflareLoadBalancerMonitor: both engines emit the monitor id and
			// its protocol (account-scoped health check referenced by pools).
			name: "CloudflareLoadBalancerMonitor",
			kind: cloudresourcekind.CloudResourceKind_CloudflareLoadBalancerMonitor,
			rawOutputs: map[string]interface{}{
				"monitor_id":   "f1aba936b94213e5b8dca0c0dbf1f9cc",
				"monitor_type": "https",
			},
			mustPopulate: []string{"monitor_id", "monitor_type"},
		},
		{
			// CloudflareWorker: both engines emit the script id and name (scalars)
			// and the custom-domain hostnames / route patterns (repeated strings).
			name: "CloudflareWorker",
			kind: cloudresourcekind.CloudResourceKind_CloudflareWorker,
			rawOutputs: map[string]interface{}{
				"script_id":               "my-worker",
				"script_name":             "my-worker",
				"custom_domain_hostnames": []interface{}{"api.example.com"},
				"route_patterns":          []interface{}{"api.example.com/*"},
			},
			mustPopulate: []string{"script_id", "script_name", "custom_domain_hostnames", "route_patterns"},
		},
		{
			// CloudflareZeroTrustAccessApplication: both engines emit the
			// application id, audience tag, protected domain, and SaaS material.
			name: "CloudflareZeroTrustAccessApplication",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustAccessApplication,
			rawOutputs: map[string]interface{}{
				"application_id":     "f174e90a-fafe-4643-bbbc-4a0ed4fc8415",
				"aud":                "8a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b",
				"domain":             "dashboard.example.com",
				"saas_client_id":     "client-abc",
				"saas_client_secret": "secret-xyz",
				"saas_public_key":    "MIIBIjANBgkqh...",
				"saas_sso_endpoint":  "https://example.cloudflareaccess.com/cdn-cgi/access/sso/saml/abc",
				"saas_idp_entity_id": "https://example.cloudflareaccess.com",
			},
			mustPopulate: []string{
				"application_id", "aud", "domain", "saas_client_id", "saas_client_secret",
				"saas_public_key", "saas_sso_endpoint", "saas_idp_entity_id",
			},
		},
		{
			// CloudflareZeroTrustAccessPolicy: both engines emit the policy id.
			name: "CloudflareZeroTrustAccessPolicy",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustAccessPolicy,
			rawOutputs: map[string]interface{}{
				"policy_id": "699d98642c564d2e855e9661899b7252",
			},
			mustPopulate: []string{"policy_id"},
		},
		{
			// CloudflareZeroTrustAccessGroup: both engines emit the group id.
			name: "CloudflareZeroTrustAccessGroup",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustAccessGroup,
			rawOutputs: map[string]interface{}{
				"group_id": "aa9d98642c564d2e855e9661899b7252",
			},
			mustPopulate: []string{"group_id"},
		},
		{
			// CloudflareZeroTrustTunnel: both engines emit flat scalar outputs --
			// tunnel id, CNAME target, the (sensitive) connector token, status, the
			// account tag, and the creation timestamp.
			name: "CloudflareZeroTrustTunnel",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustTunnel,
			rawOutputs: map[string]interface{}{
				"tunnel_id":     "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
				"tunnel_cname":  "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415.cfargotunnel.com",
				"tunnel_token":  "eyJhIjoiMDc0NzU1YTc4ZDhlIn0=",
				"tunnel_status": "healthy",
				"account_tag":   "074755a78d8e8f77c119a90a125e8a06",
				"created_on":    "2026-06-25T12:00:00Z",
			},
			mustPopulate: []string{
				"tunnel_id", "tunnel_cname", "tunnel_token",
				"tunnel_status", "account_tag", "created_on",
			},
		},
		{
			// CloudflareZeroTrustTunnelVirtualNetwork: both engines emit the virtual
			// network id and name.
			name: "CloudflareZeroTrustTunnelVirtualNetwork",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustTunnelVirtualNetwork,
			rawOutputs: map[string]interface{}{
				"virtual_network_id":   "aaaa1111-bbbb-2222-cccc-333344445555",
				"virtual_network_name": "prod-vnet",
			},
			mustPopulate: []string{"virtual_network_id", "virtual_network_name"},
		},
		{
			// CloudflareZeroTrustTunnelRoute: both engines emit the route id and the
			// advertised CIDR.
			name: "CloudflareZeroTrustTunnelRoute",
			kind: cloudresourcekind.CloudResourceKind_CloudflareZeroTrustTunnelRoute,
			rawOutputs: map[string]interface{}{
				"route_id": "b8f2e1c0-1111-2222-3333-444455556666",
				"network":  "10.0.0.0/24",
			},
			mustPopulate: []string{"route_id", "network"},
		},
		{
			// CloudflareList: both engines emit the list id, name, and kind.
			name: "CloudflareList",
			kind: cloudresourcekind.CloudResourceKind_CloudflareList,
			rawOutputs: map[string]interface{}{
				"list_id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
				"name":    "office_allowlist",
				"kind":    "ip",
			},
			mustPopulate: []string{"list_id", "name", "kind"},
		},
		{
			// CloudflareListItem: both engines emit the item id and parent list id.
			name: "CloudflareListItem",
			kind: cloudresourcekind.CloudResourceKind_CloudflareListItem,
			rawOutputs: map[string]interface{}{
				"item_id": "70c4e0c9b0e34f1a9b6f2d3c4a5b6c7d",
				"list_id": "2c0fc9fa937b11eaa1b71c4d701ab86e",
			},
			mustPopulate: []string{"item_id", "list_id"},
		},
		{
			// CloudflareTurnstileWidget: both engines emit the site key, the
			// (sensitive) secret, and timestamps.
			name: "CloudflareTurnstileWidget",
			kind: cloudresourcekind.CloudResourceKind_CloudflareTurnstileWidget,
			rawOutputs: map[string]interface{}{
				"sitekey":     "0x4AAAAAAA_examplesitekey",
				"secret":      "0x4AAAAAAA_examplesecretkey",
				"created_on":  "2026-06-25T00:00:00Z",
				"modified_on": "2026-06-25T00:00:00Z",
			},
			mustPopulate: []string{"sitekey", "secret"},
		},
		{
			// CloudflareEmailRoutingZone: both engines emit the zone id, enabled
			// flag, status, and name.
			name: "CloudflareEmailRoutingZone",
			kind: cloudresourcekind.CloudResourceKind_CloudflareEmailRoutingZone,
			rawOutputs: map[string]interface{}{
				"zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
				"enabled": "true",
				"status":  "ready",
				"name":    "example.com",
			},
			mustPopulate: []string{"zone_id", "status", "name"},
		},
		{
			// CloudflareEmailRoutingRule: both engines emit the rule id and zone id.
			name: "CloudflareEmailRoutingRule",
			kind: cloudresourcekind.CloudResourceKind_CloudflareEmailRoutingRule,
			rawOutputs: map[string]interface{}{
				"rule_id": "a1b2c3d4e5f60718293a4b5c6d7e8f90",
				"zone_id": "023e105f4ecef8ad9ca31a8372d0c353",
			},
			mustPopulate: []string{"rule_id", "zone_id"},
		},
		{
			// CloudflareEmailRoutingAddress: both engines emit the address id,
			// email, and timestamps.
			name: "CloudflareEmailRoutingAddress",
			kind: cloudresourcekind.CloudResourceKind_CloudflareEmailRoutingAddress,
			rawOutputs: map[string]interface{}{
				"address_id": "b8f2e1c0a1b2c3d4e5f60718293a4b5c",
				"email":      "ops@example.com",
				"created":    "2026-06-25T00:00:00Z",
			},
			mustPopulate: []string{"address_id", "email"},
		},
		{
			// CloudflareOriginCaCertificate: both engines emit the certificate id,
			// the certificate PEM, the (sensitive) generated private key, and expiry.
			name: "CloudflareOriginCaCertificate",
			kind: cloudresourcekind.CloudResourceKind_CloudflareOriginCaCertificate,
			rawOutputs: map[string]interface{}{
				"certificate_id": "b8f2e1c0a1b2c3d4e5f60718293a4b5c",
				"certificate":    "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
				"private_key":    "-----BEGIN PRIVATE KEY-----\nMIIE\n-----END PRIVATE KEY-----",
				"expires_on":     "2041-06-25T00:00:00Z",
			},
			mustPopulate: []string{"certificate_id", "certificate"},
		},
		{
			// CloudflareCertificatePack: both engines emit the pack id, status, and
			// primary certificate id.
			name: "CloudflareCertificatePack",
			kind: cloudresourcekind.CloudResourceKind_CloudflareCertificatePack,
			rawOutputs: map[string]interface{}{
				"certificate_pack_id": "3822ff90e3534420ac41fc7e4a1f4b07",
				"status":              "active",
				"primary_certificate": "caa875a3-b2f0-4f7e-9a1e-0d2b4c6e8f10",
			},
			mustPopulate: []string{"certificate_pack_id", "status"},
		},
		{
			// CloudflareCustomHostname: both engines emit the hostname id, status,
			// the ownership-verification records, and the creation timestamp.
			name: "CloudflareCustomHostname",
			kind: cloudresourcekind.CloudResourceKind_CloudflareCustomHostname,
			rawOutputs: map[string]interface{}{
				"custom_hostname_id":               "0d89c70f8d4f4b1aa1b5d2e3f4a5b6c7",
				"status":                           "pending",
				"ownership_verification_name":      "_cf-custom-hostname.support.acme.com",
				"ownership_verification_type":      "txt",
				"ownership_verification_value":     "1f2e3d4c5b6a7988",
				"ownership_verification_http_url":  "http://support.acme.com/.well-known/cf-custom-hostname-challenge/0d89",
				"ownership_verification_http_body": "1f2e3d4c5b6a7988",
				"verification_errors":              []interface{}{},
				"created_at":                       "2026-06-25T00:00:00Z",
			},
			mustPopulate: []string{"custom_hostname_id", "status"},
		},
		{
			// CloudflareCustomHostnameFallbackOrigin: both engines emit status and
			// timestamps for the zone's fallback origin.
			name: "CloudflareCustomHostnameFallbackOrigin",
			kind: cloudresourcekind.CloudResourceKind_CloudflareCustomHostnameFallbackOrigin,
			rawOutputs: map[string]interface{}{
				"status":     "active",
				"created_at": "2026-06-25T00:00:00Z",
				"updated_at": "2026-06-25T00:00:00Z",
				"errors":     []interface{}{},
			},
			mustPopulate: []string{"status"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateOverride(tc.kind, genericModuleDir, tc.rawOutputs)
			if err != nil {
				t.Fatalf("ValidateOverride failed: %v", err)
			}
			if len(result.SchemaErrors) != 0 {
				t.Fatalf("unexpected schema errors: %v", result.SchemaErrors)
			}
			if result.DryRun == nil {
				t.Fatal("expected a dry-run result")
			}

			// Core invariant: every emitted output lands on a proto field. A
			// regression to a flat/mismatched output name surfaces here.
			if len(result.DryRun.UnmappedOutputs) != 0 {
				t.Errorf("%s: outputs did not map onto the StackOutputs proto: %v",
					tc.kind.String(), result.DryRun.UnmappedOutputs)
			}

			populated := make(map[string]bool, len(result.DryRun.PopulatedFields))
			for _, f := range result.DryRun.PopulatedFields {
				populated[f.ProtoField] = true
			}
			for _, field := range tc.mustPopulate {
				if !populated[field] {
					t.Errorf("%s: expected proto field %q to be populated, but it was not",
						tc.kind.String(), field)
				}
			}
		})
	}
}

// TestStackOutputsConformance_DetectsFlatSecretDrift proves the guard actually
// catches the historical drift: the pre-fix Postgres tofu module emitted flat
// "password_secret_name"/"password_secret_key" outputs, which do NOT flatten onto
// the proto's password_secret{name,key} field. The guard must flag both the
// unmapped output and the unpopulated proto field.
func TestStackOutputsConformance_DetectsFlatSecretDrift(t *testing.T) {
	genericModuleDir := filepath.Join("testdata", "modules", "empty")
	kind := cloudresourcekind.CloudResourceKind_KubernetesPostgres

	flatDriftOutputs := map[string]interface{}{
		"namespace":            "gosilver-prod",
		"password_secret_name": "postgres.db-gosilver-prod-postgres.credentials.postgresql.acid.zalan.do",
		"password_secret_key":  "password",
	}

	result, err := ValidateOverride(kind, genericModuleDir, flatDriftOutputs)
	if err != nil {
		t.Fatalf("ValidateOverride failed: %v", err)
	}
	if result.DryRun == nil {
		t.Fatal("expected a dry-run result")
	}

	if len(result.DryRun.UnmappedOutputs) == 0 {
		t.Error("expected the flat password_secret_name/_key outputs to be reported as unmapped, but none were")
	}
	for _, f := range result.DryRun.PopulatedFields {
		if f.ProtoField == "password_secret" {
			t.Error("flat outputs must NOT populate the nested password_secret proto field")
		}
	}
}
