package component

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

func restoredPlatform() *v1.PlantonPlatform {
	planton := &v1.PlantonPlatform{ObjectMeta: metav1.ObjectMeta{
		Name: "planton", Namespace: "planton", UID: types.UID("1a2b3c4d-0000-4000-8000-000000000000"),
	}}
	planton.Status.Backup = &v1.BackupStatus{State: v1.BackupStateDeploying, RestoredFrom: "planton-postgres-deadbeef"}
	return planton
}

func recoveryFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(subOperatorScheme(t)).WithObjects(objs...).Build()
}

func recoveryConfig() resources.IdentityConfig {
	return resources.IdentityConfig{CRName: "planton", Namespace: "planton", Realm: "planton",
		PublicURL: "https://planton.example.com", PostgreSQL: resources.PostgreSQLConnection("planton", "planton")}
}

func recoveryJob(status batchv1.JobStatus) *batchv1.Job {
	job := resources.IdentityRecoveryAdminJob(recoveryConfig())
	job.Status = status
	return job
}

func identityDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: resources.IdentityDeploymentName("planton"), Namespace: "planton"}}
}

func objectExists(t *testing.T, c client.Client, obj client.Object, name string) bool {
	t.Helper()
	err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "planton"}, obj)
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatal(err)
	}
	return err == nil
}

func TestIdentityRecoveryPending(t *testing.T) {
	id := &Identity{}
	fresh := &v1.PlantonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "planton", Namespace: "planton"}}
	if pending, _ := id.identityRecoveryPending(context.Background(), recoveryFakeClient(t), fresh); pending {
		t.Error("a database created empty needs no recovery")
	}
	fresh.Status.Backup = &v1.BackupStatus{State: v1.BackupStateHealthy}
	if pending, _ := id.identityRecoveryPending(context.Background(), recoveryFakeClient(t), fresh); pending {
		t.Error("a database that archives but was not restored needs no recovery")
	}
	if pending, _ := id.identityRecoveryPending(context.Background(), recoveryFakeClient(t), restoredPlatform()); !pending {
		t.Error("a restored database with no completion record is pending")
	}
	marker := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: resources.IdentityRecoveryMarkerName("planton"), Namespace: "planton"}}
	if pending, _ := id.identityRecoveryPending(context.Background(), recoveryFakeClient(t, marker), restoredPlatform()); pending {
		t.Error("the completion record ends the recovery for the platform's lifetime")
	}
}

func TestRecoverAdminBeforeServer_CreatesTheCredentialAndTheJobThenWaits(t *testing.T) {
	id := &Identity{}
	c := recoveryFakeClient(t)
	planton := restoredPlatform()

	res, err := id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.Ready || res.Reason != v1.ComponentReasonDeploying || !strings.Contains(res.Message, "before the server starts") {
		t.Fatalf("the server waits for the recovery job: %+v", res)
	}
	name := resources.IdentityRecoveryAdminName("planton")
	var secret corev1.Secret
	if !objectExists(t, c, &secret, name) || len(secret.Data[resources.IdentityBootstrapAdminPasswordKey]) == 0 {
		t.Error("the recovery credential is generated before the job that reads it")
	}
	var job batchv1.Job
	if !objectExists(t, c, &job, name) {
		t.Fatal("the recovery job is applied")
	}
	if got := job.Spec.Template.Spec.Containers[0].Args[0]; got != "bootstrap-admin" {
		t.Errorf("the job runs Keycloak's recovery command, got %v", job.Spec.Template.Spec.Containers[0].Args)
	}

	// Still running on the next pass: still waiting, nothing new created.
	res, err = id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil || res == nil || res.Reason != v1.ComponentReasonDeploying {
		t.Errorf("a job without a verdict keeps the server waiting: %+v %v", res, err)
	}
}

func TestRecoverAdminBeforeServer_FailedJobIsExplainedWithItsLog(t *testing.T) {
	id := &Identity{}
	failed := recoveryJob(batchv1.JobStatus{Failed: 1, Conditions: []batchv1.JobCondition{{
		Type: batchv1.JobFailed, Status: corev1.ConditionTrue, Reason: "BackoffLimitExceeded", Message: "Job has reached the specified backoff limit",
	}}})
	c := recoveryFakeClient(t, failed)
	planton := restoredPlatform()

	res, err := id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.Reason != v1.ComponentReasonJobFailed || res.Object == nil || res.Object.Kind != "Job" {
		t.Fatalf("a failed job is the cause, named: %+v", res)
	}
	if !strings.Contains(res.Message, "kubectl logs -n planton job/planton-identity-recovery-admin") || !strings.Contains(res.Message, "BackoffLimitExceeded") {
		t.Errorf("the sentence carries the log command and the condition: %q", res.Message)
	}
	if objectExists(t, recoveryFakeClient(t, failed), &appsv1.Deployment{}, resources.IdentityDeploymentName("planton")) {
		t.Error("the server never starts on a failed recovery")
	}
}

