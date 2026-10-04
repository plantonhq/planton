package verify

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// The verifiers of the kinds that shape a tenant's sign-in: how Universal
// Login looks (Auth0Branding), how the login flow behaves (Auth0Prompt), what
// each prompt says (Auth0PromptCustomText) and inserts (Auth0PromptScreenPartials),
// and how the tenant mails (Auth0EmailProvider, Auth0EmailTemplate). Most of
// them manage a setting the tenant always has, so each states its own destroy
// contract, source-verified against the provider at the pin, and asserts the
// state its destroy leaves:
//
//   - Auth0Branding: the tenant always has branding. Destroy removes only the
//     page template (and deletes the theme, a real object), so the branding
//     still answers and the theme is gone.
//   - Auth0Prompt: the provider's delete is a no-op; the settings still answer.
//   - Auth0PromptCustomText: destroy sets the prompt's text in that language
//     to {}, so the custom text answers empty.
//   - Auth0PromptScreenPartials: destroy sets the prompt's partials to empty,
//     so the partials answer with no screen.
//   - Auth0EmailProvider: a real delete; the provider stops answering.
//   - Auth0EmailTemplate: Auth0 cannot delete a template; destroy disables
//     it, so the template answers with enabled false.

// brandingVerifier verifies Auth0Branding: the tenant's branding answers, and
// the theme the branding applied exists after deploy and is gone after
// destroy. Without a theme (theme_id empty) only the branding is checked.
type brandingVerifier struct{}

func (*brandingVerifier) IDOutput() string { return "theme_id" }

// IDOutputOptional is true: a branding without a theme reports an empty
// theme_id, and the tenant's branding is still the object to verify.
func (*brandingVerifier) IDOutputOptional() bool { return true }

func (*brandingVerifier) VerifyExists(checker ResourceChecker, themeID string) error {
	if err := requireAnswers(checker, "auth0branding", "branding", "after deploy"); err != nil {
		return err
	}
	if themeID == "" {
		return nil
	}
	return requireAnswers(checker, "auth0branding", "branding/themes/"+url.PathEscape(themeID), "after deploy (the theme)")
}

func (*brandingVerifier) VerifyAbsent(checker ResourceChecker, themeID string) error {
	if err := requireAnswers(checker, "auth0branding", "branding", "after destroy, which leaves the logo, colors and font in place"); err != nil {
		return err
	}
	if themeID == "" {
		return nil
	}
	exists, err := checker.ResourceExists("branding/themes/" + url.PathEscape(themeID))
	if err != nil {
		return errors.Wrap(err, "auth0branding: reading the theme failed")
	}
	if exists {
		return errors.Errorf("auth0branding: theme %s still exists after destroy, which deletes it", themeID)
	}
	return nil
}

// promptVerifier verifies Auth0Prompt: the tenant's prompt settings answer
// after deploy, and still answer after the provider's no-op destroy.
type promptVerifier struct{}

func (*promptVerifier) IDOutput() string { return "" }

func (*promptVerifier) VerifyExists(checker ResourceChecker, _ string) error {
	return requireAnswers(checker, "auth0prompt", "prompts", "after deploy")
}

func (*promptVerifier) VerifyAbsent(checker ResourceChecker, _ string) error {
	return requireAnswers(checker, "auth0prompt", "prompts", "after destroy, which leaves the last-applied settings in place")
}

// promptCustomTextVerifier verifies Auth0PromptCustomText by its id,
// "<prompt>::<language>": the prompt's text in that language holds words
// after deploy and is empty after destroy.
type promptCustomTextVerifier struct{}

func (*promptCustomTextVerifier) IDOutput() string { return defaultIDOutput }

func (*promptCustomTextVerifier) VerifyExists(checker ResourceChecker, id string) error {
	body, err := readCustomText(checker, id)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return errors.Errorf("auth0promptcustomtext: %s holds no words after deploy", id)
	}
	return nil
}

func (*promptCustomTextVerifier) VerifyAbsent(checker ResourceChecker, id string) error {
	body, err := readCustomText(checker, id)
	if err != nil {
		return err
	}
	if len(body) != 0 {
		return errors.Errorf("auth0promptcustomtext: %s still holds words after destroy, which resets it to {}", id)
	}
	return nil
}

