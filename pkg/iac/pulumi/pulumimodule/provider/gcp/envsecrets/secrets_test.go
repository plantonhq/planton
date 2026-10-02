package envsecrets

// Store run against mocks. What it pins: the exact Pulumi resource every
// Cloud Run kind's stack already holds -- type, logical name, and parent --
// because a renamed resource makes the engine replace the secret, and a
// replacement under the same secret id fails; the grant naming the fallback
// identity the placement asks for; and the value reaching only the version,
// marked secret.

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	typeService = "gcp:projects/service:Service"
	typeSecret  = "gcp:secretmanager/secret:Secret"
	typeVersion = "gcp:secretmanager/secretVersion:SecretVersion"
	typeMember  = "gcp:secretmanager/secretIamMember:SecretIamMember"
)

// registered is one resource Store asked the engine for: "<type> <name>",
// with " <- <parent name>" when it is parented to another resource.
type storeMocks struct {
	mu         sync.Mutex
	registered []string
	inputs     map[string]resource.PropertyMap
}

func (m *storeMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !strings.HasPrefix(args.TypeToken, "pulumi:providers:") {
		entry := args.TypeToken + " " + args.Name
		if parent := args.RegisterRPC.GetParent(); parent != "" && !strings.Contains(parent, "pulumi:pulumi:Stack") {
			entry += " <- " + parent[strings.LastIndex(parent, "::")+2:]
		}
		m.registered = append(m.registered, entry)
		m.inputs[entry] = args.Inputs.Copy()
	}
	outputs := args.Inputs.Copy()
	switch args.TypeToken {
	case typeSecret:
		outputs["project"] = resource.NewStringProperty("acme-prod")
	case typeVersion:
		outputs["version"] = resource.NewStringProperty("1")
	}
	return args.Name + "-id", outputs, nil
}

func (m *storeMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if args.Token == "gcp:organizations/getProject:getProject" {
		return resource.PropertyMap{"number": resource.NewStringProperty("123456789012")}, nil
	}
	return args.Args, nil
}

func runStore(t *testing.T, placement Placement, variables []Variable) *storeMocks {
	t.Helper()
	mocks := &storeMocks{inputs: map[string]resource.PropertyMap{}}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		provider, err := gcp.NewProvider(ctx, "gcp", &gcp.ProviderArgs{})
		if err != nil {
			return err
		}
		_, err = Store(ctx, placement, variables, provider)
		return err
	}, pulumi.WithMocks("project", "stack", mocks))
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	sort.Strings(mocks.registered)
	return mocks
}

