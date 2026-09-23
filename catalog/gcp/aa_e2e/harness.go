// Package aa_e2e implements the E2E provider harness for GCP. Like AWS (a real
// cloud account), Setup validates that the ambient Application Default
// Credentials chain can reach the test project, and resource verification runs
// through the Google Cloud REST APIs.
//
// Credentials are intentionally NOT plumbed through the stack input. The E2E
// framework builds every stack input with a nil provider config, so the IaC
// modules resolve credentials from the ambient ADC chain (locally:
// `gcloud auth application-default login`; in CI: workload identity
// federation). No static secret is ever stored on disk or in CI.
//
// The test project is resolved once at Setup (E2E_GCP_PROJECT, then
// GOOGLE_PROJECT, then the ADC credential's project) and exported as
// GOOGLE_PROJECT so both engines' subprocesses — and therefore both providers'
// default-project resolution — agree with the harness. Scenario manifests omit
// spec.project_id and ride this ambient project, so no manifest in the repo
// ever hardcodes a project id.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/gcp/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/alloydb/v1"
	apikeys "google.golang.org/api/apikeys/v2"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/bigquery/v2"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
	billingbudgets "google.golang.org/api/billingbudgets/v1"
	certificatemanager "google.golang.org/api/certificatemanager/v1"
	cloudfunctions "google.golang.org/api/cloudfunctions/v2"
	cloudidentity "google.golang.org/api/cloudidentity/v1"
	cloudkms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/cloudresourcemanager/v1"
	crmv3 "google.golang.org/api/cloudresourcemanager/v3"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
	composer "google.golang.org/api/composer/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	dataproc "google.golang.org/api/dataproc/v1"
	"google.golang.org/api/dns/v1"
	eventarc "google.golang.org/api/eventarc/v1"
	firebase "google.golang.org/api/firebase/v1beta1"
	firestore "google.golang.org/api/firestore/v1"
	"google.golang.org/api/iam/v1"
	iamv2 "google.golang.org/api/iam/v2"
	identitytoolkit "google.golang.org/api/identitytoolkit/v2"
	logging "google.golang.org/api/logging/v2"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	monitoring "google.golang.org/api/monitoring/v3"
	"google.golang.org/api/networkconnectivity/v1"
	"google.golang.org/api/option"
	orgpolicy "google.golang.org/api/orgpolicy/v2"
	pubsub "google.golang.org/api/pubsub/v1"
	"google.golang.org/api/redis/v1"
	run "google.golang.org/api/run/v2"
	secretmanager "google.golang.org/api/secretmanager/v1"
	"google.golang.org/api/spanner/v1"
	"google.golang.org/api/sqladmin/v1"
	"google.golang.org/api/storage/v1"
	htransport "google.golang.org/api/transport/http"
	"google.golang.org/api/vpcaccess/v1"
	workflows "google.golang.org/api/workflows/v1"
	"sigs.k8s.io/yaml"
)

// Harness manages the GCP E2E test lifecycle.
type Harness struct {
	services *verify.Services

	// mu guards deployed, written by VerifyDeployed and read by VerifyDestroyed.
	mu       sync.Mutex
	deployed map[string]map[string]string
}

// NewHarness creates a GCP test harness. Credentials come from the ambient ADC
// chain (see the package doc); none are passed here.
func NewHarness() *Harness {
	return &Harness{deployed: make(map[string]map[string]string)}
}

