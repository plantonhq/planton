package verify

import (
	"testing"
)

// pathChecker answers each Management API path from a fixed table, so a
// verifier that reads several paths is judged on every one of them. A path
// not in the table does not exist.
type pathChecker struct {
	objects map[string]map[string]interface{}
	asked   []string
}

func (c *pathChecker) ResourceExists(path string) (bool, error) {
	c.asked = append(c.asked, path)
	_, ok := c.objects[path]
	return ok, nil
}

func (c *pathChecker) ReadResource(path string) (map[string]interface{}, bool, error) {
	c.asked = append(c.asked, path)
	body, ok := c.objects[path]
	return body, ok, nil
}

func mustVerifier(t *testing.T, kind string) Verifier {
	t.Helper()
	v, err := GetVerifier(kind)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// Branding always exists; its theme is a real object that destroy deletes.
func TestBrandingVerifierChecksTheThemeItApplied(t *testing.T) {
	v := mustVerifier(t, "auth0branding")
	if v.IDOutput() != "theme_id" {
		t.Fatalf("IDOutput = %q, want theme_id", v.IDOutput())
	}
	if optional, ok := v.(OptionalIDVerifier); !ok || !optional.IDOutputOptional() {
		t.Fatal("a branding without a theme reports an empty theme_id, so the id output must be optional")
	}
	deployed := &pathChecker{objects: map[string]map[string]interface{}{
		"branding":                {"logo_url": "https://assets.example.com/logo.png"},
		"branding/themes/thm_abc": {"themeId": "thm_abc"},
	}}
	if err := v.VerifyExists(deployed, "thm_abc"); err != nil {
		t.Fatalf("deployed branding and theme: %v", err)
	}
	if err := v.VerifyAbsent(deployed, "thm_abc"); err == nil {
		t.Error("a theme still present after destroy passed")
	}
	destroyed := &pathChecker{objects: map[string]map[string]interface{}{"branding": {}}}
	if err := v.VerifyAbsent(destroyed, "thm_abc"); err != nil {
		t.Fatalf("branding left in place, theme deleted: %v", err)
	}
	if err := v.VerifyExists(destroyed, ""); err != nil {
		t.Fatalf("branding without a theme: %v", err)
	}
	if err := v.VerifyExists(&pathChecker{}, ""); err == nil {
		t.Error("a tenant whose branding does not answer passed")
	}
}

// The prompt settings always exist and the provider's destroy is a no-op.
func TestPromptVerifierReadsTheTenantsPromptSettings(t *testing.T) {
	v := mustVerifier(t, "auth0prompt")
	if v.IDOutput() != "" {
		t.Fatalf("IDOutput = %q, want none", v.IDOutput())
	}
	checker := &pathChecker{objects: map[string]map[string]interface{}{"prompts": {"identifier_first": true}}}
	if err := v.VerifyExists(checker, ""); err != nil {
		t.Fatalf("deployed: %v", err)
	}
	if err := v.VerifyAbsent(checker, ""); err != nil {
		t.Fatalf("settings left in place by destroy: %v", err)
	}
}

// Custom text is read by "<prompt>::<language>"; destroy resets it to {}.
func TestPromptCustomTextVerifierReadsThePromptInItsLanguage(t *testing.T) {
	v := mustVerifier(t, "auth0promptcustomtext")
	if v.IDOutput() != "id" {
		t.Fatalf("IDOutput = %q, want id", v.IDOutput())
	}
	worded := &pathChecker{objects: map[string]map[string]interface{}{
		"prompts/login/custom-text/en": {"login": map[string]interface{}{"title": "Welcome back"}},
	}}
	if err := v.VerifyExists(worded, "login::en"); err != nil {
		t.Fatalf("deployed words: %v", err)
	}
	if err := v.VerifyAbsent(worded, "login::en"); err == nil {
		t.Error("words still present after destroy passed")
	}
	reset := &pathChecker{objects: map[string]map[string]interface{}{"prompts/login/custom-text/en": {}}}
	if err := v.VerifyAbsent(reset, "login::en"); err != nil {
		t.Fatalf("text reset to {}: %v", err)
	}
	if err := v.VerifyExists(reset, "login"); err == nil {
		t.Error("an id without a language passed")
	}
}

// Screen partials are read by their prompt; destroy empties them.
func TestPromptScreenPartialsVerifierReadsThePrompt(t *testing.T) {
	v := mustVerifier(t, "auth0promptscreenpartials")
	if v.IDOutput() != "prompt_type" {
		t.Fatalf("IDOutput = %q, want prompt_type", v.IDOutput())
	}
	extended := &pathChecker{objects: map[string]map[string]interface{}{
		"prompts/signup/partials": {"signup": map[string]interface{}{"form-content-end": "<div>terms</div>"}},
	}}
	if err := v.VerifyExists(extended, "signup"); err != nil {
		t.Fatalf("deployed partials: %v", err)
	}
	emptied := &pathChecker{objects: map[string]map[string]interface{}{"prompts/signup/partials": {}}}
	if err := v.VerifyAbsent(emptied, "signup"); err != nil {
		t.Fatalf("partials emptied: %v", err)
	}
	if err := v.VerifyAbsent(extended, "signup"); err == nil {
		t.Error("partials still present after destroy passed")
	}
}

// The email provider is a real object that destroy deletes.
func TestEmailProviderVerifierExpectsDeletion(t *testing.T) {
	v := mustVerifier(t, "auth0emailprovider")
	present := &pathChecker{objects: map[string]map[string]interface{}{"emails/provider": {"name": "smtp"}}}
	if err := v.VerifyExists(present, ""); err != nil {
		t.Fatalf("deployed provider: %v", err)
	}
	if err := v.VerifyAbsent(present, ""); err == nil {
		t.Error("a provider still answering after destroy passed")
	}
	if err := v.VerifyAbsent(&pathChecker{}, ""); err != nil {
		t.Fatalf("deleted provider: %v", err)
	}
}

// Auth0 cannot delete a template: destroy disables it.
func TestEmailTemplateVerifierExpectsDisableOnDestroy(t *testing.T) {
	v := mustVerifier(t, "auth0emailtemplate")
	if v.IDOutput() != "template" {
		t.Fatalf("IDOutput = %q, want template", v.IDOutput())
	}
	enabled := &pathChecker{objects: map[string]map[string]interface{}{"email-templates/verify_email": {"enabled": true}}}
	disabled := &pathChecker{objects: map[string]map[string]interface{}{"email-templates/verify_email": {"enabled": false}}}
	if err := v.VerifyExists(enabled, "verify_email"); err != nil {
		t.Fatalf("deployed template: %v", err)
	}
	if err := v.VerifyAbsent(disabled, "verify_email"); err != nil {
		t.Fatalf("disabled template: %v", err)
	}
	if err := v.VerifyAbsent(enabled, "verify_email"); err == nil {
		t.Error("a template still enabled after destroy passed")
	}
}
