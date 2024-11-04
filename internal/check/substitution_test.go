package check

import (
	"testing"

	"github.com/errata-ai/vale/v3/internal/core"
	"github.com/errata-ai/vale/v3/internal/nlp"
)

func makeSubstitution(def baseCheck) (*Substitution, error) {
	cfg, err := core.NewConfig(&core.CLIFlags{})
	if err != nil {
		return nil, err
	}

	rule, err := NewSubstitution(cfg, def, "")
	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func TestConvertGroups(t *testing.T) {
	converted, err := convertCaptureGroups("change in(?: )?to the (.*) directory")
	if err != nil {
		t.Fatal(err)
	}

	expected := "change in(?: )?to the (?:.*) directory"
	if converted != expected {
		t.Fatalf("Expected '%s', got '%s'", expected, converted)
	}
}

func TestIsDeterministic(t *testing.T) {
	swap := map[string]interface{}{
		"extends":    "substitution",
		"name":       "Vale.Terms",
		"level":      "error",
		"message":    "Use '%s' instead of '%s'.",
		"scope":      "text",
		"ignorecase": true,
		"swap": map[string]string{
			"emnify iot supernetwork": "emnify IoT SuperNetwork",
			"emnify":                  "emnify",
		},
	}

	text := "EMnify IoT SuperNetwork"
	for i := 0; i < 100; i++ {
		rule, err := makeSubstitution(swap)
		if err != nil {
			t.Fatal(err)
		}

		actual, err := rule.Run(nlp.NewBlock(text, text, "text"), &core.File{}, &core.Config{})
		if err != nil {
			t.Fatal(err)
		}

		if len(actual) != 1 {
			t.Fatalf("expected 1 alert, found %d", len(actual))
		} else if actual[0].Match != "EMnify IoT SuperNetwork" {
			t.Fatalf("Loop %d: expected 'EMnify IoT SuperNetwork', found '%s'", i, actual[0].Match)
		}
	}
}

func TestRegex(t *testing.T) {
	swap := map[string]interface{}{
		"extends":    "substitution",
		"name":       "Vale.Terms",
		"level":      "error",
		"message":    "Use '%s' instead of '%s'.",
		"scope":      "text",
		"ignorecase": true,
		"swap": map[string]string{
			`(?:foo|bar)`: "sub",
		},
	}
	text := "foo"
	rule, err := makeSubstitution(swap)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := rule.Run(nlp.NewBlock(text, text, "text"), &core.File{}, &core.Config{})
	if err != nil {
		t.Fatal(err)
	}

	expected := "Use 'sub' instead of 'foo'."
	message := actual[0].Message
	if message != expected {
		t.Fatalf("Expected message `%s`, got `%s`", expected, message)
	}
}

func TestRegexEscapedParens(t *testing.T) {
	swap := map[string]interface{}{
		"extends":    "substitution",
		"name":       "Vale.Terms",
		"level":      "error",
		"message":    "Use '%s' instead of '%s'.",
		"scope":      "text",
		"ignorecase": true,
		"swap": map[string]string{
			`(?!\()(?:foo|bar)(?!\))?`: "sub",
		},
	}
	text := "(foo)"
	rule, err := makeSubstitution(swap)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := rule.Run(nlp.NewBlock(text, text, "text"), &core.File{}, &core.Config{})
	if err != nil {
		t.Fatal(err)
	}

	expected := "Use 'sub' instead of 'foo'."
	message := actual[0].Message
	if message != expected {
		t.Fatalf("Expected message `%s`, got `%s`", expected, message)
	}
}