func TestStoreRegistersExactlyTheResourcesTheCloudRunStacksHold(t *testing.T) {
	for _, tc := range []struct {
		placement Placement
		variables []Variable
		want      []string
	}{
		{
			placement: Placement{Kind: KindService, Location: "us-central1", Resource: "payments-api", ReplicaRegions: []string{"us-central1"}, RuntimeServiceAccount: "run@acme-prod.iam.gserviceaccount.com"},
			variables: []Variable{{ContainerIndex: 0, Container: "app", Name: "STRIPE_KEY", Value: "a"}, {ContainerIndex: 1, Name: "db.password", Value: "b"}},
			want: []string{
				typeService + " run-secretmanager.googleapis.com",
				typeSecret + " run_us-central1_payments-api_app_STRIPE_KEY",
				typeSecret + " run_us-central1_payments-api_c1_db-password",
				typeMember + " run_us-central1_payments-api_app_STRIPE_KEY <- run_us-central1_payments-api_app_STRIPE_KEY",
				typeMember + " run_us-central1_payments-api_c1_db-password <- run_us-central1_payments-api_c1_db-password",
				typeVersion + " run_us-central1_payments-api_app_STRIPE_KEY <- run_us-central1_payments-api_app_STRIPE_KEY",
				typeVersion + " run_us-central1_payments-api_c1_db-password <- run_us-central1_payments-api_c1_db-password",
			},
		},
		{
			placement: Placement{Kind: KindJob, Location: "europe-west1", Resource: "nightly-etl", ReplicaRegions: []string{"europe-west1"}},
			variables: []Variable{{ContainerIndex: 0, Container: "etl", Name: "TOKEN", Value: "a"}},
			want: []string{
				typeService + " runjob-secretmanager.googleapis.com",
				typeSecret + " runjob_europe-west1_nightly-etl_etl_TOKEN",
				typeMember + " runjob_europe-west1_nightly-etl_etl_TOKEN <- runjob_europe-west1_nightly-etl_etl_TOKEN",
				typeVersion + " runjob_europe-west1_nightly-etl_etl_TOKEN <- runjob_europe-west1_nightly-etl_etl_TOKEN",
			},
		},
		{
			placement: Placement{Kind: KindWorkerPool, Location: "us-east4", Resource: "queue-consumer", ReplicaRegions: []string{"us-east4"}},
			variables: []Variable{{ContainerIndex: 0, Container: "worker", Name: "API_KEY", Value: "a"}},
			want: []string{
				typeService + " runpool-secretmanager.googleapis.com",
				typeSecret + " runpool_us-east4_queue-consumer_worker_API_KEY",
				typeMember + " runpool_us-east4_queue-consumer_worker_API_KEY <- runpool_us-east4_queue-consumer_worker_API_KEY",
				typeVersion + " runpool_us-east4_queue-consumer_worker_API_KEY <- runpool_us-east4_queue-consumer_worker_API_KEY",
			},
		},
	} {
		got := runStore(t, tc.placement, tc.variables).registered
		sort.Strings(tc.want)
		if strings.Join(got, "\n") != strings.Join(tc.want, "\n") {
			t.Errorf("%s registered:\n  %s\nwant:\n  %s", tc.placement.Kind, strings.Join(got, "\n  "), strings.Join(tc.want, "\n  "))
		}
	}
}

func TestStoreGrantsTheFallbackIdentityThePlacementNames(t *testing.T) {
	compute := runStore(t, Placement{Kind: KindFunction, Location: "us-central1", Resource: "hook", NoContainers: true, ReplicaRegions: []string{"us-central1"}},
		[]Variable{{Name: "TOKEN", Value: "a"}})
	member := compute.inputs[typeMember+" function_us-central1_hook_TOKEN <- function_us-central1_hook_TOKEN"]["member"]
	if member.StringValue() != "serviceAccount:123456789012-compute@developer.gserviceaccount.com" {
		t.Errorf("a function without an identity grants %v", member)
	}

	agent := runStore(t, Placement{Kind: KindAgentEngine, Location: "us-central1", Resource: "helper", NoContainers: true, ReplicaRegions: []string{"us-central1"}, DefaultAccount: ReasoningEngineServiceAgent},
		[]Variable{{Name: "TOKEN", Value: "a"}})
	member = agent.inputs[typeMember+" agentengine_us-central1_helper_TOKEN <- agentengine_us-central1_helper_TOKEN"]["member"]
	if member.StringValue() != "serviceAccount:service-123456789012@gcp-sa-aiplatform-re.iam.gserviceaccount.com" {
		t.Errorf("an agent without an identity grants %v", member)
	}
}

func TestStoreKeepsTheValueOnTheVersionAloneMarkedSecret(t *testing.T) {
	const value = "sk_live_do_not_leak"
	mocks := runStore(t, Placement{Kind: KindWorkflow, Location: "us-central1", Resource: "orders", NoContainers: true, ReplicaRegions: []string{"us-central1"}, RuntimeServiceAccount: "wf@acme-prod.iam.gserviceaccount.com"},
		[]Variable{{Name: "API_TOKEN", Value: value}})
	for entry, inputs := range mocks.inputs {
		data, carries := inputs["secretData"]
		if strings.HasPrefix(entry, typeVersion) {
			if !carries || !data.IsSecret() {
				t.Errorf("the version must carry the value marked secret; got %v", data)
			}
			continue
		}
		if strings.Contains(fmt.Sprintf("%v", inputs.Mappable()), value) {
			t.Errorf("%s carries the value", entry)
		}
	}
}