func TestRecoverAdminBeforeServer_SucceededJobLetsTheServerStart(t *testing.T) {
	id := &Identity{}
	c := recoveryFakeClient(t, recoveryJob(batchv1.JobStatus{Succeeded: 1}))
	planton := restoredPlatform()
	res, err := id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil || res != nil {
		t.Fatalf("a succeeded job hands over to the server: %+v %v", res, err)
	}
}

// The pass after the Job succeeds starts the server; the pass after that
// runs this half again with the server up. It must leave the server alone,
// or the server is started and stopped every pass and the second half never
// finds it answering (seen live: a restored platform whose identity pods
// were replaced every seven seconds for ten minutes).
func TestRecoverAdminBeforeServer_SucceededJobNeverStopsTheServerAgain(t *testing.T) {
	id := &Identity{}
	c := recoveryFakeClient(t, recoveryJob(batchv1.JobStatus{Succeeded: 1}), identityDeployment())
	planton := restoredPlatform()
	res, err := id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil || res != nil {
		t.Fatalf("a succeeded job hands over to the server even when the server is already up: %+v %v", res, err)
	}
	if !objectExists(t, c, &appsv1.Deployment{}, resources.IdentityDeploymentName("planton")) {
		t.Error("the server started after the job succeeded is never stopped again")
	}
}

func TestRecoverAdminBeforeServer_ARunningServerIsStoppedFirst(t *testing.T) {
	id := &Identity{}
	c := recoveryFakeClient(t, identityDeployment())
	planton := restoredPlatform()
	res, err := id.recoverAdminBeforeServer(context.Background(), c, planton, recoveryConfig(), id.OwnerReferenceFor(planton))
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !strings.Contains(res.Message, "Restarting the identity server") {
		t.Fatalf("the recovery command needs every node stopped: %+v", res)
	}
	if objectExists(t, c, &appsv1.Deployment{}, resources.IdentityDeploymentName("planton")) {
		t.Error("the running server is stopped before the recovery job is created")
	}
	if objectExists(t, c, &batchv1.Job{}, resources.IdentityRecoveryAdminName("planton")) {
		t.Error("no job beside a live server; it is created on the next pass")
	}
}

// fakeIdentityServer is the master realm the second half talks to: the
// token endpoint knowing two users, and the users API the recovery drives.
type fakeIdentityServer struct {
	passwords map[string]string
	ids       map[string]string
	calls     []string
}

func newFakeIdentityServer(adminPassword string) *fakeIdentityServer {
	return &fakeIdentityServer{
		passwords: map[string]string{resources.IdentityBootstrapAdminUsername: adminPassword},
		ids:       map[string]string{resources.IdentityBootstrapAdminUsername: "id-admin"},
	}
}

func (f *fakeIdentityServer) withRecoveryAdmin(password string) *fakeIdentityServer {
	f.passwords[resources.IdentityRecoveryAdminUsername] = password
	f.ids[resources.IdentityRecoveryAdminUsername] = "id-recovery"
	return f
}

func (f *fakeIdentityServer) serve(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/idp/realms/master/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		user, pass := r.Form.Get("username"), r.Form.Get("password")
		f.calls = append(f.calls, "token "+user)
		if known, ok := f.passwords[user]; !ok || known != pass {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "t"})
	})
	mux.HandleFunc("/idp/admin/realms/master/users", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		var reps []map[string]any
		if id, ok := f.ids[username]; ok {
			reps = append(reps, map[string]any{"id": id, "username": username})
		}
		_ = json.NewEncoder(w).Encode(reps)
	})
	mux.HandleFunc("/idp/admin/realms/master/users/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/idp/admin/realms/master/users/")
		switch {
		case r.Method == http.MethodPut && strings.HasSuffix(rest, "/reset-password"):
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			for user, uid := range f.ids {
				if uid == strings.TrimSuffix(rest, "/reset-password") {
					f.passwords[user], _ = body["value"].(string)
				}
			}
			f.calls = append(f.calls, "reset")
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			for user, uid := range f.ids {
				if uid == rest {
					delete(f.ids, user)
					delete(f.passwords, user)
				}
			}
			f.calls = append(f.calls, "delete")
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func recoverySecret(password string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: resources.IdentityRecoveryAdminName("planton"), Namespace: "planton"},
		Data:       map[string][]byte{resources.IdentityBootstrapAdminPasswordKey: []byte(password)},
	}
}

func assertRecoveryFinished(t *testing.T, c client.Client) {
	t.Helper()
	if !objectExists(t, c, &corev1.ConfigMap{}, resources.IdentityRecoveryMarkerName("planton")) {
		t.Error("completion is recorded so the recovery runs once per platform lifetime")
	}
	if objectExists(t, c, &corev1.Secret{}, resources.IdentityRecoveryAdminName("planton")) || objectExists(t, c, &batchv1.Job{}, resources.IdentityRecoveryAdminName("planton")) {
		t.Error("the recovery credential and job are removed once the admin is re-established")
	}
}

