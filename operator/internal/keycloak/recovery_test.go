package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeMasterRealm is the slice of the Admin API the recovery touches: the
// master token endpoint and the master realm's users. It records every call
// so the tests pin the exact sequence, and it knows two users -- the real
// admin with its source-platform password, and the recovery admin the
// recovery command created.
type fakeMasterRealm struct {
	t         *testing.T
	passwords map[string]string // username -> password accepted at the token endpoint
	ids       map[string]string // username -> user id
	calls     []string
}

func newFakeMasterRealm(t *testing.T) *fakeMasterRealm {
	return &fakeMasterRealm{
		t:         t,
		passwords: map[string]string{"admin": "source-platform-password", "planton-recovery": "recovery-password"},
		ids:       map[string]string{"admin": "id-admin", "planton-recovery": "id-recovery"},
	}
}

func (f *fakeMasterRealm) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/idp/realms/master/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		user, pass := r.Form.Get("username"), r.Form.Get("password")
		f.calls = append(f.calls, "token "+user)
		if f.passwords[user] != pass {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid user credentials"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-for-" + user})
	})
	mux.HandleFunc("/idp/admin/realms/master/users", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		f.calls = append(f.calls, "find "+username)
		var reps []Representation
		if id, ok := f.ids[username]; ok {
			reps = append(reps, Representation{"id": id, "username": username})
		}
		_ = json.NewEncoder(w).Encode(reps)
	})
	mux.HandleFunc("/idp/admin/realms/master/users/", func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, "/idp/admin/realms/master/users/")
		switch {
		case r.Method == http.MethodPut && strings.HasSuffix(rest, "/reset-password"):
			id := strings.TrimSuffix(rest, "/reset-password")
			var body Representation
			_ = json.NewDecoder(r.Body).Decode(&body)
			f.calls = append(f.calls, "reset "+id)
			if body["type"] != "password" || body["temporary"] != false {
				f.t.Errorf("reset-password body must be a permanent password: %v", body)
			}
			for user, uid := range f.ids {
				if uid == id {
					f.passwords[user], _ = body["value"].(string)
				}
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			f.calls = append(f.calls, "delete "+rest)
			for user, uid := range f.ids {
				if uid == rest {
					delete(f.ids, user)
					delete(f.passwords, user)
				}
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	return mux
}

func TestReestablishAdmin_ResetsTheAdminThenRemovesItself(t *testing.T) {
	realm := newFakeMasterRealm(t)
	srv := httptest.NewServer(realm.handler())
	defer srv.Close()

	err := ReestablishAdmin(context.Background(), ReestablishAdminInput{
		HTTPClient:       srv.Client(),
		ServerRoot:       srv.URL + "/idp",
		RecoveryUsername: "planton-recovery",
		RecoveryPassword: "recovery-password",
		AdminUsername:    "admin",
		AdminPassword:    "this-install-password",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"token planton-recovery", "find admin", "reset id-admin", "find planton-recovery", "delete id-recovery"}
	if strings.Join(realm.calls, ", ") != strings.Join(want, ", ") {
		t.Errorf("call sequence\n got: %v\nwant: %v", realm.calls, want)
	}
	if realm.passwords["admin"] != "this-install-password" {
		t.Errorf("admin's password is this install's now, got %q", realm.passwords["admin"])
	}
	if _, stillThere := realm.ids["planton-recovery"]; stillThere {
		t.Error("the recovery admin is removed once the real admin is re-established")
	}

	// The real admin authenticates with the new password afterwards.
	admin := NewAdminClient(srv.Client(), srv.URL+"/idp")
	if err := admin.Authenticate(context.Background(), "admin", "this-install-password"); err != nil {
		t.Errorf("admin should sign in with this install's password: %v", err)
	}
}

func TestReestablishAdmin_RefusedRecoveryCredentialIsTyped(t *testing.T) {
	realm := newFakeMasterRealm(t)
	srv := httptest.NewServer(realm.handler())
	defer srv.Close()

	err := ReestablishAdmin(context.Background(), ReestablishAdminInput{
		HTTPClient:       srv.Client(),
		ServerRoot:       srv.URL + "/idp",
		RecoveryUsername: "planton-recovery",
		RecoveryPassword: "wrong",
		AdminUsername:    "admin",
		AdminPassword:    "this-install-password",
	})
	if err == nil || !IsCredentialRefused(err) {
		t.Fatalf("a refused recovery credential is a typed refusal the caller can name, got %v", err)
	}
	if !strings.Contains(err.Error(), `recovery admin "planton-recovery"`) {
		t.Errorf("the error names the recovery admin: %v", err)
	}
	if len(realm.calls) != 1 {
		t.Errorf("nothing is touched after a refused sign-in: %v", realm.calls)
	}
}

func TestReestablishAdmin_NoAdminInTheRestoredRealmIsTyped(t *testing.T) {
	realm := newFakeMasterRealm(t)
	delete(realm.ids, "admin")
	srv := httptest.NewServer(realm.handler())
	defer srv.Close()

	err := ReestablishAdmin(context.Background(), ReestablishAdminInput{
		HTTPClient:       srv.Client(),
		ServerRoot:       srv.URL + "/idp",
		RecoveryUsername: "planton-recovery",
		RecoveryPassword: "recovery-password",
		AdminUsername:    "admin",
		AdminPassword:    "this-install-password",
	})
	if !IsAdminNotFound(err) {
		t.Fatalf("a realm without the operator's admin is a typed answer, got %v", err)
	}
	if _, stillThere := realm.ids["planton-recovery"]; !stillThere {
		t.Error("the recovery admin stays when there is no admin to hand over to -- a person needs it")
	}
}

func TestIsCredentialRefused_OnlyFor401(t *testing.T) {
	if !IsCredentialRefused(&TokenRefusedError{Status: 401}) {
		t.Error("401 is the credential being refused")
	}
	if IsCredentialRefused(&TokenRefusedError{Status: 503}) {
		t.Error("503 is the server not answering, not a refusal")
	}
	if IsCredentialRefused(context.Canceled) {
		t.Error("an unrelated error is not a refusal")
	}
}
