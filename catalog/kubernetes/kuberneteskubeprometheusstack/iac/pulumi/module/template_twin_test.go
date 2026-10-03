package module

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The OpenTofu module carries the same message template as a heredoc in its
// locals.tf (notification_template). The live notifications scenario proves
// both engines deliver one message, but only for the alerts it fires; this
// test holds the two texts byte for byte on every run, so a change to one
// (a title fallback, a body line) cannot ship without the other.
func TestOpenTofuTemplateIsThePulumiTemplatesTwin(t *testing.T) {
	locals, err := os.ReadFile(filepath.Join("..", "..", "tf", "locals.tf"))
	if err != nil {
		t.Fatal(err)
	}
	heredoc := regexp.MustCompile(`(?s)notification_template = <<EOT\n(.*?)EOT\n`).FindSubmatch(locals)
	if heredoc == nil {
		t.Fatal("locals.tf has no notification_template heredoc")
	}
	tofu := strings.NewReplacer(
		"${local.notification_env_label}", "environment",
		"${local.notification_component_label}", "component",
	).Replace(string(heredoc[1]))
	if pulumi := notificationTemplate("environment", "component"); tofu != pulumi {
		t.Errorf("the two engines' message templates differ\nOpenTofu:\n%s\nPulumi:\n%s", tofu, pulumi)
	}
}