func TestRecoverAdminAfterServer_ReestablishesTheAdminAndFinishes(t *testing.T) {
	id := &Identity{}
	server := newFakeIdentityServer("source-platform-password").withRecoveryAdmin("recovery-password")
	srv := server.serve(t)
	c := recoveryFakeClient(t, recoverySecret("recovery-password"), recoveryJob(batchv1.JobStatus{Succeeded: 1}))
	planton := restoredPlatform()

	res, err := id.recoverAdminAfterServer(context.Background(), c, planton, srv.URL+"/idp", "this-install-password", srv.Client())
	if err != nil || res != nil {
		t.Fatalf("a re-established admin hands over to the realm reconciler: %+v %v", res, err)
	}
	if server.passwords[resources.IdentityBootstrapAdminUsername] != "this-install-password" {
		t.Error("the master admin carries this install's password now")
	}
	if _, stillThere := server.ids[resources.IdentityRecoveryAdminUsername]; stillThere {
		t.Error("the recovery admin is removed")
	}
	assertRecoveryFinished(t, c)
}

func TestRecoverAdminAfterServer_AnAdminThatAlreadyWorksOnlyRecords(t *testing.T) {
	id := &Identity{}
	server := newFakeIdentityServer("this-install-password")
	srv := server.serve(t)
	c := recoveryFakeClient(t, recoverySecret("recovery-password"))
	planton := restoredPlatform()

	res, err := id.recoverAdminAfterServer(context.Background(), c, planton, srv.URL+"/idp", "this-install-password", srv.Client())
	if err != nil || res != nil {
		t.Fatalf("nothing to recover: %+v %v", res, err)
	}
	if strings.Join(server.calls, ",") != "token admin" {
		t.Errorf("the admin's own sign-in is the only call: %v", server.calls)
	}
	assertRecoveryFinished(t, c)
}

func TestRecoverAdminAfterServer_RefusedAdminWithNoRecoveryInFlightRestartsTheServer(t *testing.T) {
	id := &Identity{}
	srv := newFakeIdentityServer("source-platform-password").serve(t)
	c := recoveryFakeClient(t, identityDeployment())
	planton := restoredPlatform()

	res, err := id.recoverAdminAfterServer(context.Background(), c, planton, srv.URL+"/idp", "this-install-password", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !strings.Contains(res.Message, "Restarting the identity server") {
		t.Fatalf("a server that started before the operator knew to recover is stopped: %+v", res)
	}
	if objectExists(t, c, &appsv1.Deployment{}, resources.IdentityDeploymentName("planton")) {
		t.Error("the server is stopped so the recovery command can run")
	}
}

func TestRecoverAdminAfterServer_NoAdminInTheRealmIsRefusedInWords(t *testing.T) {
	id := &Identity{}
	server := newFakeIdentityServer("source-platform-password").withRecoveryAdmin("recovery-password")
	delete(server.ids, resources.IdentityBootstrapAdminUsername)
	delete(server.passwords, resources.IdentityBootstrapAdminUsername)
	srv := server.serve(t)
	c := recoveryFakeClient(t, recoverySecret("recovery-password"))
	planton := restoredPlatform()

	res, err := id.recoverAdminAfterServer(context.Background(), c, planton, srv.URL+"/idp", "this-install-password", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.Reason != v1.ComponentReasonConfigurationRefused || !strings.Contains(res.Message, "no admin named") || !strings.Contains(res.Message, "planton-recovery") {
		t.Fatalf("a realm without the operator's admin is refused in words that name the way out: %+v", res)
	}
	if !objectExists(t, c, &corev1.Secret{}, resources.IdentityRecoveryAdminName("planton")) {
		t.Error("the recovery credential stays for the person who has to act")
	}
}

func TestRecoverAdminAfterServer_RefusedRecoveryCredentialNamesTheJobLog(t *testing.T) {
	id := &Identity{}
	srv := newFakeIdentityServer("source-platform-password").serve(t) // no recovery admin exists
	c := recoveryFakeClient(t, recoverySecret("recovery-password"))
	planton := restoredPlatform()

	res, err := id.recoverAdminAfterServer(context.Background(), c, planton, srv.URL+"/idp", "this-install-password", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.Reason != v1.ComponentReasonReconcileFailed || !strings.Contains(res.Message, "kubectl logs -n planton job/planton-identity-recovery-admin") {
		t.Fatalf("both credentials refused: the job's log is the next step: %+v", res)
	}
}

func TestRecoverAdminAfterServer_ServerNotAnsweringIsLeftToTheRealmReconciler(t *testing.T) {
	id := &Identity{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer srv.Close()
	c := recoveryFakeClient(t, recoverySecret("recovery-password"))
	res, err := id.recoverAdminAfterServer(context.Background(), c, restoredPlatform(), srv.URL+"/idp", "x", srv.Client())
	if err != nil || res != nil {
		t.Fatalf("a server that is not answering is the realm reconciler's to report, in its own words: %+v %v", res, err)
	}
}