// Setup resolves the test project, exports GOOGLE_PROJECT for the IaC
// subprocesses, loads ADC, and confirms the project is reachable via a
// side-effect-free cloudresourcemanager projects.get call.
func (h *Harness) Setup(ctx context.Context) error {
	creds, err := google.FindDefaultCredentials(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return errors.Wrap(err, "failed to load GCP Application Default Credentials "+
			"(locally: `gcloud auth application-default login`; in CI: workload identity federation)")
	}

	project := firstNonEmpty(os.Getenv("E2E_GCP_PROJECT"), os.Getenv("GOOGLE_PROJECT"), creds.ProjectID)
	if project == "" {
		return errors.New("no GCP test project resolved: set E2E_GCP_PROJECT (or GOOGLE_PROJECT), " +
			"or use ADC credentials that carry a project")
	}

	// Both the Terraform google provider and the Pulumi gcp provider resolve
	// their default project from GOOGLE_PROJECT, and both E2E spawners rebuild
	// the subprocess environment from this process at call time — so this single
	// export is what lets scenario manifests omit spec.project_id.
	if err := os.Setenv("GOOGLE_PROJECT", project); err != nil {
		return errors.Wrap(err, "failed to export GOOGLE_PROJECT")
	}
	// A few kinds must NAME the project in the spec rather than inherit it
	// from the provider -- a tag key's owner is "exactly one of organization
	// or project", so a project-owned key cannot leave the arm empty. Those
	// fixtures reference the resolved project through the
	// ${E2E_ENV:PLANTON_E2E_GCP_PROJECT_ID} token (the env-token prefix the
	// scenario loader admits), the same mechanism as the GCS agent below.
	if err := os.Setenv("PLANTON_E2E_GCP_PROJECT_ID", project); err != nil {
		return errors.Wrap(err, "failed to export PLANTON_E2E_GCP_PROJECT_ID")
	}

	// Every verifier probe names the test project as its quota project --
	// the same posture the IaC modules take with user_project_override on
	// every provider call. Some Google APIs (API Keys, Identity Toolkit, App
	// Check, the Firebase Management API on some methods) refuse a
	// user-credential call that carries no quota project with 403 "requires
	// a quota project, which is not set by default". The Go client libraries
	// fall back to the ADC file's own quota_project_id, but a fresh
	// `gcloud auth application-default login` leaves that field EMPTY when
	// the account cannot bill quota to gcloud's configured project, so a
	// harness that relied on it worked or failed by accident of the
	// developer's machine. Naming the project here makes the attribution
	// explicit under every credential mode, including workload identity in
	// CI, and it is what the modules already do for the resources they
	// create. Every API a verifier probes is enabled on the test project by
	// the module under test (or is a project-level API every project has),
	// so the header can never fail a probe that would otherwise pass.
	clientOpts := []option.ClientOption{option.WithQuotaProject(project)}

	crmService, err := cloudresourcemanager.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudresourcemanager client")
	}
	iamService, err := iam.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create iam client")
	}
	computeService, err := compute.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create compute client")
	}
	storageService, err := storage.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create storage client")
	}

	// The project's GCS service agent (service-{project_number}@
	// gs-project-accounts.iam.gserviceaccount.com) is the identity GCS
	// publishes bucket notifications as, and it needs
	// roles/pubsub.publisher granted BEFORE a notification config can be
	// created. The email is project-specific, so committed fixtures
	// reference it through the ${E2E_ENV:PLANTON_E2E_GCS_AGENT_EMAIL}
	// token; this lookup (side-effect-free, and it lazily provisions the
	// agent — GCP creates it on first read) is what makes those fixtures
	// deployable against any test project.
	gcsAgent, err := storageService.Projects.ServiceAccount.Get(project).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "failed to resolve the GCS service agent for project %s", project)
	}
	if err := os.Setenv("PLANTON_E2E_GCS_AGENT_EMAIL", gcsAgent.EmailAddress); err != nil {
		return errors.Wrap(err, "failed to export PLANTON_E2E_GCS_AGENT_EMAIL")
	}
	sqlAdminService, err := sqladmin.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create sqladmin client")
	}
	redisService, err := redis.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create redis client")
	}
	containerService, err := container.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create container client")
	}
	runService, err := run.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create run client")
	}
	alloyDBService, err := alloydb.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create alloydb client")
	}
	dnsService, err := dns.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create dns client")
	}
	spannerService, err := spanner.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create spanner client")
	}
	bigQueryService, err := bigquery.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create bigquery client")
	}
	vpcAccessService, err := vpcaccess.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create vpcaccess client")
	}
	cloudFunctionsService, err := cloudfunctions.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudfunctions client")
	}
	networkConnectivityService, err := networkconnectivity.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create networkconnectivity client")
	}
	bigtableAdminService, err := bigtableadmin.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create bigtableadmin client")
	}
	firestoreService, err := firestore.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create firestore admin client")
	}
	dataprocService, err := dataproc.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create dataproc client")
	}
	composerService, err := composer.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create composer client")
	}
	pubsubService, err := pubsub.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create pubsub client")
	}
	cloudKmsService, err := cloudkms.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudkms client")
	}
	cloudTasksService, err := cloudtasks.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudtasks client")
	}
	cloudSchedulerService, err := cloudscheduler.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudscheduler client")
	}
	artifactRegistryService, err := artifactregistry.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create artifactregistry client")
	}
	certificateManagerService, err := certificatemanager.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create certificatemanager client")
	}
	monitoringService, err := monitoring.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create monitoring client")
	}
	// Dashboards are served by the Monitoring API's v1 surface — a
	// DIFFERENT API version from the v3 client above, with its own typed
	// client on the same pinned google.golang.org/api line.
	monitoringDashboardsService, err := monitoringv1.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create monitoring dashboards (v1) client")
	}
	secretManagerService, err := secretmanager.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create secretmanager client")
	}
	loggingService, err := logging.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create logging client")
	}
	identityToolkitService, err := identitytoolkit.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create identitytoolkit client")
	}
	iamV2Service, err := iamv2.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create iam v2 client")
	}
	workflowsService, err := workflows.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create workflows client")
	}
	eventarcService, err := eventarc.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create eventarc client")
	}
	firebaseService, err := firebase.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create firebase management client")
	}
	apiKeysService, err := apikeys.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create api keys client")
	}
	// Resource Manager v3 serves folders and the tag family (keys, values,
	// bindings), which the v1 client above predates; Organization Policy v2
	// serves policies and custom constraints.
	crmV3Service, err := crmv3.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudresourcemanager v3 client")
	}
	orgPolicyService, err := orgpolicy.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create orgpolicy client")
	}
	// Cloud Billing budgets live on the billing account; Cloud Identity
	// groups live under a customer -- neither is project-scoped, and the
	// kinds that use them stay deferred until the harness identity holds
	// the account- and customer-level roles.
	billingBudgetsService, err := billingbudgets.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create billingbudgets client")
	}
	cloudIdentityService, err := cloudidentity.NewService(ctx, clientOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cloudidentity client")
	}
	// ADC-authenticated plain HTTP client for services whose typed Go
	// client is not in the pinned google.golang.org/api line (Vertex AI,
	// Discovery Engine, Model Armor, Document AI, Cloud TPU, Memorystore for
	// Valkey, ...) -- verifiers reach it only through googleRestGet, for
	// GET probes. Built through the same transport the typed clients use so
	// it carries the same quota project header.
	restClient, _, err := htransport.NewClient(ctx,
		append([]option.ClientOption{option.WithScopes(cloudresourcemanager.CloudPlatformScope)}, clientOpts...)...)
	if err != nil {
		return errors.Wrap(err, "failed to create ADC-authenticated REST client")
	}

	gotProject, err := crmService.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "GCP credential validation failed: cannot reach test project %q "+
			"(cloudresourcemanager projects.get)", project)
	}

	fmt.Printf("  [gcp] authenticated against project %s (%s)\n", gotProject.ProjectId, gotProject.LifecycleState)

	h.services = &verify.Services{
		Project:              project,
		Crm:                  crmService,
		Iam:                  iamService,
		Compute:              computeService,
		Storage:              storageService,
		SqlAdmin:             sqlAdminService,
		Redis:                redisService,
		Container:            containerService,
		Run:                  runService,
		AlloyDB:              alloyDBService,
		DNS:                  dnsService,
		Spanner:              spannerService,
		BigQuery:             bigQueryService,
		VpcAccess:            vpcAccessService,
		Functions:            cloudFunctionsService,
		NetworkConnectivity:  networkConnectivityService,
		BigtableAdmin:        bigtableAdminService,
		Firestore:            firestoreService,
		Dataproc:             dataprocService,
		Composer:             composerService,
		PubSub:               pubsubService,
		CloudKms:             cloudKmsService,
		CloudTasks:           cloudTasksService,
		CloudScheduler:       cloudSchedulerService,
		ArtifactRegistry:     artifactRegistryService,
		CertificateManager:   certificateManagerService,
		Monitoring:           monitoringService,
		MonitoringDashboards: monitoringDashboardsService,
		SecretManager:        secretManagerService,
		Logging:              loggingService,
		IdentityToolkit:      identityToolkitService,
		IamV2:                iamV2Service,
		Workflows:            workflowsService,
		Eventarc:             eventarcService,
		Firebase:             firebaseService,
		ApiKeys:              apiKeysService,
		CrmV3:                crmV3Service,
		OrgPolicy:            orgPolicyService,
		BillingBudgets:       billingBudgetsService,
		CloudIdentity:        cloudIdentityService,
		RestClient:           restClient,
	}
	return nil
}