func readCustomText(checker ResourceChecker, id string) (map[string]interface{}, error) {
	prompt, language, ok := strings.Cut(id, "::")
	if !ok || prompt == "" || language == "" {
		return nil, errors.Errorf("auth0promptcustomtext: id %q is not <prompt>::<language>", id)
	}
	body, exists, err := checker.ReadResource(fmt.Sprintf("prompts/%s/custom-text/%s", url.PathEscape(prompt), url.PathEscape(language)))
	if err != nil {
		return nil, errors.Wrap(err, "auth0promptcustomtext: reading the custom text failed")
	}
	if !exists {
		return nil, errors.Errorf("auth0promptcustomtext: prompt %s in language %s did not answer", prompt, language)
	}
	return body, nil
}

// promptScreenPartialsVerifier verifies Auth0PromptScreenPartials by its
// prompt: the prompt carries partials after deploy and none after destroy.
type promptScreenPartialsVerifier struct{}

func (*promptScreenPartialsVerifier) IDOutput() string { return "prompt_type" }

func (*promptScreenPartialsVerifier) VerifyExists(checker ResourceChecker, prompt string) error {
	partials, err := readPartials(checker, prompt)
	if err != nil {
		return err
	}
	if len(partials) == 0 {
		return errors.Errorf("auth0promptscreenpartials: prompt %s carries no partials after deploy", prompt)
	}
	return nil
}

func (*promptScreenPartialsVerifier) VerifyAbsent(checker ResourceChecker, prompt string) error {
	partials, err := readPartials(checker, prompt)
	if err != nil {
		return err
	}
	if len(partials) != 0 {
		return errors.Errorf("auth0promptscreenpartials: prompt %s still carries partials after destroy, which empties them", prompt)
	}
	return nil
}

func readPartials(checker ResourceChecker, prompt string) (map[string]interface{}, error) {
	if prompt == "" {
		return nil, errors.New("auth0promptscreenpartials: the prompt_type output is empty")
	}
	body, exists, err := checker.ReadResource(fmt.Sprintf("prompts/%s/partials", url.PathEscape(prompt)))
	if err != nil {
		return nil, errors.Wrap(err, "auth0promptscreenpartials: reading the partials failed")
	}
	if !exists {
		return nil, errors.Errorf("auth0promptscreenpartials: prompt %s did not answer", prompt)
	}
	return body, nil
}

// emailProviderVerifier verifies Auth0EmailProvider: the tenant's email
// provider answers after deploy and is gone after destroy.
type emailProviderVerifier struct{}

func (*emailProviderVerifier) IDOutput() string { return "" }

func (*emailProviderVerifier) VerifyExists(checker ResourceChecker, _ string) error {
	return requireAnswers(checker, "auth0emailprovider", "emails/provider", "after deploy")
}

func (*emailProviderVerifier) VerifyAbsent(checker ResourceChecker, _ string) error {
	exists, err := checker.ResourceExists("emails/provider")
	if err != nil {
		return errors.Wrap(err, "auth0emailprovider: reading the email provider failed")
	}
	if exists {
		return errors.New("auth0emailprovider: the tenant's email provider still answers after destroy, which deletes it")
	}
	return nil
}

// emailTemplateVerifier verifies Auth0EmailTemplate by its template name: the
// template is enabled after deploy and disabled after destroy.
type emailTemplateVerifier struct{}

func (*emailTemplateVerifier) IDOutput() string { return "template" }

func (*emailTemplateVerifier) VerifyExists(checker ResourceChecker, template string) error {
	return requireTemplateEnabled(checker, template, true, "after deploy")
}

func (*emailTemplateVerifier) VerifyAbsent(checker ResourceChecker, template string) error {
	return requireTemplateEnabled(checker, template, false, "after destroy, which disables it")
}

func requireTemplateEnabled(checker ResourceChecker, template string, want bool, when string) error {
	if template == "" {
		return errors.New("auth0emailtemplate: the template output is empty")
	}
	body, exists, err := checker.ReadResource("email-templates/" + url.PathEscape(template))
	if err != nil {
		return errors.Wrap(err, "auth0emailtemplate: reading the template failed")
	}
	if !exists {
		return errors.Errorf("auth0emailtemplate: template %s did not answer %s", template, when)
	}
	if enabled, _ := body["enabled"].(bool); enabled != want {
		return errors.Errorf("auth0emailtemplate: template %s is enabled=%t %s, want %t", template, enabled, when, want)
	}
	return nil
}

// requireAnswers asserts a per-tenant Management API object answers.
func requireAnswers(checker ResourceChecker, kind, path, when string) error {
	exists, err := checker.ResourceExists(path)
	if err != nil {
		return errors.Wrapf(err, "%s: reading %s failed", kind, path)
	}
	if !exists {
		return errors.Errorf("%s: %s did not answer %s", kind, path, when)
	}
	return nil
}