// Teardown is a no-op. Each scenario destroys its own resources in the DESTROY
// phase and confirms removal in VERIFY-CLN.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed confirms the component's resource exists via its registered
// verifier. GCP identifiers are frequently compound (an IAM grant is a
// project+role+member tuple), so the whole string-ified output set is stored
// and handed to the verifier rather than a single id.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	strOutputs := stringOutputs(outputs)
	if strOutputs[v.IDOutputKey()] == "" {
		return errors.Errorf("no %q in outputs for %s -- cannot verify", v.IDOutputKey(), component)
	}

	h.mu.Lock()
	h.deployed[componentKey(ctx, component)] = strOutputs
	h.mu.Unlock()

	return v.VerifyExists(ctx, h.services, strOutputs)
}

// VerifyDestroyed confirms the previously deployed resource no longer exists.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	h.mu.Lock()
	outputs := h.deployed[componentKey(ctx, component)]
	h.mu.Unlock()

	if len(outputs) == 0 {
		return errors.Errorf("no stored outputs for %s -- VerifyDeployed may not have run", component)
	}
	return v.VerifyAbsent(ctx, h.services, outputs)
}

// VerifyExpectedDeployFailure implements the framework's optional
// DeployFailureVerifier capability (expected-deploy-failure lanes, for
// substrates that gate resource creation on workload health). Stack outputs
// do not exist on a failed deploy, so identity comes from the scenario
// manifest: the service name (metadata.name) and region (spec.region). The
// kind's verifier must itself opt in via the local
// verify.DeployFailureVerifier interface.
func (h *Harness) VerifyExpectedDeployFailure(ctx context.Context, tc *provider.ComponentTestContext, expectation string, deployErr error) error {
	v, err := verify.GetVerifier(tc.Component)
	if err != nil {
		return err
	}
	dfv, ok := v.(verify.DeployFailureVerifier)
	if !ok {
		return errors.Errorf("component %q's verifier does not implement verify.DeployFailureVerifier -- the scenario expects a deploy failure (%s) it cannot attribute", tc.Component, expectation)
	}

	manifestPath, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	if manifestPath == "" {
		manifestPath = tc.ManifestPath
	}
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrap(err, "reading the scenario manifest for failure attribution")
	}
	var manifest struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Spec struct {
			Region string `json:"region"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &manifest); err != nil {
		return errors.Wrap(err, "parsing the scenario manifest for failure attribution")
	}
	if manifest.Metadata.Name == "" || manifest.Spec.Region == "" {
		return errors.Errorf("the scenario manifest must carry metadata.name and spec.region for failure attribution (got name=%q region=%q)",
			manifest.Metadata.Name, manifest.Spec.Region)
	}
	if err := dfv.VerifyExpectedDeployFailure(ctx, h.services, manifest.Metadata.Name, manifest.Spec.Region, expectation, deployErr); err != nil {
		return err
	}

	// Store the manifest-derived identity where VerifyDeployed would have:
	// the expected-failure lifecycle has no VERIFY-RES, but its VERIFY-CLN
	// still must prove the destroyed-after-failed-create resource is GONE,
	// through the same absence probe every normal lane uses.
	h.mu.Lock()
	h.deployed[componentKey(ctx, tc.Component)] = map[string]string{
		"service_short_name": manifest.Metadata.Name,
		"region":             manifest.Spec.Region,
		"project_id":         h.services.Project,
	}
	h.mu.Unlock()
	return nil
}

// stringOutputs flattens stack outputs to strings, tolerating non-string scalars.
func stringOutputs(outputs map[string]interface{}) map[string]string {
	result := make(map[string]string, len(outputs))
	for key, value := range outputs {
		if s, ok := value.(string); ok {
			result[key] = s
			continue
		}
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

// componentKey combines the manifest path (from context) with the component name
// so concurrent scenarios of the same component type do not collide in the map.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
